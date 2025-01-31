package eventprocessor

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

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

func TestDeDupeJetStreamClient(t *testing.T) {
	t.Parallel()

	// Create test logger
	logger := zaptest.NewLogger(t)
	defer logger.Sync()

	cfg := Config{
		URL:           "nats://nats:4222",
		Token:         "test-token",
		MaxReconnects: 5,
		ReconnectWait: time.Second,
		Logger:        logger,
	}

	t.Run("NewDeDupeJetStreamClient", func(t *testing.T) {
		t.Parallel()
		client, err := NewDedupJetStreamClient(cfg, jetstream.StreamConfig{
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
		testCfg := cfg
		testCfg.Logger = logger

		client, err := NewDedupJetStreamClient(testCfg, jetstream.StreamConfig{
			Name:       "TEST_DEDUPE_2",
			Subjects:   []string{"test.dedupe2.>"},
			Duplicates: time.Minute,
		})
		require.NoError(t, err)
		defer client.Close(context.Background())

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Publish 100 messages to the stream
		go func() {
			for range 2 {
				for i := 0; i < 100; i++ {
					err = client.PublishToStream(ctx, "test.dedupe2."+strconv.Itoa(i), []byte("test-"+strconv.Itoa(i)))
					require.NoError(t, err)
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
		default:
			// Check logs until we find the 100th message
			require.Eventually(t, func() bool {
				entries := logs.All()
				count := 0
				for _, entry := range entries {
					if count == 100 && len(entries) > 200 {
						break
					}
					if sub, ok := entry.ContextMap()["subject"].(string); ok {
						msg := strings.SplitAfterN(sub, ".", 3)[2]
						seq, err := strconv.ParseInt(msg, 10, 64)
						if err != nil {
							t.Fatal(err)
						}

						if seq > int64(count) {
							count = int(seq) + 1
						}
					}
				}

				return true
			}, 5*time.Second, 100*time.Millisecond, "did not receive all 100 messages or stream deduplication is not working")

			// Verify log messages contain expected fields
			logEntries := logs.All()
			require.True(t, len(logEntries) > 0, "expected some log messages")

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
		invalidCfg := cfg
		invalidCfg.URL = invalidNatsURL
		_, err := NewDedupJetStreamClient(invalidCfg, jetstream.StreamConfig{
			Name:     "TEST_DEDUPE_3",
			Subjects: []string{"test.dedupe3.>"},
		})
		assert.Error(t, err)
	})
}
