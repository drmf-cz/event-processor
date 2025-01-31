package eventprocessor

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// DeDupeJetStreamClient implements NATS JetStream with deduplication.
type DedupJetStreamClient struct {
	*JetStreamClient
	config Config
	logger *zap.Logger
}

// NewDedupJetStreamClient creates a new NATS JetStream client with deduplication.
func NewDedupJetStreamClient(cfg Config, streamConfig jetstream.StreamConfig) (*DedupJetStreamClient, error) {
	js, err := NewJetStreamClient(cfg, streamConfig)
	if err != nil {
		return nil, err
	}

	if cfg.Logger == nil {
		return nil, fmt.Errorf("logger is required in config")
	}

	return &DedupJetStreamClient{
		JetStreamClient: js,
		config:          cfg,
		logger:          cfg.Logger,
	}, nil
}

// PublishToStream publishes a message to a stream
func (c *DedupJetStreamClient) PublishToStream(ctx context.Context, topic string, data []byte) error {
	_, err := c.js.Publish(ctx, topic, data)
	return err
}

// DeduplicateConsumer creates a durable pull consumer with deduplication for the stream
func (c *DedupJetStreamClient) DeduplicateConsumer(ctx context.Context, name string) (jetstream.ConsumeContext, error) {
	c.logger.Info("creating deduplicated consumer",
		zap.String("stream", c.streamConfig.Name),
		zap.String("name", name),
		zap.Duration("dedupe_window", c.streamConfig.Duplicates),
	)

	consumerConfig := jetstream.ConsumerConfig{
		Name:          name,
		Durable:       name,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
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
	return consumer.Consume(func(msg jetstream.Msg) {
		defer msg.Ack()
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
}

// Close closes the NATS connection
func (c *DedupJetStreamClient) Close(ctx context.Context) error {
	return c.JetStreamClient.Close(ctx)
}
