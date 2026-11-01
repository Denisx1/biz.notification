package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc     *nats.Conn
	closed chan struct{}
}

func Connect(cfg Config) (*Client, error) {
	closed := make(chan struct{})

	nc, err := nats.Connect(cfg.URL,
		nats.Name(cfg.Name),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(time.Second),
		nats.ClosedHandler(func(*nats.Conn) { close(closed) }),
	)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	return &Client{nc: nc, closed: closed}, nil
}

func (c *Client) Conn() *nats.Conn { return c.nc }

// Shutdown дожидается отправки буфера и закрывает соединение.
func (c *Client) Shutdown(timeout time.Duration) error {
	if err := c.nc.Drain(); err != nil {
		return fmt.Errorf("nats drain: %w", err)
	}
	select {
	case <-c.closed:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("nats drain: timeout")
	}
}
