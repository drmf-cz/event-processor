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
	config       *Config
	stream       jetstream.Stream
	streamConfig jetstream.StreamConfig
	logger       *zap.Logger
}

// NewJetStreamClient creates a new NATS JetStream client.
func NewJetStreamClient(cfg *Config, streamConfig jetstream.StreamConfig) (*JetStreamClient, error) {
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

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()

		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	stream, err := js.CreateStream(context.Background(), streamConfig)
	if err != nil {
		nc.Close()

		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return &JetStreamClient{
		conn:         nc,
		js:           js,
		mu:           sync.RWMutex{},
		config:       cfg,
		streamConfig: streamConfig,
		stream:       stream,
		logger:       cfg.Logger,
	}, nil
}

// PublishToStream publishes a message to a stream.
func (c *JetStreamClient) PublishToStream(ctx context.Context, topic string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	_, err := c.js.Publish(ctx, topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// CreateConsumer creates a durable pull consumer for the stream.
// Returns a ConsumeContext that must be used to receive messages.
func (c *JetStreamClient) CreateConsumer(ctx context.Context, name string) (jetstream.ConsumeContext, error) { //nolint: ireturn
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	consumerConfig := jetstream.ConsumerConfig{ //nolint: exhaustruct
		Name:               name,
		Durable:            name,
		DeliverPolicy:      jetstream.DeliverAllPolicy,
		AckPolicy:          jetstream.AckExplicitPolicy,
		Description:        fmt.Sprintf("Consumer %s for stream %s", name, c.streamConfig.Name),
		MaxRequestBatch:    DefaultMaxRequestBatch,
		MaxRequestExpires:  c.config.ReconnectWait,
		MaxRequestMaxBytes: DefaultMaxRequestMaxBytes,
		InactiveThreshold:  c.config.ReconnectWait * DefaultInactiveThresholdMultiplier,
	}

	consumer, err := c.stream.CreateOrUpdateConsumer(ctx, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	// Create consume context with options
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		defer func() {
			if err := msg.Ack(); err != nil {
				c.logger.Error("failed to acknowledge message", zap.Error(err))
			}
		}()
		meta, err := msg.Metadata()
		if err != nil {
			c.logger.Error("failed to get metadata", zap.Error(err))

			return
		}

		c.logger.Info("received message",
			zap.Uint64("consumer_sequence", meta.Sequence.Consumer),
			zap.String("subject", msg.Subject()))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consume context: %w", err)
	}

	return cc, nil
}

// Close closes the NATS connection and cleans up resources.
func (c *JetStreamClient) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.js.DeleteStream(ctx, c.streamConfig.Name); err != nil {
		c.logger.Error("failed to delete stream", zap.Error(err))
	}

	c.conn.Close()

	return nil
}
