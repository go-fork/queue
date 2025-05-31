package queue

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.fork.vn/config"
	"go.fork.vn/di"
	"go.fork.vn/scheduler"
)

type ServiceProvider interface {
	di.ServiceProvider

	setupQueueScheduledTasks(schedulerManager scheduler.Manager, container *di.Container)

	cleanupFailedJobs(manager Manager, container *di.Container)

	retryFailedJobs(manager Manager, container *di.Container)
}

// serviceProvider triển khai interface di.serviceProvider cho các dịch vụ queue.
//
// Provider này tự động hóa việc đăng ký các dịch vụ queue trong một container
// dependency injection, thiết lập client và server với các giá trị mặc định hợp lý.
type serviceProvider struct {
	providers []string
}

// NewServiceProvider tạo một provider dịch vụ queue mới với cấu hình mặc định.
//
// Sử dụng hàm này để tạo một provider có thể được đăng ký với
// một instance di.Container.
//
// Trả về:
//   - di.ServiceProvider: một service provider cho queue
//
// Ví dụ:
//
//	app := myapp.New()
//	app.Register(queue.NewServiceProvider())
func NewServiceProvider() di.ServiceProvider {
	return &serviceProvider{
		providers: []string{"queue"},
	}
}

// Register đăng ký các dịch vụ queue với container của ứng dụng.
//
// Phương thức này:
//   - Tạo một queue manager
//   - Đăng ký client, server và các thành phần cần thiết khác
//   - Đăng ký tất cả vào container DI
//
// Tham số:
//   - app: interface{} - instance của ứng dụng cung cấp Container()
//
// Ứng dụng phải triển khai:
//   - Container() *di.Container
func (p *serviceProvider) Register(app interface{}) {
	// Trích xuất container từ ứng dụng
	if appWithContainer, ok := app.(interface {
		Container() *di.Container
	}); ok {
		c := appWithContainer.Container()
		if c == nil {
			return // Không xử lý nếu container nil
		}

		// Đăng ký scheduler service provider trước nếu chưa có
		if _, err := c.Make("scheduler"); err != nil {
			schedulerProvider := scheduler.NewServiceProvider()
			schedulerProvider.Register(app)
		}

		// Kiểm tra xem container đã có config manager chưa
		configManager := c.MustMake("config").(config.Manager)
		if configManager != nil {
			// Cố gắng load cấu hình từ config manager
			queueConfig := DefaultConfig()
			if err := configManager.UnmarshalKey("queue", &queueConfig); err != nil {
				panic(fmt.Sprintf("Failed to load queue config: %v", err))
			}
			// Tạo một queue manager mới với container để truy cập Redis provider
			manager := NewManagerWithContainer(queueConfig, c)
			c.Instance("queue", manager)         // Dịch vụ queue manager chung
			c.Instance("queue.manager", manager) // Direct instance instead of alias
		} else {
			panic("Config manager is not available in the container")
		}
	}
}

// Boot thực hiện thiết lập sau đăng ký cho dịch vụ queue.
//
// Phương thức này sẽ:
//   - Khởi động scheduler để xử lý các delayed/scheduled tasks
//   - Thiết lập các task định kỳ cho queue maintenance
//
// Tham số:
//   - app: interface{} - instance của ứng dụng
func (p *serviceProvider) Boot(app interface{}) {
	if app == nil {
		return
	}

	// Trích xuất container từ ứng dụng
	if appWithContainer, ok := app.(interface {
		Container() *di.Container
	}); ok {
		c := appWithContainer.Container()
		if c == nil {
			return // Không xử lý nếu container nil
		}

		// Lấy scheduler manager
		schedulerManager := c.MustMake("scheduler").(scheduler.Manager)

		// Lấy queue manager để có thể set scheduler cho server
		queueManager := c.MustMake("queue").(Manager)

		// Set up scheduler trên server nếu có
		if server := queueManager.Server(); server != nil {
			server.SetScheduler(schedulerManager)
		}

		// Thiết lập các task định kỳ cho queue
		p.setupQueueScheduledTasks(schedulerManager, c)

		// Khởi động scheduler nếu chưa chạy
		if !schedulerManager.IsRunning() {
			schedulerManager.StartAsync()
		}
	}
}

