package nats

import "context"

// Message не зависит от брокера.
type Message struct {
	Subject string
	Data    []byte
	Headers map[string]string
}

// Handler реализует бизнес-логика. Ошибка передаётся в ErrorHandler подписчика.
type Handler interface {
	Handle(ctx context.Context, msg Message) error
}

type HandlerFunc func(ctx context.Context, msg Message) error

func (f HandlerFunc) Handle(ctx context.Context, msg Message) error { return f(ctx, msg) }

// Subscriber запускается в main и блокируется до отмены ctx.
// Реализаций может быть сколько угодно (NATS, Kafka, тестовая).
type Subscriber interface {
	Run(ctx context.Context) error
}
