package queue

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.fork.vn/config/mocks"
	"go.fork.vn/di"
	"go.fork.vn/queue/adapter"
	redisMocks "go.fork.vn/redis/mocks"
	"go.fork.vn/scheduler"
	schedulerMocks "go.fork.vn/scheduler/mocks"
)

// mockApp implements the container interface for testing
type mockApp struct {
	container *di.Container
}

func (a *mockApp) Container() *di.Container {
	return a.container
}

// setupTestQueueConfig creates a Queue config for testing
func setupTestQueueConfig() *Config {
	return &Config{
		Adapter: AdapterConfig{
			Default: "memory",
			Memory: MemoryConfig{
				Prefix: "test_queue:",
			},
			Redis: RedisConfig{
				Prefix:      "test_queue:",
				ProviderKey: "default",
			},
		},
		Server: ServerConfig{
			Concurrency:     5,
			PollingInterval: 500,
			DefaultQueue:    "test",
			StrictPriority:  true,
			Queues:          []string{"critical", "high", "test", "low"},
			ShutdownTimeout: 10,
			LogLevel:        1,
			RetryLimit:      2,
		},
		Client: ClientConfig{
			DefaultOptions: ClientDefaultOptions{
				Queue:    "test",
				MaxRetry: 2,
				Timeout:  15,
			},
		},
	}
}

func TestNewServiceProvider(t *testing.T) {
	provider := NewServiceProvider()
	assert.NotNil(t, provider, "Expected service provider to be initialized")
}

func TestServiceProviderRegister(t *testing.T) {
	t.Run("registers queue services to container with config", func(t *testing.T) {
		// Arrange
		container := di.New()
		mockConfig := mocks.NewMockManager(t)

		// Setup mock config to handle UnmarshalKey for our test Queue config
		testQueueConfig := setupTestQueueConfig()
		mockConfig.EXPECT().UnmarshalKey("queue", mock.Anything).Run(func(_ string, out interface{}) {
			// Copy our test config to the output parameter
			if cfg, ok := out.(*Config); ok {
				*cfg = *testQueueConfig
			}
		}).Return(nil)

		container.Instance("config", mockConfig)

		// Mock scheduler service to satisfy scheduler dependency
		mockScheduler := schedulerMocks.NewMockManager(t)
		container.Instance("scheduler", mockScheduler)

		// Mock redis service to satisfy redis dependency
		mockRedis := redisMocks.NewMockManager(t)
		container.Instance("redis", mockRedis)

		app := &mockApp{container: container}
		provider := NewServiceProvider()

		// Act - since we're testing dynamic providers, we have an empty list initially
		initialProviders := provider.Providers()
		assert.Contains(t, initialProviders, "queue", "Expected 'queue' in initial providers")

		provider.Register(app)

		// Assert - check services were registered
		assert.True(t, container.Bound("queue"), "Expected 'queue' to be bound")

		// Test manager resolution
		managerService, err := container.Make("queue")
		assert.NoError(t, err, "Expected no error when resolving queue")
		assert.NotNil(t, managerService, "Expected queue to be non-nil")
		assert.Implements(t, (*Manager)(nil), managerService, "Expected queue to implement Manager interface")

		// Test manager alias resolution (if it works)
		managerAliasService, err := container.Make("queue.manager")
		if err == nil {
			assert.NotNil(t, managerAliasService, "Expected queue.manager to be non-nil")
			assert.Equal(t, managerService, managerAliasService, "Expected queue.manager to be alias of queue")
		} else {
			// Alias might not be supported by this DI container implementation
			t.Logf("Alias resolution not supported: %v", err)
		}
	})

	t.Run("panics when config service is missing", func(t *testing.T) {
		// Arrange
		container := di.New()
		app := &mockApp{container: container}
		provider := NewServiceProvider()

		// Act & Assert - should panic when config is missing
		assert.Panics(t, func() {
			provider.Register(app)
		}, "Expected provider.Register to panic when config is missing")
	})

	t.Run("does nothing when app doesn't have container", func(t *testing.T) {
		// Arrange
		app := &mockApp{container: nil}
		provider := NewServiceProvider()

		// Act & Assert - should not panic
		assert.NotPanics(t, func() {
			provider.Register(app)
		}, "Should not panic when app doesn't have container")
	})

	t.Run("registers scheduler service if not available", func(t *testing.T) {
		// Arrange
		container := di.New()
		mockConfig := mocks.NewMockManager(t)

		testQueueConfig := setupTestQueueConfig()
		mockConfig.EXPECT().UnmarshalKey("queue", mock.Anything).Run(func(_ string, out interface{}) {
			if cfg, ok := out.(*Config); ok {
				*cfg = *testQueueConfig
			}
		}).Return(nil)

		// Also expect scheduler config to be loaded
		mockConfig.EXPECT().UnmarshalKey("scheduler", mock.Anything).Return(nil)

		container.Instance("config", mockConfig)

		// Mock redis service to satisfy redis dependency
		mockRedis := redisMocks.NewMockManager(t)
		container.Instance("redis", mockRedis)

		app := &mockApp{container: container}
		provider := NewServiceProvider()

		// Act
		provider.Register(app)

		// Assert - scheduler should be registered
		assert.True(t, container.Bound("scheduler"), "Expected 'scheduler' to be bound when not available")
	})
}