// setupQueueScheduledTasks thiết lập các task định kỳ cho queue system
func (p *serviceProvider) setupQueueScheduledTasks(schedulerManager scheduler.Manager, container *di.Container) {
	// Task dọn dẹp failed jobs cũ (chạy mỗi giờ)
	schedulerManager.Every(1).Hours().Do(func() {
		queueManager := container.MustMake("queue").(Manager)
		p.cleanupFailedJobs(queueManager, container)
	})

	// Task retry failed jobs (chạy mỗi 5 phút)
	schedulerManager.Every(5).Minutes().Do(func() {
		queueManager := container.MustMake("queue").(Manager)
		p.retryFailedJobs(queueManager, container)
	})
}

// cleanupFailedJobs dọn dẹp các failed jobs cũ
func (p *serviceProvider) cleanupFailedJobs(manager Manager, container *di.Container) {
	ctx := context.Background()

	// Lấy config manager để lấy danh sách queue
	configManager := container.MustMake("config").(config.Manager)
	var queueConfig Config
	if err := configManager.UnmarshalKey("queue", &queueConfig); err != nil {
		log.Printf("Failed to load queue config for cleanup: %v", err)
		return
	}

	// Lấy danh sách các queue từ cấu hình
	queues := queueConfig.Server.Queues
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	// Xử lý cleanup cho từng queue
	for _, queueName := range queues {
		deadLetterQueueName := fmt.Sprintf("%s:dead", queueName)

		// Đếm số lượng dead letter tasks
		adapter := manager.Adapter("")
		size, err := adapter.Size(ctx, deadLetterQueueName)
		if err != nil {
			log.Printf("Failed to get size of dead letter queue %s: %v", deadLetterQueueName, err)
			continue
		}

		if size == 0 {
			continue // Không có gì để cleanup
		}

		log.Printf("Cleaning up dead letter queue %s with %d tasks", deadLetterQueueName, size)

		// Xóa các dead letter tasks cũ hơn 7 ngày
		cleanupCount := 0
		cutoffTime := time.Now().AddDate(0, 0, -7) // 7 ngày trước

		// Tạo temporary queue để lưu các tasks không bị xóa
		tempQueueName := fmt.Sprintf("%s:dead:temp", queueName)

		// Xử lý từng task trong dead letter queue
		for i := int64(0); i < size; i++ {
			var deadLetterTask DeadLetterTask
			if err := adapter.Dequeue(ctx, deadLetterQueueName, &deadLetterTask); err != nil {
				log.Printf("Failed to dequeue dead letter task: %v", err)
				break
			}

			// Kiểm tra xem task có cũ hơn 7 ngày không
			if deadLetterTask.FailedAt.Before(cutoffTime) {
				// Task cũ, xóa (không đưa vào temp queue)
				cleanupCount++
				log.Printf("Cleaned up old dead letter task %s (failed at %v)", deadLetterTask.Task.ID, deadLetterTask.FailedAt)
			} else {
				// Task còn mới, giữ lại
				if err := adapter.Enqueue(ctx, tempQueueName, &deadLetterTask); err != nil {
					log.Printf("Failed to preserve dead letter task %s: %v", deadLetterTask.Task.ID, err)
				}
			}
		}

		// Xóa dead letter queue cũ và đổi tên temp queue
		if err := adapter.Clear(ctx, deadLetterQueueName); err != nil {
			log.Printf("Failed to clear dead letter queue %s: %v", deadLetterQueueName, err)
		}

		// Di chuyển các task từ temp queue về dead letter queue
		tempSize, _ := adapter.Size(ctx, tempQueueName)
		for i := int64(0); i < tempSize; i++ {
			var deadLetterTask DeadLetterTask
			if err := adapter.Dequeue(ctx, tempQueueName, &deadLetterTask); err != nil {
				break
			}
			adapter.Enqueue(ctx, deadLetterQueueName, &deadLetterTask)
		}

		// Xóa temp queue
		adapter.Clear(ctx, tempQueueName)

		if cleanupCount > 0 {
			log.Printf("Cleaned up %d old dead letter tasks from queue %s", cleanupCount, queueName)
		}
	}
}

