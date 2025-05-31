# Usage Guide - Queue Package

## Cài đặt

```bash
go get go.fork.vn/queue@v0.1.0
```

## Import

```go
import "go.fork.vn/queue"
```

## Quick Start

### 1. Sử dụng cơ bản

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "go.fork.vn/queue"
)

func main() {
    // Tạo queue manager
    queueManager := queue.NewQueue()
    
    // Khởi tạo connection
    ctx := context.Background()
    if err := queueManager.Connect(ctx); err != nil {
        log.Fatal("Lỗi kết nối queue:", err)
    }
    
    // Đăng ký worker cho job
    queueManager.Job("send-email", func(ctx context.Context, payload queue.Payload) error {
        email := payload.GetString("email")
        subject := payload.GetString("subject")
        
        fmt.Printf("Sending email to %s with subject: %s\n", email, subject)
        
        // Simulate email sending
        time.Sleep(100 * time.Millisecond)
        return nil
    })
    
    // Dispatch job
    payload := queue.NewPayload().
        Set("email", "user@example.com").
        Set("subject", "Welcome!")
        
    job := queue.NewJob("send-email", payload)
    
    if err := queueManager.Dispatch(ctx, job); err != nil {
        log.Fatal("Lỗi dispatch job:", err)
    }
    
    fmt.Println("Job dispatched successfully!")
}
```

### 2. Sử dụng với Dependency Injection

```go
package main

import (
    "context"
    "log"
    
    "go.fork.vn/di"
    "go.fork.vn/queue"
)

func main() {
    // Tạo DI container
    container := di.NewContainer()
    
    // Đăng ký Queue service provider
    container.Register(queue.NewServiceProvider())
    
    // Boot container
    if err := container.Boot(); err != nil {
        log.Fatal("Lỗi boot container:", err)
    }
    
    // Resolve queue manager
    var queueManager queue.Manager
    if err := container.Resolve(&queueManager); err != nil {
        log.Fatal("Lỗi resolve queue manager:", err)
    }
    
    // Sử dụng queue manager
    ctx := context.Background()
    if err := queueManager.Connect(ctx); err != nil {
        log.Fatal("Lỗi kết nối queue:", err)
    }
    
    // Register jobs và workers...
}
```

## Chi tiết sử dụng

### Job Management

#### Tạo và dispatch job

```go
// Tạo payload
payload := queue.NewPayload().
    Set("user_id", 123).
    Set("action", "welcome").
    Set("data", map[string]interface{}{
        "name": "John Doe",
        "email": "john@example.com",
    })

// Tạo job
job := queue.NewJob("process-user", payload)

// Set options (tùy chọn)
job.SetPriority(queue.PriorityHigh)
job.SetDelay(5 * time.Minute)
job.SetMaxRetries(3)

// Dispatch
ctx := context.Background()
if err := queueManager.Dispatch(ctx, job); err != nil {
    log.Printf("Lỗi dispatch job: %v", err)
}
```

#### Tạo job với delay

```go
// Job chạy sau 10 phút
delayedJob := queue.NewJob("delayed-task", payload).
    SetDelay(10 * time.Minute)

if err := queueManager.Dispatch(ctx, delayedJob); err != nil {
    log.Printf("Lỗi dispatch delayed job: %v", err)
}
```

#### Tạo job với priority

```go
// Job priority cao
highPriorityJob := queue.NewJob("urgent-task", payload).
    SetPriority(queue.PriorityHigh)

// Job priority thấp
lowPriorityJob := queue.NewJob("background-task", payload).
    SetPriority(queue.PriorityLow)
```

### Worker Registration

#### Worker đơn giản

```go
queueManager.Job("send-notification", func(ctx context.Context, payload queue.Payload) error {
    userID := payload.GetInt("user_id")
    message := payload.GetString("message")
    
    // Logic gửi notification
    return sendNotification(userID, message)
})
```

#### Worker với error handling

```go
queueManager.Job("process-payment", func(ctx context.Context, payload queue.Payload) error {
    paymentID := payload.GetString("payment_id")
    
    // Xử lý payment
    if err := processPayment(paymentID); err != nil {
        // Log error
        log.Printf("Payment processing failed for %s: %v", paymentID, err)
        
        // Return error để trigger retry
        return fmt.Errorf("payment processing failed: %w", err)
    }
    
    return nil
})
```

#### Worker với timeout

```go
queueManager.Job("long-running-task", func(ctx context.Context, payload queue.Payload) error {
    // Tạo context với timeout
    taskCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Xử lý task với timeout
    return performLongRunningTask(taskCtx, payload)
})
```

### Failed Job Handling

#### Xử lý failed jobs

```go
// Lấy danh sách failed jobs
failedJobs, err := queueManager.GetFailedJobs(ctx)
if err != nil {
    log.Printf("Lỗi lấy failed jobs: %v", err)
    return
}

