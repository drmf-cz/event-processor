package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// DedupJetStreamClient implements NATS JetStream with deduplication.
type DedupJetStreamClient struct {
	*JetStreamClient
	config *Config
	logger *zap.Logger
}

// NewDedupJetStreamClient creates a new NATS JetStream client with deduplication.
func NewDedupJetStreamClient(cfg *Config, streamConfig jetstream.StreamConfig) (*DedupJetStreamClient, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	js, err := NewJetStreamClient(cfg, streamConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream client: %w", err)
	}

	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required in config: %w", ErrInvalidConfig)
	}

	return &DedupJetStreamClient{
		JetStreamClient: js,
		config:          cfg,
		logger:          cfg.Logger,
	}, nil
}

// PublishToStream publishes a message to a stream.
func (c *DedupJetStreamClient) PublishToStream(ctx context.Context, topic string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	_, err := c.js.Publish(ctx, topic, data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// DeduplicateConsumer creates a durable pull consumer with deduplication for the stream.
// Returns a ConsumeContext that must be used to receive messages.
func (c *DedupJetStreamClient) DeduplicateConsumer(ctx context.Context, name string) (jetstream.ConsumeContext, error) { //nolint: ireturn
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context error: %w", err)
	}

	c.logger.Info("creating deduplicated consumer",
		zap.String("stream", c.streamConfig.Name),
		zap.String("name", name),
		zap.Duration("dedupe_window", c.streamConfig.Duplicates),
	)

	consumerConfig := jetstream.ConsumerConfig{ //nolint: exhaustruct
		Name:               name,
		Durable:            name,
		DeliverPolicy:      jetstream.DeliverAllPolicy,
		AckPolicy:          jetstream.AckExplicitPolicy,
		Description:        fmt.Sprintf("Deduplicated consumer %s for stream %s", name, c.streamConfig.Name),
		MaxRequestBatch:    DefaultMaxRequestBatch,
		MaxRequestExpires:  c.config.ReconnectWait,
		MaxRequestMaxBytes: DefaultMaxRequestMaxBytes,
		InactiveThreshold:  c.config.ReconnectWait * DefaultInactiveThresholdMultiplier,
	}

	consumer, err := c.stream.CreateOrUpdateConsumer(ctx, consumerConfig)
	if err != nil {
		c.logger.Error("failed to create consumer",
			zap.String("name", name),
			zap.Error(err),
		)

		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	c.logger.Info("consumer created successfully",
		zap.String("name", name),
	)

	// Create consume context with options
	cc, err := consumer.Consume(func(msg jetstream.Msg) {
		defer func() {
			if err := msg.Ack(); err != nil {
				c.logger.Error("failed to acknowledge message", zap.Error(err))
			}
		}()
		meta, err := msg.Metadata()
		if err != nil {
			c.logger.Error("failed to get metadata",
				zap.Error(err),
				zap.String("subject", msg.Subject()),
			)

			return
		}

		c.logger.Info("received deduplicated message",
			zap.Uint64("sequence", meta.Sequence.Consumer),
			zap.String("subject", msg.Subject()),
			zap.String("consumer", name),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consume context: %w", err)
	}

	return cc, nil
}

// Close closes the NATS connection and cleans up resources.
func (c *DedupJetStreamClient) Close(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}

	if err := c.js.DeleteStream(ctx, c.streamConfig.Name); err != nil {
		c.logger.Error("failed to delete stream", zap.Error(err))
	}

	return c.JetStreamClient.Close(ctx)
}
