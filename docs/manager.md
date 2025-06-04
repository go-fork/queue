# Queue Manager - Trung tâm quản lý hệ thống

Queue Manager là thành phần trung tâm quản lý tất cả các thành phần khác trong hệ thống queue. Manager cung cấp factory pattern để tạo và quản lý các services như Client, Server, Adapters và Scheduler.

## Manager Interface

**Vị trí**: [`manager.go`](../manager.go)

```go
type Manager interface {
    // RedisClient trả về Redis client
    RedisClient() redisClient.UniversalClient
    
    // MemoryAdapter trả về memory queue adapter
    MemoryAdapter() adapter.QueueAdapter
    
    // RedisAdapter trả về redis queue adapter  
    RedisAdapter() adapter.QueueAdapter
    
    // Adapter trả về queue adapter dựa trên cấu hình
    Adapter(name string) adapter.QueueAdapter
    
    // Client trả về Client
    Client() Client
    
    // Server trả về Server
    Server() Server
    
    // Scheduler trả về Scheduler manager để lên lịch tasks
    Scheduler() scheduler.Manager
    
    // SetScheduler thiết lập scheduler manager từ bên ngoài
    SetScheduler(scheduler scheduler.Manager)
}
```

## Khởi tạo Manager

### 1. Manager với cấu hình mặc định

```go
import "go.fork.vn/queue"

func main() {
    // Sử dụng cấu hình mặc định
    config := queue.DefaultConfig()
    
    // Tạo manager
    manager := queue.NewManager(config)
    
    // Sử dụng các services
    client := manager.Client()
    server := manager.Server()
}
```

### 2. Manager với DI Container

```go
import (
    "go.fork.vn/di"
    "go.fork.vn/queue"
)

func main() {
    // Tạo DI container với Redis provider
    container := di.NewContainer()
    
    // Đăng ký Redis services...
    setupRedisProvider(container)
    
    // Tạo manager với container
    config := queue.DefaultConfig()
    config.Adapter.Default = "redis"
    
    manager := queue.NewManagerWithContainer(config, container)
    
    // Manager sẽ tự động sử dụng Redis từ container
    client := manager.Client()
    server := manager.Server()
}
```

### 3. Manager với custom configuration

```go
func createProductionManager() queue.Manager {
    config := queue.Config{
        Adapter: queue.AdapterConfig{
            Default: "redis",
            Redis: queue.RedisConfig{
                Prefix:      "prod_queue:",
                ProviderKey: "primary_redis",
            },
        },
        Server: queue.ServerConfig{
            Concurrency:     50,
            PollingInterval: 500,
            DefaultQueue:    "default",
            StrictPriority:  true,
            Queues:          []string{"urgent", "high", "normal", "low", "batch"},
            ShutdownTimeout: 120,
            LogLevel:        1, // Info
            RetryLimit:      5,
        },
        Client: queue.ClientConfig{
            DefaultOptions: queue.ClientDefaultOptions{
                Queue:    "normal",
                MaxRetry: 3,
                Timeout:  60,
            },
        },
    }
    
    return queue.NewManager(config)
}
```

## Adapter Management

### 1. Memory Adapter

```go
func main() {
    manager := queue.NewManager(queue.DefaultConfig())
    
    // Lấy memory adapter
    memoryAdapter := manager.MemoryAdapter()
    
    // Sử dụng trực tiếp
    ctx := context.Background()
    task := &queue.Task{
        ID:   "task-1",
        Name: "test_task",
        Payload: []byte(`{"message": "hello"}`),
    }
    
    err := memoryAdapter.Enqueue(ctx, "test_queue", task)
    if err != nil {
        log.Fatal(err)
    }
    
    // Dequeue task
    dequeuedTask, err := memoryAdapter.Dequeue(ctx, "test_queue")
    if err != nil {
        log.Fatal(err)
    }
}
```

