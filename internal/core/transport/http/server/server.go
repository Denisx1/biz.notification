package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	core_logger "notification/internal/core/logger"
	core_http_middleware "notification/internal/core/transport/http/middlware"

	httpSwagger "github.com/swaggo/http-swagger/v2"
	"go.uber.org/zap"
)

type HTTPServer struct {
	mux        *http.ServeMux // мультипексор сущность что бы распознать чреез каикие миддлвары нааж пропустить нашу сузщность
	config     Config
	log        *core_logger.Logger
	middleware []core_http_middleware.Middleware
}

func NewHTTPServer(config Config, log *core_logger.Logger, middleware ...core_http_middleware.Middleware) *HTTPServer {
	return &HTTPServer{
		mux:        http.NewServeMux(),
		config:     config,
		log:        log,
		middleware: middleware,
	}
}

//	func (h *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
//		for _, router := range routers {
//			prefix := "/api" + string(router.apiVersion)
//			h.mux.Handle(
//				prefix+"/",
//				http.StripPrefix(prefix, router),
//			)
//		}
//	}
func (s *HTTPServer) RegisterAPIRouters(routers ...*APIVersionRouter) {
	for _, router := range routers {
		// 1. ИСПРАВЛЕНО: Добавлен пропущенный слэш между "api" и версией.
		// Теперь получится "/api/v1", а не "/apiv1"
		prefix := fmt.Sprintf("/api/%s", string(router.apiVersion))

		// 2. ИСПРАВЛЕНО: В Handle передаем префикс СО слэшем на конце ("/api/v1/"),
		// а в StripPrefix передаем префикс БЕЗ слэша на конце ("/api/v1").
		// Это заставит Go оставлять начальный слэш у путей (будет "/users", а не "users").
		s.mux.Handle(
			prefix+"/",
			http.StripPrefix(prefix, router.WithMiddleware()),
		)
	}
}
func (s *HTTPServer) RegisterSwagger() {
	s.mux.Handle(
		"/swagger/",
		httpSwagger.Handler(
			httpSwagger.URL("/swagger/doc.json"),
			httpSwagger.DefaultModelsExpandDepth(-1),
		),
	)
	s.mux.HandleFunc(
		"/swagger/doc.json",
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-type", "application/json")
			w.WriteHeader(http.StatusOK)
		},
	)
}
func (s *HTTPServer) Run(ctx context.Context) error {
	mux := core_http_middleware.ChainMiddleware(s.mux, s.middleware...)
	server := &http.Server{
		Addr:    s.config.Addr,
		Handler: mux,
	}
	ch := make(chan error, 1)
	go func() {
		defer close(ch)
		s.log.Warn("Start http server", zap.String("addr", s.config.Addr))
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()
	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("Listen and serve HTTP:%w", err)
		}
	case <-ctx.Done():
		s.log.Warn("Shutdown HTTP server")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.ShutDownTimeOut,
		)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		s.log.Warn("HTTP server stopped")
	}
	return nil
}
