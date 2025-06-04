# Service Provider - Tích hợp với hệ thống DI

Service Provider là thành phần chịu trách nhiệm đăng ký và khởi tạo các dịch vụ queue trong hệ thống Dependency Injection. Tài liệu này mô tả chi tiết cách sử dụng và tùy chỉnh Service Provider.

## Service Provider Interface

**Vị trí**: [`provider.go`](../provider.go)

```go
type ServiceProvider interface {
    di.ServiceProvider
    
    // Thiết lập các tác vụ lên lịch cho queue
    setupQueueScheduledTasks(schedulerManager scheduler.Manager, container di.Container)
    
    // Dọn dẹp các jobs thất bại
    cleanupFailedJobs(manager Manager, container di.Container)
    
    // Retry các jobs thất bại
    retryFailedJobs(manager Manager, container di.Container)
}
```

Provider triển khai interface `di.ServiceProvider` chuẩn với các methods:
- `Boot(app di.Application) error` - Khởi tạo các services
- `Providers() []string` - Danh sách services được cung cấp
- `Requires() []string` - Dependencies cần thiết

## Khởi tạo và sử dụng

### 1. Sử dụng cơ bản

```go
import (
    "go.fork.vn/di"
    "go.fork.vn/queue"
)

func main() {
    // Tạo DI container
    container := di.NewContainer()
    
    // Tạo application
    app := di.NewApplication(container)
    
    // Đăng ký Queue Service Provider
    queueProvider := queue.NewServiceProvider()
    app.Register(queueProvider)
    
    // Boot application
    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
    
    // Sử dụng services
    manager, _ := container.Make("queue")
    client := manager.(queue.Manager).Client()
    server := manager.(queue.Manager).Server()
}
```

### 2. Với configuration file

```go
import (
    "go.fork.vn/config"
    "go.fork.vn/di"
    "go.fork.vn/queue"
)

func main() {
    container := di.NewContainer()
    app := di.NewApplication(container)
    
    // Đăng ký Config Provider trước
    configProvider := config.NewServiceProvider()
    app.Register(configProvider)
    
    // Đăng ký Queue Provider
    queueProvider := queue.NewServiceProvider()
    app.Register(queueProvider)
    
    // Boot application - sẽ load config từ file
    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

## Services được đăng ký

Provider tự động đăng ký các services sau vào DI container:

### 1. queue.config

```go
// Tự động load từ config file hoặc sử dụng default
container.Singleton("queue.config", func() (queue.Config, error) {
    return loadQueueConfig(container)
})
```

**Sử dụng:**
```go
config, _ := container.Make("queue.config")
queueConfig := config.(queue.Config)
```

### 2. queue

```go
// Manager chính quản lý tất cả các thành phần
container.Singleton("queue", func() (queue.Manager, error) {
    config, _ := container.Make("queue.config")
    return queue.NewManagerWithContainer(config.(queue.Config), container), nil
})
```

**Sử dụng:**
```go
manager, _ := container.Make("queue")
queueManager := manager.(queue.Manager)
```

### 3. queue.client

```go
// Client để gửi tasks vào queue
container.Singleton("queue.client", func() (queue.Client, error) {
    manager, _ := container.Make("queue")
    return manager.(queue.Manager).Client(), nil
})
```

**Sử dụng:**
```go
client, _ := container.Make("queue.client")
queueClient := client.(queue.Client)

// Enqueue task
_, err := queueClient.Enqueue("send_email", emailData)
```

### 4. queue.server

```go
// Server để xử lý tasks từ queue
container.Singleton("queue.server", func() (queue.Server, error) {
    manager, _ := container.Make("queue")
    return manager.(queue.Manager).Server(), nil
})
```

**Sử dụng:**
```go
server, _ := container.Make("queue.server")
queueServer := server.(queue.Server)

// Đăng ký handlers
queueServer.RegisterHandler("send_email", emailHandler)
queueServer.Start()
```

## Dependencies và Integration

### 1. Redis Provider Integration

Queue provider tự động tích hợp với Redis provider nếu có:

```yaml
# config.yaml
queue:
  adapter:
    default: "redis"
    redis:
      provider_key: "default"  # Tham chiếu đến redis provider

redis:
  default:
    host: "localhost"
    port: 6379
    db: 0
```

**Provider setup:**
```go
import (
    "go.fork.vn/redis"
    "go.fork.vn/queue"
)

func main() {
    container := di.NewContainer()
    app := di.NewApplication(container)
    
    // Đăng ký Redis Provider trước
    redisProvider := redis.NewServiceProvider()
    app.Register(redisProvider)
    
    // Queue Provider sẽ tự động sử dụng Redis services
    queueProvider := queue.NewServiceProvider()
    app.Register(queueProvider)
    
    app.Boot()
}
```

### 2. Scheduler Provider Integration

Tích hợp với Scheduler provider cho delayed tasks:

```yaml
# config.yaml
scheduler:
  auto_start: true
  distributed_lock:
    enabled: true
    provider: "redis"
