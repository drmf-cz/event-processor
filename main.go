package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint: gosec
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/drmf-cz/event-processor/pkg/eventprocessor/nats"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

// Default configuration values.
const (
	DefaultPprofPort = 6060
)

func setupPprof(logger *zap.Logger) {
	pprofPort := DefaultPprofPort
	if portStr := os.Getenv("PPROF_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			pprofPort = port
		} else {
			logger.Warn("Invalid PPROF_PORT value, using default",
				zap.String("value", portStr),
				zap.Int("default", DefaultPprofPort),
			)
		}
	}

	go func() {
		addr := fmt.Sprintf(":%d", pprofPort)
		logger.Info("Starting pprof server", zap.String("addr", addr))
		if err := http.ListenAndServe(addr, nil); err != nil { //nolint:gosec
			logger.Error("pprof server failed", zap.Error(err))
		}
	}()
}

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

	// Setup pprof if enabled
	if os.Getenv("ENABLE_PPROF") == "true" {
		setupPprof(cfg.Logger)
	}

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
