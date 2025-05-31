# Queue Provider

Queue Provider là giải pháp xử lý hàng đợi và tác vụ nền đơn giản nhưng mạnh mẽ cho ứng dụng Go, với tích hợp hoàn chỉnh scheduler và khả năng xử lý tác vụ phức tạp.

## Tính năng nổi bật

- **Triển khai đơn giản**: Dễ bảo trì và mở rộng với kiến trúc module hóa
- **Dual Adapter Support**: Hỗ trợ Redis và Memory adapter cho mọi môi trường
- **Redis Provider Integration**: Tích hợp hoàn chỉnh với Redis Provider để centralize Redis configuration
- **Enhanced Redis Features**: Priority queues, TTL support, pipeline operations, monitoring
- **Scheduler Integration**: Tích hợp hoàn chỉnh với Scheduler Provider để xử lý delayed/scheduled tasks
- **Advanced Task Management**: Hỗ trợ retry logic, dead letter queue và task tracking
- **Configuration-driven**: Cấu hình hoàn toàn qua file config với struct validation
- **Worker Model**: Xử lý tác vụ đa luồng với concurrency control
- **Queue Priority**: Hỗ trợ multiple queues với strict priority
- **Batch Processing**: API batch processing cho hiệu suất cao
- **Maintenance Tasks**: Tự động cleanup và retry failed jobs
- **DI Integration**: Tích hợp hoàn chỉnh với DI container

## Cài đặt

Để cài đặt Queue Provider, bạn có thể sử dụng lệnh go get:

```bash
go get go.fork.vn/queue
```

## Cách sử dụng

### 1. Đăng ký Service Provider

#### Cách đơn giản (Auto-configuration)

```go
package main

import (
    "go.fork.vn/di"
    "go.fork.vn/config"
    "go.fork.vn/redis"
    "go.fork.vn/scheduler"
    "go.fork.vn/queue"
)

func main() {
    app := di.New()
    
    // Đăng ký các providers cần thiết
    app.Register(config.NewServiceProvider()) // Required cho cấu hình
    app.Register(redis.NewServiceProvider())  // Required cho Redis adapter
    app.Register(scheduler.NewServiceProvider()) // Required cho delayed tasks
    app.Register(queue.NewServiceProvider())
    
    // Boot ứng dụng - tự động cấu hình từ file config
    app.Boot()
    
    // Giữ ứng dụng chạy để worker có thể xử lý tác vụ
    select {}
}
```

#### Cấu hình thông qua file config

```yaml
# config/app.yaml
queue:
  adapter:
    default: "redis"  # hoặc "memory"
    memory:
      prefix: "queue:"
    redis:
      prefix: "queue:"
      provider_key: "default"  # Sử dụng Redis provider với key "default"
  
  server:
    concurrency: 10
    polling_interval: 1000  # milliseconds
    default_queue: "default"
    strict_priority: true
    queues: ["critical", "high", "default", "low"]
    shutdown_timeout: 30  # seconds
    log_level: 1
    retry_limit: 3
  
  client:
    default_options:
      queue: "default"
      max_retry: 3
      timeout: 30  # minutes

# Cấu hình Redis trong redis section
redis:
  default:  # Redis provider key được reference từ queue config
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    cluster:
      enabled: false
      hosts: ["localhost:7000", "localhost:7001"]

scheduler:
  auto_start: true
  distributed_lock:
    enabled: true  # Cho môi trường distributed
```

### 2. Thêm tác vụ vào hàng đợi (Producer)

```go
// Lấy queue manager từ container
container := app.Container()
manager := container.MustMake("queue").(queue.Manager)

// Hoặc lấy trực tiếp client từ container
client := container.MustMake("queue.client").(queue.Client)

// Thêm tác vụ ngay lập tức với options
payload := map[string]interface{}{
    "user_id": 123,
    "email":   "user@example.com",
    "action":  "welcome",
}

taskInfo, err := client.Enqueue("email:welcome", payload,
    queue.WithQueue("emails"),       // Chỉ định queue
    queue.WithMaxRetry(5),          // Tối đa 5 lần retry
    queue.WithTimeout(2*time.Minute), // Timeout sau 2 phút
    queue.WithTaskID("welcome-123"), // Custom task ID
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Đã thêm tác vụ: %s vào queue: %s\n", taskInfo.ID, taskInfo.Queue)

// Thêm tác vụ delayed (chạy sau 5 phút)
taskInfo, err = client.EnqueueIn("reminder:task", 5*time.Minute, payload,
    queue.WithQueue("notifications"),
)
if err != nil {
    log.Fatal(err)
}

// Thêm tác vụ scheduled (chạy vào thời điểm cụ thể)
processAt := time.Date(2025, 6, 1, 9, 0, 0, 0, time.Local)
taskInfo, err = client.EnqueueAt("report:generate", processAt, payload,
    queue.WithQueue("reports"),
    queue.WithMaxRetry(3),
)
if err != nil {
    log.Fatal(err)
}
```

