# Task - Quản lý Tác vụ

## Giới thiệu

Module `Task` là thành phần cốt lõi của Go Queue library, định nghĩa cấu trúc và hành vi của các tác vụ trong hệ thống hàng đợi. Nó cung cấp các struct, method và option để tạo, cấu hình và quản lý tác vụ một cách linh hoạt.

## Cấu trúc Task

### Task struct

```go
type Task struct {
    ID         string    // Định danh duy nhất của tác vụ
    Name       string    // Tên loại tác vụ
    Payload    []byte    // Dữ liệu tác vụ dưới dạng bytes
    Queue      string    // Tên hàng đợi chứa tác vụ
    MaxRetry   int       // Số lần thử lại tối đa
    RetryCount int       // Số lần đã thử lại
    CreatedAt  time.Time // Thời điểm tạo tác vụ
    ProcessAt  time.Time // Thời điểm xử lý tác vụ
}
```

**Tham chiếu:** `task.go` dòng 9-28

### TaskInfo struct

```go
type TaskInfo struct {
    ID        string    // Định danh duy nhất
    Name      string    // Tên loại tác vụ
    Queue     string    // Tên hàng đợi
    MaxRetry  int       // Số lần thử lại tối đa
    State     string    // Trạng thái tác vụ
    CreatedAt time.Time // Thời điểm tạo
    ProcessAt time.Time // Thời điểm xử lý
}
```

**Tham chiếu:** `task.go` dòng 34-49

## Tạo Task

### Tạo task cơ bản

```go
package main

import (
    "encoding/json"
    "github.com/go-fork/queue"
)

type EmailData struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func main() {
    // Tạo dữ liệu cho task
    emailData := EmailData{
        To:      "user@example.com",
        Subject: "Chào mừng!",
        Body:    "Chào mừng bạn đến với dịch vụ của chúng tôi",
    }
    
    payload, _ := json.Marshal(emailData)
    
    // Tạo task mới
    task := queue.NewTask("send_email", payload)
    
    fmt.Println("Task created:", task.Name)
}
```

**Tham chiếu:** `task.go` dòng 51-56

### Giải mã dữ liệu task

```go
func processEmailTask(task *queue.Task) error {
    var emailData EmailData
    
    // Giải mã payload thành struct
    if err := task.Unmarshal(&emailData); err != nil {
        return fmt.Errorf("failed to unmarshal task data: %w", err)
    }
    
    // Xử lý email
    fmt.Printf("Sending email to: %s\n", emailData.To)
    fmt.Printf("Subject: %s\n", emailData.Subject)
    
    return nil
}
```

**Tham chiếu:** `task.go` dòng 30-32

## Task Options

### TaskOptions struct

```go
type TaskOptions struct {
    Queue     string        // Tên hàng đợi
    MaxRetry  int          // Số lần thử lại tối đa
    Timeout   time.Duration // Thời gian timeout
    Deadline  time.Time     // Thời hạn hoàn thành
    Delay     time.Duration // Thời gian trì hoãn
    ProcessAt time.Time     // Thời điểm xử lý cụ thể
    TaskID    string        // ID tùy chỉnh
}
```

**Tham chiếu:** `task.go` dòng 73-87

### Các Option Functions

#### WithQueue - Chỉ định hàng đợi

```go
// Đưa task vào hàng đợi cụ thể
client.Enqueue("send_email", payload, 
    queue.WithQueue("email_queue"))
```

**Tham chiếu:** `task.go` dòng 89-94

#### WithMaxRetry - Cấu hình retry

```go
// Cho phép thử lại tối đa 5 lần
client.Enqueue("process_payment", payload,
    queue.WithMaxRetry(5))
```

**Tham chiếu:** `task.go` dòng 96-101

#### WithTimeout - Đặt timeout

```go
// Task phải hoàn thành trong 10 phút
client.Enqueue("heavy_task", payload,
    queue.WithTimeout(10 * time.Minute))
```

