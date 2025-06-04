package queue

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.fork.vn/queue/adapter"
	schedulerMocks "go.fork.vn/scheduler/mocks"
)

// TestNewServer tests the creation of a server with Redis adapter
func TestNewServer(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "default",
		Queues:          []string{"default", "emails", "critical"},
	}

	srv := NewServer(redisClient, opts)

	assert.NotNil(t, srv, "Server should not be nil")
	queueServer, ok := srv.(*queueServer)
	assert.True(t, ok, "Server should be of type *queueServer")
	assert.Equal(t, opts.Concurrency, queueServer.options.Concurrency)
}

// TestNewServerWithAdapter tests the creation of a server with a custom adapter
func TestNewServerWithAdapter(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	assert.NotNil(t, server, "Server should not be nil")
	queueServer, ok := server.(*queueServer)
	assert.True(t, ok, "Server should be of type *queueServer")
	assert.Equal(t, opts.Concurrency, queueServer.options.Concurrency)
}

// TestServerRegisterHandler tests registering a single handler
func TestServerRegisterHandler(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:  2,
		DefaultQueue: "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts).(*queueServer)

	// Register a handler
	handler := func(ctx context.Context, task *Task) error {
		return nil
	}

	server.RegisterHandler("test_task", handler)

	// Verify handler was registered
	value, ok := server.handlers.Load("test_task")
	assert.True(t, ok, "Handler should be stored in handlers map")
	assert.NotNil(t, value, "Handler function should not be nil")
}

// TestServerRegisterHandlers tests registering multiple handlers at once
func TestServerRegisterHandlers(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:  2,
		DefaultQueue: "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts).(*queueServer)

	// Create some handlers
	handler1 := func(ctx context.Context, task *Task) error { return nil }
	handler2 := func(ctx context.Context, task *Task) error { return nil }

	// Register multiple handlers
	handlers := map[string]HandlerFunc{
		"task1": handler1,
		"task2": handler2,
	}

	server.RegisterHandlers(handlers)

	// Verify handlers were registered
	value1, ok1 := server.handlers.Load("task1")
	assert.True(t, ok1, "Handler for task1 should be stored")
	assert.NotNil(t, value1, "Handler function for task1 should not be nil")

	value2, ok2 := server.handlers.Load("task2")
	assert.True(t, ok2, "Handler for task2 should be stored")
	assert.NotNil(t, value2, "Handler function for task2 should not be nil")
}

// TestServerOptionsDefaults tests that default values are applied correctly
func TestServerOptionsDefaults(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	// Create server with no options specified
	opts := ServerOptions{}

	server := NewServerWithAdapter(memoryAdapter, opts).(*queueServer)

	// Check that the default queue was applied
	assert.Contains(t, server.queues, "default", "Default queue should be 'default'")
}

// TestServerStart tests the Start method
func TestServerStart(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:  2,
		DefaultQueue: "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Start the server
	err := server.Start()
	assert.NoError(t, err, "Start should not return an error")

	// Try starting again - should fail
	err = server.Start()
	assert.Error(t, err, "Starting an already started server should return an error")
	assert.Contains(t, err.Error(), "already started", "Error should mention server is already started")

	// Clean up
	_ = server.Stop()
}

// TestServerStop tests the Stop method
func TestServerStop(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:  2,
		DefaultQueue: "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Start the server first
	err := server.Start()
	require.NoError(t, err, "Start should not return an error")

	// Stop the server
	err = server.Stop()
	assert.NoError(t, err, "Stop should not return an error")

	// Try stopping again - should fail
	err = server.Stop()
	assert.Error(t, err, "Stopping an already stopped server should return an error")
	assert.Contains(t, err.Error(), "not started", "Error should mention server is not started")
}

// TestServerWithRedisAdapter tests server with Redis adapter using redismock
func TestServerWithRedisAdapter(t *testing.T) {
	// Create redis client and mock
	redisClient, _ := redismock.NewClientMock()

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "test-queue",
		ShutdownTimeout: 5 * time.Second,
	}

	server := NewServer(redisClient, opts)

	assert.NotNil(t, server, "Server should not be nil")

	// Register a handler
	server.RegisterHandler("test_task", func(ctx context.Context, task *Task) error {
		return nil
	})

	// We can test start/stop cycle
	err := server.Start()
	assert.NoError(t, err, "Start should not return an error")

	err = server.Stop()
	assert.NoError(t, err, "Stop should not return an error")
}

