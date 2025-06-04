# Queue Client - Gửi tác vụ vào hàng đợi

Queue Client là thành phần chịu trách nhiệm gửi các tác vụ vào hàng đợi để xử lý. Tài liệu này mô tả chi tiết cách sử dụng Client để enqueue tasks với các tùy chọn khác nhau.

## Client Interface

**Vị trí**: [`client.go`](../client.go)

```go
type Client interface {
    // Enqueue đưa một tác vụ vào hàng đợi để xử lý ngay lập tức
    Enqueue(taskName string, payload interface{}, opts ...Option) (*TaskInfo, error)
    
    // EnqueueContext tương tự Enqueue nhưng với context
    EnqueueContext(ctx context.Context, taskName string, payload interface{}, opts ...Option) (*TaskInfo, error)
    
    // EnqueueIn đưa một tác vụ vào hàng đợi để xử lý sau một khoảng thời gian
    EnqueueIn(taskName string, delay time.Duration, payload interface{}, opts ...Option) (*TaskInfo, error)
    
    // EnqueueAt đưa một tác vụ vào hàng đợi để xử lý vào một thời điểm cụ thể
    EnqueueAt(taskName string, processAt time.Time, payload interface{}, opts ...Option) (*TaskInfo, error)
    
    // Close đóng kết nối của client
    Close() error
}
```

## Khởi tạo Client

### 1. Qua Manager (Recommended)

```go
import "go.fork.vn/queue"

func main() {
    // Tạo config
    config := queue.DefaultConfig()
    
    // Tạo manager
    manager := queue.NewManager(config)
    
    // Lấy client từ manager
    client := manager.Client()
    
    // Sử dụng client...
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
    })
    
    // Tạo queue client
    client := queue.NewClient(redisClient)
    
    // Sử dụng client...
    defer client.Close()
}
```

### 3. Với Universal Redis Client

```go
import (
    "github.com/redis/go-redis/v9"
    "go.fork.vn/queue"
)

func main() {
    // Universal client (hỗ trợ cả single và cluster)
    universalClient := redis.NewUniversalClient(&redis.UniversalOptions{
        Addrs: []string{"localhost:6379"},
    })
    
    client := queue.NewClientWithUniversalClient(universalClient)
    defer client.Close()
}
```

### 4. Với Custom Adapter

```go
import (
    "go.fork.vn/queue"
    "go.fork.vn/queue/adapter"
)

func main() {
    // Sử dụng memory adapter
    memoryAdapter := adapter.NewMemoryQueue("queue:")
    client := queue.NewClientWithAdapter(memoryAdapter)
    
    // Hoặc custom adapter
    // customAdapter := &MyCustomAdapter{}
    // client := queue.NewClientWithAdapter(customAdapter)
}
```

## Enqueue Methods

### 1. Enqueue - Xử lý ngay lập tức

```go
// Enqueue cơ bản
taskInfo, err := client.Enqueue("send_email", map[string]string{
    "to":      "user@example.com",
    "subject": "Welcome!",
    "body":    "Thanks for signing up",
})

if err != nil {
    log.Fatal(err)
}

fmt.Printf("Task ID: %s, Queue: %s\n", taskInfo.ID, taskInfo.Queue)
```

### 2. EnqueueContext - Với context cancellation

```go
import (
    "context"
    "time"
)

func main() {
    // Context với timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    taskInfo, err := client.EnqueueContext(ctx, "process_data", userData)
    if err != nil {
        if err == context.DeadlineExceeded {
            log.Println("Enqueue timeout")
        } else {
            log.Printf("Enqueue error: %v", err)
        }
        return
    }
    
    log.Printf("Task enqueued: %s", taskInfo.ID)
}
```

### 3. EnqueueIn - Delay execution

```go
import "time"

// Gửi email reminder sau 1 giờ
taskInfo, err := client.EnqueueIn(
    "send_reminder",
    time.Hour,
    map[string]interface{}{
        "user_id": 123,
        "message": "Don't forget to complete your profile",
    },
)

// Gửi báo cáo hàng tuần vào chủ nhật
taskInfo, err := client.EnqueueIn(
    "weekly_report", 
    7*24*time.Hour,  // 7 ngày
    reportData,
)
```

### 4. EnqueueAt - Schedule cụ thể

