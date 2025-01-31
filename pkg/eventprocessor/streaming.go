package eventprocessor

import (
	"fmt"
	"sync"

	"github.com/nats-io/stan.go"
)

// StreamingClient implements NATS Streaming functionality
type StreamingClient struct {
	conn   stan.Conn
	mu     sync.RWMutex
	config Config
}

// NewStreamingClient creates a new NATS Streaming client
func NewStreamingClient(cfg Config) (*StreamingClient, error) {
	sc, err := stan.Connect(
		cfg.ClusterID,
		cfg.ClientID,
		stan.NatsURL(cfg.URL),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS Streaming: %w", err)
	}

	return &StreamingClient{conn: sc, config: cfg}, nil
}

func (c *StreamingClient) Publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn.Publish(subject, data)
}

func (c *StreamingClient) Subscribe(subject string, handler func([]byte)) error {
	_, err := c.conn.Subscribe(subject, func(msg *stan.Msg) {
		handler(msg.Data)
	}, stan.DurableName("durable-"+subject))
	return err
}

func (c *StreamingClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}
