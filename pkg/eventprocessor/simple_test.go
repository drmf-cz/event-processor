package eventprocessor

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleNatsClient(t *testing.T) {
	cfg := Config{
		URL:           "nats://localhost:4222",
		Token:         "development-token-123",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
	}

	t.Run("NewSimpleNatsClient", func(t *testing.T) {
		client, err := NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("PublishSubscribe", func(t *testing.T) {
		client, err := NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		msgChan := make(chan []byte, 1)
		err = client.Subscribe("test.subject", func(data []byte) {
			msgChan <- data
		})
		require.NoError(t, err)

		testMsg := []byte("test message")
		err = client.Publish("test.subject", testMsg)
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
		_, err := NewSimpleNatsClient(invalidCfg)
		assert.Error(t, err)
	})
}
