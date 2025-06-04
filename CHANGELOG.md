# Changelog

## [Unreleased]

### Added
- **GitHub Workflows**: Complete CI/CD pipeline with automated testing, release management, and dependency updates
  - CI workflow for automated testing and build verification
  - Release workflow for automated version management and GitHub releases
  - Dependency update workflow with automated PR creation
- **GitHub Configuration**: CODEOWNERS file for code review assignments
- **Issue Templates**: 
  - Bug report template with structured information collection
  - Feature request template for enhancement proposals
- **Pull Request Template**: Standardized PR description format
- **Funding Configuration**: GitHub sponsors and funding information
- **Dependabot**: Automated dependency updates with security monitoring
- **Release Automation Scripts**: 
  - Archive release script for release management
  - Release template creation script for consistent releases
  - Documentation for automation scripts usage

### Changed
- **Dependencies Upgrade**: Upgraded all direct dependencies to latest versions with breaking changes compatibility
  - `go.fork.vn/config`: v0.1.0 → v0.1.3 (Added ServiceProvider with Boot, Requires, Providers methods)
  - `go.fork.vn/di`: v0.1.0 → v0.1.3 (Enhanced ServiceProvider interface with new methods)
  - `go.fork.vn/redis`: v0.1.0 → v0.1.2 (Updated ServiceProvider interface compatibility)
  - `go.fork.vn/scheduler`: v0.1.0 → v0.1.1 (Configuration-driven approach with auto-start support)

### Fixed
- **Breaking Changes Compatibility**: Updated code to work with new ServiceProvider interfaces
  - Fixed ServiceProvider implementations to include new required methods (Boot, Requires, Providers)
  - Updated DI container usage to match new interface specifications
  - Maintained backward compatibility where possible

### Dependencies
- **Core Dependencies**:
  - `github.com/go-redis/redismock/v9` v9.2.0 - Redis mocking for testing
  - `github.com/redis/go-redis/v9` v9.9.0 - Redis Go client
  - `github.com/stretchr/testify` v1.10.0 - Testing framework
  - `go.fork.vn/config` v0.1.3 - Configuration management (upgraded)
  - `go.fork.vn/di` v0.1.3 - Dependency injection (upgraded)
  - `go.fork.vn/redis` v0.1.2 - Redis provider integration (upgraded)
  - `go.fork.vn/scheduler` v0.1.1 - Task scheduling (upgraded)

- **Indirect Dependencies**:
  - `github.com/cespare/xxhash/v2` v2.3.0 - Hash functions
  - `github.com/fsnotify/fsnotify` v1.9.0 - File system notifications
  - `github.com/go-co-op/gocron` v1.37.0 - Cron job scheduling
  - `github.com/google/uuid` v1.6.0 - UUID generation
  - `github.com/spf13/viper` v1.20.1 - Configuration management
  - `golang.org/x/sys` v0.33.0 - System calls
  - `golang.org/x/text` v0.25.0 - Text processing

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
- Repository located at `github.com/Fork/queue`
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

[Unreleased]: github.com/go-fork/queue/compare/v0.1.0...HEAD
[v0.1.0]: github.com/go-fork/queue/releases/tag/v0.1.0
