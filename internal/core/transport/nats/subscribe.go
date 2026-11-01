package nats

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Subscribe struct {
	nc      *nats.Conn
	cfg     Config
	handler Handler
}

func New(nc *nats.Conn, cfg Config, h Handler) *Subscribe {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(subject string, err error) {
			log.Printf("nats handler %s: %v", subject, err)
		}
	}
	return &Subscribe{nc: nc, cfg: cfg, handler: h}
}

func (s *Subscribe) Run(ctx context.Context) error {
	handlerCtx := context.WithoutCancel(ctx)

	cb := func(m *nats.Msg) {
		if err := s.handler.Handle(handlerCtx, toMessage(m)); err != nil {
			s.cfg.ErrorHandler(m.Subject, err)
		}
	}

	var (
		sub *nats.Subscription
		err error
	)
	if s.cfg.Queue != "" {
		sub, err = s.nc.QueueSubscribe(s.cfg.Subject, s.cfg.Queue, cb)
	} else {
		sub, err = s.nc.Subscribe(s.cfg.Subject, cb)
	}
	if err != nil {
		return fmt.Errorf("nats subscribe %s: %w", s.cfg.Subject, err)
	}

	<-ctx.Done()

	// Дренится только подписка: доработает уже принятые сообщения.
	if err := sub.Drain(); err != nil {
		return fmt.Errorf("nats drain sub: %w", err)
	}
	return nil
}
func toMessage(m *nats.Msg) Message {
	hs := make(map[string]string, len(m.Header))
	for k := range m.Header {
		hs[k] = m.Header.Get(k)
	}
	return Message{Subject: m.Subject, Data: m.Data, Headers: hs}
}
