# Queue Server - Xử lý tác vụ từ hàng đợi

Queue Server là thành phần chịu trách nhiệm lấy và xử lý các tác vụ từ hàng đợi. Tài liệu này mô tả chi tiết cách thiết lập, cấu hình và vận hành server để xử lý các tác vụ một cách hiệu quả.

## Server Interface

**Vị trí**: [`server.go`](../server.go)

```go
type Server interface {
    // RegisterHandler đăng ký handler để xử lý một loại task cụ thể
    RegisterHandler(taskName string, handler HandlerFunc)
    
    // Start khởi động server và bắt đầu xử lý tasks
    Start() error
    
    // Stop dừng server một cách graceful
    Stop() error
    
    // IsRunning kiểm tra server có đang chạy hay không
    IsRunning() bool
    
    // GetScheduler trả về scheduler manager
    GetScheduler() scheduler.Manager
    
    // SetScheduler thiết lập scheduler từ bên ngoài
    SetScheduler(scheduler scheduler.Manager)
}
```

## Server Options

```go
type ServerOptions struct {
    Concurrency      int           // Số worker xử lý đồng thời
    PollingInterval  int           // Khoảng thời gian polling (ms)
    DefaultQueue     string        // Queue mặc định
    StrictPriority   bool          // Ưu tiên nghiêm ngặt queues
    Queues           []string      // Danh sách queues theo thứ tự ưu tiên
    ShutdownTimeout  time.Duration // Timeout khi graceful shutdown
    LogLevel         int           // Mức độ logging
    RetryLimit       int           // Số lần retry tối đa
}
```

## Khởi tạo Server

### 1. Qua Manager (Recommended)

```go
import "go.fork.vn/queue"

func main() {
    // Tạo config
    config := queue.DefaultConfig()
    config.Server.Concurrency = 10
    config.Server.Queues = []string{"critical", "high", "default", "low"}
    
    // Tạo manager
    manager := queue.NewManager(config)
    
    // Lấy server từ manager
    server := manager.Server()
    
    // Sử dụng server...
}
```

### 2. Trực tiếp với Redis

```go
import (
    "github.com/redis/go-redis/v9"
    "go.fork.vn/queue"
)

func main() {
    // Tạo Redis client
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
        DB:   0,
    })
    
    // Cấu hình server
    opts := queue.ServerOptions{
        Concurrency:     10,
        PollingInterval: 1000,
        DefaultQueue:    "default",
        StrictPriority:  true,
        Queues:          []string{"critical", "high", "default", "low"},
        ShutdownTimeout: 30 * time.Second,
        LogLevel:        1, // Info
        RetryLimit:      3,
    }
    
    // Tạo server
    server := queue.NewServer(redisClient, opts)
}
```

### 3. Với Custom Adapter

```go
import (
    "go.fork.vn/queue"
    "go.fork.vn/queue/adapter"
)

func main() {
    // Memory adapter cho development/testing
    memoryAdapter := adapter.NewMemoryQueue("queue:")
    
    opts := queue.ServerOptions{
        Concurrency:     5,
        PollingInterval: 500,
        DefaultQueue:    "default",
        Queues:          []string{"default"},
        ShutdownTimeout: 10 * time.Second,
    }
    
    server := queue.NewServerWithAdapter(memoryAdapter, opts)
}
```

## Handler Functions

### Handler Function Signature

```go
type HandlerFunc func(ctx context.Context, task *Task) error
```

### 1. Basic Handler

```go
func main() {
    server := getServer() // Khởi tạo server
    
    // Handler đơn giản
    server.RegisterHandler("send_email", func(ctx context.Context, task *queue.Task) error {
        // Parse payload
        var emailData struct {
            To      string `json:"to"`
            Subject string `json:"subject"`
            Body    string `json:"body"`
        }
        
        if err := task.Unmarshal(&emailData); err != nil {
            return fmt.Errorf("invalid email payload: %w", err)
        }
        
        // Xử lý gửi email
        err := sendEmail(emailData.To, emailData.Subject, emailData.Body)
        if err != nil {
            return fmt.Errorf("failed to send email: %w", err)
        }
        
        log.Printf("Email sent to %s: %s", emailData.To, emailData.Subject)
        return nil
    })
}
```

