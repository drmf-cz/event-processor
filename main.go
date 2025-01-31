package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Matovidlo/event-processor/pkg/eventprocessor"
)

func main() {
	cfg := eventprocessor.Config{
		URL:           os.Getenv("NATS_URL"),
		ClusterID:     "event-processor-cluster",
		ClientID:      "example-client",
		Token:         "development-token-123",
		DeDupeWindow:  time.Minute,
		MaxReconnects: 5,
		ReconnectWait: time.Second * 5,
	}

	// Example with Simple NATS
	simpleClient, err := eventprocessor.NewSimpleNatsClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create simple NATS client: %v", err)
	}
	defer simpleClient.Close()

	// Example with JetStream
	jsClient, err := eventprocessor.NewJetStreamClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create JetStream client: %v", err)
	}
	defer jsClient.Close()

	// Example with Deduplication
	dedupeClient, err := eventprocessor.NewDeDupeJetStreamClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create DeDupe client: %v", err)
	}
	defer dedupeClient.Close()

	// Example with Streaming
	streamingClient, err := eventprocessor.NewStreamingClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Streaming client: %v", err)
	}
	defer streamingClient.Close()

	// Subscribe to messages
	err = simpleClient.Subscribe("simple.events", func(data []byte) {
		log.Printf("Simple client received: %s", string(data))
	})
	if err != nil {
		log.Printf("Failed to subscribe with simple client: %v", err)
	}

	err = jsClient.Subscribe("js.events", func(data []byte) {
		log.Printf("JetStream client received: %s", string(data))
	})
	if err != nil {
		log.Printf("Failed to subscribe with JetStream client: %v", err)
	}

	err = dedupeClient.Subscribe("dedupe.events", func(data []byte) {
		log.Printf("DeDupe client received: %s", string(data))
	})
	if err != nil {
		log.Printf("Failed to subscribe with DeDupe client: %v", err)
	}

	err = streamingClient.Subscribe("streaming.events", func(data []byte) {
		log.Printf("Streaming client received: %s", string(data))
	})
	if err != nil {
		log.Printf("Failed to subscribe with Streaming client: %v", err)
	}

	// Publish example messages
	go func() {
		for i := 0; ; i++ {
			message := []byte(time.Now().String())

			if err := simpleClient.Publish("simple.events", message); err != nil {
				log.Printf("Failed to publish with simple client: %v", err)
			}

			if err := jsClient.Publish("js.events", message); err != nil {
				log.Printf("Failed to publish with JetStream client: %v", err)
			}

			if err := dedupeClient.Publish("dedupe.events", message); err != nil {
				log.Printf("Failed to publish with DeDupe client: %v", err)
			}

			if err := streamingClient.Publish("streaming.events", message); err != nil {
				log.Printf("Failed to publish with Streaming client: %v", err)
			}

			time.Sleep(time.Second)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	log.Println("Shutting down...")
}
