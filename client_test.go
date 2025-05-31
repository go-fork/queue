package queue

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.fork.vn/queue/adapter"
)

// TestNewClient tests creation of client with Redis adapter
func TestNewClient(t *testing.T) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClient(redisClient)

	assert.NotNil(t, client, "Client should not be nil")
}

// TestNewMemoryClient tests creation of client with memory adapter
func TestNewMemoryClient(t *testing.T) {
	client := NewMemoryClient()

	assert.NotNil(t, client, "Client should not be nil")
}

// TestNewClientWithAdapter tests creation of client with a custom adapter
func TestNewClientWithAdapter(t *testing.T) {
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	assert.NotNil(t, client, "Client should not be nil")
}

// TestNewClientWithUniversalClient tests creation of client with UniversalClient
func TestNewClientWithUniversalClient(t *testing.T) {
	// Test with standard Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	client := NewClientWithUniversalClient(redisClient)
	assert.NotNil(t, client, "Client should not be nil")

	// Test with cluster client (which is not a standard *redis.Client)
	// This should trigger the fallback code path
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:6379"},
	})
	clusterQueueClient := NewClientWithUniversalClient(clusterClient)
	assert.NotNil(t, clusterQueueClient, "Client should not be nil")

	// Test basic operation to ensure it's usable
	err := clusterQueueClient.Close()
	assert.NoError(t, err, "Close should not return an error")
}

// TestClientEnqueue tests the Enqueue function
func TestClientEnqueue(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Enqueue a task
	payload := map[string]string{"key": "value"}
	taskInfo, err := client.Enqueue("test_task", payload)

	require.NoError(t, err, "Enqueue should not return an error")
	assert.NotNil(t, taskInfo, "TaskInfo should not be nil")
	assert.Equal(t, "test_task", taskInfo.Name, "Task name should match")
	assert.Equal(t, "default", taskInfo.Queue, "Default queue should be 'default'")

	// Verify the task was actually enqueued
	queueItems, err := memoryAdapter.Size(context.Background(), "default:pending")
	require.NoError(t, err, "Getting queue size should not return an error")
	assert.Equal(t, int64(1), queueItems, "Queue should have 1 item")
}

// TestClientEnqueueWithOptions tests the Enqueue function with options
func TestClientEnqueueWithOptions(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Enqueue a task with custom options
	payload := map[string]string{"key": "value"}
	taskID := "custom-id"
	queueName := "custom-queue"
	maxRetry := 5

	taskInfo, err := client.Enqueue(
		"test_task",
		payload,
		WithTaskID(taskID),
		WithQueue(queueName),
		WithMaxRetry(maxRetry),
	)

	require.NoError(t, err, "Enqueue should not return an error")
	assert.NotNil(t, taskInfo, "TaskInfo should not be nil")
	assert.Equal(t, "test_task", taskInfo.Name, "Task name should match")
	assert.Equal(t, taskID, taskInfo.ID, "Task ID should match custom ID")
	assert.Equal(t, queueName, taskInfo.Queue, "Queue should match custom queue")
	assert.Equal(t, maxRetry, taskInfo.MaxRetry, "MaxRetry should match custom value")

	// Verify the task was actually enqueued in the correct queue
	queueItems, err := memoryAdapter.Size(context.Background(), queueName+":pending")
	require.NoError(t, err, "Getting queue size should not return an error")
	assert.Equal(t, int64(1), queueItems, "Queue should have 1 item")
}

// TestClientEnqueueContext tests the EnqueueContext function
func TestClientEnqueueContext(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Create a context
	ctx := context.Background()

	// Enqueue a task with context
	payload := map[string]string{"key": "value"}
	taskInfo, err := client.EnqueueContext(ctx, "test_task", payload)

	require.NoError(t, err, "EnqueueContext should not return an error")
	assert.NotNil(t, taskInfo, "TaskInfo should not be nil")
	assert.Equal(t, "test_task", taskInfo.Name, "Task name should match")
}

// TestClientEnqueueIn tests the EnqueueIn function
func TestClientEnqueueIn(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Enqueue a task with delay
	payload := map[string]string{"key": "value"}
	delay := 5 * time.Minute
	now := time.Now()

	taskInfo, err := client.EnqueueIn("test_task", delay, payload)

	require.NoError(t, err, "EnqueueIn should not return an error")
	assert.NotNil(t, taskInfo, "TaskInfo should not be nil")
	assert.Equal(t, "test_task", taskInfo.Name, "Task name should match")

	// Verify ProcessAt is in the future
	assert.True(t, taskInfo.ProcessAt.After(now), "ProcessAt should be in the future")
	assert.True(t, taskInfo.ProcessAt.After(now.Add(delay-time.Second)), "ProcessAt should be approximately delay time in the future")
}

// TestClientEnqueueAt tests the EnqueueAt function
func TestClientEnqueueAt(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Enqueue a task at a specific time
	payload := map[string]string{"key": "value"}
	futureTime := time.Now().Add(10 * time.Minute)

	taskInfo, err := client.EnqueueAt("test_task", futureTime, payload)

	require.NoError(t, err, "EnqueueAt should not return an error")
	assert.NotNil(t, taskInfo, "TaskInfo should not be nil")
	assert.Equal(t, "test_task", taskInfo.Name, "Task name should match")
	assert.Equal(t, futureTime.Unix(), taskInfo.ProcessAt.Unix(), "ProcessAt should match the specified time")
}

// TestClientEnqueueWithInvalidPayload tests error handling for invalid payload
func TestClientEnqueueWithInvalidPayload(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Create an invalid payload (a function cannot be marshaled to JSON)
	invalidPayload := func() {}

	// Attempt to enqueue with invalid payload
	_, err := client.Enqueue("test_task", invalidPayload)

	assert.Error(t, err, "Enqueue with invalid payload should return an error")
	assert.Contains(t, err.Error(), "failed to marshal payload", "Error should mention payload marshaling issue")
}

// TestClientClose tests the Close function
func TestClientClose(t *testing.T) {
	// Create a memory adapter and client
	memoryAdapter := adapter.NewMemoryQueue("test:")
	client := NewClientWithAdapter(memoryAdapter)

	// Close should not error
	err := client.Close()
	assert.NoError(t, err, "Close should not return an error")
}

// TestGenerateID tests the ID generation function
func TestGenerateID(t *testing.T) {
	// Test basic ID generation
	id1 := generateID()
	id2 := generateID()

	assert.NotEmpty(t, id1, "Generated ID should not be empty")
	assert.NotEqual(t, id1, id2, "Generated IDs should be unique")
	assert.Len(t, id1, 32, "Generated ID should be 32 characters (16 bytes hex)")

	// Generate multiple IDs
	for i := 0; i < 10; i++ {
		id := generateID()
		assert.NotEmpty(t, id, "Generated ID should not be empty")
		assert.GreaterOrEqual(t, len(id), 8, "Generated ID should be at least 8 characters")
	}

	// Note: We can't easily mock crypto/rand.Read to test the fallback path directly
	// but the function should always return a valid ID regardless
}