### 2. Handler với Context

```go
server.RegisterHandler("process_file", func(ctx context.Context, task *queue.Task) error {
    // Kiểm tra context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err() // Trả về context.Canceled hoặc context.DeadlineExceeded
    default:
    }
    
    var fileData struct {
        FilePath string `json:"file_path"`
        UserID   int    `json:"user_id"`
    }
    
    if err := task.Unmarshal(&fileData); err != nil {
        return err
    }
    
    // Long-running process với context checking
    for i := 0; i < 100; i++ {
        select {
        case <-ctx.Done():
            log.Println("File processing cancelled")
            return ctx.Err()
        default:
        }
        
        // Process chunk
        if err := processFileChunk(fileData.FilePath, i); err != nil {
            return err
        }
        
        // Simulate work
        time.Sleep(100 * time.Millisecond)
    }
    
    log.Printf("File processed successfully: %s", fileData.FilePath)
    return nil
})
```

### 3. Handler với Retry Logic

```go
server.RegisterHandler("api_call", func(ctx context.Context, task *queue.Task) error {
    var apiData struct {
        URL     string            `json:"url"`
        Method  string            `json:"method"`
        Headers map[string]string `json:"headers"`
        Body    string            `json:"body"`
    }
    
    if err := task.Unmarshal(&apiData); err != nil {
        return err
    }
    
    // Temporary error - sẽ retry
    if isTemporaryError(err) {
        return fmt.Errorf("temporary API error (will retry): %w", err)
    }
    
    // Permanent error - không retry
    if isPermanentError(err) {
        log.Printf("Permanent error for task %s: %v", task.ID, err)
        return nil // Trả về nil để không retry
    }
    
    // Success
    return makeAPICall(ctx, apiData)
})

func isTemporaryError(err error) bool {
    // HTTP 5xx, network timeout, etc.
    return strings.Contains(err.Error(), "timeout") ||
           strings.Contains(err.Error(), "connection refused")
}

func isPermanentError(err error) bool {
    // HTTP 4xx (except 429), validation error, etc.
    return strings.Contains(err.Error(), "400") ||
           strings.Contains(err.Error(), "401") ||
           strings.Contains(err.Error(), "404")
}
```

### 4. Handler với Structured Logging

```go
import (
    "log/slog"
    "go.fork.vn/queue"
)

server.RegisterHandler("user_notification", func(ctx context.Context, task *queue.Task) error {
    logger := slog.With(
        "task_id", task.ID,
        "task_name", task.Name,
        "queue", task.Queue,
        "retry_count", task.RetryCount,
    )
    
    logger.Info("Starting task processing")
    
    var notifData struct {
        UserID  int    `json:"user_id"`
        Type    string `json:"type"`
        Message string `json:"message"`
    }
    
    if err := task.Unmarshal(&notifData); err != nil {
        logger.Error("Failed to unmarshal task payload", "error", err)
        return err
    }
    
    logger.Info("Processing notification", "user_id", notifData.UserID, "type", notifData.Type)
    
    // Process notification
    if err := sendNotification(notifData); err != nil {
        logger.Error("Failed to send notification", "error", err)
        return err
    }
    
    logger.Info("Notification sent successfully")
    return nil
})
```

## Server Lifecycle

### 1. Khởi động Server

```go
func main() {
    server := setupServer()
    
    // Đăng ký handlers
    registerHandlers(server)
    
    // Graceful shutdown
    setupGracefulShutdown(server)
    
    // Khởi động server
    log.Println("Starting queue server...")
    if err := server.Start(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func registerHandlers(server queue.Server) {
    server.RegisterHandler("send_email", emailHandler)
    server.RegisterHandler("process_file", fileHandler)
    server.RegisterHandler("generate_report", reportHandler)
    server.RegisterHandler("cleanup_data", cleanupHandler)
}
```

### 2. Graceful Shutdown

