package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

func setupClients(cfg *nats.Config) (
	*nats.SimpleNatsClient,
	*nats.JetStreamClient,
	*nats.DedupJetStreamClient,
	error,
) {
	simpleClient, err := nats.NewSimpleNatsClient(cfg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create simple client: %w", err)
	}

	jsClient, err := nats.NewJetStreamClient(cfg, jetstream.StreamConfig{ //nolint: exhaustruct
		Name:     "TEST_JETSREAM",
		Subjects: []string{"test.jetstream1.>"},
	})
	if err != nil {
		simpleClient.Close(context.Background())

		return nil, nil, nil, fmt.Errorf("failed to create jetstream client: %w", err)
	}

	dedupeClient, err := nats.NewDedupJetStreamClient(cfg, jetstream.StreamConfig{ //nolint: exhaustruct
		Name:       "TEST_DEDUPE",
		Subjects:   []string{"test.dedupe1.>"},
		Duplicates: time.Minute,
	})
	if err != nil {
		simpleClient.Close(context.Background())
		jsClient.Close(context.Background())

		return nil, nil, nil, fmt.Errorf("failed to create dedupe client: %w", err)
	}

	return simpleClient, jsClient, dedupeClient, nil
}

func setupSubscriptions(
	logger *zap.Logger,
	simpleClient *nats.SimpleNatsClient,
) {
	if err := simpleClient.Subscribe("simple.events", func(data []byte) {
		logger.Info("Simple client received message", zap.String("data", string(data)))
	}); err != nil {
		logger.Error("Failed to subscribe with simple client", zap.Error(err))
	}
}

func publishMessages(
	logger *zap.Logger,
	simpleClient *nats.SimpleNatsClient,
) {
	for {
		message := []byte(time.Now().String())

		if err := simpleClient.PublishToStream(context.Background(), "simple.events", message); err != nil {
			logger.Error("Failed to publish with simple client", zap.Error(err))
		}

		time.Sleep(time.Second)
	}
}

func main() {
	cfg := nats.NewConfig()
	defer func() {
		if err := cfg.Logger.Sync(); err != nil {
			panic("failed to sync logger: " + err.Error())
		}
	}()

	simpleClient, jsClient, dedupeClient, err := setupClients(cfg)
	if err != nil {
		cfg.Logger.Fatal("Failed to setup clients", zap.Error(err))
	}
	defer func() {
		simpleClient.Close(context.Background())
		jsClient.Close(context.Background())
		dedupeClient.Close(context.Background())
	}()

	setupSubscriptions(cfg.Logger, simpleClient)

	go publishMessages(cfg.Logger, simpleClient)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	cfg.Logger.Info("Shutting down...")
}
