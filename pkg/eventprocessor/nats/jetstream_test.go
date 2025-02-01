package nats_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const (
	defaultNatsURL = "nats://nats:4222"
	testTimeout    = 5 * time.Second
	messageCount   = 100
)

func TestJetStreamClient(t *testing.T) {
	t.Parallel()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = defaultNatsURL // fallback for local testing
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := &nats.Config{
		Logger:        logger,
		URL:           natsURL,
		Token:         "test-token",
		CredsFile:     "",
		MaxReconnects: nats.DefaultMaxReconnects,
		ReconnectWait: time.Second * nats.DefaultReconnectWaitSeconds,
	}

	t.Run("NewJetStreamClient", func(t *testing.T) {
		t.Parallel()
		client, err := nats.NewJetStreamClient(cfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:       "TEST_JETSTREAM_1",
			Subjects:   []string{"test.jetstream1.>"},
			Duplicates: 100 * time.Millisecond,
		})
		require.NoError(t, err)
		require.NotNil(t, client)
		defer client.Close(context.Background())
	})

	t.Run("ConsumerTest", func(t *testing.T) {
		t.Parallel()

		// Create test logger with observer for assertions
		core, logs := observer.New(zap.InfoLevel)
		logger := zap.New(core)
		testCfg := *cfg
		testCfg.Logger = logger

		client, err := nats.NewJetStreamClient(&testCfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:       "TEST_JETSTREAM_2",
			Subjects:   []string{"test.jetstream2.>"},
			Duplicates: 100 * time.Millisecond,
		})
		require.NoError(t, err)
		defer client.Close(context.Background())

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		// Publish messages to the stream
		for i := range messageCount {
			err := client.PublishToStream(ctx, "test.jetstream2."+strconv.Itoa(i), []byte("data"))
			require.NoError(t, err)
		}

		// Create consumer and start consuming
		cc, err := client.CreateConsumer(ctx, "test-jetstream2")
		require.NoError(t, err)
		require.NotNil(t, cc)
		defer cc.Stop()

		// Wait for all messages to be processed
		require.Eventually(t, func() bool {
			entries := logs.All()
			for _, entry := range entries {
				if seq, ok := entry.ContextMap()["consumer_sequence"].(uint64); ok && seq == messageCount {
					return true
				}
			}

			return false
		}, testTimeout, 100*time.Millisecond, "did not receive all messages or stream is already filled")
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		t.Parallel()
		invalidCfg := &nats.Config{
			URL:           "nats://invalid:4222",
			Token:         cfg.Token,
			CredsFile:     "",
			MaxReconnects: cfg.MaxReconnects,
			ReconnectWait: cfg.ReconnectWait,
			Logger:        cfg.Logger,
		}
		_, err := nats.NewJetStreamClient(invalidCfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:     "TEST_JETSTREAM_3",
			Subjects: []string{"test.jetstream3.>"},
		})
		assert.Error(t, err)
	})
}