```go
import (
    "os"
    "os/signal"
    "syscall"
)

func setupGracefulShutdown(server queue.Server) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    go func() {
        <-c
        log.Println("Received shutdown signal, stopping server...")
        
        if err := server.Stop(); err != nil {
            log.Printf("Error stopping server: %v", err)
        } else {
            log.Println("Server stopped gracefully")
        }
        
        os.Exit(0)
    }()
}
```

### 3. Health Checks

```go
import "net/http"

func setupHealthCheck(server queue.Server) {
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        if server.IsRunning() {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte("OK"))
        } else {
            w.WriteHeader(http.StatusServiceUnavailable)
            w.Write([]byte("Server not running"))
        }
    })
    
    go http.ListenAndServe(":8080", nil)
}
```

## Scheduled Tasks Integration

### Thiết lập Scheduler

```go
import "go.fork.vn/scheduler"

func main() {
    server := setupServer()
    
    // Tạo scheduler
    schedulerManager := scheduler.NewScheduler()
    server.SetScheduler(schedulerManager)
    
    // Thiết lập scheduled tasks
    setupScheduledTasks(schedulerManager)
    
    server.Start()
}

func setupScheduledTasks(scheduler scheduler.Manager) {
    // Task chạy mỗi 5 phút
    scheduler.Every(5).Minutes().Do(func() {
        log.Println("Running maintenance task...")
        // Logic maintenance
    })
    
    // Task chạy hàng ngày lúc 2:00 AM
    scheduler.Every(1).Day().At("02:00").Do(func() {
        log.Println("Running daily cleanup...")
        // Logic cleanup
    })
    
    // Task chạy mỗi thứ 2 lúc 9:00 AM
    scheduler.Every(1).Monday().At("09:00").Do(func() {
        log.Println("Running weekly report...")
        // Logic báo cáo tuần
    })
}
```

## Error Handling và Recovery

### 1. Global Error Handler

```go
type ErrorHandler struct {
    logger *slog.Logger
}

func (h *ErrorHandler) HandleError(task *queue.Task, err error) {
    h.logger.Error("Task processing failed",
        "task_id", task.ID,
        "task_name", task.Name,
        "retry_count", task.RetryCount,
        "max_retry", task.MaxRetry,
        "error", err,
    )
    
    // Send alert nếu task đã retry nhiều lần
    if task.RetryCount >= task.MaxRetry-1 {
        h.sendAlert(task, err)
    }
}

func (h *ErrorHandler) sendAlert(task *queue.Task, err error) {
    // Send to monitoring system
    // Send email notification
    // Log to special error queue
}
```

### 2. Dead Letter Queue

```go
server.RegisterHandler("failed_task_handler", func(ctx context.Context, task *queue.Task) error {
    // Handler cho tasks đã fail nhiều lần
    log.Printf("Processing failed task: %s", task.ID)
    
    // Log chi tiết hoặc gửi alert
    sendFailureNotification(task)
    
    // Có thể attempt manual recovery
    if shouldAttemptRecovery(task) {
        return retryTaskWithDifferentStrategy(task)
    }
    
    // Hoặc store vào persistent storage để manual review
    return storeForManualReview(task)
})
```

## Monitoring và Metrics

### 1. Built-in Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    tasksProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "queue_tasks_processed_total",
            Help: "Total number of processed tasks",
        },
        []string{"task_name", "queue", "status"},
    )
    
    taskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "queue_task_duration_seconds",
            Help: "Task processing duration",
        },
        []string{"task_name", "queue"},
    )
    
    queueLength = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "queue_length",
            Help: "Current queue length",
        },
        []string{"queue"},
    )
)

func instrumentedHandler(taskName string, handler queue.HandlerFunc) queue.HandlerFunc {
    return func(ctx context.Context, task *queue.Task) error {
        start := time.Now()
        
        err := handler(ctx, task)
        
        duration := time.Since(start)
        status := "success"
        if err != nil {
            status = "error"
        }
        
        tasksProcessed.WithLabelValues(taskName, task.Queue, status).Inc()
        taskDuration.WithLabelValues(taskName, task.Queue).Observe(duration.Seconds())
        
        return err
    }
}
```

### 2. Custom Middleware

```go
type Middleware func(queue.HandlerFunc) queue.HandlerFunc