**Tham chiếu:** `task.go` dòng 103-108

#### WithDelay - Trì hoãn xử lý

```go
// Trì hoãn xử lý 1 giờ
client.Enqueue("reminder_email", payload,
    queue.WithDelay(1 * time.Hour))
```

**Tham chiếu:** `task.go` dòng 117-122

#### WithProcessAt - Lên lịch cụ thể

```go
// Xử lý vào lúc 9:00 sáng ngày mai
tomorrow9AM := time.Now().Add(24 * time.Hour).Truncate(time.Hour).Add(9 * time.Hour)
client.Enqueue("daily_report", payload,
    queue.WithProcessAt(tomorrow9AM))
```

**Tham chiếu:** `task.go` dòng 124-129

#### WithTaskID - ID tùy chỉnh

```go
// Sử dụng ID tùy chỉnh để tránh trùng lặp
userID := "user123"
client.Enqueue("user_notification", payload,
    queue.WithTaskID(fmt.Sprintf("notify_%s", userID)))
```

**Tham chiếu:** `task.go` dòng 131-136

## Kết hợp Options

### Ví dụ phức tạp

```go
func scheduleMaintenanceTask() {
    // Dữ liệu maintenance
    maintenanceData := MaintenanceJob{
        Type:        "database_cleanup",
        Database:    "production",
        TablePrefix: "temp_",
    }
    
    payload, _ := json.Marshal(maintenanceData)
    
    // Lên lịch maintenance vào 2:00 AM với nhiều option
    maintenanceTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour).Add(2 * time.Hour)
    
    info, err := client.Enqueue("maintenance", payload,
        queue.WithQueue("maintenance_queue"),    // Hàng đợi chuyên dụng
        queue.WithProcessAt(maintenanceTime),    // Thời gian cụ thể
        queue.WithMaxRetry(1),                   // Chỉ thử lại 1 lần
        queue.WithTimeout(2 * time.Hour),        // Timeout 2 giờ
        queue.WithTaskID("daily_maintenance"),   // ID duy nhất
    )
    
    if err != nil {
        log.Printf("Failed to schedule maintenance: %v", err)
        return
    }
    
    log.Printf("Maintenance scheduled: %s", info.String())
}
```

## Xử lý Task Info

### In thông tin task

```go
func printTaskInfo(info *queue.TaskInfo) {
    fmt.Printf("Task Information:\n")
    fmt.Printf("  ID: %s\n", info.ID)
    fmt.Printf("  Name: %s\n", info.Name)
    fmt.Printf("  Queue: %s\n", info.Queue)
    fmt.Printf("  State: %s\n", info.State)
    fmt.Printf("  Max Retry: %d\n", info.MaxRetry)
    fmt.Printf("  Created: %s\n", info.CreatedAt.Format(time.RFC3339))
    fmt.Printf("  Process At: %s\n", info.ProcessAt.Format(time.RFC3339))
    
    // Hoặc sử dụng String() method
    fmt.Println(info.String())
}
```

**Tham chiếu:** `task.go` dòng 58-60

## Best Practices

### 1. Đặt tên Task có ý nghĩa

```go
// Tốt - tên mô tả rõ chức năng
client.Enqueue("send_welcome_email", payload)
client.Enqueue("process_payment_refund", payload)
client.Enqueue("generate_monthly_report", payload)

// Tránh - tên chung chung
client.Enqueue("task1", payload)
client.Enqueue("process", payload)
```

### 2. Sử dụng struct cho Payload phức tạp

```go
// Định nghĩa struct rõ ràng
type OrderProcessingData struct {
    OrderID     string    `json:"order_id"`
    CustomerID  string    `json:"customer_id"`
    Items       []Item    `json:"items"`
    TotalAmount float64   `json:"total_amount"`
    CreatedAt   time.Time `json:"created_at"`
}

// Tạo và marshal payload
orderData := OrderProcessingData{
    OrderID:     "ORD-12345",
    CustomerID:  "CUST-67890",
    Items:       items,
    TotalAmount: 199.99,
    CreatedAt:   time.Now(),
}

payload, err := json.Marshal(orderData)
if err != nil {
    return fmt.Errorf("failed to marshal order data: %w", err)
}
```

