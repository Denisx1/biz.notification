package main

import (
	"context"
	"fmt"
	core_config "notification/internal/core/config"
	core_logger "notification/internal/core/logger"
	core_pgx_pool "notification/internal/core/repository/postgres/pool/pgx"
	core_http_middleware "notification/internal/core/transport/http/middlware"
	core_http_server "notification/internal/core/transport/http/server"
	"notification/internal/core/transport/nats"
	receiver_postgres_repository "notification/internal/features/receiver/repository"
	receiver_service "notification/internal/features/receiver/service"
	receiver_transport_http "notification/internal/features/receiver/transport/http"
	sender_service "notification/internal/features/sender/service"
	sender_transport_nats "notification/internal/features/sender/transport/nats"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	cfg := core_config.NewConfigMust()
	time.Local = cfg.TimeZone

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	logger, err := core_logger.NewLoger(core_logger.NewConfigMust())
	if err != nil {
		fmt.Println("Failed to init application logger:", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Debug("Application time zone", zap.Any("zone", time.Local))
	logger.Debug("initializing postgres connection pool")

	pool, err := core_pgx_pool.NewPool(
		ctx, core_pgx_pool.NewConfigMust(),
	)
	if err != nil {
		logger.Fatal("failed to init postgres connection pool", zap.Error(err))
	}
	defer pool.Close()

	natsCfg := nats.NewNatsConfigMust()

	natsClient, err := nats.Connect(*natsCfg)
	if err != nil {
		logger.Fatal("failed to connect to nats", zap.Error(err))
	}
	// Регистрируется после pool.Close, значит выполнится раньше него.
	defer func() {
		if err := natsClient.Shutdown(5 * time.Second); err != nil {
			logger.Error("nats shutdown", zap.Error(err))
		}
	}()

	pub := nats.NewPublisher(natsClient.Conn())

	logger.Debug("Initializing feature!", zap.String("feature", "tasks"))

	receiverRepository := receiver_postgres_repository.NewReceiverRepository(pool)
	receiverService := receiver_service.NewReceiverService(receiverRepository, pool, pub)
	receiverTransportHttp := receiver_transport_http.NewReceiverHTTPHandler(receiverService)

	senderService := sender_service.NewSenderService(receiverRepository, cfg.BaseSMS, cfg.APIKey)
	senderHandler := sender_transport_nats.NewNatsHandler(senderService)

	subscriber := nats.New(natsClient.Conn(), *natsCfg, senderHandler)
	fmt.Println(subscriber)

	httpServer := core_http_server.NewHTTPServer(
		core_http_server.NewConfigMust(),
		logger,
		core_http_middleware.CORS(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(logger),
		core_http_middleware.Trace(),
		core_http_middleware.Panic(),
	)
	apiVersionRouterv1 := core_http_server.NewAPIVersionRouter(core_http_server.ApiVersion1)
	apiVersionRouterv1.RegisterRoutes(receiverTransportHttp.Routes()...)

	httpServer.RegisterAPIRouters(
		apiVersionRouterv1,
		// apiVersionRouterv2,
	)
	httpServer.RegisterSwagger()

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := subscriber.Run(gctx); err != nil {
			return fmt.Errorf("nats subscriber: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := httpServer.Run(gctx); err != nil {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("application stopped with error", zap.Error(err))
	}
}
