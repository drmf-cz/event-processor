package eventprocessor_test

import (
	"os"
	"testing"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleNatsClient(t *testing.T) {
	t.Parallel()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // fallback for local testing
	}

	cfg := eventprocessor.Config{
		URL:           natsURL,
		Token:         "test-token",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
	}

	t.Run("NewSimpleNatsClient", func(t *testing.T) {
		t.Parallel()
		client, err := eventprocessor.NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close()
	})

	t.Run("PublishSubscribe", func(t *testing.T) {
		t.Parallel()
		client, err := eventprocessor.NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		defer client.Close()

		msgChan := make(chan []byte, 1)
		err = client.Subscribe("test.simple.subject", func(data []byte) {
			msgChan <- data
		})
		require.NoError(t, err)

		testMsg := []byte("test message")
		err = client.Publish("test.simple.subject", testMsg)
		require.NoError(t, err)

		select {
		case receivedMsg := <-msgChan:
			assert.Equal(t, testMsg, receivedMsg)
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		t.Parallel()
		invalidCfg := cfg
		invalidCfg.URL = "nats://invalid:4222"
		_, err := eventprocessor.NewSimpleNatsClient(invalidCfg)
		assert.Error(t, err)
	})
}
