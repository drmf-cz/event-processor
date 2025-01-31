package eventprocessor

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// DeDupeJetStreamClient implements NATS JetStream with deduplication
type DeDupeJetStreamClient struct {
	*JetStreamClient
	dedupeWindow time.Duration
}

// NewDeDupeJetStreamClient creates a new NATS JetStream client with deduplication
func NewDeDupeJetStreamClient(cfg Config) (*DeDupeJetStreamClient, error) {
	js, err := NewJetStreamClient(cfg)
	if err != nil {
		return nil, err
	}

	return &DeDupeJetStreamClient{
		JetStreamClient: js,
		dedupeWindow:    cfg.DeDupeWindow,
	}, nil
}

func (c *DeDupeJetStreamClient) Publish(subject string, data []byte) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Add message ID for deduplication
	msgID := fmt.Sprintf("%d", time.Now().UnixNano())
	_, err := c.js.Publish(subject, data, nats.MsgId(msgID))
	return err
}

func (c *DeDupeJetStreamClient) Subscribe(subject string, handler func([]byte)) error {
	// Configure consumer with deduplication
	_, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		handler(msg.Data)
	}, nats.DeliverAll(), nats.AckWait(c.dedupeWindow))
	return err
}
