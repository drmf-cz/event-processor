package nats

import (
	"context"
	"errors"
	"os"
	"time"

	"go.uber.org/zap"
)

// Common errors.
var (
	// ErrInvalidConfig is returned when the configuration is invalid.
	ErrInvalidConfig = errors.New("invalid configuration")
)

// EventProcessor defines the interface for different event processing strategies.
// It provides methods for publishing messages to streams and managing connections.
type EventProcessor interface {
	// PublishToStream publishes a message to a stream.
	// ctx provides context for the operation
	// topic is the stream or subject to publish to
	// data is the message payload to publish
	// Returns an error if the publish operation fails
	PublishToStream(ctx context.Context, topic string, data []byte) error

	// Close gracefully shuts down the event processor and its connections.
	// ctx provides context for the shutdown operation
	// Returns an error if the shutdown fails
	Close(ctx context.Context) error
}

// NewConfig creates a new configuration with values from environment.
// It initializes a production logger and sets default values for reconnection parameters.
func NewConfig() *Config {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return &Config{
		URL:           os.Getenv("NATS_URL"),
		Token:         os.Getenv("NATS_TOKEN"),
		CredsFile:     os.Getenv("NATS_CREDS"),
		MaxReconnects: DefaultMaxReconnects,
		ReconnectWait: time.Second * DefaultReconnectWaitSeconds,
		Logger:        logger,
	}
}
