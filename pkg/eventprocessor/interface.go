package eventprocessor

import (
	"context"
	"errors"
	"os"
	"time"

	"go.uber.org/zap"
)

// Common errors
var (
	ErrInvalidConfig = errors.New("invalid configuration")
)

// EventProcessor defines the interface for different event processing strategies.
type EventProcessor interface {
	// PublishToStream publishes a message to a stream.
	PublishToStream(ctx context.Context, topic string, data []byte) error
	// Close closes the connection.
	Close(ctx context.Context) error
}

// Config holds common configuration for event processors.
type Config struct {
	URL           string
	Token         string // Deprecated: Use CredsFile for JWT authentication
	CredsFile     string // Path to the credentials file for JWT authentication
	MaxReconnects int
	ReconnectWait time.Duration
	Logger        *zap.Logger
}

// NewConfig creates a new configuration with values from environment.
func NewConfig() *Config {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	return &Config{
		URL:           os.Getenv("NATS_URL"),
		CredsFile:     os.Getenv("NATS_CREDS"),
		MaxReconnects: 5,
		ReconnectWait: time.Second * 5,
		Logger:        logger,
	}
}
