package queue

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.fork.vn/queue/adapter"
	"go.fork.vn/scheduler"
)

// HandlerFunc là một hàm xử lý tác vụ.
type HandlerFunc func(ctx context.Context, task *Task) error

// ServerOptions chứa các tùy chọn cấu hình cho server.
type ServerOptions struct {
	// Concurrency xác định số lượng worker xử lý tác vụ song song.
	Concurrency int

	// PollingInterval xác định thời gian chờ giữa các lần kiểm tra tác vụ (tính bằng mili giây).
	PollingInterval int

	// DefaultQueue xác định tên queue mặc định nếu không có queue nào được chỉ định.
	DefaultQueue string

	// StrictPriority xác định liệu có nên ưu tiên nghiêm ngặt giữa các hàng đợi.
	StrictPriority bool

	// Queues xác định danh sách các queue cần lắng nghe theo thứ tự ưu tiên.
	Queues []string

	// ShutdownTimeout xác định thời gian chờ để các worker hoàn tất tác vụ khi dừng server.
	ShutdownTimeout time.Duration

	// LogLevel xác định mức log.
	LogLevel int

	// RetryLimit xác định số lần thử lại tối đa cho tác vụ bị lỗi.
	RetryLimit int
}

// Server là interface cho việc xử lý tác vụ từ hàng đợi.
type Server interface {
	// RegisterHandler đăng ký một handler cho một loại tác vụ.
	RegisterHandler(taskName string, handler HandlerFunc)

	// RegisterHandlers đăng ký nhiều handler cùng một lúc.
	RegisterHandlers(handlers map[string]HandlerFunc)

	// Start bắt đầu xử lý tác vụ (worker).
	Start() error

	// Stop dừng xử lý tác vụ.
	Stop() error

	// SetScheduler thiết lập scheduler cho server để xử lý delayed tasks.
	SetScheduler(scheduler scheduler.Manager)

	// GetScheduler trả về scheduler hiện tại.
	GetScheduler() scheduler.Manager
}

// queueServer triển khai interface Server.
type queueServer struct {
	queue           adapter.QueueAdapter
	scheduler       scheduler.Manager
	handlers        sync.Map
	started         bool
	stopCh          chan struct{}
	workerDoneCh    chan struct{}
	schedulerDoneCh chan struct{}
	mu              sync.Mutex
	options         ServerOptions
	queues          []string
}

// NewServer tạo một Server mới.
func NewServer(redisClient redis.UniversalClient, opts ServerOptions) Server {
	// Khởi tạo với client thông thường
	redisStdClient, ok := redisClient.(*redis.Client)
	if !ok {
		// Fallback cho các trường hợp client không phải là *redis.Client
		redisStdClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
		})
	}

	queue := adapter.NewRedisQueue(redisStdClient, "queue:")

	// Sử dụng danh sách các queue từ cấu hình
	queues := opts.Queues

	// Nếu không có queue nào được chỉ định, sử dụng queue mặc định
	if len(queues) == 0 {
		defaultQueue := "default"
		if opts.DefaultQueue != "" {
			defaultQueue = opts.DefaultQueue
		}
		queues = append(queues, defaultQueue)
	}

	return &queueServer{
		queue:           queue,
		handlers:        sync.Map{},
		started:         false,
		stopCh:          make(chan struct{}),
		workerDoneCh:    make(chan struct{}),
		schedulerDoneCh: make(chan struct{}),
		options:         opts,
		queues:          queues,
	}
}

// NewServerWithAdapter tạo một Server mới với adapter QueueAdapter được cung cấp.
func NewServerWithAdapter(adapter adapter.QueueAdapter, opts ServerOptions) Server {
	// Sử dụng danh sách các queue từ cấu hình
	queues := opts.Queues

	// Nếu không có queue nào được chỉ định, sử dụng queue mặc định
	if len(queues) == 0 {
		defaultQueue := "default"
		if opts.DefaultQueue != "" {
			defaultQueue = opts.DefaultQueue
		}
		queues = append(queues, defaultQueue)
	}

	return &queueServer{
		queue:           adapter,
		handlers:        sync.Map{},
		started:         false,
		stopCh:          make(chan struct{}),
		workerDoneCh:    make(chan struct{}),
		schedulerDoneCh: make(chan struct{}),
		options:         opts,
		queues:          queues,
	}
}

// RegisterHandler đăng ký một handler cho một loại tác vụ.
func (s *queueServer) RegisterHandler(taskName string, handler HandlerFunc) {
	s.handlers.Store(taskName, handler)
}

// RegisterHandlers đăng ký nhiều handler cùng một lúc.
func (s *queueServer) RegisterHandlers(handlers map[string]HandlerFunc) {
	for taskName, handler := range handlers {
		s.RegisterHandler(taskName, handler)
	}
}

