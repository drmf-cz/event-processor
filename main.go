package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

func createConfig(logger *zap.Logger) *eventprocessor.Config {
	return &eventprocessor.Config{
		URL:           os.Getenv("NATS_URL"),
		Token:         "test-token",
		MaxReconnects: 5,
		ReconnectWait: time.Second * 5,
		Logger:        logger,
	}
}

func setupClients(cfg *eventprocessor.Config) (
	*eventprocessor.SimpleNatsClient,
	*eventprocessor.JetStreamClient,
	*eventprocessor.DedupJetStreamClient,
	error,
) {
	simpleClient, err := eventprocessor.NewSimpleNatsClient(*cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	jsClient, err := eventprocessor.NewJetStreamClient(*cfg, jetstream.StreamConfig{
		Name:     "TEST_JETSREAM",
		Subjects: []string{"test.jetstream1.>"},
	})
	if err != nil {
		simpleClient.Close()
		return nil, nil, nil, err
	}

	dedupeClient, err := eventprocessor.NewDedupJetStreamClient(*cfg, jetstream.StreamConfig{
		Name:       "TEST_DEDUPE",
		Subjects:   []string{"test.dedupe1.>"},
		Duplicates: time.Minute,
	})
	if err != nil {
		simpleClient.Close()
		jsClient.Close(context.Background())
		return nil, nil, nil, err
	}

	return simpleClient, jsClient, dedupeClient, nil
}

func setupSubscriptions(
	logger *zap.Logger,
	simpleClient *eventprocessor.SimpleNatsClient,
	jsClient *eventprocessor.JetStreamClient,
	dedupeClient *eventprocessor.DedupJetStreamClient,
) {
	if err := simpleClient.Subscribe("simple.events", func(data []byte) {
		logger.Info("Simple client received message", zap.String("data", string(data)))
	}); err != nil {
		logger.Error("Failed to subscribe with simple client", zap.Error(err))
	}
}

func publishMessages(
	logger *zap.Logger,
	simpleClient *eventprocessor.SimpleNatsClient,
	jsClient *eventprocessor.JetStreamClient,
	dedupeClient *eventprocessor.DedupJetStreamClient,
) {
	for {
		message := []byte(time.Now().String())

		if err := simpleClient.Publish("simple.events", message); err != nil {
			logger.Error("Failed to publish with simple client", zap.Error(err))
		}

		time.Sleep(time.Second)
	}
}

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	cfg := createConfig(logger)

	simpleClient, jsClient, dedupeClient, err := setupClients(cfg)
	if err != nil {
		logger.Fatal("Failed to setup clients", zap.Error(err))
	}
	defer simpleClient.Close()
	defer jsClient.Close(context.Background())
	defer dedupeClient.Close(context.Background())

	setupSubscriptions(logger, simpleClient, jsClient, dedupeClient)

	go publishMessages(logger, simpleClient, jsClient, dedupeClient)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Info("Shutting down...")
}
