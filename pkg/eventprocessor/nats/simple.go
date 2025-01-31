package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

// SimpleNatsClient implements the EventProcessor interface with basic NATS functionality.
type SimpleNatsClient struct {
	conn   *nats.Conn
	config *Config
}

// NewSimpleNatsClient creates a new NATS client with the provided configuration.
// Returns an error if the connection cannot be established.
func NewSimpleNatsClient(cfg *Config) (*SimpleNatsClient, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

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

// PublishToStream implements the EventProcessor interface.
// It publishes a message to the specified NATS subject.
func (c *SimpleNatsClient) PublishToStream(ctx context.Context, subject string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := c.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close implements the EventProcessor interface.
// It gracefully closes the NATS connection.
func (c *SimpleNatsClient) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	c.conn.Close()

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
