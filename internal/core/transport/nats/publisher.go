package nats

import (
	"context"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher { return &Publisher{nc: nc} }

func (p *Publisher) Publish(ctx context.Context, subject string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := p.nc.Publish(subject, data); err != nil {
		return err
	}
	log.Printf("nats sent: subject=%s size=%d", subject, len(data))
	return nil
}
