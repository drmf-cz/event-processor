package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"go.uber.org/zap"
)

// Handler handles HTTP requests for the event processor API
type Handler struct {
	logger   *zap.Logger
	jsClient *nats.JetStreamClient
}

// NewHandler creates a new Handler instance
func NewHandler(logger *zap.Logger, jsClient *nats.JetStreamClient) *Handler {
	return &Handler{
		logger:   logger,
		jsClient: jsClient,
	}
}

// PublishMessage handles POST requests to publish messages to NATS
func (h *Handler) PublishMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		h.logger.Error("Failed to unmarshal request body", zap.Error(err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := h.jsClient.PublishToStream(r.Context(), "test.jetstream1.messages", []byte(msg.Data)); err != nil {
		h.logger.Error("Failed to publish message", zap.Error(err))
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Response{Status: "published"}) //nolint:errcheck
}

// GetLatestMessage handles GET requests to receive the latest message from NATS
func (h *Handler) GetLatestMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a channel to receive the message
	msgChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	// Create a consumer with a callback that sends the message to our channel
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	// Create a consumer that will receive one message and send it to our channel
	consumer, err := h.jsClient.CreateConsumer(ctx, "latest-message-consumer")
	if err != nil {
		h.logger.Error("Failed to create consumer", zap.Error(err))
		http.Error(w, "Failed to create consumer", http.StatusInternalServerError)
		return
	}

	// Start consuming messages
	go func() {
		defer close(msgChan)
		defer close(errChan)
		defer consumer.Stop()

		// Wait for a message
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
		case msg := <-consumer.Messages(): // Note: This is a placeholder as Messages() is not implemented
			if err := msg.Ack(); err != nil {
				h.logger.Error("Failed to acknowledge message", zap.Error(err))
			}
			msgChan <- msg.Data()
		}
	}()

	// Wait for either a message, error, or timeout
	select {
	case <-ctx.Done():
		w.WriteHeader(http.StatusNoContent)
		return
	case err := <-errChan:
		if err == context.DeadlineExceeded {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.logger.Error("Error receiving message", zap.Error(err))
		http.Error(w, "Failed to receive message", http.StatusInternalServerError)
		return
	case msg := <-msgChan:
		response := Message{
			Data: string(msg),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response) //nolint:errcheck
	}
}
