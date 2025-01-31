package eventprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeDupeJetStreamClient(t *testing.T) {
	cfg := Config{
		URL:           "nats://localhost:4222",
		Token:         "development-token-123",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
		DeDupeWindow:  time.Second * 30,
	}

	t.Run("NewDeDupeJetStreamClient", func(t *testing.T) {
		client, err := NewDeDupeJetStreamClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("PublishSubscribeWithDedupe", func(t *testing.T) {
		client, err := NewDeDupeJetStreamClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		msgChan := make(chan []byte, 2)
		err = client.Subscribe("test.dedupe.subject", func(data []byte) {
			msgChan <- data
		})
		require.NoError(t, err)

		testMsg := []byte("test dedupe message")

		// Send same message twice quickly
		err = client.Publish("test.dedupe.subject", testMsg)
		require.NoError(t, err)
		err = client.Publish("test.dedupe.subject", testMsg)
		require.NoError(t, err)

		// Should only receive one message due to deduplication
		select {
		case receivedMsg := <-msgChan:
			assert.Equal(t, testMsg, receivedMsg)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for first message")
		}

		// Should not receive second message
		select {
		case <-msgChan:
			t.Fatal("received duplicate message")
		case <-time.After(time.Second):
			// This is expected
		}
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		invalidCfg := cfg
		invalidCfg.URL = "nats://invalid:4222"
		_, err := NewDeDupeJetStreamClient(invalidCfg)
		assert.Error(t, err)
	})
}