**Đặc điểm Memory Adapter:**
- ✅ **Tốc độ cao**: Truy cập trực tiếp bộ nhớ
- ✅ **Không dependencies**: Không cần external services
- ✅ **Thread-safe**: Sử dụng mutex cho concurrency
- ❌ **Không persistent**: Mất dữ liệu khi restart
- ❌ **Single process**: Không chia sẻ giữa multiple instances

### 2. Redis Adapter

```go
func main() {
    // Cấu hình Redis adapter
    config := queue.DefaultConfig()
    config.Adapter.Default = "redis"
    
    // Với DI container có Redis provider
    container := setupRedisContainer()
    manager := queue.NewManagerWithContainer(config, container)
    
    // Lấy Redis adapter
    redisAdapter := manager.RedisAdapter()
    
    // Redis adapter tự động sử dụng connection từ provider
    ctx := context.Background()
    task := &queue.Task{
        ID:   "redis-task-1",
        Name: "persistent_task",
        Payload: []byte(`{"data": "persistent"}`),
    }
    
    err := redisAdapter.Enqueue(ctx, "persistent_queue", task)
    if err != nil {
        log.Fatal(err)
    }
}

func setupRedisContainer() di.Container {
    container := di.NewContainer()
    
    // Đăng ký Redis manager
    container.Singleton("redis", func() (*redis.Manager, error) {
        return redis.NewManager(redisConfig), nil
    })
    
    return container
}
```

**Đặc điểm Redis Adapter:**
- ✅ **Persistent**: Dữ liệu được lưu trữ bền vững
- ✅ **Distributed**: Multiple instances có thể share queue
- ✅ **Scalable**: Horizontal scaling support
- ✅ **Rich features**: Sorted sets cho delayed tasks
- ❌ **Network dependency**: Cần Redis server
- ❌ **Complexity**: Cần quản lý Redis infrastructure

### 3. Dynamic Adapter Selection

```go
func main() {
    manager := queue.NewManager(queue.DefaultConfig())
    
    // Chọn adapter theo environment
    var adapter adapter.QueueAdapter
    
    if os.Getenv("ENV") == "production" {
        adapter = manager.Adapter("redis")
    } else {
        adapter = manager.Adapter("memory")
    }
    
    // Hoặc chọn theo tên
    adapter = manager.Adapter("redis")  // Redis adapter
    adapter = manager.Adapter("memory") // Memory adapter
    adapter = manager.Adapter("")       // Default adapter từ config
}
```

## Service Management

### 1. Client Management

```go
func main() {
    manager := queue.NewManager(config)
    
    // Client tự động sử dụng adapter mặc định
    client := manager.Client()
    
    // Enqueue tasks
    client.Enqueue("send_email", emailData)
    client.Enqueue("process_file", fileData, queue.WithQueue("files"))
    
    // Client sẽ tự động cleanup khi manager bị garbage collected
}
```

### 2. Server Management

```go
func main() {
    manager := queue.NewManager(config)
    
    // Server với cấu hình từ config
    server := manager.Server()
    
    // Đăng ký handlers
    server.RegisterHandler("send_email", emailHandler)
    server.RegisterHandler("process_file", fileHandler)
    
    // Khởi động server
    server.Start()
}
```

### 3. Scheduler Integration

```go
import "go.fork.vn/scheduler"

func main() {
    manager := queue.NewManager(config)
    
    // Lấy hoặc tạo scheduler
    scheduler := manager.Scheduler()
    
    // Thiết lập scheduled tasks
    scheduler.Every(5).Minutes().Do(func() {
        log.Println("Periodic maintenance task")
    })
    
    // Hoặc set custom scheduler
    customScheduler := scheduler.NewCustomScheduler()
    manager.SetScheduler(customScheduler)
}
```

## Redis Client Management

### 1. Automatic Redis Integration

