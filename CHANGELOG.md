# Changelog

## [Unreleased]

## v0.1.0 - 2025-05-31

### Added
- **Queue Management System**: Comprehensive queue management and task processing system for Go applications
- **Multiple Queue Adapters**: Support for Memory, Redis, and Redis Provider-based queues
- **Redis Integration**: Advanced Redis features with priority queues, TTL support, and batch operations
- **Priority Queues**: Redis Sorted Sets implementation for task prioritization
- **Batch Operations**: Pipeline support for high-throughput scenarios
- **TTL Support**: Temporary tasks with automatic expiration
- **Queue Monitoring**: Real-time queue statistics and health monitoring
- **DI Integration**: Seamless integration with Dependency Injection container
- **Configuration-Driven**: Flexible configuration through config management
- **Thread-Safe**: Safe concurrent queue operations and task processing
- **Resource Management**: Automatic cleanup and proper resource disposal
- **Error Resilience**: Robust error handling and recovery mechanisms
- **Health Checks**: Built-in health monitoring and connectivity checks
- **Development Tools**: Utilities for queue flushing and debugging
- **Memory Optimization**: Efficient memory usage for high-volume scenarios
- **Production Ready**: Advanced configuration templates and best practices

### Queue Features
- **Task Enqueuing**: Standard task queuing with JSON serialization
- **Priority Support**: EnqueueWithPriority for task prioritization
- **TTL Management**: EnqueueWithTTL for temporary tasks
- **Batch Processing**: EnqueueWithPipeline and MultiDequeue operations
- **Queue Information**: GetQueueInfo for monitoring and statistics
- **Development Utilities**: FlushQueues for testing and development

### Technical Details
- Initial release as standalone module `go.fork.vn/queue`
- Repository located at `github.com/go-fork/queue`
- Built with Go 1.23.9
- Full test coverage with comprehensive documentation
- Redis Provider integration for centralized Redis management
- Clean separation between queue logic and Redis connection
- QueueRedisAdapter interface for Redis-specific functionality
- Complete migration guides and production configuration samples

### Dependencies
- `go.fork.vn/di`: Dependency injection integration
- `go.fork.vn/config`: Configuration management
- `go.fork.vn/redis`: Redis connectivity and provider integration
- `go.fork.vn/scheduler`: Scheduler integration for queue processing

# Changelog

## [Unreleased]

## v0.1.0 - 2025-05-31

### Added
- **Queue Management System**: Comprehensive queue management and task processing system for Go applications
- **Multiple Queue Adapters**: Support for Memory, Redis, and Redis Provider-based queues
- **Redis Integration**: Advanced Redis features with priority queues, TTL support, and batch operations
- **Priority Queues**: Redis Sorted Sets implementation for task prioritization
- **Batch Operations**: Pipeline support for high-throughput scenarios
- **TTL Support**: Temporary tasks with automatic expiration
- **Queue Monitoring**: Real-time queue statistics and health monitoring
- **DI Integration**: Seamless integration with Dependency Injection container
- **Configuration-Driven**: Flexible configuration through config management
- **Thread-Safe**: Safe concurrent queue operations and task processing
- **Resource Management**: Automatic cleanup and proper resource disposal
- **Error Resilience**: Robust error handling and recovery mechanisms
- **Health Checks**: Built-in health monitoring and connectivity checks
- **Development Tools**: Utilities for queue flushing and debugging
- **Memory Optimization**: Efficient memory usage for high-volume scenarios
- **Production Ready**: Advanced configuration templates and best practices
- **Worker System**: Complete worker implementation with scheduler integration
- **Delayed Tasks**: Support for delayed and scheduled task execution
- **Batch Processing**: Enhanced batch processing capabilities for high-throughput scenarios
- **Message Filtering**: Advanced message filtering capabilities
- **Dead Letter Queue**: Support for failed message handling
- **Task Status Tracking**: Comprehensive task state management in distributed environments
- **Retry Logic**: Configurable retry strategies with backoff mechanisms
- **Asynchronous Processing**: Efficient asynchronous message processing
- **Task Serialization**: Robust payload serialization and deserialization
- **Server Component**: Dedicated server component for queue task processing

### Queue Features
- **Task Enqueuing**: Standard task queuing with JSON serialization
- **Priority Support**: EnqueueWithPriority for task prioritization
- **TTL Management**: EnqueueWithTTL for temporary tasks
- **Batch Processing**: EnqueueWithPipeline and MultiDequeue operations
- **Queue Information**: GetQueueInfo for monitoring and statistics
- **Development Utilities**: FlushQueues for testing and development
- **Immediate Execution**: Support for immediate task execution
- **Scheduled Tasks**: Time-based and interval-based task scheduling
- **Time-specific Scheduling**: Execute tasks at specific times

### Technical Details
- Initial release as standalone module `go.fork.vn/queue`
- Repository located at `github.com/go-fork/queue`
- Built with Go 1.23.9
- Full test coverage with comprehensive documentation
- Redis Provider integration for centralized Redis management
- Clean separation between queue logic and Redis connection
- QueueRedisAdapter interface for Redis-specific functionality
- Complete migration guides and production configuration samples
- Memory adapter for development environments
- Redis adapter optimized for production environments
- Client API with simple task enqueueing interface
- Worker model with configurable processing strategies

### Dependencies
- `go.fork.vn/di`: Dependency injection integration
- `go.fork.vn/config`: Configuration management
- `go.fork.vn/redis`: Redis connectivity and provider integration
- `go.fork.vn/scheduler`: Scheduler integration for queue processing

[Unreleased]: https://github.com/go-fork/queue/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/go-fork/queue/releases/tag/v0.1.0