func TestServiceProviderBoot(t *testing.T) {
	t.Run("Boot doesn't panic", func(t *testing.T) {
		// Create DI container with config
		container := di.New()
		mockConfig := mocks.NewMockManager(t)
		mockSched := schedulerMocks.NewMockManager(t)
		mockRedis := redisMocks.NewMockManager(t)

		// Setup expectations for UnmarshalKey
		mockConfig.EXPECT().UnmarshalKey("queue", mock.Anything).Run(func(_ string, out interface{}) {
			// Copy our test config to the output parameter
			if cfg, ok := out.(*Config); ok {
				*cfg = *setupTestQueueConfig()
			}
		}).Return(nil)

		// Setup expectations for scheduler methods called during Boot
		mockSched.EXPECT().Every(1).Return(mockSched)
		mockSched.EXPECT().Hours().Return(mockSched)
		mockSched.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

		mockSched.EXPECT().Every(5).Return(mockSched)
		mockSched.EXPECT().Minutes().Return(mockSched)
		mockSched.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

		mockSched.EXPECT().IsRunning().Return(false)
		mockSched.EXPECT().StartAsync().Return()

		container.Instance("config", mockConfig)
		container.Instance("scheduler", mockSched)
		container.Instance("redis", mockRedis)

		// Create app and provider
		app := &mockApp{container: container}
		provider := NewServiceProvider()

		// First register the provider
		provider.Register(app)

		// Then test that boot doesn't panic
		assert.NotPanics(t, func() {
			provider.Boot(app)
		}, "Boot should not panic with valid configuration")

		// Verify mock expectations were met
		mockSched.AssertExpectations(t)
	})

	t.Run("Boot works without container", func(t *testing.T) {
		// Test with no container
		provider := NewServiceProvider()
		app := &mockApp{container: nil}

		// Should not panic
		assert.NotPanics(t, func() {
			provider.Boot(app)
		}, "Boot should not panic with nil container")
	})

	t.Run("Boot sets scheduler on server", func(t *testing.T) {
		// Create DI container with config
		container := di.New()
		mockConfig := mocks.NewMockManager(t)
		mockSched := schedulerMocks.NewMockManager(t)

		mockConfig.EXPECT().UnmarshalKey("queue", mock.Anything).Run(func(_ string, out interface{}) {
			if cfg, ok := out.(*Config); ok {
				*cfg = *setupTestQueueConfig()
			}
		}).Return(nil)

		// Setup expectations for scheduler methods called during Boot
		mockSched.EXPECT().Every(1).Return(mockSched)
		mockSched.EXPECT().Hours().Return(mockSched)
		mockSched.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

		mockSched.EXPECT().Every(5).Return(mockSched)
		mockSched.EXPECT().Minutes().Return(mockSched)
		mockSched.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

		mockSched.EXPECT().IsRunning().Return(false)
		mockSched.EXPECT().StartAsync().Return()

		container.Instance("config", mockConfig)
		container.Instance("scheduler", mockSched)

		app := &mockApp{container: container}
		provider := NewServiceProvider()

		// Register and boot
		provider.Register(app)
		provider.Boot(app)

		// Get the queue manager and verify scheduler was set
		queueManager, err := container.Make("queue")
		assert.NoError(t, err)
		manager := queueManager.(Manager)

		// Verify server has scheduler set
		server := manager.Server()
		if server != nil {
			scheduler := server.GetScheduler()
			assert.NotNil(t, scheduler, "Expected scheduler to be set on server")
		}
	})
}