```go
func main() {
    // Manager tự động lấy Redis client từ DI container
    container := setupRedisContainer()
    
    config := queue.DefaultConfig()
    config.Adapter.Redis.ProviderKey = "primary_redis"
    
    manager := queue.NewManagerWithContainer(config, container)
    
    // Redis client tự động được cấu hình
    redisClient := manager.RedisClient()
    
    // Có thể sử dụng trực tiếp Redis client
    ctx := context.Background()
    result := redisClient.Ping(ctx)
    log.Printf("Redis ping: %v", result.Val())
}
```

### 2. Fallback Redis Client

```go
// Trong manager.go implementation
func (m *manager) RedisClient() redisClient.UniversalClient {
    if m.redisClient == nil {
        // Thử lấy từ DI container trước
        if m.container != nil {
            providerKey := m.config.Adapter.Redis.ProviderKey
            if providerKey == "" {
                providerKey = "redis"
            }
            
            if redisService, err := m.container.Make(providerKey); err == nil {
                if redisManager, ok := redisService.(redis.Manager); ok {
                    if universalClient, err := redisManager.UniversalClient(); err == nil {
                        m.redisClient = *universalClient
                    }
                }
            }
        }
        
        // Fallback: tạo client mặc định
        if m.redisClient == nil {
            m.redisClient = redisClient.NewClient(&redisClient.Options{
                Addr: "localhost:6379",
            })
        }
    }
    return m.redisClient
}
```

## Configuration Management

### 1. Environment-based Configuration

```go
func createManagerFromEnv() queue.Manager {
    config := queue.DefaultConfig()
    
    // Override từ environment variables
    if adapter := os.Getenv("QUEUE_ADAPTER"); adapter != "" {
        config.Adapter.Default = adapter
    }
    
    if concurrency := os.Getenv("QUEUE_CONCURRENCY"); concurrency != "" {
        if c, err := strconv.Atoi(concurrency); err == nil {
            config.Server.Concurrency = c
        }
    }
    
    if prefix := os.Getenv("QUEUE_REDIS_PREFIX"); prefix != "" {
        config.Adapter.Redis.Prefix = prefix
    }
    
    return queue.NewManager(config)
}
```

### 2. Multi-environment Configuration

```go
type EnvironmentConfig struct {
    Development queue.Config
    Staging     queue.Config
    Production  queue.Config
}

func createManagerForEnv(env string) queue.Manager {
    configs := EnvironmentConfig{
        Development: queue.Config{
            Adapter: queue.AdapterConfig{Default: "memory"},
            Server: queue.ServerConfig{
                Concurrency: 2,
                LogLevel:   0, // Debug
            },
        },
        Staging: queue.Config{
            Adapter: queue.AdapterConfig{Default: "redis"},
            Server: queue.ServerConfig{
                Concurrency: 10,
                LogLevel:   1, // Info
            },
        },
        Production: queue.Config{
            Adapter: queue.AdapterConfig{Default: "redis"},
            Server: queue.ServerConfig{
                Concurrency: 50,
                LogLevel:   2, // Warning
            },
        },
    }
    
    var config queue.Config
    switch env {
    case "development":
        config = configs.Development
    case "staging":
        config = configs.Staging
    case "production":
        config = configs.Production
    default:
        config = queue.DefaultConfig()
    }
    
    return queue.NewManager(config)
}
```

## Advanced Usage Patterns

### 1. Multi-Queue Management

