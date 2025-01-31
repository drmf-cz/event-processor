package eventprocessor

import "time"

// EventProcessor defines the interface for different event processing strategies
type EventProcessor interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, handler func([]byte)) error
	Close() error
}

// Config holds common configuration for event processors
type Config struct {
	URL           string
	ClusterID     string
	ClientID      string
	Token         string
	DeDupeWindow  time.Duration
	MaxReconnects int
	ReconnectWait time.Duration
}
