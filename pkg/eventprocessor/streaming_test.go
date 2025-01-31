package eventprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamingClient(t *testing.T) {
	cfg := Config{
		URL:           "nats://localhost:4222",
		ClusterID:     "event-processor-cluster",
		ClientID:      "test-client",
		Token:         "development-token-123",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
	}

	t.Run("NewStreamingClient", func(t *testing.T) {
		client, err := NewStreamingClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("PublishSubscribe", func(t *testing.T) {
		client, err := NewStreamingClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		msgChan := make(chan []byte, 1)
		err = client.Subscribe("test.streaming.subject", func(data []byte) {
			msgChan <- data
		})
		require.NoError(t, err)

		testMsg := []byte("test streaming message")
		err = client.Publish("test.streaming.subject", testMsg)
		require.NoError(t, err)

		select {
		case receivedMsg := <-msgChan:
			assert.Equal(t, testMsg, receivedMsg)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("DurableSubscription", func(t *testing.T) {
		client, err := NewStreamingClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		subject := "test.durable.subject"
		msgChan := make(chan []byte, 1)

		// First subscription
		err = client.Subscribe(subject, func(data []byte) {
			msgChan <- data
		})
		require.NoError(t, err)

		// Publish message
		testMsg := []byte("test durable message")
		err = client.Publish(subject, testMsg)
		require.NoError(t, err)

		select {
		case receivedMsg := <-msgChan:
			assert.Equal(t, testMsg, receivedMsg)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.URL = "nats://invalid:4222"
		_, err := NewStreamingClient(invalidCfg)
		assert.Error(t, err)
	})
}