func LoggingMiddleware(logger *slog.Logger) Middleware {
    return func(next queue.HandlerFunc) queue.HandlerFunc {
        return func(ctx context.Context, task *queue.Task) error {
            start := time.Now()
            
            logger.Info("Task started",
                "task_id", task.ID,
                "task_name", task.Name,
                "queue", task.Queue,
            )
            
            err := next(ctx, task)
            
            duration := time.Since(start)
            
            if err != nil {
                logger.Error("Task failed",
                    "task_id", task.ID,
                    "duration", duration,
                    "error", err,
                )
            } else {
                logger.Info("Task completed",
                    "task_id", task.ID,
                    "duration", duration,
                )
            }
            
            return err
        }
    }
}

func WithMiddleware(handler queue.HandlerFunc, middlewares ...Middleware) queue.HandlerFunc {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

// Sử dụng
server.RegisterHandler("send_email", 
    WithMiddleware(emailHandler,
        LoggingMiddleware(logger),
        MetricsMiddleware(),
        TracingMiddleware(),
    ),
)
```

## Best Practices

### 1. Handler Design

```go
// ✅ Good: Idempotent handler
server.RegisterHandler("update_user_score", func(ctx context.Context, task *queue.Task) error {
    var data struct {
        UserID int `json:"user_id"`
        Score  int `json:"score"`
    }
    
    if err := task.Unmarshal(&data); err != nil {
        return err
    }
    
    // Idempotent operation - có thể chạy nhiều lần với kết quả giống nhau
    return updateUserScore(data.UserID, data.Score)
})

// ❌ Bad: Non-idempotent handler
server.RegisterHandler("increment_counter", func(ctx context.Context, task *queue.Task) error {
    // Có thể bị increment nhiều lần nếu retry
    return incrementCounter()
})
```

### 2. Resource Management

```go
server.RegisterHandler("process_file", func(ctx context.Context, task *queue.Task) error {
    var fileData struct {
        FilePath string `json:"file_path"`
    }
    
    if err := task.Unmarshal(&fileData); err != nil {
        return err
    }
    
    // Mở file
    file, err := os.Open(fileData.FilePath)
    if err != nil {
        return err
    }
    defer file.Close() // Luôn cleanup resources
    
    // Process file...
    return processFile(ctx, file)
})
```

### 3. Configuration Management

```go
type ServerConfig struct {
    Development bool
    LogLevel    string
    Timeout     time.Duration
}

func NewProductionServer(config ServerConfig) queue.Server {
    opts := queue.ServerOptions{
        Concurrency:     50,
        PollingInterval: 1000,
        StrictPriority:  true,
        Queues:          []string{"critical", "high", "default", "low"},
        ShutdownTimeout: 60 * time.Second,
        LogLevel:        1, // Info
        RetryLimit:      5,
    }
    
    if config.Development {
        opts.Concurrency = 2
        opts.LogLevel = 0 // Debug
        opts.ShutdownTimeout = 10 * time.Second
    }
    
    return queue.NewServer(getRedisClient(), opts)
}
```

### 4. Testing

```go
func TestEmailHandler(t *testing.T) {
    // Sử dụng memory adapter cho testing
    adapter := adapter.NewMemoryQueue("test:")
    
    opts := queue.ServerOptions{
        Concurrency: 1,
        DefaultQueue: "test",
    }
    
    server := queue.NewServerWithAdapter(adapter, opts)
    
    // Register test handler
    server.RegisterHandler("send_email", emailHandler)
    
    // Test data
    task := &queue.Task{
        ID:   "test-task",
        Name: "send_email",
        Payload: []byte(`{"to":"test@example.com","subject":"Test"}`),
    }
    
    // Execute handler directly
    ctx := context.Background()
    err := emailHandler(ctx, task)
    
    assert.NoError(t, err)
    // Verify side effects...
}
```