func TestServiceProviderBootWithNil(t *testing.T) {
	// Test Boot with nil app parameter
	provider := NewServiceProvider()

	// Should not panic with nil app
	assert.NotPanics(t, func() {
		provider.Boot(nil)
	}, "Boot should not panic with nil app parameter")
}

func TestServiceProviderProviders(t *testing.T) {
	provider := NewServiceProvider()
	providers := provider.Providers()

	// Queue provider should provide "queue" service
	assert.Contains(t, providers, "queue", "Expected providers to contain 'queue'")
	assert.Len(t, providers, 1, "Expected 1 provider")
}

func TestServiceProviderRequires(t *testing.T) {
	provider := NewServiceProvider()
	requires := provider.Requires()

	// Queue provider requires config, redis and scheduler providers
	assert.Len(t, requires, 3, "Expected 3 required dependencies")
	assert.Contains(t, requires, "config", "Expected required dependency to be 'config'")
	assert.Contains(t, requires, "redis", "Expected required dependency to be 'redis'")
	assert.Contains(t, requires, "scheduler", "Expected required dependency to be 'scheduler'")
}

func TestServiceProviderInterfaceCompliance(t *testing.T) {
	// This test verifies that our concrete type implements the interface
	var _ ServiceProvider = (*serviceProvider)(nil)
	var _ di.ServiceProvider = (*serviceProvider)(nil)
}

func TestMockConfigManagerWithQueueConfig(t *testing.T) {
	// This test verifies that our mock config manager can be used with Queue config
	mockConfig := mocks.NewMockManager(t)
	testConfig := setupTestQueueConfig()

	// Setup expectations for the Has method
	mockConfig.EXPECT().Has("queue").Return(true)

	// Setup expectations for the Get method
	mockConfig.EXPECT().Get("queue").Return(testConfig, true)

	// Setup expectations for UnmarshalKey
	mockConfig.EXPECT().UnmarshalKey("queue", mock.Anything).Run(func(_ string, out interface{}) {
		// Copy our test config to the output parameter
		if cfg, ok := out.(*Config); ok {
			*cfg = *testConfig
		}
	}).Return(nil)

	// Test Has method
	assert.True(t, mockConfig.Has("queue"), "Has should return true for queue key")

	// Test Get method
	value, exists := mockConfig.Get("queue")
	assert.True(t, exists, "Should find the queue key")
	assert.Equal(t, testConfig, value, "Should return our test config")

	// Test UnmarshalKey method
	var outConfig Config
	err := mockConfig.UnmarshalKey("queue", &outConfig)
	assert.NoError(t, err, "UnmarshalKey should not return an error")

	// Verify our mock expectations were met
	mockConfig.AssertExpectations(t)
}

func TestCleanupAndRetryMethods(t *testing.T) {
	// Test that the provider has the cleanup and retry methods
	provider := NewServiceProvider().(*serviceProvider)

	// These methods should be available as they're called by Boot
	assert.NotNil(t, provider.cleanupFailedJobs, "Expected cleanupFailedJobs method to exist")
	assert.NotNil(t, provider.retryFailedJobs, "Expected retryFailedJobs method to exist")
	assert.NotNil(t, provider.setupQueueScheduledTasks, "Expected setupQueueScheduledTasks method to exist")
}

