package eventprocessor

import (
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

// JetStreamClient implements NATS JetStream functionality
type JetStreamClient struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	mu     sync.RWMutex
	config Config
}

// NewJetStreamClient creates a new NATS JetStream client
func NewJetStreamClient(cfg Config) (*JetStreamClient, error) {
	opts := []nats.Option{
		nats.Token(cfg.Token),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &JetStreamClient{conn: nc, js: js, config: cfg}, nil
}

func (c *JetStreamClient) Publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, err := c.js.Publish(subject, data)
	return err
}

func (c *JetStreamClient) Subscribe(subject string, handler func([]byte)) error {
	_, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	}, nats.DeliverAll())
	return err
}

func (c *JetStreamClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn.Close()
	return nil
}