// TestServerWithCustomQueues tests server with custom queue names
func TestServerWithCustomQueues(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	customQueues := []string{"high", "medium", "low"}
	opts := ServerOptions{
		Queues: customQueues,
	}

	server := NewServerWithAdapter(memoryAdapter, opts).(*queueServer)

	// Verify the queues were set correctly
	assert.Equal(t, len(customQueues), len(server.queues), "Server should have the correct number of queues")
	for i, queue := range customQueues {
		assert.Equal(t, queue, server.queues[i], "Queue name should match")
	}
}

// TestNewServerWithUniversalClient tests server creation with a non-standard Redis client
func TestNewServerWithUniversalClient(t *testing.T) {
	// Create a Redis Cluster client (which is not a standard *redis.Client)
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:6379"},
	})

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "cluster-queue",
	}

	// This should trigger the fallback code path
	server := NewServer(clusterClient, opts)

	assert.NotNil(t, server, "Server should not be nil")

	// Start and stop to verify it's functional
	err := server.Start()
	assert.NoError(t, err, "Start should not return an error")

	err = server.Stop()
	assert.NoError(t, err, "Stop should not return an error")
}

// TestSchedulerMock tests the server with a mocked scheduler
func TestSchedulerMock(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:  2,
		DefaultQueue: "default",
	}

	// Create a mock scheduler
	schedulerMock := schedulerMocks.NewMockManager(t)

	// Set up the server with the mock scheduler
	server := NewServerWithAdapter(memoryAdapter, opts)
	server.SetScheduler(schedulerMock)

	// Expectations on the mock scheduler - sử dụng EXPECT() syntax
	schedulerMock.EXPECT().IsRunning().Return(false).Maybe()
	schedulerMock.EXPECT().StartAsync().Maybe()

	// Expect scheduler setup for delayed task processing
	schedulerMock.EXPECT().Every(mock.AnythingOfType("uint64")).Return(schedulerMock).Maybe()
	schedulerMock.EXPECT().Seconds().Return(schedulerMock).Maybe()
	schedulerMock.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil).Maybe()

	// Register a handler
	server.RegisterHandler("test_task", func(ctx context.Context, task *Task) error {
		return nil
	})

	// Start the server
	err := server.Start()
	assert.NoError(t, err, "Start should not return an error")

	// Stop the server
	err = server.Stop()
	assert.NoError(t, err, "Stop should not return an error")

	// Assert that the expectations were met
	schedulerMock.AssertExpectations(t)
}

// TestServerSchedulerIntegration tests scheduler integration with queue server
func TestServerSchedulerIntegration(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Tạo mock scheduler
	mockScheduler := schedulerMocks.NewMockManager(t)

	// Set scheduler
	server.SetScheduler(mockScheduler)

	// Verify scheduler được set
	scheduler := server.GetScheduler()
	assert.Same(t, mockScheduler, scheduler, "Scheduler should be set correctly")
}

// TestServerSchedulerDelayedTasks tests scheduler handling delayed tasks
func TestServerSchedulerDelayedTasks(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Tạo mock scheduler với expectations
	mockScheduler := schedulerMocks.NewMockManager(t)

	// Expect scheduler được start khi server start
	mockScheduler.EXPECT().IsRunning().Return(false)
	mockScheduler.EXPECT().StartAsync()

	// Expect scheduler setup delayed task processing
	mockScheduler.EXPECT().Every(mock.AnythingOfType("uint64")).Return(mockScheduler)
	mockScheduler.EXPECT().Seconds().Return(mockScheduler)
	mockScheduler.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

	// Set scheduler
	server.SetScheduler(mockScheduler)

	// Start server (should trigger scheduler setup)
	err := server.Start()
	assert.NoError(t, err, "Server should start without error")

	// Stop server
	err = server.Stop()
	assert.NoError(t, err, "Server should stop without error")

	// Verify all expectations were met
	mockScheduler.AssertExpectations(t)
}

// TestServerWithoutScheduler tests server operation without scheduler
func TestServerWithoutScheduler(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:     5,
		PollingInterval: 1000,
		DefaultQueue:    "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Test server hoạt động bình thường không có scheduler
	scheduler := server.GetScheduler()
	assert.Nil(t, scheduler, "Scheduler should be nil by default")

	// Server vẫn có thể start/stop
	err := server.Start()
	assert.NoError(t, err, "Server should start without scheduler")

	err = server.Stop()
	assert.NoError(t, err, "Server should stop without scheduler")
}