for _, job := range failedJobs {
    fmt.Printf("Failed job: %s, Error: %s\n", job.GetName(), job.GetLastError())
    
    // Retry job
    if err := queueManager.RetryJob(ctx, job.GetID()); err != nil {
        log.Printf("Lỗi retry job %s: %v", job.GetID(), err)
    }
}
```

#### Flush failed jobs

```go
// Xóa tất cả failed jobs
if err := queueManager.FlushFailedJobs(ctx); err != nil {
    log.Printf("Lỗi flush failed jobs: %v", err)
}
```

### Queue Statistics

```go
// Lấy thống kê queue
stats, err := queueManager.GetStats(ctx)
if err != nil {
    log.Printf("Lỗi lấy stats: %v", err)
    return
}

fmt.Printf("Pending jobs: %d\n", stats.Pending)
fmt.Printf("Processing jobs: %d\n", stats.Processing)
fmt.Printf("Completed jobs: %d\n", stats.Completed)
fmt.Printf("Failed jobs: %d\n", stats.Failed)
```

### Configuration với config package

```go
package main

import (
    "context"
    "log"
    
    "go.fork.vn/config"
    "go.fork.vn/di"
    "go.fork.vn/queue"
)

func main() {
    // Tạo container
    container := di.NewContainer()
    
    // Đăng ký config service provider
    container.Register(config.NewServiceProvider())
    
    // Đăng ký queue service provider  
    container.Register(queue.NewServiceProvider())
    
    // Boot container
    if err := container.Boot(); err != nil {
        log.Fatal("Lỗi boot container:", err)
    }
    
    // Resolve services
    var cfg config.Manager
    var queueManager queue.Manager
    
    container.Resolve(&cfg)
    container.Resolve(&queueManager)
    
    // Queue sẽ tự động đọc config từ config manager
    ctx := context.Background()
    if err := queueManager.Connect(ctx); err != nil {
        log.Fatal("Lỗi kết nối queue:", err)
    }
}
```

### File cấu hình mẫu (config.yaml)

```yaml
queue:
  driver: "redis"
  connection: "default"
  
  redis:
    host: "localhost"
    port: 6379
    db: 0
    password: ""
    
  options:
    max_retries: 3
    retry_delay: "5s"
    worker_count: 5
    timeout: "30s"
    
  queues:
    default:
      priority: 1
      worker_count: 3
    high_priority:
      priority: 10
      worker_count: 2
    low_priority:
      priority: 0
      worker_count: 1
```

## Advanced Usage

### Custom Job Processor

```go
type CustomProcessor struct {
    logger log.Logger
}

func (p *CustomProcessor) Process(ctx context.Context, job queue.Job) error {
    p.logger.Info("Processing job", "job_id", job.GetID(), "job_name", job.GetName())
    
    // Custom processing logic
    payload := job.GetPayload()
    
    switch job.GetName() {
    case "email":
        return p.processEmail(ctx, payload)
    case "sms":
        return p.processSMS(ctx, payload)
    default:
        return fmt.Errorf("unknown job type: %s", job.GetName())
    }
}

func (p *CustomProcessor) processEmail(ctx context.Context, payload queue.Payload) error {
    // Email processing logic
    return nil
}

func (p *CustomProcessor) processSMS(ctx context.Context, payload queue.Payload) error {
    // SMS processing logic
    return nil
}

// Đăng ký custom processor
processor := &CustomProcessor{logger: logger}
queueManager.SetProcessor("email", processor)
queueManager.SetProcessor("sms", processor)
```

### Middleware cho Jobs

```go
// Logging middleware
loggingMiddleware := func(next queue.JobHandler) queue.JobHandler {
    return func(ctx context.Context, payload queue.Payload) error {
        start := time.Now()
        
        log.Printf("Starting job processing")
        err := next(ctx, payload)
        
        duration := time.Since(start)
        if err != nil {
            log.Printf("Job failed after %v: %v", duration, err)
        } else {
            log.Printf("Job completed in %v", duration)
        }
        
        return err
    }
}

// Metric middleware
metricMiddleware := func(next queue.JobHandler) queue.JobHandler {
    return func(ctx context.Context, payload queue.Payload) error {
        // Record metrics
        metrics.Increment("jobs.started")
        
        err := next(ctx, payload)
        
        if err != nil {
            metrics.Increment("jobs.failed")
        } else {
            metrics.Increment("jobs.completed")
        }
        
        return err
    }
}

// Sử dụng middleware
queueManager.Job("process-order", 
    loggingMiddleware(
        metricMiddleware(func(ctx context.Context, payload queue.Payload) error {
            // Job logic
            return processOrder(ctx, payload)
        }),
    ),
)
```

### Batch Job Processing

```go
// Dispatch multiple jobs
jobs := []queue.Job{
    queue.NewJob("send-email", emailPayload1),
    queue.NewJob("send-email", emailPayload2),
    queue.NewJob("send-sms", smsPayload),
}

if err := queueManager.DispatchBatch(ctx, jobs); err != nil {
    log.Printf("Lỗi dispatch batch: %v", err)
}
```

### Job Chains

```go
// Tạo job chain
chain := queue.NewChain().
    Add("validate-data", validatePayload).
    Add("process-payment", processPayload).
    Add("send-confirmation", confirmPayload)

