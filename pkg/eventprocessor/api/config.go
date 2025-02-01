package api

import "time"

// Config holds the configuration for the API server
type Config struct {
	Port            int           `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

// DefaultConfig returns a default configuration for the API server
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		ShutdownTimeout: 5 * time.Second,
	}
}
