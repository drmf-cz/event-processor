package nats_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSimpleNatsClient(t *testing.T) {
	t.Parallel()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // fallback for local testing
	}

	logger, _ := zap.NewProduction()
	cfg := &nats.Config{
		URL:           natsURL,
		Token:         "test-token",
		CredsFile:     "",
		MaxReconnects: nats.DefaultMaxReconnects,
		ReconnectWait: time.Second * nats.DefaultReconnectWaitSeconds,
		Logger:        logger,
	}

	ctx := context.Background()

	t.Run("NewSimpleNatsClient", func(t *testing.T) {
		t.Parallel()
		client, err := nats.NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(ctx)
	})

	t.Run("PublishSubscribe", func(t *testing.T) {
		t.Parallel()
		client, err := nats.NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		defer client.Close(ctx)

		testMsg := []byte("test message")
		err = client.PublishToStream(ctx, "test.simple.subject", testMsg)
		require.NoError(t, err)
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		t.Parallel()
		invalidCfg := *cfg
		invalidCfg.URL = "nats://invalid:4222"
		_, err := nats.NewSimpleNatsClient(&invalidCfg)
		assert.Error(t, err)
	})

	t.Run("NilConfig", func(t *testing.T) {
		t.Parallel()
		_, err := nats.NewSimpleNatsClient(nil)
		assert.ErrorIs(t, err, nats.ErrInvalidConfig)
	})

	t.Run("CanceledContext", func(t *testing.T) {
		t.Parallel()
		client, err := nats.NewSimpleNatsClient(cfg)
		require.NoError(t, err)
		defer client.Close(ctx)

		canceledCtx, cancel := context.WithCancel(ctx)
		cancel()

		err = client.PublishToStream(canceledCtx, "test.subject", []byte("test"))
		assert.Error(t, err)
	})
}
