package eventprocessor

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

// SimpleNatsClient implements basic NATS functionality.
type SimpleNatsClient struct {
	conn   *nats.Conn
	config Config
}

// NewSimpleNatsClient creates a new NATS client.
func NewSimpleNatsClient(cfg Config) (*SimpleNatsClient, error) {
	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	// Use credentials file if provided, fallback to token
	if cfg.CredsFile != "" {
		opts = append(opts, nats.UserCredentials(cfg.CredsFile))
	} else if cfg.Token != "" {
		opts = append(opts, nats.Token(cfg.Token))
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &SimpleNatsClient{conn: nc, config: cfg}, nil
}

func (c *SimpleNatsClient) Publish(subject string, data []byte) error {
	err := c.conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (c *SimpleNatsClient) Subscribe(subject string, handler func([]byte)) error {
	_, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	return nil
}

func (c *SimpleNatsClient) Close() error {
	c.conn.Close()

	return nil
}