// Start bắt đầu xử lý tác vụ (worker).
func (s *queueServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return fmt.Errorf("server already started")
	}

	s.started = true
	log.Println("Starting queue worker server...")

	// Khởi động workers để xử lý immediate tasks
	s.startWorkers()

	// Thiết lập scheduler để xử lý delayed tasks nếu có
	if s.scheduler != nil {
		s.setupDelayedTaskScheduler()
	}

	log.Printf("Queue worker server started with %d workers", s.options.Concurrency)
	return nil
}

// Stop dừng xử lý tác vụ.
func (s *queueServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return fmt.Errorf("server not started")
	}

	log.Println("Stopping queue worker server...")

	// Gửi tín hiệu stop cho tất cả workers (chỉ close nếu chưa closed)
	select {
	case <-s.stopCh:
		// Channel đã được closed rồi
	default:
		close(s.stopCh)
	}

	// Chờ workers hoàn thành trong thời gian ShutdownTimeout
	done := make(chan struct{})
	go func() {
		// Chờ tất cả workers dừng
		for i := 0; i < s.options.Concurrency; i++ {
			<-s.workerDoneCh
		}
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers stopped gracefully")
	case <-time.After(s.options.ShutdownTimeout):
		log.Println("Shutdown timeout reached, forcing stop")
	}

	// Dừng scheduler nếu có
	if s.scheduler != nil && s.scheduler.IsRunning() {
		s.scheduler.Stop()
	}

	s.started = false
	log.Println("Queue worker server stopped")
	return nil
}

// SetScheduler thiết lập scheduler cho server để xử lý delayed tasks.
func (s *queueServer) SetScheduler(sched scheduler.Manager) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scheduler = sched
}

// GetScheduler trả về scheduler hiện tại.
func (s *queueServer) GetScheduler() scheduler.Manager {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.scheduler
}

// startWorkers khởi động các worker để xử lý immediate tasks
func (s *queueServer) startWorkers() {
	// Khởi tạo worker done channel
	s.workerDoneCh = make(chan struct{})

	for i := 0; i < s.options.Concurrency; i++ {
		go s.workerLoop(i)
	}
}

// setupDelayedTaskScheduler thiết lập scheduler để xử lý delayed tasks
func (s *queueServer) setupDelayedTaskScheduler() {
	if s.scheduler == nil {
		return
	}

	// Thiết lập task định kỳ để kiểm tra delayed tasks (mỗi 30 giây)
	s.scheduler.Every(uint64(30)).Seconds().Do(func() {
		if s.started {
			s.processDelayedTasks()
		}
	})

	// Khởi động scheduler nếu chưa chạy
	if !s.scheduler.IsRunning() {
		s.scheduler.StartAsync()
	}
}

// workerLoop là vòng lặp chính của một worker
func (s *queueServer) workerLoop(workerID int) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d recovered from panic: %v", workerID, r)
		}
		// Thông báo worker đã dừng
		s.workerDoneCh <- struct{}{}
	}()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-s.stopCh:
			log.Printf("Worker %d stopping", workerID)
			return
		default:
			// Thử lấy task từ queue
			if task := s.fetchTask(); task != nil {
				s.processTask(workerID, task)
			} else {
				// Không có task, chờ một chút trước khi thử lại
				time.Sleep(time.Duration(s.options.PollingInterval) * time.Millisecond)
			}
		}
	}
}

// fetchTask lấy task từ queue
func (s *queueServer) fetchTask() *Task {
	ctx := context.Background()

	// Thử lấy task từ các queue theo thứ tự ưu tiên
	for _, queueName := range s.queues {
		var task Task
		// Client enqueues to {queueName}:pending, so we need to dequeue from there
		pendingQueueName := fmt.Sprintf("%s:pending", queueName)
		if err := s.queue.Dequeue(ctx, pendingQueueName, &task); err == nil {
			log.Printf("Successfully dequeued task %s from queue %s", task.ID, pendingQueueName)
			return &task
		} else {
			// Log the error for debugging, but only if it's not "queue is empty"
			if err.Error() != "queue is empty: "+pendingQueueName {
				log.Printf("Error dequeuing from queue %s: %v", pendingQueueName, err)
			}
		}
	}

	return nil
}