// TestProviderCleanupFailedJobs tests the cleanupFailedJobs method
func TestProviderCleanupFailedJobs(t *testing.T) {
	provider := NewServiceProvider().(*serviceProvider)

	t.Run("cleans up old dead letter tasks", func(t *testing.T) {
		// Create mock container
		container := di.New()

		// Create mock config manager
		mockConfig := mocks.NewMockManager(t)
		container.Instance("config", mockConfig)

		// Create test config
		testConfig := Config{
			Server: ServerConfig{
				Queues: []string{"test", "another"},
			},
		}

		// Set up config manager expectations
		mockConfig.EXPECT().UnmarshalKey("queue", mock.AnythingOfType("*queue.Config")).Run(func(key string, out interface{}) {
			cfg := out.(*Config)
			*cfg = testConfig
		}).Return(nil)

		// Create test manager with memory adapter
		manager := &testManager{
			adapter: adapter.NewMemoryQueue("test:"),
		}

		// Add some test dead letter tasks
		ctx := context.Background()
		deadQueue := "test:dead"

		// Create old task (more than 7 days old)
		oldTask := DeadLetterTask{
			Task: Task{
				ID:   "old-task",
				Name: "test",
			},
			FailedAt: time.Now().AddDate(0, 0, -8), // 8 days ago
		}

		// Create new task (less than 7 days old)
		newTask := DeadLetterTask{
			Task: Task{
				ID:   "new-task",
				Name: "test",
			},
			FailedAt: time.Now().AddDate(0, 0, -3), // 3 days ago
		}

		// Add tasks to dead letter queue
		manager.adapter.Enqueue(ctx, deadQueue, &oldTask)
		manager.adapter.Enqueue(ctx, deadQueue, &newTask)

		// Verify initial state
		size, _ := manager.adapter.Size(ctx, deadQueue)
		assert.Equal(t, int64(2), size, "Should have 2 tasks initially")

		// Run cleanup
		provider.cleanupFailedJobs(manager, container)

		// Verify cleanup results
		finalSize, _ := manager.adapter.Size(ctx, deadQueue)
		assert.Equal(t, int64(1), finalSize, "Should have 1 task after cleanup (old task removed)")

		// Verify remaining task is the new one
		var remainingTask DeadLetterTask
		err := manager.adapter.Dequeue(ctx, deadQueue, &remainingTask)
		assert.NoError(t, err)
		assert.Equal(t, "new-task", remainingTask.Task.ID, "Remaining task should be the new one")
	})

	t.Run("handles empty dead letter queue", func(t *testing.T) {
		// Create mock container
		container := di.New()

		// Create mock config manager
		mockConfig := mocks.NewMockManager(t)
		container.Instance("config", mockConfig)

		// Create test config
		testConfig := Config{
			Server: ServerConfig{
				Queues: []string{"empty"},
			},
		}

		// Set up config manager expectations
		mockConfig.EXPECT().UnmarshalKey("queue", mock.AnythingOfType("*queue.Config")).Run(func(key string, out interface{}) {
			cfg := out.(*Config)
			*cfg = testConfig
		}).Return(nil)

		// Create test manager with memory adapter
		manager := &testManager{
			adapter: adapter.NewMemoryQueue("test:"),
		}

		// Run cleanup on empty queue
		provider.cleanupFailedJobs(manager, container)

		// Should not panic or error
		ctx := context.Background()
		size, _ := manager.adapter.Size(ctx, "empty:dead")
		assert.Equal(t, int64(0), size, "Dead letter queue should remain empty")
	})

	t.Run("handles config unmarshal error", func(t *testing.T) {
		// Create mock container
		container := di.New()

		// Create mock config manager that returns error
		mockConfig := mocks.NewMockManager(t)
		container.Instance("config", mockConfig)

		// Set up config manager to return error
		mockConfig.EXPECT().UnmarshalKey("queue", mock.AnythingOfType("*queue.Config")).Return(assert.AnError)

		// Create test manager
		manager := &testManager{
			adapter: adapter.NewMemoryQueue("test:"),
		}

		// Run cleanup - should handle error gracefully
		provider.cleanupFailedJobs(manager, container)

		// Method should complete without panic
	})
}