### 3. Cấu hình Retry hợp lý

```go
// Task quan trọng - nhiều lần retry
client.Enqueue("process_payment", payload,
    queue.WithMaxRetry(5),
    queue.WithQueue("critical"))

// Task không quan trọng - ít retry
client.Enqueue("send_newsletter", payload,
    queue.WithMaxRetry(1),
    queue.WithQueue("notifications"))

// Task một lần - không retry
client.Enqueue("generate_one_time_token", payload,
    queue.WithMaxRetry(0))
```

### 4. Sử dụng Queue phù hợp

```go
// Phân loại theo priority
client.Enqueue("urgent_alert", payload,
    queue.WithQueue("high_priority"))

client.Enqueue("weekly_cleanup", payload,
    queue.WithQueue("low_priority"))

// Phân loại theo chức năng
client.Enqueue("send_email", payload,
    queue.WithQueue("email_queue"))

client.Enqueue("resize_image", payload,
    queue.WithQueue("media_processing"))
```

### 5. Quản lý thời gian hiệu quả

```go
// Lên lịch hàng ngày
dailyTime := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour).Add(6 * time.Hour) // 6 AM tomorrow

// Lên lịch hàng tuần
weeklyTime := time.Now().AddDate(0, 0, 7-int(time.Now().Weekday())) // Next Sunday

// Trì hoãn ngắn hạn
client.Enqueue("retry_failed_request", payload,
    queue.WithDelay(5 * time.Minute))

// Timeout phù hợp với task
client.Enqueue("quick_validation", payload,
    queue.WithTimeout(30 * time.Second))

client.Enqueue("batch_processing", payload,
    queue.WithTimeout(1 * time.Hour))
```

## Error Handling

### Xử lý lỗi khi tạo task

```go
func createTaskSafely(name string, data interface{}) (*queue.TaskInfo, error) {
    // Marshal payload
    payload, err := json.Marshal(data)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal task data: %w", err)
    }
    
    // Validate task name
    if name == "" {
        return nil, fmt.Errorf("task name cannot be empty")
    }
    
    // Enqueue với error handling
    info, err := client.Enqueue(name, payload,
        queue.WithMaxRetry(3),
        queue.WithTimeout(30 * time.Minute))
    
    if err != nil {
        return nil, fmt.Errorf("failed to enqueue task %s: %w", name, err)
    }
    
    return info, nil
}
```

### Xử lý lỗi khi unmarshal

```go
func processTask(task *queue.Task) error {
    switch task.Name {
    case "send_email":
        var emailData EmailData
        if err := task.Unmarshal(&emailData); err != nil {
            return fmt.Errorf("invalid email data: %w", err)
        }
        return processEmail(emailData)
        
    case "process_payment":
        var paymentData PaymentData
        if err := task.Unmarshal(&paymentData); err != nil {
            return fmt.Errorf("invalid payment data: %w", err)
        }
        return processPayment(paymentData)
        
    default:
        return fmt.Errorf("unknown task type: %s", task.Name)
    }
}
```

## Testing

### Unit test cho task creation

```go
func TestNewTask(t *testing.T) {
    // Test data
    name := "test_task"
    payload := []byte(`{"message": "hello"}`)
    
    // Tạo task
    task := queue.NewTask(name, payload)
    
    // Assertions
    assert.Equal(t, name, task.Name)
    assert.Equal(t, payload, task.Payload)
    assert.NotZero(t, task.CreatedAt)
    assert.NotZero(t, task.ProcessAt)
}

func TestTaskUnmarshal(t *testing.T) {
    // Test data
    testData := struct {
        Message string `json:"message"`
        Count   int    `json:"count"`
    }{
        Message: "test",
        Count:   42,
    }
    
    payload, _ := json.Marshal(testData)
    task := queue.NewTask("test", payload)
    
    // Unmarshal
    var result struct {
        Message string `json:"message"`
        Count   int    `json:"count"`
    }
    
    err := task.Unmarshal(&result)
    
    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, testData.Message, result.Message)
    assert.Equal(t, testData.Count, result.Count)
}
```

