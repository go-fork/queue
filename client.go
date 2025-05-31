package queue

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.fork.vn/queue/adapter"
)

// Client là interface cho việc đưa tác vụ vào hàng đợi.
type Client interface {
	// Enqueue đưa một tác vụ vào hàng đợi để xử lý ngay lập tức.
	Enqueue(taskName string, payload interface{}, opts ...Option) (*TaskInfo, error)

	// EnqueueContext tương tự Enqueue nhưng với context.
	EnqueueContext(ctx context.Context, taskName string, payload interface{}, opts ...Option) (*TaskInfo, error)

	// EnqueueIn đưa một tác vụ vào hàng đợi để xử lý sau một khoảng thời gian.
	EnqueueIn(taskName string, delay time.Duration, payload interface{}, opts ...Option) (*TaskInfo, error)

	// EnqueueAt đưa một tác vụ vào hàng đợi để xử lý vào một thời điểm cụ thể.
	EnqueueAt(taskName string, processAt time.Time, payload interface{}, opts ...Option) (*TaskInfo, error)

	// Close đóng kết nối của client.
	Close() error
}

// client triển khai interface Client.
type client struct {
	queue       adapter.QueueAdapter
	defaultOpts *TaskOptions
}

// NewClient tạo một Client mới với Redis.
func NewClient(redisClient *redis.Client) Client {
	queue := adapter.NewRedisQueue(redisClient, "queue:")
	return &client{
		queue:       queue,
		defaultOpts: GetDefaultOptions(),
	}
}

// NewClientWithAdapter tạo một Client mới với adapter QueueAdapter.
func NewClientWithAdapter(adapter adapter.QueueAdapter) Client {
	return &client{
		queue:       adapter,
		defaultOpts: GetDefaultOptions(),
	}
}

// NewClientWithUniversalClient tạo một Client mới với Redis UniversalClient.
func NewClientWithUniversalClient(redisClient redis.UniversalClient) Client {
	// Khởi tạo với client thông thường
	redisStdClient, ok := redisClient.(*redis.Client)
	if !ok {
		// Fallback cho các trường hợp client không phải là *redis.Client
		redisStdClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	}

	queue := adapter.NewRedisQueue(redisStdClient, "queue:")
	return &client{
		queue:       queue,
		defaultOpts: GetDefaultOptions(),
	}
}

// NewMemoryClient tạo một Client mới với bộ nhớ trong.
func NewMemoryClient() Client {
	queue := adapter.NewMemoryQueue("queue:")
	return &client{
		queue:       queue,
		defaultOpts: GetDefaultOptions(),
	}
}

// Enqueue đưa một tác vụ vào hàng đợi để xử lý ngay lập tức.
func (c *client) Enqueue(taskName string, payload interface{}, opts ...Option) (*TaskInfo, error) {
	return c.EnqueueContext(context.Background(), taskName, payload, opts...)
}

// EnqueueContext tương tự Enqueue nhưng với context.
func (c *client) EnqueueContext(ctx context.Context, taskName string, payload interface{}, opts ...Option) (*TaskInfo, error) {
	options := ApplyOptions(opts...)

	task := &Task{
		ID:         generateID(),
		Name:       taskName,
		Queue:      options.Queue,
		MaxRetry:   options.MaxRetry,
		CreatedAt:  time.Now(),
		ProcessAt:  time.Now(),
		RetryCount: 0,
	}

	// Nếu có ID tùy chỉnh
	if options.TaskID != "" {
		task.ID = options.TaskID
	}

	// Tạo payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}
	task.Payload = payloadBytes

	// Đưa vào hàng đợi
	queueName := fmt.Sprintf("%s:pending", task.Queue)
	if err := c.queue.Enqueue(ctx, queueName, task); err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Lưu thông tin thời gian xử lý nếu khác thời điểm hiện tại
	if !options.ProcessAt.IsZero() && options.ProcessAt.After(time.Now()) {
		scheduledQueue := fmt.Sprintf("%s:scheduled", task.Queue)
		scheduledTask := &scheduledTask{
			TaskID:    task.ID,
			ProcessAt: options.ProcessAt,
		}
		if err := c.queue.Enqueue(ctx, scheduledQueue, scheduledTask); err != nil {
			return nil, fmt.Errorf("failed to schedule task: %w", err)
		}
	}

	var processTime time.Time
	if options.ProcessAt.IsZero() {
		processTime = task.CreatedAt
	} else {
		processTime = options.ProcessAt
	}

	return &TaskInfo{
		ID:        task.ID,
		Name:      task.Name,
		Queue:     task.Queue,
		MaxRetry:  task.MaxRetry,
		State:     "pending",
		CreatedAt: task.CreatedAt,
		ProcessAt: processTime,
	}, nil
}

// EnqueueIn đưa một tác vụ vào hàng đợi để xử lý sau một khoảng thời gian.
func (c *client) EnqueueIn(taskName string, delay time.Duration, payload interface{}, opts ...Option) (*TaskInfo, error) {
	processAt := time.Now().Add(delay)
	opts = append(opts, WithProcessAt(processAt))
	return c.EnqueueContext(context.Background(), taskName, payload, opts...)
}

// EnqueueAt đưa một tác vụ vào hàng đợi để xử lý vào một thời điểm cụ thể.
func (c *client) EnqueueAt(taskName string, processAt time.Time, payload interface{}, opts ...Option) (*TaskInfo, error) {
	opts = append(opts, WithProcessAt(processAt))
	return c.EnqueueContext(context.Background(), taskName, payload, opts...)
}

// Close đóng kết nối của client.
func (c *client) Close() error {
	// Không có gì để đóng với adapter hiện tại
	return nil
}

// scheduledTask đại diện cho một tác vụ đã được lên lịch.
type scheduledTask struct {
	TaskID    string    `json:"task_id"`
	ProcessAt time.Time `json:"process_at"`
}

// generateID tạo một ID ngẫu nhiên cho tác vụ.
func generateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if random generation fails
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}