```go
type MultiQueueManager struct {
    managers map[string]queue.Manager
}

func NewMultiQueueManager() *MultiQueueManager {
    return &MultiQueueManager{
        managers: make(map[string]queue.Manager),
    }
}

func (m *MultiQueueManager) AddQueue(name string, config queue.Config) {
    m.managers[name] = queue.NewManager(config)
}

func (m *MultiQueueManager) GetClient(queueName string) queue.Client {
    if manager, exists := m.managers[queueName]; exists {
        return manager.Client()
    }
    return nil
}

func (m *MultiQueueManager) GetServer(queueName string) queue.Server {
    if manager, exists := m.managers[queueName]; exists {
        return manager.Server()
    }
    return nil
}

// Sử dụng
func main() {
    multiManager := NewMultiQueueManager()
    
    // Email queue với Redis
    emailConfig := queue.DefaultConfig()
    emailConfig.Adapter.Default = "redis"
    emailConfig.Adapter.Redis.Prefix = "email_queue:"
    multiManager.AddQueue("email", emailConfig)
    
    // File processing queue với memory (fast)
    fileConfig := queue.DefaultConfig()
    fileConfig.Adapter.Default = "memory"
    multiManager.AddQueue("files", fileConfig)
    
    // Sử dụng
    emailClient := multiManager.GetClient("email")
    fileClient := multiManager.GetClient("files")
}
```

### 2. Manager với Monitoring

```go
type MonitoredManager struct {
    queue.Manager
    metrics *Metrics
}

type Metrics struct {
    EnqueueCount    int64
    DequeueCount    int64
    ProcessingTime  time.Duration
    ErrorCount      int64
}

func NewMonitoredManager(config queue.Config) *MonitoredManager {
    return &MonitoredManager{
        Manager: queue.NewManager(config),
        metrics: &Metrics{},
    }
}

func (m *MonitoredManager) Client() queue.Client {
    baseClient := m.Manager.Client()
    return &MonitoredClient{
        Client:  baseClient,
        metrics: m.metrics,
    }
}

type MonitoredClient struct {
    queue.Client
    metrics *Metrics
}

func (c *MonitoredClient) Enqueue(taskName string, payload interface{}, opts ...queue.Option) (*queue.TaskInfo, error) {
    start := time.Now()
    taskInfo, err := c.Client.Enqueue(taskName, payload, opts...)
    
    atomic.AddInt64(&c.metrics.EnqueueCount, 1)
    if err != nil {
        atomic.AddInt64(&c.metrics.ErrorCount, 1)
    }
    
    duration := time.Since(start)
    atomic.AddInt64((*int64)(&c.metrics.ProcessingTime), int64(duration))
    
    return taskInfo, err
}

func (m *MonitoredManager) GetMetrics() Metrics {
    return Metrics{
        EnqueueCount:   atomic.LoadInt64(&m.metrics.EnqueueCount),
        DequeueCount:   atomic.LoadInt64(&m.metrics.DequeueCount),
        ProcessingTime: time.Duration(atomic.LoadInt64((*int64)(&m.metrics.ProcessingTime))),
        ErrorCount:     atomic.LoadInt64(&m.metrics.ErrorCount),
    }
}
```

### 3. Manager Pool

```go
type ManagerPool struct {
    managers []queue.Manager
    current  int64
    mu       sync.RWMutex
}

func NewManagerPool(configs []queue.Config) *ManagerPool {
    managers := make([]queue.Manager, len(configs))
    for i, config := range configs {
        managers[i] = queue.NewManager(config)
    }
    
    return &ManagerPool{
        managers: managers,
    }
}

func (p *ManagerPool) GetManager() queue.Manager {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    // Round-robin selection
    index := atomic.AddInt64(&p.current, 1) % int64(len(p.managers))
    return p.managers[index]
}

func (p *ManagerPool) GetClient() queue.Client {
    return p.GetManager().Client()
}

func (p *ManagerPool) BroadcastToAllServers(handler func(queue.Server)) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    for _, manager := range p.managers {
        handler(manager.Server())
    }
}
```

## Testing với Manager

### 1. Test Setup