// retryFailedJobs thử lại các failed jobs
func (p *serviceProvider) retryFailedJobs(manager Manager, container *di.Container) {
	ctx := context.Background()

	// Lấy config manager để lấy danh sách queue
	configManager := container.MustMake("config").(config.Manager)
	var queueConfig Config
	if err := configManager.UnmarshalKey("queue", &queueConfig); err != nil {
		log.Printf("Failed to load queue config for retry: %v", err)
		return
	}

	// Lấy danh sách các queue từ cấu hình
	queues := queueConfig.Server.Queues
	if len(queues) == 0 {
		queues = []string{"default"}
	}

	// Xử lý retry cho từng queue
	for _, queueName := range queues {
		retryQueueName := fmt.Sprintf("%s:retry", queueName)
		pendingQueueName := fmt.Sprintf("%s:pending", queueName)

		// Lấy adapter để làm việc với queue
		adapter := manager.Adapter("")

		// Đếm số lượng retry tasks
		size, err := adapter.Size(ctx, retryQueueName)
		if err != nil {
			log.Printf("Failed to get size of retry queue %s: %v", retryQueueName, err)
			continue
		}

		if size == 0 {
			continue // Không có task nào cần retry
		}

		log.Printf("Processing retry queue %s with %d tasks", retryQueueName, size)

		// Tạo temporary queue để lưu các tasks chưa đến hạn retry
		tempQueueName := fmt.Sprintf("%s:retry:temp", queueName)
		retryCount := 0

		// Xử lý từng task trong retry queue
		for i := int64(0); i < size; i++ {
			var task Task
			if err := adapter.Dequeue(ctx, retryQueueName, &task); err != nil {
				log.Printf("Failed to dequeue retry task: %v", err)
				break
			}

			// Kiểm tra xem task đã đến hạn retry chưa
			if time.Now().After(task.ProcessAt) || time.Now().Equal(task.ProcessAt) {
				// Task đã đến hạn retry, đưa vào pending queue
				if err := adapter.Enqueue(ctx, pendingQueueName, &task); err != nil {
					log.Printf("Failed to move retry task %s to pending queue: %v", task.ID, err)
					// Nếu không thể enqueue, đưa lại vào retry queue
					adapter.Enqueue(ctx, tempQueueName, &task)
				} else {
					retryCount++
					log.Printf("Moved retry task %s to pending queue %s", task.ID, pendingQueueName)
				}
			} else {
				// Task chưa đến hạn retry, giữ lại trong retry queue
				if err := adapter.Enqueue(ctx, tempQueueName, &task); err != nil {
					log.Printf("Failed to preserve retry task %s: %v", task.ID, err)
				}
			}
		}

		// Xóa retry queue cũ và đổi tên temp queue
		if err := adapter.Clear(ctx, retryQueueName); err != nil {
			log.Printf("Failed to clear retry queue %s: %v", retryQueueName, err)
		}

		// Di chuyển các task từ temp queue về retry queue
		tempSize, _ := adapter.Size(ctx, tempQueueName)
		for i := int64(0); i < tempSize; i++ {
			var task Task
			if err := adapter.Dequeue(ctx, tempQueueName, &task); err != nil {
				break
			}
			adapter.Enqueue(ctx, retryQueueName, &task)
		}

		// Xóa temp queue
		adapter.Clear(ctx, tempQueueName)

		if retryCount > 0 {
			log.Printf("Moved %d retry tasks to pending queue %s", retryCount, queueName)
		}
	}
}

// Requires trả về danh sách các service provider mà provider này phụ thuộc.
//
// Queue provider phụ thuộc vào config, redis và scheduler provider.
//
// Trả về:
//   - []string: danh sách các service provider khác mà provider này yêu cầu
func (p *serviceProvider) Requires() []string {
	return []string{"config", "redis", "scheduler"}
}

// Providers trả về danh sách các dịch vụ được cung cấp bởi provider.
//
// Trả về:
//   - []string: danh sách các khóa dịch vụ mà provider này cung cấp
func (p *serviceProvider) Providers() []string {
	return p.providers
}
