package eventprocessor

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

// SimpleNatsClient implements basic NATS functionality
type SimpleNatsClient struct {
	conn *nats.Conn
	mu   sync.RWMutex
}

// NewSimpleNatsClient creates a new NATS client
func NewSimpleNatsClient(cfg Config) (*SimpleNatsClient, error) {
	opts := []nats.Option{
		nats.Token(cfg.Token),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	return &SimpleNatsClient{conn: nc}, nil
}

func (c *SimpleNatsClient) Publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn.Publish(subject, data)
}

func (c *SimpleNatsClient) Subscribe(subject string, handler func([]byte)) error {
	_, err := c.conn.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	})
	return err
}

func (c *SimpleNatsClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
	return nil
}