// processTask xử lý một task
func (s *queueServer) processTask(workerID int, task *Task) {
	log.Printf("Worker %d processing task: %s (type: %s)", workerID, task.ID, task.Name)

	// Tìm handler cho task
	handlerInterface, exists := s.handlers.Load(task.Name)
	if !exists {
		log.Printf("No handler found for task type: %s", task.Name)
		// Move to dead letter queue since we can't process this task
		s.moveToDeadLetterQueue(task, fmt.Errorf("no handler found for task type: %s", task.Name))
		return
	}

	handler, ok := handlerInterface.(HandlerFunc)
	if !ok {
		log.Printf("Invalid handler type for task: %s", task.Name)
		return
	}

	// Xử lý task với timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	start := time.Now()
	err := handler(ctx, task)
	duration := time.Since(start)

	if err != nil {
		log.Printf("Worker %d failed to process task %s: %v (took %v)", workerID, task.ID, err, duration)
		s.handleFailedTask(task, err)
	} else {
		log.Printf("Worker %d completed task %s successfully (took %v)", workerID, task.ID, duration)
	}
}

// processDelayedTasks xử lý các delayed tasks đã đến hạn
func (s *queueServer) processDelayedTasks() {
	ctx := context.Background()

	// Xử lý các scheduled tasks từ mỗi queue
	for _, queueName := range s.queues {
		scheduledQueueName := fmt.Sprintf("%s:scheduled", queueName)
		pendingQueueName := fmt.Sprintf("%s:pending", queueName)

		// Lấy và xử lý các scheduled tasks đã đến hạn
		for {
			var scheduledTask scheduledTask
			if err := s.queue.Dequeue(ctx, scheduledQueueName, &scheduledTask); err != nil {
				// Không có scheduled task nào hoặc có lỗi, chuyển sang queue tiếp theo
				break
			}

			// Kiểm tra xem task đã đến hạn chưa
			if time.Now().Before(scheduledTask.ProcessAt) {
				// Task chưa đến hạn, đưa lại vào scheduled queue
				s.queue.Enqueue(ctx, scheduledQueueName, &scheduledTask)
				break
			}

			// Task đã đến hạn, tạo task thực sự và đưa vào pending queue
			task := &Task{
				ID:         scheduledTask.TaskID,
				CreatedAt:  time.Now(),
				ProcessAt:  time.Now(),
				RetryCount: 0,
			}

			// Đưa task vào pending queue để worker xử lý
			if err := s.queue.Enqueue(ctx, pendingQueueName, task); err != nil {
				log.Printf("Failed to move scheduled task %s to pending queue: %v", scheduledTask.TaskID, err)
			} else {
				log.Printf("Moved scheduled task %s to pending queue %s", scheduledTask.TaskID, pendingQueueName)
			}
		}
	}
}

// handleFailedTask xử lý task bị lỗi
func (s *queueServer) handleFailedTask(task *Task, err error) {
	ctx := context.Background()

	// Tăng retry count
	task.RetryCount++

	log.Printf("Task %s failed (attempt %d/%d): %v", task.ID, task.RetryCount, task.MaxRetry, err)

	// Kiểm tra xem có thể retry không
	if task.RetryCount < task.MaxRetry {
		// Tính toán thời gian delay cho retry (exponential backoff)
		retryDelay := time.Duration(task.RetryCount*task.RetryCount) * time.Minute
		task.ProcessAt = time.Now().Add(retryDelay)

		// Đưa task vào retry queue để xử lý lại sau
		retryQueueName := fmt.Sprintf("%s:retry", task.Queue)
		if enqueueErr := s.queue.Enqueue(ctx, retryQueueName, task); enqueueErr != nil {
			log.Printf("Failed to enqueue task %s for retry: %v", task.ID, enqueueErr)
			// Nếu không thể enqueue để retry, đưa vào dead letter queue
			s.moveToDeadLetterQueue(task, enqueueErr)
		} else {
			log.Printf("Task %s scheduled for retry %d in %v", task.ID, task.RetryCount, retryDelay)
		}
	} else {
		// Đã vượt quá số lần retry, đưa vào dead letter queue
		log.Printf("Task %s exceeded max retry limit (%d), moving to dead letter queue", task.ID, task.MaxRetry)
		s.moveToDeadLetterQueue(task, err)
	}
}

// moveToDeadLetterQueue đưa task vào dead letter queue
func (s *queueServer) moveToDeadLetterQueue(task *Task, reason error) {
	ctx := context.Background()

	// Tạo dead letter task với thông tin lỗi
	deadLetterTask := &DeadLetterTask{
		Task:     *task,
		Reason:   reason.Error(),
		FailedAt: time.Now(),
	}

	deadLetterQueueName := fmt.Sprintf("%s:dead", task.Queue)
	if err := s.queue.Enqueue(ctx, deadLetterQueueName, deadLetterTask); err != nil {
		log.Printf("Failed to move task %s to dead letter queue: %v", task.ID, err)
	} else {
		log.Printf("Task %s moved to dead letter queue %s", task.ID, deadLetterQueueName)
	}
}

// DeadLetterTask đại diện cho một task trong dead letter queue
type DeadLetterTask struct {
	Task     Task      `json:"task"`
	Reason   string    `json:"reason"`
	FailedAt time.Time `json:"failed_at"`
}