### Integration test với options

```go
func TestTaskOptions(t *testing.T) {
    // Setup client
    client := setupTestClient(t)
    
    testCases := []struct {
        name    string
        options []queue.Option
        want    func(*queue.TaskInfo) bool
    }{
        {
            name:    "with custom queue",
            options: []queue.Option{queue.WithQueue("test_queue")},
            want:    func(info *queue.TaskInfo) bool { return info.Queue == "test_queue" },
        },
        {
            name:    "with max retry",
            options: []queue.Option{queue.WithMaxRetry(5)},
            want:    func(info *queue.TaskInfo) bool { return info.MaxRetry == 5 },
        },
        {
            name: "with delay",
            options: []queue.Option{queue.WithDelay(1 * time.Hour)},
            want: func(info *queue.TaskInfo) bool {
                return info.ProcessAt.After(time.Now().Add(59 * time.Minute))
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            info, err := client.Enqueue("test_task", []byte("{}"), tc.options...)
            
            assert.NoError(t, err)
            assert.True(t, tc.want(info))
        })
    }
}
```

## Performance Tips

### 1. Tối ưu payload size

```go
// Tránh payload quá lớn
type EfficientTaskData struct {
    ID          string `json:"id"`           // Chỉ lưu ID
    Type        string `json:"type"`         // Loại processing
    ConfigHash  string `json:"config_hash"`  // Hash của config thay vì toàn bộ config
}

// Thay vì lưu toàn bộ data, chỉ lưu reference
type HeavyTaskData struct {
    DataURL     string `json:"data_url"`     // URL để tải data
    ProcessType string `json:"process_type"`
    Checksum    string `json:"checksum"`
}
```

### 2. Batch processing

```go
// Xử lý nhiều item cùng lúc
type BatchTaskData struct {
    BatchID string   `json:"batch_id"`
    ItemIDs []string `json:"item_ids"`
    Type    string   `json:"type"`
}

func createBatchTask(items []Item) error {
    // Chia thành các batch nhỏ
    batchSize := 100
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        batchData := BatchTaskData{
            BatchID: fmt.Sprintf("batch_%d_%d", i, end-1),
            ItemIDs: extractIDs(batch),
            Type:    "batch_process",
        }
        
        payload, _ := json.Marshal(batchData)
        client.Enqueue("batch_processor", payload)
    }
    
    return nil
}
```

## Monitoring

### Task metrics

```go
type TaskMetrics struct {
    mu            sync.RWMutex
    tasksCreated  int64
    tasksEnqueued int64
    tasksFailed   int64
}

func (m *TaskMetrics) RecordTaskCreated(taskName string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasksCreated++
    
    // Log to monitoring system
    log.Printf("Task created: %s, total: %d", taskName, m.tasksCreated)
}

func (m *TaskMetrics) RecordTaskEnqueued(info *queue.TaskInfo) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.tasksEnqueued++
    
    // Send metrics to monitoring
    metrics.Counter("queue.tasks.enqueued").
        WithTag("queue", info.Queue).
        WithTag("task_name", info.Name).
        Increment()
}
```

## Kết luận

Module Task cung cấp một API linh hoạt và mạnh mẽ để quản lý tác vụ trong Go Queue library. Với các option functions và cấu trúc dữ liệu được thiết kế tốt, bạn có thể dễ dàng tạo, cấu hình và theo dõi các tác vụ theo nhu cầu cụ thể của ứng dụng.