```

**Provider setup:**
```go
import (
    "go.fork.vn/scheduler"
    "go.fork.vn/queue"
)

func main() {
    container := di.NewContainer()
    app := di.NewApplication(container)
    
    // Đăng ký Scheduler Provider
    schedulerProvider := scheduler.NewServiceProvider()
    app.Register(schedulerProvider)
    
    // Queue Provider sẽ tự động setup scheduled tasks
    queueProvider := queue.NewServiceProvider()
    app.Register(queueProvider)
    
    app.Boot()
}
```

## Scheduled Tasks Setup

Provider tự động thiết lập các tác vụ lên lịch:

### 1. Cleanup Failed Jobs

```go
func (p *serviceProvider) cleanupFailedJobs(manager Manager, container di.Container) {
    // Chạy mỗi ngày lúc 2:00 AM
    // Dọn dẹp các jobs thất bại cũ hơn 7 ngày
    
    server := manager.Server()
    scheduler := server.GetScheduler()
    
    scheduler.Every(1).Day().At("02:00").Do(func() {
        log.Println("Starting failed jobs cleanup...")
        
        // Logic cleanup
        cutoffTime := time.Now().AddDate(0, 0, -7) // 7 ngày trước
        count := cleanupFailedJobsBefore(cutoffTime)
        
        log.Printf("Cleaned up %d failed jobs", count)
    })
}
```

### 2. Retry Failed Jobs

```go
func (p *serviceProvider) retryFailedJobs(manager Manager, container di.Container) {
    // Chạy mỗi 30 phút
    // Retry các jobs thất bại có thể recover
    
    server := manager.Server()
    scheduler := server.GetScheduler()
    
    scheduler.Every(30).Minutes().Do(func() {
        log.Println("Starting failed jobs retry...")
        
        // Logic retry
        count := retryRecoverableJobs()
        
        if count > 0 {
            log.Printf("Retried %d recoverable jobs", count)
        }
    })
}
```

### 3. Queue Monitoring

```go
func (p *serviceProvider) setupQueueScheduledTasks(schedulerManager scheduler.Manager, container di.Container) {
    // Monitor queue lengths mỗi 5 phút
    schedulerManager.Every(5).Minutes().Do(func() {
        manager, _ := container.Make("queue")
        queueManager := manager.(queue.Manager)
        
        // Check queue lengths
        queues := []string{"critical", "high", "default", "low"}
        for _, queueName := range queues {
            adapter := queueManager.Adapter("redis")
            size, _ := adapter.Size(context.Background(), queueName)
            
            // Alert nếu queue quá dài
            if size > 1000 {
                sendQueueAlert(queueName, size)
            }
        }
    })
}
```

## Custom Service Provider

### 1. Extending Base Provider

```go
type CustomQueueProvider struct {
    *queue.ServiceProvider
    customConfig CustomConfig
}

func NewCustomQueueProvider(config CustomConfig) *CustomQueueProvider {
    return &CustomQueueProvider{
        ServiceProvider: queue.NewServiceProvider(),
        customConfig:    config,
    }
}

func (p *CustomQueueProvider) Boot(app di.Application) error {
    container := app.Container()
    
    // Boot base provider trước
    if err := p.ServiceProvider.Boot(app); err != nil {
        return err
    }
    
    // Đăng ký custom services
    container.Singleton("queue.custom.processor", func() (*CustomProcessor, error) {
        manager, _ := container.Make("queue")
        return NewCustomProcessor(manager.(queue.Manager), p.customConfig), nil
    })
    
    return nil
}
```

### 2. Custom Services

```go
type CustomProcessor struct {
    manager queue.Manager
    config  CustomConfig
}

func (p *CustomProcessor) SetupCustomHandlers() {
    server := p.manager.Server()
    
    // Custom handlers với business logic riêng
    server.RegisterHandler("custom_task", p.handleCustomTask)
    server.RegisterHandler("priority_task", p.handlePriorityTask)
}

