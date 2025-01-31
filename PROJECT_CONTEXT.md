# Project Context

## Architecture Overview

### Core Components

1. **Event Processing Interface**
   - Common interface for all event processors
   - Unified publish/subscribe methods
   - Standardized error handling
   - Thread-safe operations

2. **Client Implementations**
   ```
   pkg/eventprocessor/
   ├── interface.go     # Core interfaces and types
   ├── simple.go        # Basic NATS implementation
   ├── jetstream.go     # JetStream functionality
   ├── dedupe.go        # Deduplication logic
   └── streaming.go     # NATS Streaming (STAN)
   ```

### Design Decisions

1. **Interface-First Design**
   - Common `EventProcessor` interface enables easy swapping of implementations
   - Consistent error handling across all clients
   - Simplified testing through interface mocking

2. **Thread Safety**
   - All clients implement mutex-based synchronization
   - Read/Write locks for optimal performance
   - Safe concurrent access to shared resources

3. **Configuration Management**
   - Centralized configuration structure
   - Environment variable support
   - Flexible client options

4. **Error Handling Strategy**
   - Error wrapping with context
   - Graceful connection management
   - Automatic reconnection handling

## Implementation Details

### Message Deduplication
- Uses message IDs based on timestamps
- Configurable deduplication window
- JetStream-based implementation for persistence

### Streaming Patterns
1. **Simple Pub/Sub**
   - In-memory message delivery
   - Best for high-throughput, low-latency
   - No persistence guarantees

2. **JetStream**
   - Persistent message storage
   - Stream replay capabilities
   - Enhanced delivery guarantees

3. **NATS Streaming**
   - Durable subscriptions
   - At-least-once delivery
   - Message history and replay

### Testing Strategy
- Table-driven tests
- Integration tests with real NATS server
- Comprehensive error case coverage
- Performance benchmarks

## Infrastructure

### Development Environment
```yaml
Services:
- NATS Server (core messaging)
- NATS Streaming (durable messaging)
- Kafka (alternative messaging)
- Go Application (event processor)
```

### Monitoring and Observability
- NATS monitoring endpoints
- pprof profiling support
- Metrics collection points
- Health check endpoints

## Future Considerations

1. **Scalability**
   - Horizontal scaling of consumers
   - Message partitioning
   - Load balancing strategies

2. **Reliability**
   - Circuit breaker implementation
   - Dead letter queues
   - Message retry policies

3. **Performance**
   - Message batching
   - Connection pooling
   - Optimized serialization

4. **Monitoring**
   - Prometheus metrics
   - Tracing integration
   - Enhanced logging

## Development Guidelines

### Code Structure
- Maximum 250 lines per file
- Clear separation of concerns
- Comprehensive documentation
- Consistent error handling

### Testing Requirements
- 80%+ test coverage
- Integration tests
- Performance benchmarks
- Error case validation

### Performance Targets
- Sub-millisecond publish latency
- High throughput capability
- Minimal memory allocation
- Efficient connection usage 