### 3. Xử lý tác vụ từ hàng đợi (Consumer)

```go
// Lấy queue server từ container
server := container.MustMake("queue.server").(queue.Server)

// Đăng ký handler cho email tasks
server.RegisterHandler("email:welcome", func(ctx context.Context, task *queue.Task) error {
    var payload map[string]interface{}
    if err := task.Unmarshal(&payload); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    userID := int(payload["user_id"].(float64))
    email := payload["email"].(string)
    
    log.Printf("Gửi email chào mừng đến %s (ID: %d)", email, userID)
    
    // Xử lý logic gửi email ở đây...
    // Có thể return error để trigger retry mechanism
    if !sendWelcomeEmail(email) {
        return fmt.Errorf("failed to send email to %s", email)
    }
    
    return nil
})

// Đăng ký handlers cho các loại tác vụ khác với error handling
server.RegisterHandler("reminder:task", func(ctx context.Context, task *queue.Task) error {
    log.Printf("Processing reminder task: %s", task.ID)
    
    // Check context timeout
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process reminder logic
        return processReminder(task)
    }
})

server.RegisterHandler("report:generate", func(ctx context.Context, task *queue.Task) error {
    log.Printf("Generating report: %s", task.ID)
    return generateReport(task)
})

// Đăng ký nhiều handlers cùng một lúc
server.RegisterHandlers(map[string]queue.HandlerFunc{
    "notification:push": handlePushNotification,
    "order:process":     handleOrderProcessing,
    "data:cleanup":      handleDataCleanup,
})

// Server tự động start khi ứng dụng boot
// Nhưng bạn cũng có thể control thủ công:
// err := server.Start()
// if err != nil {
//     log.Fatal(err)
// }

// Graceful shutdown
// defer server.Stop()
```

### 4. Tích hợp với Scheduler (Tính năng mới)

Queue Provider hiện đã tích hợp hoàn chỉnh với Scheduler Provider để xử lý các tác vụ phức tạp:

```go
// Lấy scheduler từ manager
scheduler := manager.Scheduler()

// Schedule tasks mà sẽ enqueue jobs vào queue
scheduler.Every(5).Minutes().Do(func() {
    // Task định kỳ mỗi 5 phút
    client.Enqueue("maintenance:cleanup", map[string]interface{}{
        "type": "temporary_files",
        "timestamp": time.Now(),
    }, queue.WithQueue("maintenance"))
})

// Daily report generation
scheduler.Every(1).Days().At("09:00").Do(func() {
    client.Enqueue("report:daily", map[string]interface{}{
        "date": time.Now().Format("2006-01-02"),
        "type": "sales_summary",
    }, queue.WithQueue("reports"))
})

// Weekly backup với cron expression
scheduler.Cron("0 2 * * 0").Do(func() { // Chủ nhật 2:00 AM
    client.Enqueue("backup:weekly", map[string]interface{}{
        "week": time.Now().Format("2006-W02"),
    }, queue.WithQueue("maintenance"))
})

// Distributed scheduling (chỉ chạy trên 1 instance trong cluster)
scheduler.Every(10).Minutes().Tag("distributed").Do(func() {
    client.Enqueue("monitor:health_check", nil, queue.WithQueue("monitoring"))
})
```

### 5. Tùy chọn nâng cao khi thêm tác vụ

Queue Provider v0.0.3 cung cấp nhiều options linh hoạt:

