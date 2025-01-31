package nats

import (
	"time"

	"go.uber.org/zap"
)

// Config holds common configuration for event processors.
type Config struct {
	// URL is the NATS server URL
	URL string
	// Token is the authentication token (deprecated)
	Token string // Deprecated: Use CredsFile for JWT authentication
	// CredsFile is the path to the credentials file for JWT authentication
	CredsFile string
	// MaxReconnects is the maximum number of reconnection attempts
	MaxReconnects int
	// ReconnectWait is the duration to wait between reconnection attempts
	ReconnectWait time.Duration
	// Logger is the configured zap logger instance
	Logger *zap.Logger
}