if err := queueManager.DispatchChain(ctx, chain); err != nil {
    log.Printf("Lỗi dispatch chain: %v", err)
}
```

## Error Handling

### Retry Strategy

```go
// Custom retry strategy
retryStrategy := queue.RetryStrategy{
    MaxRetries: 5,
    BackoffStrategy: queue.ExponentialBackoff{
        InitialDelay: 1 * time.Second,
        MaxDelay:     5 * time.Minute,
        Multiplier:   2,
    },
}

job := queue.NewJob("retry-job", payload).
    SetRetryStrategy(retryStrategy)
```

### Dead Letter Queue

```go
// Job sẽ được chuyển vào dead letter queue sau khi retry hết
queueManager.Job("risky-job", func(ctx context.Context, payload queue.Payload) error {
    // Logic có thể fail
    return riskyOperation(payload)
})

// Xử lý dead letter queue
deadJobs, err := queueManager.GetDeadJobs(ctx)
if err != nil {
    log.Printf("Lỗi lấy dead jobs: %v", err)
}

for _, job := range deadJobs {
    log.Printf("Dead job: %s, Final error: %s", job.GetName(), job.GetLastError())
}
```

## Best Practices

### 1. Job Design

```go
// ✅ Good: Idempotent job
queueManager.Job("send-welcome-email", func(ctx context.Context, payload queue.Payload) error {
    userID := payload.GetInt("user_id")
    
    // Check if email already sent
    if emailAlreadySent(userID) {
        return nil // Job is idempotent
    }
    
    return sendWelcomeEmail(userID)
})

// ❌ Bad: Non-idempotent job
queueManager.Job("increment-counter", func(ctx context.Context, payload queue.Payload) error {
    // This will cause issues if retried
    counter++
    return nil
})
```

### 2. Resource Management

```go
queueManager.Job("process-file", func(ctx context.Context, payload queue.Payload) error {
    filePath := payload.GetString("file_path")
    
    // Open file
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close() // Always close resources
    
    // Process file
    return processFile(file)
})
```

### 3. Context Handling

```go
queueManager.Job("external-api-call", func(ctx context.Context, payload queue.Payload) error {
    // Respect context cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Use context in HTTP calls
    req, err := http.NewRequestWithContext(ctx, "POST", apiURL, body)
    if err != nil {
        return err
    }
    
    resp, err := httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return processResponse(resp)
})
```

## Testing

### Unit Testing Jobs

```go
func TestSendEmailJob(t *testing.T) {
    // Tạo test payload
    payload := queue.NewPayload().
        Set("email", "test@example.com").
        Set("subject", "Test Subject")
    
    // Mock dependencies
    mockEmailService := &MockEmailService{}
    
    // Create job handler
    handler := createSendEmailHandler(mockEmailService)
    
    // Execute job
    ctx := context.Background()
    err := handler(ctx, payload)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, mockEmailService.WasCalled())
}
```

### Integration Testing

```go
func TestQueueIntegration(t *testing.T) {
    // Setup test queue
    queueManager := queue.NewQueue()
    
    // Use test configuration
    queueManager.SetConfig(map[string]interface{}{
        "driver": "memory", // Use in-memory driver for tests
    })
    
    ctx := context.Background()
    err := queueManager.Connect(ctx)
    require.NoError(t, err)
    
    // Register test job
    var processedPayload queue.Payload
    queueManager.Job("test-job", func(ctx context.Context, payload queue.Payload) error {
        processedPayload = payload
        return nil
    })
    
    // Dispatch job
    testPayload := queue.NewPayload().Set("test", "value")
    job := queue.NewJob("test-job", testPayload)
    
    err = queueManager.Dispatch(ctx, job)
    require.NoError(t, err)
    
    // Wait for processing
    time.Sleep(100 * time.Millisecond)
    
    // Assert
    assert.Equal(t, "value", processedPayload.GetString("test"))
}
```

## Troubleshooting

### Common Issues

1. **Job không được xử lý**
   - Kiểm tra worker có được đăng ký chính xác
   - Kiểm tra connection đến queue backend
   - Kiểm tra log để tìm errors

2. **Job bị retry liên tục**
   - Kiểm tra job logic có throw error không cần thiết
   - Điều chỉnh retry strategy
   - Kiểm tra external dependencies

3. **Performance issues**
   - Tăng số lượng workers
   - Optimize job logic
   - Sử dụng batch processing cho jobs tương tự

### Debug Mode

```go
// Enable debug logging
queueManager.SetDebug(true)

// Hoặc qua config
config:
  queue:
    debug: true
    log_level: "debug"
```

### Monitoring

```go
// Health check
health := queueManager.Health(ctx)
if !health.IsHealthy() {
    log.Printf("Queue unhealthy: %v", health.GetErrors())
}

// Performance metrics
metrics := queueManager.GetMetrics(ctx)
fmt.Printf("Average processing time: %v\n", metrics.AvgProcessingTime)
fmt.Printf("Throughput: %d jobs/min\n", metrics.Throughput)
```
