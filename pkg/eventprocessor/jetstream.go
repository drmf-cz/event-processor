package eventprocessor

import (
	"context"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// JetStreamClient implements NATS JetStream functionality.
type JetStreamClient struct {
	conn         *nats.Conn
	js           jetstream.JetStream
	mu           sync.RWMutex
	config       Config
	stream       jetstream.Stream
	streamConfig jetstream.StreamConfig
	logger       *zap.Logger
}

// NewJetStreamClient creates a new NATS JetStream client.
func NewJetStreamClient(cfg Config, streamConfig jetstream.StreamConfig) (*JetStreamClient, error) {
	opts := []nats.Option{
		nats.Token(cfg.Token),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	}

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()

		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	stream, err := js.CreateStream(context.Background(), streamConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return &JetStreamClient{conn: nc, js: js, config: cfg, streamConfig: streamConfig, stream: stream, logger: cfg.Logger}, nil
}

// PublishToStream publishes a message to a stream.
func (c *JetStreamClient) PublishToStream(ctx context.Context, topic string, data []byte) error {
	_, err := c.js.Publish(ctx, topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// CreateConsumer creates a durable pull consumer for the stream.
func (c *JetStreamClient) CreateConsumer(ctx context.Context, name string) (jetstream.ConsumeContext, error) {
	consumerConfig := jetstream.ConsumerConfig{
		Name:          name,
		Durable:       name,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
	}

	consumer, err := c.stream.CreateOrUpdateConsumer(ctx, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create consume context with options
	return consumer.Consume(func(msg jetstream.Msg) {
		defer func() {
			if err := msg.Ack(); err != nil {
				c.logger.Error("failed to acknowledge message", zap.Error(err))
			}
		}()
		meta, err := msg.Metadata()
		if err != nil {
			c.logger.Error("failed to get metadata",
				zap.Error(err))

			return
		}

		c.logger.Info("received message",
			zap.Uint64("consumer_sequence", meta.Sequence.Consumer),
			zap.String("subject", msg.Subject()))
	})
}

// Close closes the NATS connection.
func (c *JetStreamClient) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.js.DeleteStream(ctx, c.streamConfig.Name)
	c.conn.Close()

	return nil
}