```go
// Tất cả options có thể dùng khi enqueue
taskInfo, err := client.Enqueue("image:resize", payload,
    queue.WithQueue("media"),                    // Chỉ định queue
    queue.WithMaxRetry(5),                      // Số lần retry tối đa
    queue.WithTimeout(10*time.Minute),          // Timeout cho task
    queue.WithTaskID("resize-user-123-photo"),  // Custom task ID
    queue.WithDelay(30*time.Second),            // Delay trước khi xử lý
    queue.WithDeadline(time.Now().Add(1*time.Hour)), // Deadline tuyệt đối
)
if err != nil {
    log.Fatal(err)
}

// Batch enqueue cho hiệu suất cao
tasks := []map[string]interface{}{
    {"user_id": 1, "action": "welcome"},
    {"user_id": 2, "action": "welcome"},
    {"user_id": 3, "action": "welcome"},
}

for i, task := range tasks {
    client.Enqueue("email:welcome", task,
        queue.WithQueue("emails"),
        queue.WithTaskID(fmt.Sprintf("welcome-%d", i)),
        queue.WithMaxRetry(3),
    )
}

// Process trong ưu tiên queues
highPriorityTask, _ := client.Enqueue("order:urgent", orderData,
    queue.WithQueue("critical"),
    queue.WithMaxRetry(5),
    queue.WithTimeout(30*time.Second),
)

lowPriorityTask, _ := client.Enqueue("analytics:update", analyticsData,
    queue.WithQueue("low"),
    queue.WithMaxRetry(1),
    queue.WithTimeout(5*time.Minute),
)
```

### 6. Sử dụng Memory Adapter (cho môi trường phát triển)

```go
// Memory adapter tự động được sử dụng khi cấu hình default là "memory"
// Hoặc có thể khởi tạo trực tiếp:

package main

import (
    "go.fork.vn/di"
    "go.fork.vn/queue"
    "go.fork.vn/config"
    "go.fork.vn/redis"
)

func main() {
    app := di.New()
    app.Register(config.NewServiceProvider())
    app.Register(redis.NewServiceProvider())  // Required cho Redis adapter
    app.Register(queue.NewServiceProvider())
    
    // Cấu hình sử dụng memory adapter trong config/app.yaml:
    // queue:
    //   adapter:
    //     default: "memory"
    //     memory:
    //       prefix: "test_queue:"
    
    app.Boot()
    
    // Memory adapter không cần Redis và phù hợp cho:
    // - Unit testing
    // - Development environment
    // - Prototype/demo applications
    
    container := app.Container()
    client := container.MustMake("queue.client").(queue.Client)
    
    // Sử dụng giống hệt như Redis adapter
    client.Enqueue("test:task", "test payload")
}
```

### 7. Redis Provider Integration (Tính năng v0.0.5)

Queue Provider hiện đã tích hợp hoàn chỉnh với Redis Provider để centralize Redis configuration và cung cấp advanced Redis features:

```go
// Redis configuration được quản lý bởi Redis Provider
// config/app.yaml
redis:
  default:  # Redis instance key
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    max_retries: 3
    dial_timeout: 5
    read_timeout: 3
    write_timeout: 3
    pool_size: 10

queue:
  adapter:
    default: "redis"
    redis:
      prefix: "queue:"
      provider_key: "default"  # Reference Redis provider key

// Sử dụng advanced Redis features
package main

import (
    "context"
    "time"
    "go.fork.vn/di"
    "go.fork.vn/config"
    "go.fork.vn/redis"
    "go.fork.vn/queue"
    "go.fork.vn/queue/adapter"
)

func main() {
    app := di.New()
    app.Register(config.NewServiceProvider())
    app.Register(redis.NewServiceProvider())
    app.Register(queue.NewServiceProvider())
    app.Boot()

    container := app.Container()
    manager := container.MustMake("queue").(queue.Manager)
    
    // Lấy Redis queue adapter để sử dụng advanced features
    redisAdapter := manager.RedisAdapter()
    
    // Type assertion để access Redis-specific methods
    if redisQueue, ok := redisAdapter.(adapter.QueueRedisAdapter); ok {
        ctx := context.Background()
        
        // Priority queue operations
        err := redisQueue.EnqueueWithPriority(ctx, "tasks", &task, 10) // Priority 10
        if err != nil {
            log.Fatal(err)
        }
        
        // Dequeue from priority queue
        var priorityTask queue.Task
        err = redisQueue.DequeueFromPriority(ctx, "tasks", &priorityTask)
        if err != nil {
            log.Printf("No priority tasks available: %v", err)
        }
        
        // TTL support
        err = redisQueue.EnqueueWithTTL(ctx, "temporary", &task, 1*time.Hour)
        if err != nil {
            log.Fatal(err)
        }
        
        // Batch operations với pipeline
        tasks := []*queue.Task{&task1, &task2, &task3}
        err = redisQueue.EnqueueWithPipeline(ctx, "batch", tasks)
        if err != nil {
            log.Fatal(err)
        }
        
        // Multi-dequeue
        results, err := redisQueue.MultiDequeue(ctx, "batch", 5) // Lấy tối đa 5 tasks
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("Dequeued %d tasks", len(results))
        
        // Queue monitoring
        info, err := redisQueue.GetQueueInfo(ctx, "tasks")
        if err != nil {
            log.Fatal(err)
        }
        log.Printf("Queue info: %+v", info)
        
        // Health check
        if err := redisQueue.Ping(ctx); err != nil {
            log.Printf("Redis connection issue: %v", err)
        }
        
        // Development/testing utilities
        if err := redisQueue.FlushQueues(ctx, []string{"test", "debug"}); err != nil {
            log.Printf("Failed to flush queues: %v", err)
        }
    }
}
```