// TestServerSchedulerLifecycle tests scheduler lifecycle management
func TestServerSchedulerLifecycle(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")

	opts := ServerOptions{
		Concurrency:     2,
		PollingInterval: 500,
		DefaultQueue:    "default",
	}

	server := NewServerWithAdapter(memoryAdapter, opts)

	// Tạo mock scheduler
	mockScheduler := schedulerMocks.NewMockManager(t)

	// Expect scheduler lifecycle calls
	mockScheduler.EXPECT().IsRunning().Return(false).Once()
	mockScheduler.EXPECT().StartAsync().Once()
	mockScheduler.EXPECT().Every(mock.AnythingOfType("uint64")).Return(mockScheduler).Once()
	mockScheduler.EXPECT().Seconds().Return(mockScheduler).Once()
	mockScheduler.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil).Once()

	// When stopping
	mockScheduler.EXPECT().IsRunning().Return(true).Once()
	mockScheduler.EXPECT().Stop().Once()

	// Set scheduler và test lifecycle
	server.SetScheduler(mockScheduler)

	// Start server
	err := server.Start()
	require.NoError(t, err)

	// Give some time for async operations
	time.Sleep(100 * time.Millisecond)

	// Stop server
	err = server.Stop()
	require.NoError(t, err)

	// Verify all expectations
	mockScheduler.AssertExpectations(t)
}

// TestServerStartStopLifecycle tests server start/stop lifecycle
func TestServerStartStopLifecycle(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")
	opts := ServerOptions{
		Concurrency:     1,
		DefaultQueue:    "test",
		ShutdownTimeout: 1 * time.Second,
		PollingInterval: 100,
	}
	server := NewServerWithAdapter(memoryAdapter, opts)

	// Test that server can start
	err := server.Start()
	assert.NoError(t, err)

	// Test that server can stop
	err = server.Stop()
	assert.NoError(t, err)

	// Test that double stop returns error
	err = server.Stop()
	assert.Error(t, err, "Second stop should return error")
}

// TestServerTaskProcessingBasic tests basic task processing functionality
func TestServerTaskProcessingBasic(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")
	opts := ServerOptions{
		Concurrency:     1,
		DefaultQueue:    "test",
		ShutdownTimeout: 2 * time.Second,
		PollingInterval: 10, // Fast polling for test
	}
	server := NewServerWithAdapter(memoryAdapter, opts)

	processed := make(chan string, 1)
	handler := func(ctx context.Context, task *Task) error {
		processed <- task.ID
		return nil
	}

	server.RegisterHandler("simple_task", handler)

	// Create and enqueue a task
	task := &Task{
		ID:        "test-simple-1",
		Name:      "simple_task",
		Queue:     "test",
		Payload:   []byte(`{"test": true}`),
		CreatedAt: time.Now(),
		ProcessAt: time.Now(),
	}

	ctx := context.Background()
	err := memoryAdapter.Enqueue(ctx, "test:pending", task)
	assert.NoError(t, err)

	// Start server
	err = server.Start()
	assert.NoError(t, err)

	// Wait for processing with timeout
	select {
	case taskID := <-processed:
		assert.Equal(t, "test-simple-1", taskID, "Should process correct task")
	case <-time.After(1 * time.Second):
		t.Fatal("Task processing timed out")
	}

	// Clean stop
	err = server.Stop()
	assert.NoError(t, err)
}

// TestServerGetSetScheduler tests scheduler getter and setter
func TestServerGetSetScheduler(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")
	opts := ServerOptions{DefaultQueue: "test"}
	server := NewServerWithAdapter(memoryAdapter, opts)

	// Initially should be nil
	scheduler := server.GetScheduler()
	assert.Nil(t, scheduler, "Scheduler should be nil initially")

	// Set a mock scheduler
	mockScheduler := schedulerMocks.NewMockManager(t)
	server.SetScheduler(mockScheduler)

	// Get scheduler should return the mock
	scheduler = server.GetScheduler()
	assert.Equal(t, mockScheduler, scheduler, "Should return the set scheduler")
}

// TestServerHandlerNotFound tests task processing when handler is not found
func TestServerHandlerNotFound(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")
	opts := ServerOptions{
		Concurrency:     1,
		DefaultQueue:    "test",
		PollingInterval: 10,
		ShutdownTimeout: 1 * time.Second,
	}
	server := NewServerWithAdapter(memoryAdapter, opts)

	// Create task with no registered handler
	task := &Task{
		ID:        "test-no-handler",
		Name:      "unknown_task",
		Queue:     "test",
		Payload:   []byte(`{}`),
		CreatedAt: time.Now(),
		ProcessAt: time.Now(),
	}

	ctx := context.Background()
	err := memoryAdapter.Enqueue(ctx, "test:pending", task)
	assert.NoError(t, err)

	// Start and quickly stop server
	err = server.Start()
	assert.NoError(t, err)

	// Give minimal time for processing
	time.Sleep(50 * time.Millisecond)

	err = server.Stop()
	assert.NoError(t, err)
	// Test should not panic or error - server should handle missing handlers gracefully
}
