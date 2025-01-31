# Event Processor

A Go-based event processing system demonstrating various NATS streaming patterns and configurations. This repository serves as both a learning resource and a practical implementation of real-time event processing systems.

## Overview

This repository contains:
- Multiple NATS client implementations
- JetStream integration
- Message deduplication
- Durable subscriptions
- Comprehensive test suite
- Docker-based development environment

## Client Implementations

1. **Simple NATS Client**
   - Basic publish/subscribe functionality
   - Thread-safe operations
   - Connection management

2. **JetStream Client**
   - Persistent message storage
   - Stream management
   - Enhanced delivery guarantees

3. **Deduplication Client**
   - Message ID-based deduplication
   - Configurable deduplication window
   - Built on JetStream capabilities

4. **Streaming Client**
   - Durable subscriptions
   - Message persistence
   - At-least-once delivery

## Development

### Prerequisites
- Go 1.22+
- Docker and Docker Compose
- NATS Server 2.10+

### Quick Start
```bash
# Start the infrastructure
docker-compose up -d

# Run the application
go run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run specific client tests
go test ./pkg/eventprocessor -run TestSimpleNatsClient
```

## Educational Materials

This repository is part of a larger educational initiative:

- [Workshop: Building a Real-Time Event Processing System](workshop_ideas.md)
  - Hands-on implementation
  - Best practices
  - Performance considerations
  - Error handling patterns

- [Lecture: Inside Go Runtime and Event Processing](lecture_ideas.md)
  - Garbage collection impact
  - Goroutine scheduling
  - Memory optimization
  - Performance profiling

## Configuration

The system supports various configuration options through:
- NATS configuration (`config/nats.conf`)
- Kafka configuration (`config/kafka.conf`)
- Environment variables
- Docker Compose settings

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details. 