### 8. Failed Jobs và Dead Letter Queue (Tính năng nâng cao)

Queue Provider v0.0.3 có hệ thống xử lý lỗi tiên tiến:

```go
// Failed jobs được tự động retry với exponential backoff
server.RegisterHandler("risky:task", func(ctx context.Context, task *queue.Task) error {
    // Giả lập task có thể fail
    if rand.Float32() < 0.3 { // 30% chance fail
        return fmt.Errorf("simulated failure")
    }
    
    log.Printf("Task %s completed successfully", task.ID)
    return nil
})

// Tasks sẽ được retry tối đa theo cấu hình (mặc định 3 lần)
// Delay giữa các lần retry tăng theo exponential backoff:
// - Retry 1: 1 minute
// - Retry 2: 4 minutes  
// - Retry 3: 9 minutes

// Sau khi vượt quá retry limit, task sẽ được chuyển vào Dead Letter Queue
// Dead Letter Queue có thể được monitor và xử lý thủ công

// Hệ thống maintenance tự động:
// - Cleanup dead letter tasks cũ hơn 7 ngày (chạy mỗi giờ)
// - Retry failed tasks đủ điều kiện (chạy mỗi 5 phút)
// - Xử lý delayed tasks đã đến hạn (chạy mỗi 30 giây)
```

### 8. Monitoring và Debugging

```go
// TaskInfo cung cấp thông tin chi tiết về task
taskInfo, err := client.Enqueue("debug:task", payload)
if err == nil {
    log.Printf("Task created: %s", taskInfo.String())
    // Output: Task ID: abc-123, Name: debug:task, Queue: default, 
    //         State: pending, Created: 2025-05-28T10:30:00Z
}

// Server logging tự động theo dõi:
// - Task processing time
// - Worker performance  
// - Retry attempts
// - Failed task reasons
// - Queue sizes

// Có thể tùy chỉnh log level trong config:
// queue:
//   server:
//     log_level: 2  # 0=SILENT, 1=ERROR, 2=INFO, 3=DEBUG
```

### 9. Production Best Practices

```go
// Cấu hình production-ready
// config/production.yaml
/*
queue:
  adapter:
    default: "redis"
    redis:
      prefix: "myapp_queue:"
      provider_key: "default"  # Reference Redis provider key
  
  server:
    concurrency: 50              # Điều chỉnh theo CPU cores
    polling_interval: 500        # Giảm cho high-throughput
    strict_priority: true        # Đảm bảo critical tasks được ưu tiên
    queues: ["critical", "high", "default", "low", "bulk"]
    shutdown_timeout: 60         # Đủ thời gian cho graceful shutdown
    retry_limit: 5              # Tăng retry cho production
    
  client:
    default_options:
      queue: "default"
      max_retry: 3
      timeout: 15               # 15 phút timeout mặc định

# Redis cấu hình riêng biệt trong Redis provider
redis:
  default:
    host: "redis-cluster.internal"
    port: 6379
    password: "${REDIS_PASSWORD}"
    db: 0
    cluster:
      enabled: true
      hosts:
        - "redis-1.internal:6379"
        - "redis-2.internal:6379"
        - "redis-3.internal:6379"

scheduler:
  auto_start: true
  distributed_lock:
    enabled: true               # Bắt buộc cho production cluster
  options:
    key_prefix: "myapp_scheduler:"
    lock_duration: 120          # 2 phút
    max_retries: 5
    retry_delay: 500
*/

// Graceful shutdown handling
func setupGracefulShutdown(server queue.Server) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Shutting down gracefully...")
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := server.Stop(); err != nil {
            log.Printf("Error during shutdown: %v", err)
        }
        
        log.Println("Shutdown complete")
        os.Exit(0)
    }()
}
```
