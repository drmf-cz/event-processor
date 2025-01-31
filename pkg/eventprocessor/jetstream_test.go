package eventprocessor_test

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestJetStreamClient(t *testing.T) {
	t.Parallel()

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222" // fallback for local testing
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	cfg := eventprocessor.Config{
		Logger:        logger,
		URL:           natsURL,
		Token:         "test-token",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
	}

	t.Run("NewJetStreamClient", func(t *testing.T) {
		t.Parallel()
		client, err := eventprocessor.NewJetStreamClient(cfg, jetstream.StreamConfig{
			Name:       "TEST_JETSREAM",
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
		testCfg := cfg
		testCfg.Logger = logger

		client, err := eventprocessor.NewJetStreamClient(testCfg, jetstream.StreamConfig{
			Name:       "TEST_JETSREAM",
			Subjects:   []string{"test.jetstream2.>"},
			Duplicates: 100 * time.Millisecond,
		})
		require.NoError(t, err)
		defer client.Close(context.Background())

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Publish 100 messages to the stream
		go func() {
			for i := 0; i < 100; i++ {
				_, err = client.PublishToStream(ctx, "test.jetstream2."+strconv.Itoa(i), []byte("data"))
				require.NoError(t, err)
			}
		}()

		// Create consumer and start consuming
		cc, err := client.CreateConsumer(ctx, "test-consumer")
		require.NoError(t, err)
		require.NotNil(t, cc)
		defer cc.Stop()

		// Wait for all messages to be processed
		select {
		case <-ctx.Done():
			t.Fatal("context deadline exceeded")
		default:
			// Check logs until we find the 100th message
			require.Eventually(t, func() bool {
				entries := logs.All()
				for _, entry := range entries {
					if seq, ok := entry.ContextMap()["consumer_sequence"].(uint64); ok && seq == 100 {
						return true
					}
				}
				return false
			}, 5*time.Second, 100*time.Millisecond, "did not receive all 100 messages or stream is already filled")
		}
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		t.Parallel()
		invalidCfg := eventprocessor.Config{
			URL:           "nats://invalid:4222",
			Token:         cfg.Token,
			MaxReconnects: cfg.MaxReconnects,
			ReconnectWait: cfg.ReconnectWait,
		}
		_, err := eventprocessor.NewJetStreamClient(invalidCfg, jetstream.StreamConfig{
			Name:     "TEST_JETSREAM",
			Subjects: []string{"test.jetstream3.>"},
		})
		assert.Error(t, err)
	})
}
