package eventprocessor

const (
	// DefaultMaxReconnects is the default number of reconnection attempts.
	DefaultMaxReconnects = 5
	// DefaultReconnectWaitSeconds is the default wait time between reconnection attempts in seconds.
	DefaultReconnectWaitSeconds = 5

	// DefaultMaxRequestBatch is the default number of messages to request in a batch.
	DefaultMaxRequestBatch = 100
	// DefaultMaxRequestMaxBytes is the default maximum bytes to request (1MB).
	DefaultMaxRequestMaxBytes = 1024 * 1024
	// DefaultInactiveThresholdMultiplier is the multiplier for inactive threshold.
	DefaultInactiveThresholdMultiplier = 2
)
