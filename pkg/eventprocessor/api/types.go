package api

// Message represents the structure of messages sent to NATS
type Message struct {
	Data string `json:"data"`
}

// Response represents a generic API response
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
