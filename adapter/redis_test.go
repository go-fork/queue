package adapter

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestNewRedisQueue(t *testing.T) {
	client, _ := redismock.NewClientMock()

	// Test với prefix tùy chỉnh
	queue1 := NewRedisQueue(client, "custom:")
	assert.NotNil(t, queue1)

	// Test với prefix rỗng (sử dụng default)
	queue2 := NewRedisQueue(client, "")
	assert.NotNil(t, queue2)
}

func TestRedisQueueEnqueue(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	item := testItem{
		ID:      "123",
		Message: "test message",
		Time:    time.Now(),
	}

	// Chuẩn bị expected json bytes
	jsonBytes, _ := json.Marshal(item)

	// Mock Redis RPUSH
	mock.ExpectRPush("test:test-queue", jsonBytes).SetVal(1)

	// Thực thi
	err := queue.Enqueue(ctx, queueName, item)

	// Kiểm tra
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueDequeue(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	expectedItem := testItem{
		ID:      "123",
		Message: "test message",
		Time:    time.Now().UTC().Truncate(time.Second),
	}

	// Chuẩn bị json bytes
	jsonBytes, _ := json.Marshal(expectedItem)

	// Mock Redis LPOP - match the actual implementation which uses LPOP not BLPOP
	mock.ExpectLPop("test:test-queue").SetVal(string(jsonBytes))

	// Thực thi
	var result testItem
	err := queue.Dequeue(ctx, queueName, &result)

	// Kiểm tra
	assert.NoError(t, err)
	assert.Equal(t, expectedItem.ID, result.ID)
	assert.Equal(t, expectedItem.Message, result.Message)
	assert.Equal(t, expectedItem.Time.Unix(), result.Time.Unix())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueDequeueEmptyQueue(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "empty-queue"

	// Mock Redis LPOP returns nil for empty queue
	mock.ExpectLPop("test:empty-queue").SetErr(redis.Nil)

	// Thực thi
	var result testItem
	err := queue.Dequeue(ctx, queueName, &result)

	// Kiểm tra
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueEnqueueBatch(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-batch"

	items := []interface{}{
		testItem{ID: "1", Message: "message 1", Time: time.Now()},
		testItem{ID: "2", Message: "message 2", Time: time.Now()},
	}

	// Chuyển đổi items thành json
	jsonItems := make([]interface{}, len(items))
	for i, item := range items {
		bytes, _ := json.Marshal(item)
		jsonItems[i] = bytes
	}

	// Mock Redis RPUSH với nhiều giá trị
	mock.ExpectRPush("test:test-batch", jsonItems...).SetVal(int64(len(items)))

	// Thực thi
	err := queue.EnqueueBatch(ctx, queueName, items)

	// Kiểm tra
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueSize(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-size"

	// Mock Redis LLEN
	mock.ExpectLLen("test:test-size").SetVal(5)

	// Thực thi
	size, err := queue.Size(ctx, queueName)

	// Kiểm tra
	assert.NoError(t, err)
	assert.Equal(t, int64(5), size)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueIsEmpty(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-empty"

	// Trường hợp queue rỗng
	mock.ExpectLLen("test:test-empty").SetVal(0)

	// Thực thi
	empty, err := queue.IsEmpty(ctx, queueName)

	// Kiểm tra
	assert.NoError(t, err)
	assert.True(t, empty)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Trường hợp queue không rỗng
	mock.ExpectLLen("test:test-empty").SetVal(3)

	// Thực thi
	empty, err = queue.IsEmpty(ctx, queueName)

	// Kiểm tra
	assert.NoError(t, err)
	assert.False(t, empty)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueClear(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-clear"

	// Mock Redis DEL
	mock.ExpectDel("test:test-clear").SetVal(1)

	// Thực thi
	err := queue.Clear(ctx, queueName)

	// Kiểm tra
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueDequeueWithTimeout(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"
	timeout := 5 * time.Second

	expectedItem := testItem{
		ID:      "123",
		Message: "test message",
		Time:    time.Now().UTC().Truncate(time.Second),
	}

	// Chuẩn bị json bytes
	jsonBytes, _ := json.Marshal(expectedItem)

	// Mock Redis BLPOP với timeout
	mock.ExpectBLPop(timeout, "test:test-queue").SetVal([]string{"test:test-queue", string(jsonBytes)})

	// Thực thi
	var result testItem
	err := queue.(*redisQueue).DequeueWithTimeout(ctx, queueName, timeout, &result)

	// Kiểm tra
	assert.NoError(t, err)
	assert.Equal(t, expectedItem.ID, result.ID)
	assert.Equal(t, expectedItem.Message, result.Message)
	assert.Equal(t, expectedItem.Time.Unix(), result.Time.Unix())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueDequeueWithTimeoutError(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "empty-queue"
	timeout := 1 * time.Second

	// Mock Redis BLPOP returns nil (timeout)
	mock.ExpectBLPop(timeout, "test:empty-queue").SetErr(redis.Nil)

	// Thực thi
	var result testItem
	err := queue.(*redisQueue).DequeueWithTimeout(ctx, queueName, timeout, &result)

	// Kiểm tra
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout waiting")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisQueueDequeueWithTimeoutInvalidResponse(t *testing.T) {
	// Chuẩn bị
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "invalid-queue"
	timeout := 1 * time.Second

	// Mock Redis BLPOP returns invalid response format
	mock.ExpectBLPop(timeout, "test:invalid-queue").SetVal([]string{"test:invalid-queue"}) // Missing value

	// Thực thi
	var result testItem
	err := queue.(*redisQueue).DequeueWithTimeout(ctx, queueName, timeout, &result)

	// Kiểm tra
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected response format")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// TestRedisQueueEnqueueError tests error cases for Enqueue
func TestRedisQueueEnqueueError(t *testing.T) {
	// Setup
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	// Case 1: Redis error
	jsonBytes, _ := json.Marshal(testItem{ID: "123"})
	mock.ExpectRPush("test:test-queue", jsonBytes).SetErr(redis.ErrClosed)

	err := queue.Enqueue(ctx, queueName, testItem{ID: "123"})
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)

	// Case 2: JSON marshaling error (channels can't be marshaled)
	badItem := make(chan int)
	err = queue.Enqueue(ctx, queueName, badItem)
	assert.Error(t, err)
}

// TestRedisQueueDequeueError tests error cases for Dequeue
func TestRedisQueueDequeueError(t *testing.T) {
	// Setup
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	// Case 1: Redis error
	mock.ExpectLPop("test:test-queue").SetErr(redis.ErrClosed)

	var result testItem
	err := queue.Dequeue(ctx, queueName, &result)
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
}

// TestRedisQueueEnqueueBatchError tests error cases for EnqueueBatch
func TestRedisQueueEnqueueBatchError(t *testing.T) {
	// Setup
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	// Case 1: Redis error
	jsonBytes, _ := json.Marshal(testItem{ID: "1"})
	mock.ExpectRPush("test:test-queue", jsonBytes).SetErr(redis.ErrClosed)

	err := queue.EnqueueBatch(ctx, queueName, []interface{}{testItem{ID: "1"}})
	assert.Error(t, err)

	// Case 2: JSON marshaling error
	items := []interface{}{
		testItem{ID: "1"},
		make(chan int), // Will cause marshal error
	}
	err = queue.EnqueueBatch(ctx, queueName, items)
	assert.Error(t, err)
}

// TestRedisQueueIsEmptyError tests error cases for IsEmpty
func TestRedisQueueIsEmptyError(t *testing.T) {
	// Setup
	client, mock := redismock.NewClientMock()
	queue := NewRedisQueue(client, "test:")
	ctx := context.Background()
	queueName := "test-queue"

	// Case 1: Redis error
	mock.ExpectLLen("test:test-queue").SetErr(redis.ErrClosed)

	_, err := queue.IsEmpty(ctx, queueName)
	assert.Error(t, err)
	assert.Equal(t, redis.ErrClosed, err)
}