func (p *CustomProcessor) handleCustomTask(ctx context.Context, task *queue.Task) error {
    // Custom logic
    return nil
}
```

## Configuration Loading

### 1. Default Configuration

```go
func (p *serviceProvider) loadDefaultConfig() queue.Config {
    return queue.Config{
        Adapter: queue.AdapterConfig{
            Default: "memory",
            Memory: queue.MemoryConfig{
                Prefix: "queue:",
            },
            Redis: queue.RedisConfig{
                Prefix:      "queue:",
                ProviderKey: "redis",
            },
        },
        Server: queue.ServerConfig{
            Concurrency:     10,
            PollingInterval: 1000,
            DefaultQueue:    "default",
            StrictPriority:  true,
            Queues:          []string{"critical", "high", "default", "low"},
            ShutdownTimeout: 30,
            LogLevel:        1,
            RetryLimit:      3,
        },
        Client: queue.ClientConfig{
            DefaultOptions: queue.ClientDefaultOptions{
                Queue:    "default",
                MaxRetry: 3,
                Timeout:  30,
            },
        },
    }
}
```

### 2. Environment-based Configuration

```go
func (p *serviceProvider) loadConfig(container di.Container) (queue.Config, error) {
    // Thử load từ config provider trước
    if configService, err := container.Make("config"); err == nil {
        if configManager, ok := configService.(config.Manager); ok {
            var queueConfig queue.Config
            if err := configManager.UnmarshalKey("queue", &queueConfig); err == nil {
                return queueConfig, nil
            }
        }
    }
    
    // Fallback: load từ environment variables
    config := p.loadDefaultConfig()
    
    if adapter := os.Getenv("QUEUE_ADAPTER"); adapter != "" {
        config.Adapter.Default = adapter
    }
    
    if concurrency := os.Getenv("QUEUE_CONCURRENCY"); concurrency != "" {
        if c, err := strconv.Atoi(concurrency); err == nil {
            config.Server.Concurrency = c
        }
    }
    
    return config, nil
}
```

## Service Resolution

### 1. Type-based Resolution

```go
// Trong service khác
type EmailService struct {
    client queue.Client
}

func NewEmailService(container di.Container) *EmailService {
    client, _ := container.Make("queue.client")
    return &EmailService{
        client: client.(queue.Client),
    }
}
```

### 2. Interface-based Resolution

```go
// Đăng ký interface binding
container.Bind("queue.ClientInterface", func() (queue.Client, error) {
    return container.Make("queue.client")
})

// Sử dụng
type UserService struct {
    queueClient queue.Client
}

func NewUserService(container di.Container) *UserService {
    client, _ := container.Make("queue.ClientInterface")
    return &UserService{
        queueClient: client.(queue.Client),
    }
}
```

## Testing với Service Provider

### 1. Test Setup

```go
func setupTestContainer() di.Container {
    container := di.NewContainer()
    app := di.NewApplication(container)
    
    // Override config cho testing
    container.Singleton("queue.config", func() (queue.Config, error) {
        config := queue.DefaultConfig()
        config.Adapter.Default = "memory"  // Sử dụng memory cho test
        config.Server.Concurrency = 1
        return config, nil
    })
    
    // Đăng ký provider
    queueProvider := queue.NewServiceProvider()
    app.Register(queueProvider)
    app.Boot()
    
    return container
}

func TestQueueIntegration(t *testing.T) {
    container := setupTestContainer()
    
    client, _ := container.Make("queue.client")
    server, _ := container.Make("queue.server")
    
    queueClient := client.(queue.Client)
    queueServer := server.(queue.Server)
    
    // Test workflow
    queueServer.RegisterHandler("test_task", func(ctx context.Context, task *queue.Task) error {
        return nil
    })
    
    _, err := queueClient.Enqueue("test_task", map[string]string{"key": "value"})
    assert.NoError(t, err)
}
```

### 2. Mock Services

```go
func setupMockContainer() di.Container {
    container := di.NewContainer()
    
    // Mock queue client
    container.Singleton("queue.client", func() (queue.Client, error) {
        return &MockClient{}, nil
    })
    
    return container
}

type MockClient struct{}

func (m *MockClient) Enqueue(taskName string, payload interface{}, opts ...queue.Option) (*queue.TaskInfo, error) {
    return &queue.TaskInfo{ID: "mock-id"}, nil
}
```

## Best Practices

### 1. Provider Organization

```go
// Tách riêng providers cho từng domain
type QueueServiceProvider struct {
    // Core queue functionality
}

type QueueSchedulerProvider struct {
    // Scheduled tasks và cron jobs
}

type QueueMonitoringProvider struct {
    // Metrics và monitoring
}
```

### 2. Configuration Management

```go
// Sử dụng struct tags cho validation
type QueueConfig struct {
    Adapter AdapterConfig `mapstructure:"adapter" validate:"required"`
    Server  ServerConfig  `mapstructure:"server" validate:"required"`
}

func (p *serviceProvider) validateConfig(config QueueConfig) error {
    validator := validator.New()
    return validator.Struct(config)
}
```

### 3. Error Handling

```go
func (p *serviceProvider) Boot(app di.Application) error {
    container := app.Container()
    
    // Graceful degradation
    container.Singleton("queue", func() (queue.Manager, error) {
        config, err := p.loadConfig(container)
        if err != nil {
            log.Printf("Failed to load queue config, using defaults: %v", err)
            config = queue.DefaultConfig()
        }
        
        return queue.NewManagerWithContainer(config, container), nil
    })
    
    return nil
}
```