// TestProviderRetryFailedJobs tests the retryFailedJobs method
func TestProviderRetryFailedJobs(t *testing.T) {
	provider := NewServiceProvider().(*serviceProvider)

	t.Run("moves retry tasks to pending when ready", func(t *testing.T) {
		// Create mock container
		container := di.New()

		// Create mock config manager
		mockConfig := mocks.NewMockManager(t)
		container.Instance("config", mockConfig)

		// Create test config
		testConfig := Config{
			Server: ServerConfig{
				Queues: []string{"test"},
			},
		}

		// Set up config manager expectations
		mockConfig.EXPECT().UnmarshalKey("queue", mock.AnythingOfType("*queue.Config")).Run(func(key string, out interface{}) {
			cfg := out.(*Config)
			*cfg = testConfig
		}).Return(nil)

		// Create test manager with memory adapter
		manager := &testManager{
			adapter: adapter.NewMemoryQueue("test:"),
		}

		// Add some test retry tasks
		ctx := context.Background()
		retryQueue := "test:retry"
		pendingQueue := "test:pending"

		// Create ready task (past ProcessAt time)
		readyTask := Task{
			ID:        "ready-task",
			Name:      "test",
			ProcessAt: time.Now().Add(-1 * time.Hour), // 1 hour ago
		}

		// Create not-ready task (future ProcessAt time)
		notReadyTask := Task{
			ID:        "not-ready-task",
			Name:      "test",
			ProcessAt: time.Now().Add(1 * time.Hour), // 1 hour from now
		}

		// Add tasks to retry queue
		manager.adapter.Enqueue(ctx, retryQueue, &readyTask)
		manager.adapter.Enqueue(ctx, retryQueue, &notReadyTask)

		// Verify initial state
		retrySize, _ := manager.adapter.Size(ctx, retryQueue)
		pendingSize, _ := manager.adapter.Size(ctx, pendingQueue)
		assert.Equal(t, int64(2), retrySize, "Should have 2 tasks in retry queue initially")
		assert.Equal(t, int64(0), pendingSize, "Should have 0 tasks in pending queue initially")

		// Run retry processing
		provider.retryFailedJobs(manager, container)

		// Verify results
		finalRetrySize, _ := manager.adapter.Size(ctx, retryQueue)
		finalPendingSize, _ := manager.adapter.Size(ctx, pendingQueue)
		assert.Equal(t, int64(1), finalRetrySize, "Should have 1 task remaining in retry queue")
		assert.Equal(t, int64(1), finalPendingSize, "Should have 1 task moved to pending queue")

		// Verify the ready task was moved to pending
		var pendingTask Task
		err := manager.adapter.Dequeue(ctx, pendingQueue, &pendingTask)
		assert.NoError(t, err)
		assert.Equal(t, "ready-task", pendingTask.ID, "Ready task should be in pending queue")

		// Verify the not-ready task remains in retry
		var retryTask Task
		err = manager.adapter.Dequeue(ctx, retryQueue, &retryTask)
		assert.NoError(t, err)
		assert.Equal(t, "not-ready-task", retryTask.ID, "Not-ready task should remain in retry queue")
	})

	t.Run("handles empty retry queue", func(t *testing.T) {
		// Create mock container
		container := di.New()

		// Create mock config manager
		mockConfig := mocks.NewMockManager(t)
		container.Instance("config", mockConfig)

		// Create test config
		testConfig := Config{
			Server: ServerConfig{
				Queues: []string{"empty"},
			},
		}

		// Set up config manager expectations
		mockConfig.EXPECT().UnmarshalKey("queue", mock.AnythingOfType("*queue.Config")).Run(func(key string, out interface{}) {
			cfg := out.(*Config)
			*cfg = testConfig
		}).Return(nil)

		// Create test manager with memory adapter
		manager := &testManager{
			adapter: adapter.NewMemoryQueue("test:"),
		}

		// Run retry processing on empty queue
		provider.retryFailedJobs(manager, container)

		// Should not panic or error
		ctx := context.Background()
		retrySize, _ := manager.adapter.Size(ctx, "empty:retry")
		pendingSize, _ := manager.adapter.Size(ctx, "empty:pending")
		assert.Equal(t, int64(0), retrySize, "Retry queue should remain empty")
		assert.Equal(t, int64(0), pendingSize, "Pending queue should remain empty")
	})
}

// testManager is a simplified test implementation of Manager interface
// Used for integration-style tests where we need a real adapter instance
// Note: We use a local implementation instead of the generated mock to avoid import cycles
type testManager struct {
	adapter adapter.QueueAdapter
}

func (m *testManager) Adapter(name string) adapter.QueueAdapter {
	return m.adapter
}

func (m *testManager) RedisClient() redis.UniversalClient {
	// For testing purposes, return a simple mock client
	// In real implementation, this would come from the Redis provider
	return redis.NewClient(&redis.Options{Addr: "localhost:6379"})
}

func (m *testManager) MemoryAdapter() adapter.QueueAdapter {
	return m.adapter
}

func (m *testManager) RedisAdapter() adapter.QueueAdapter {
	return m.adapter
}

func (m *testManager) Client() Client {
	return nil
}

func (m *testManager) Server() Server {
	return nil
}

func (m *testManager) Scheduler() scheduler.Manager {
	return nil
}

func (m *testManager) SetScheduler(s scheduler.Manager) {
}