```go
import "time"

// Gửi vào 9h sáng mai
tomorrow := time.Now().Add(24 * time.Hour)
scheduleTime := time.Date(
    tomorrow.Year(),
    tomorrow.Month(), 
    tomorrow.Day(),
    9, 0, 0, 0,  // 9:00:00
    tomorrow.Location(),
)

taskInfo, err := client.EnqueueAt("daily_report", scheduleTime, reportData)

// Gửi birthday email vào đúng ngày sinh nhật
birthday := time.Date(2024, 12, 25, 10, 0, 0, 0, time.UTC)
taskInfo, err := client.EnqueueAt("birthday_email", birthday, userBirthdayData)
```

## Task Options

### Option Types

```go
// Định nghĩa trong client.go
type Option func(*TaskOptions)

type TaskOptions struct {
    Queue    string
    MaxRetry int
    Timeout  time.Duration
}
```

### Available Options

#### WithQueue - Chỉ định queue

```go
// Gửi tác vụ quan trọng vào queue ưu tiên cao
client.Enqueue("urgent_task", data, queue.WithQueue("critical"))

// Gửi tác vụ batch processing vào queue riêng
client.Enqueue("batch_process", data, queue.WithQueue("batch"))

// Gửi email vào queue chuyên dụng
client.Enqueue("send_email", emailData, queue.WithQueue("emails"))
```

#### WithMaxRetry - Số lần retry

```go
// Task quan trọng, retry nhiều lần
client.Enqueue("payment_process", paymentData, 
    queue.WithMaxRetry(10))

// Task không quan trọng, không retry
client.Enqueue("analytics_track", trackData, 
    queue.WithMaxRetry(0))

// Task thông thường
client.Enqueue("user_notification", notifData, 
    queue.WithMaxRetry(3))
```

#### WithTimeout - Timeout cho task

```go
// Task xử lý nhanh
client.Enqueue("quick_validation", data, 
    queue.WithTimeout(30*time.Second))

// Task xử lý lâu
client.Enqueue("video_processing", videoData, 
    queue.WithTimeout(30*time.Minute))

// Task batch processing
client.Enqueue("data_migration", migrationData, 
    queue.WithTimeout(2*time.Hour))
```

#### Kết hợp multiple options

```go
// Task ưu tiên cao, retry nhiều, timeout dài
client.Enqueue("critical_payment", paymentData,
    queue.WithQueue("critical"),
    queue.WithMaxRetry(5),
    queue.WithTimeout(10*time.Minute),
)
```

## Payload Handling

### 1. JSON Serializable Data

```go
// Struct
type UserData struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

user := UserData{ID: 123, Name: "John", Email: "john@example.com"}
client.Enqueue("process_user", user)

// Map
data := map[string]interface{}{
    "action": "send_welcome_email",
    "user_id": 123,
    "template": "welcome_new_user",
}
client.Enqueue("email_task", data)

// Slice
items := []string{"item1", "item2", "item3"}
client.Enqueue("process_items", items)
```

### 2. Complex Payloads

```go
type ComplexPayload struct {
    UserID      int                    `json:"user_id"`
    Action      string                 `json:"action"`
    Metadata    map[string]interface{} `json:"metadata"`
    Files       []FileInfo             `json:"files"`
    Settings    *UserSettings          `json:"settings,omitempty"`
    ProcessedAt time.Time             `json:"processed_at"`
}

type FileInfo struct {
    Path string `json:"path"`
    Size int64  `json:"size"`
    Type string `json:"type"`
}

type UserSettings struct {
    Theme       string `json:"theme"`
    Notifications bool `json:"notifications"`
}

payload := ComplexPayload{
    UserID: 123,
    Action: "process_files",
    Metadata: map[string]interface{}{
        "source": "upload",
        "batch_id": "batch_001",
    },
    Files: []FileInfo{
        {Path: "/uploads/file1.jpg", Size: 1024000, Type: "image"},
        {Path: "/uploads/file2.pdf", Size: 2048000, Type: "document"},
    },
    Settings: &UserSettings{
        Theme: "dark",
        Notifications: true,
    },
    ProcessedAt: time.Now(),
}

client.Enqueue("file_processor", payload)
```

### 3. Binary Data

```go
// Để gửi binary data, encode thành base64
import "encoding/base64"

binaryData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
encodedData := base64.StdEncoding.EncodeToString(binaryData)

payload := map[string]string{
    "type": "image",
    "data": encodedData,
}

client.Enqueue("process_binary", payload)
```

