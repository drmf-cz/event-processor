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
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
)

const (
	invalidNatsURL = "nats://invalid:4222"
)

func TestDeDupeJetStreamClient(t *testing.T) { //nolint: gocognit, funlen
	t.Parallel()

	const (
		testTimeout  = 5 * time.Second
		messageCount = 100
	)

	// Create test logger
	logger := zaptest.NewLogger(t)
	_ = logger.Sync() // Ignore sync errors in tests

	cfg := &nats.Config{
		URL:           os.Getenv("NATS_URL"),
		Token:         "test-token",
		CredsFile:     "",
		MaxReconnects: nats.DefaultMaxReconnects,
		ReconnectWait: time.Second * nats.DefaultReconnectWaitSeconds,
		Logger:        logger,
	}

	t.Run("NewDeDupeJetStreamClient", func(t *testing.T) {
		t.Parallel()
		client, err := nats.NewDedupJetStreamClient(cfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:       "TEST_DEDUPE_1",
			Subjects:   []string{"test.dedupe1.>"},
			Duplicates: time.Minute,
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

		client, err := nats.NewDedupJetStreamClient(&testCfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:       "TEST_DEDUPE_2",
			Subjects:   []string{"test.dedupe2.>"},
			Duplicates: time.Minute,
		})
		require.NoError(t, err)
		defer client.Close(context.Background())

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		// Create a channel to signal message publishing completion
		done := make(chan struct{})

		// Publish messages to the stream
		go func() {
			defer close(done)
			for i := range messageCount {
				if err := client.PublishToStream(ctx, "test.dedupe2."+strconv.Itoa(i), []byte("test-"+strconv.Itoa(i))); err != nil {
					t.Errorf("failed to publish message: %v", err)

					return
				}
			}
		}()

		// Create consumer and start consuming
		cc, err := client.DeduplicateConsumer(ctx, "test-consumer")
		require.NoError(t, err)
		require.NotNil(t, cc)
		defer cc.Stop()

		// Wait for all messages to be processed
		select {
		case <-ctx.Done():
			t.Fatal("context deadline exceeded")
		case <-done:
			// Check logs until we find the last message
			require.Eventually(t, func() bool {
				entries := logs.All()
				for _, entry := range entries {
					if seq, ok := entry.ContextMap()["sequence"].(uint64); ok && seq == messageCount {
						return true
					}
				}

				return false
			}, testTimeout, 100*time.Millisecond, "did not receive all messages or stream is already filled")

			// Verify log messages contain expected fields
			logEntries := logs.All()
			require.NotEmpty(t, logEntries, "expected some log messages")

			for _, entry := range logEntries {
				if entry.Message == "received deduplicated message" {
					fields := entry.ContextMap()
					require.Contains(t, fields, "sequence")
					require.Contains(t, fields, "subject")
				}
			}
		}
	})

	t.Run("InvalidConnection", func(t *testing.T) {
		t.Parallel()
		invalidCfg := *cfg
		invalidCfg.URL = invalidNatsURL
		_, err := nats.NewDedupJetStreamClient(&invalidCfg, jetstream.StreamConfig{ //nolint: exhaustruct
			Name:     "TEST_DEDUPE_3",
			Subjects: []string{"test.dedupe3.>"},
		})
		assert.Error(t, err)
	})
}