```go
func setupTestManager() queue.Manager {
    config := queue.DefaultConfig()
    config.Adapter.Default = "memory"  // Sử dụng memory cho test
    config.Server.Concurrency = 1      // Single worker cho predictable testing
    config.Server.LogLevel = 0         // Debug logging
    
    return queue.NewManager(config)
}

func TestManagerIntegration(t *testing.T) {
    manager := setupTestManager()
    
    client := manager.Client()
    server := manager.Server()
    
    // Test workflow
    var processed bool
    server.RegisterHandler("test_task", func(ctx context.Context, task *queue.Task) error {
        processed = true
        return nil
    })
    
    go server.Start()
    defer server.Stop()
    
    _, err := client.Enqueue("test_task", map[string]string{"test": "data"})
    assert.NoError(t, err)
    
    // Wait for processing
    time.Sleep(100 * time.Millisecond)
    assert.True(t, processed)
}
```

### 2. Mock Manager

```go
type MockManager struct {
    clientFunc  func() queue.Client
    serverFunc  func() queue.Server
    adapterFunc func(string) adapter.QueueAdapter
}

func (m *MockManager) Client() queue.Client {
    if m.clientFunc != nil {
        return m.clientFunc()
    }
    return &MockClient{}
}

func (m *MockManager) Server() queue.Server {
    if m.serverFunc != nil {
        return m.serverFunc()
    }
    return &MockServer{}
}

func (m *MockManager) Adapter(name string) adapter.QueueAdapter {
    if m.adapterFunc != nil {
        return m.adapterFunc(name)
    }
    return &MockAdapter{}
}

// Sử dụng trong test
func TestWithMockManager(t *testing.T) {
    mockManager := &MockManager{
        clientFunc: func() queue.Client {
            return &MockClient{
                enqueueFunc: func(taskName string, payload interface{}, opts ...queue.Option) (*queue.TaskInfo, error) {
                    return &queue.TaskInfo{ID: "mock-id"}, nil
                },
            }
        },
    }
    
    service := NewMyService(mockManager)
    err := service.ProcessSomething()
    assert.NoError(t, err)
}
```

## Best Practices

### 1. Manager Lifecycle

```go
type Application struct {
    manager queue.Manager
    server  queue.Server
}

func NewApplication(config queue.Config) *Application {
    manager := queue.NewManager(config)
    
    return &Application{
        manager: manager,
        server:  manager.Server(),
    }
}

func (app *Application) Start() error {
    // Setup handlers
    app.setupHandlers()
    
    // Start server
    return app.server.Start()
}

func (app *Application) Stop() error {
    return app.server.Stop()
}

func (app *Application) setupHandlers() {
    app.server.RegisterHandler("task1", handler1)
    app.server.RegisterHandler("task2", handler2)
}
```

### 2. Configuration Validation

```go
func validateConfig(config queue.Config) error {
    if config.Server.Concurrency <= 0 {
        return fmt.Errorf("concurrency must be positive")
    }
    
    if config.Server.PollingInterval <= 0 {
        return fmt.Errorf("polling interval must be positive")
    }
    
    if config.Adapter.Default == "redis" && config.Adapter.Redis.ProviderKey == "" {
        return fmt.Errorf("redis provider key required for redis adapter")
    }
    
    return nil
}

func NewValidatedManager(config queue.Config) (queue.Manager, error) {
    if err := validateConfig(config); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }
    
    return queue.NewManager(config), nil
}
```

### 3. Resource Management

```go
type ManagedQueue struct {
    manager queue.Manager
    server  queue.Server
    client  queue.Client
    
    done chan struct{}
    wg   sync.WaitGroup
}

func NewManagedQueue(config queue.Config) *ManagedQueue {
    manager := queue.NewManager(config)
    
    return &ManagedQueue{
        manager: manager,
        server:  manager.Server(),
        client:  manager.Client(),
        done:    make(chan struct{}),
    }
}

func (mq *ManagedQueue) Start() error {
    mq.wg.Add(1)
    go func() {
        defer mq.wg.Done()
        mq.server.Start()
    }()
    
    return nil
}

func (mq *ManagedQueue) Stop() error {
    close(mq.done)
    
    if err := mq.server.Stop(); err != nil {
        return err
    }
    
    if err := mq.client.Close(); err != nil {
        return err
    }
    
    mq.wg.Wait()
    return nil
}
```