## Error Handling

### 1. Enqueue Errors

```go
taskInfo, err := client.Enqueue("task_name", payload)
if err != nil {
    switch {
    case errors.Is(err, queue.ErrInvalidPayload):
        log.Printf("Invalid payload: %v", err)
        
    case errors.Is(err, queue.ErrQueueUnavailable):
        log.Printf("Queue unavailable: %v", err)
        // Có thể retry hoặc fallback
        
    case errors.Is(err, queue.ErrConnectionLost):
        log.Printf("Connection lost: %v", err)
        // Reconnect logic
        
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}

log.Printf("Task enqueued successfully: %s", taskInfo.ID)
```

### 2. Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

// Goroutine có thể cancel operation
go func() {
    time.Sleep(2 * time.Second)
    cancel()
}()

taskInfo, err := client.EnqueueContext(ctx, "long_task", data)
if err != nil {
    if err == context.Canceled {
        log.Println("Operation was cancelled")
    } else {
        log.Printf("Error: %v", err)
    }
}
```

### 3. Retry Logic

```go
func enqueueWithRetry(client queue.Client, taskName string, payload interface{}, maxRetries int) (*queue.TaskInfo, error) {
    var lastErr error
    
    for i := 0; i < maxRetries; i++ {
        taskInfo, err := client.Enqueue(taskName, payload)
        if err == nil {
            return taskInfo, nil
        }
        
        lastErr = err
        
        // Exponential backoff
        backoff := time.Duration(i*i) * time.Second
        time.Sleep(backoff)
    }
    
    return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

## Best Practices

### 1. Client Lifecycle

```go
type Service struct {
    client queue.Client
}

func NewService(manager queue.Manager) *Service {
    return &Service{
        client: manager.Client(),
    }
}

func (s *Service) ProcessUser(user User) error {
    // Sử dụng client
    _, err := s.client.Enqueue("process_user", user)
    return err
}

func (s *Service) Close() error {
    return s.client.Close()
}
```

### 2. Dependency Injection

```go
// Với DI container
func registerServices(container di.Container) {
    container.Singleton("queue.client", func(manager queue.Manager) queue.Client {
        return manager.Client()
    })
    
    container.Singleton("user.service", func(client queue.Client) *UserService {
        return NewUserService(client)
    })
}
```

### 3. Configuration-based Queues

```go
type QueueConfig struct {
    UserTasks     string `mapstructure:"user_tasks"`
    EmailTasks    string `mapstructure:"email_tasks"`
    BatchTasks    string `mapstructure:"batch_tasks"`
    CriticalTasks string `mapstructure:"critical_tasks"`
}

type TaskService struct {
    client queue.Client
    config QueueConfig
}

func (s *TaskService) ProcessUser(user User) error {
    return s.enqueue(s.config.UserTasks, "process_user", user)
}

func (s *TaskService) SendEmail(email EmailData) error {
    return s.enqueue(s.config.EmailTasks, "send_email", email)
}

func (s *TaskService) enqueue(queueName, taskName string, payload interface{}) error {
    _, err := s.client.Enqueue(taskName, payload, queue.WithQueue(queueName))
    return err
}
```

### 4. Monitoring và Metrics

```go
import (
    "go.fork.vn/queue"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    enqueueCounter = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "queue_enqueue_total",
            Help: "Total number of enqueued tasks",
        },
        []string{"task_name", "queue", "status"},
    )
)

func (s *TaskService) enqueueWithMetrics(taskName string, payload interface{}, opts ...queue.Option) (*queue.TaskInfo, error) {
    taskInfo, err := s.client.Enqueue(taskName, payload, opts...)
    
    status := "success"
    if err != nil {
        status = "error"
    }
    
    queueName := "default"  // Extract từ opts hoặc config
    enqueueCounter.WithLabelValues(taskName, queueName, status).Inc()
    
    return taskInfo, err
}
```

### 5. Testing

```go
func TestTaskService(t *testing.T) {
    // Sử dụng memory adapter cho testing
    memoryAdapter := adapter.NewMemoryQueue("test:")
    client := queue.NewClientWithAdapter(memoryAdapter)
    
    service := NewTaskService(client)
    
    // Test enqueue
    err := service.ProcessUser(testUser)
    assert.NoError(t, err)
    
    // Verify task được enqueue
    assert.False(t, memoryAdapter.IsEmpty(context.Background(), "default"))
}
```
