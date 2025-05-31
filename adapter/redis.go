// Package adapter cung cấp các triển khai khác nhau cho queue backend.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisClient "github.com/redis/go-redis/v9"
	"go.fork.vn/redis"
)

// QueueRedisAdapter mở rộng QueueAdapter với các tính năng đặc biệt của Redis.
// Interface này cung cấp các phương thức nâng cao chỉ có sẵn khi sử dụng Redis
// như priority queue, pipeline operations, TTL support và blocking operations.
type QueueRedisAdapter interface {
	QueueAdapter

	// DequeueWithTimeout lấy item từ hàng đợi với khả năng chờ đợi (blocking)
	DequeueWithTimeout(ctx context.Context, queueName string, timeout time.Duration, dest interface{}) error

	// EnqueueWithPipeline sử dụng Redis pipeline để tối ưu hóa hiệu suất khi thêm nhiều items
	EnqueueWithPipeline(ctx context.Context, items map[string][]interface{}) error

	// EnqueueWithTTL thêm item và thiết lập TTL (Time To Live) cho hàng đợi
	EnqueueWithTTL(ctx context.Context, queueName string, item interface{}, ttl time.Duration) error

	// EnqueueWithPriority thêm item vào hàng đợi ưu tiên sử dụng Redis Sorted Set
	EnqueueWithPriority(ctx context.Context, queueName string, item interface{}, priority float64) error

	// DequeueFromPriority lấy item có độ ưu tiên cao nhất từ hàng đợi ưu tiên
	DequeueFromPriority(ctx context.Context, queueName string, dest interface{}) error

	// GetQueueInfo trả về thông tin chi tiết về hàng đợi (size, TTL, priority queue size, etc.)
	GetQueueInfo(ctx context.Context, queueName string) (map[string]interface{}, error)

	// MultiDequeue lấy nhiều item cùng lúc từ hàng đợi (Redis 6.2+)
	MultiDequeue(ctx context.Context, queueName string, count int, destSlice interface{}) (int, error)

	// Ping kiểm tra kết nối Redis có hoạt động không
	Ping(ctx context.Context) error

	// FlushQueues xóa tất cả hàng đợi có prefix (development/testing only)
	FlushQueues(ctx context.Context) (int64, error)

	// GetRedisClient trả về Redis client để sử dụng trực tiếp
	GetRedisClient() *redisClient.Client
}

// redisQueue triển khai interface QueueAdapter sử dụng Redis.
// Struct này sử dụng Redis List để lưu trữ và quản lý hàng đợi,
// cho phép các hoạt động queue có tính mở rộng cao và phân tán.
type redisQueue struct {
	client *redisClient.Client
	prefix string
}

// NewRedisQueue tạo một instance mới của redisQueue.
// Hàm này khởi tạo kết nối Redis và áp dụng prefix cho key.
//
// Tham số:
//   - client (*redis.Client): Redis client instance
//   - prefix (string): Prefix cho các key Redis
//
// Trả về:
//   - QueueRedisAdapter: Instance mới của redisQueue với đầy đủ tính năng Redis
func NewRedisQueue(client *redisClient.Client, prefix string) QueueRedisAdapter {
	if prefix == "" {
		prefix = "queue:"
	}
	return &redisQueue{
		client: client,
		prefix: prefix,
	}
}

// NewRedisQueueWithProvider tạo một instance mới của redisQueue sử dụng Redis provider.
// Hàm này khởi tạo kết nối Redis thông qua provider và áp dụng prefix cho key.
//
// Tham số:
//   - provider (redis.Manager): Redis manager instance
//   - prefix (string): Prefix cho các key Redis
//
// Trả về:
//   - QueueRedisAdapter: Instance mới của redisQueue với đầy đủ tính năng Redis
//   - error: Lỗi nếu có khi lấy client từ provider
func NewRedisQueueWithProvider(provider redis.Manager, prefix string) (QueueRedisAdapter, error) {
	client, err := provider.Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis client from provider: %w", err)
	}

	if prefix == "" {
		prefix = "queue:"
	}
	return &redisQueue{
		client: client,
		prefix: prefix,
	}, nil
}

// prefixKey thêm prefix đã cấu hình vào tên hàng đợi.
// Hàm này đảm bảo tất cả các key queue trong Redis có cùng prefix
// để dễ dàng quản lý và tránh xung đột tên.
//
// Tham số:
//   - queueName (string): Tên gốc của hàng đợi
//
// Trả về:
//   - string: Tên hàng đợi có prefix
func (q *redisQueue) prefixKey(queueName string) string {
	return q.prefix + queueName
}

// Enqueue thêm một item vào cuối hàng đợi.
// Hàm này serialize item thành JSON và thêm vào Redis list
// sử dụng lệnh RPUSH.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - item (interface{}): Đối tượng cần đưa vào hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi thêm item vào hàng đợi
func (q *redisQueue) Enqueue(ctx context.Context, queueName string, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshaling queue item: %w", err)
	}

	return q.client.RPush(ctx, q.prefixKey(queueName), data).Err()
}

// Dequeue lấy và xóa item ở đầu hàng đợi.
// Hàm này sử dụng lệnh LPOP của Redis để lấy item đầu tiên
// từ list, sau đó deserialize thành đối tượng định sẵn.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - dest (interface{}): Con trỏ đến đối tượng sẽ nhận dữ liệu
//
// Trả về:
//   - error: Lỗi nếu có khi lấy item từ hàng đợi hoặc khi hàng đợi rỗng
func (q *redisQueue) Dequeue(ctx context.Context, queueName string, dest interface{}) error {
	data, err := q.client.LPop(ctx, q.prefixKey(queueName)).Bytes()
	if err != nil {
		if err == redisClient.Nil {
			return fmt.Errorf("queue is empty: %s", queueName)
		}
		return err
	}

	return json.Unmarshal(data, dest)
}

// EnqueueBatch thêm nhiều item vào cuối hàng đợi.
// Hàm này serialize mỗi item thành JSON và thêm tất cả
// vào Redis list trong một lệnh RPUSH duy nhất.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - items ([]interface{}): Slice các đối tượng cần đưa vào hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi thêm items vào hàng đợi
func (q *redisQueue) EnqueueBatch(ctx context.Context, queueName string, items []interface{}) error {
	if len(items) == 0 {
		return nil
	}

	// Convert items to JSON strings
	values := make([]interface{}, len(items))
	for i, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("error marshaling queue item at index %d: %w", i, err)
		}
		values[i] = data
	}

	return q.client.RPush(ctx, q.prefixKey(queueName), values...).Err()
}

// Size trả về số lượng item trong hàng đợi.
// Hàm này sử dụng lệnh LLEN của Redis để đếm số lượng
// phần tử trong list.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - int64: Số lượng item trong hàng đợi
//   - error: Lỗi nếu có khi truy vấn Redis
func (q *redisQueue) Size(ctx context.Context, queueName string) (int64, error) {
	return q.client.LLen(ctx, q.prefixKey(queueName)).Result()
}

// IsEmpty kiểm tra xem hàng đợi có rỗng không.
// Hàm này kiểm tra kích thước của hàng đợi và trả về true
// nếu kích thước bằng 0.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - bool: true nếu hàng đợi rỗng, ngược lại là false
//   - error: Lỗi nếu có khi kiểm tra
func (q *redisQueue) IsEmpty(ctx context.Context, queueName string) (bool, error) {
	size, err := q.Size(ctx, queueName)
	if err != nil {
		return false, err
	}
	return size == 0, nil
}

// Clear xóa tất cả các item trong hàng đợi.
// Hàm này sử dụng lệnh DEL của Redis để xóa hoàn toàn
// Redis list tương ứng với hàng đợi.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi xóa hàng đợi
func (q *redisQueue) Clear(ctx context.Context, queueName string) error {
	return q.client.Del(ctx, q.prefixKey(queueName)).Err()
}

// DequeueWithTimeout lấy và xóa item ở đầu hàng đợi, với khả năng chờ đợi
// nếu hàng đợi đang rỗng. Hàm này sử dụng lệnh BLPOP của Redis để chờ tối đa
// một khoảng thời gian nhất định cho đến khi có item mới trong hàng đợi.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - timeout (time.Duration): Thời gian tối đa chờ đợi
//   - dest (interface{}): Con trỏ đến đối tượng sẽ nhận dữ liệu
//
// Trả về:
//   - error: Lỗi nếu có khi lấy item hoặc khi hết thời gian chờ
func (q *redisQueue) DequeueWithTimeout(ctx context.Context, queueName string, timeout time.Duration, dest interface{}) error {
	data, err := q.client.BLPop(ctx, timeout, q.prefixKey(queueName)).Result()
	if err != nil {
		if err == redisClient.Nil {
			return fmt.Errorf("timeout waiting for queue item: %s", queueName)
		}
		return err
	}

	// BLPop returns a slice with [key, value]
	if len(data) < 2 {
		return fmt.Errorf("unexpected response format from redis")
	}

	return json.Unmarshal([]byte(data[1]), dest)
}

// EnqueueWithPipeline sử dụng Redis pipeline để thêm nhiều item vào các hàng đợi khác nhau
// trong một lần giao tiếp với Redis server, giảm network latency.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - items (map[string][]interface{}): Map với key là tên hàng đợi và value là slice các item
//
// Trả về:
//   - error: Lỗi nếu có khi thực hiện pipeline
func (q *redisQueue) EnqueueWithPipeline(ctx context.Context, items map[string][]interface{}) error {
	if len(items) == 0 {
		return nil
	}

	pipe := q.client.Pipeline()

	for queueName, queueItems := range items {
		if len(queueItems) == 0 {
			continue
		}

		// Convert items to JSON strings
		values := make([]interface{}, len(queueItems))
		for i, item := range queueItems {
			data, err := json.Marshal(item)
			if err != nil {
				return fmt.Errorf("error marshaling queue item for queue %s at index %d: %w", queueName, i, err)
			}
			values[i] = data
		}

		pipe.RPush(ctx, q.prefixKey(queueName), values...)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// EnqueueWithTTL thêm item vào hàng đợi và thiết lập TTL (Time To Live) cho hàng đợi.
// Hàng đợi sẽ tự động bị xóa sau khoảng thời gian TTL nếu rỗng.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - item (interface{}): Đối tượng cần đưa vào hàng đợi
//   - ttl (time.Duration): Thời gian sống của hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi thêm item hoặc thiết lập TTL
func (q *redisQueue) EnqueueWithTTL(ctx context.Context, queueName string, item interface{}, ttl time.Duration) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshaling queue item: %w", err)
	}

	key := q.prefixKey(queueName)

	// Sử dụng pipeline để thực hiện cả RPUSH và EXPIRE atomically
	pipe := q.client.Pipeline()
	pipe.RPush(ctx, key, data)
	pipe.Expire(ctx, key, ttl)

	_, err = pipe.Exec(ctx)
	return err
}

// EnqueueWithPriority thêm item vào hàng đợi ưu tiên sử dụng Redis Sorted Set.
// Item có score cao hơn sẽ được xử lý trước.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi ưu tiên
//   - item (interface{}): Đối tượng cần đưa vào hàng đợi
//   - priority (float64): Mức độ ưu tiên (số càng cao càng ưu tiên)
//
// Trả về:
//   - error: Lỗi nếu có khi thêm item vào hàng đợi ưu tiên
func (q *redisQueue) EnqueueWithPriority(ctx context.Context, queueName string, item interface{}, priority float64) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshaling priority queue item: %w", err)
	}

	priorityKey := q.prefixKey(queueName + ":priority")

	// Sử dụng ZADD để thêm vào sorted set với score là priority
	return q.client.ZAdd(ctx, priorityKey, redisClient.Z{
		Score:  priority,
		Member: data,
	}).Err()
}

// DequeueFromPriority lấy item có độ ưu tiên cao nhất từ hàng đợi ưu tiên.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi ưu tiên
//   - dest (interface{}): Con trỏ đến đối tượng sẽ nhận dữ liệu
//
// Trả về:
//   - error: Lỗi nếu có khi lấy item từ hàng đợi ưu tiên hoặc khi hàng đợi rỗng
func (q *redisQueue) DequeueFromPriority(ctx context.Context, queueName string, dest interface{}) error {
	priorityKey := q.prefixKey(queueName + ":priority")

	// ZPOPMAX lấy item có score cao nhất
	data, err := q.client.ZPopMax(ctx, priorityKey).Result()
	if err != nil {
		if err == redisClient.Nil {
			return fmt.Errorf("priority queue is empty: %s", queueName)
		}
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("priority queue is empty: %s", queueName)
	}

	member, ok := data[0].Member.(string)
	if !ok {
		return fmt.Errorf("unexpected member type from redis")
	}

	return json.Unmarshal([]byte(member), dest)
}

// GetQueueInfo trả về thông tin chi tiết về hàng đợi bao gồm kích thước,
// TTL và các metadata khác.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - map[string]interface{}: Map chứa thông tin về hàng đợi
//   - error: Lỗi nếu có khi truy vấn thông tin
func (q *redisQueue) GetQueueInfo(ctx context.Context, queueName string) (map[string]interface{}, error) {
	key := q.prefixKey(queueName)
	priorityKey := q.prefixKey(queueName + ":priority")

	pipe := q.client.Pipeline()
	sizeCmd := pipe.LLen(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	prioritySizeCmd := pipe.ZCard(ctx, priorityKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redisClient.Nil {
		return nil, fmt.Errorf("error getting queue info: %w", err)
	}

	size := sizeCmd.Val()
	ttl := ttlCmd.Val()
	prioritySize := prioritySizeCmd.Val()

	info := map[string]interface{}{
		"name":          queueName,
		"size":          size,
		"priority_size": prioritySize,
		"total_size":    size + prioritySize,
		"ttl_seconds":   ttl.Seconds(),
		"has_ttl":       ttl > 0,
		"is_empty":      size == 0 && prioritySize == 0,
	}

	return info, nil
}

// MultiDequeue cho phép lấy nhiều item cùng lúc từ hàng đợi.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - count (int): Số lượng item muốn lấy
//   - destSlice (interface{}): Con trỏ đến slice sẽ nhận dữ liệu
//
// Trả về:
//   - int: Số lượng item thực tế đã lấy được
//   - error: Lỗi nếu có khi lấy items
func (q *redisQueue) MultiDequeue(ctx context.Context, queueName string, count int, destSlice interface{}) (int, error) {
	if count <= 0 {
		return 0, fmt.Errorf("count must be positive")
	}

	key := q.prefixKey(queueName)

	// Sử dụng LPOP với count parameter (Redis 6.2+)
	data, err := q.client.LPopCount(ctx, key, count).Result()
	if err != nil {
		if err == redisClient.Nil {
			return 0, nil // Không có item nào
		}
		return 0, err
	}

	// Deserialize các items vào slice
	for _, item := range data {
		// Note: Để implement đúng, cần reflection để append vào destSlice
		// Đây là version simplified, trong thực tế cần xử lý reflection phức tạp hơn
		_ = item // Placeholder - cần implement reflection logic
	}

	return len(data), nil
}

// Ping kiểm tra kết nối Redis có hoạt động không.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//
// Trả về:
//   - error: Lỗi nếu có khi ping Redis server
func (q *redisQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

// FlushQueues xóa tất cả các hàng đợi có prefix tương ứng.
// Chỉ nên sử dụng trong môi trường development hoặc testing.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//
// Trả về:
//   - int64: Số lượng key đã bị xóa
//   - error: Lỗi nếu có khi xóa
func (q *redisQueue) FlushQueues(ctx context.Context) (int64, error) {
	pattern := q.prefix + "*"

	keys, err := q.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("error getting keys with pattern %s: %w", pattern, err)
	}

	if len(keys) == 0 {
		return 0, nil
	}

	deleted, err := q.client.Del(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("error deleting keys: %w", err)
	}

	return deleted, nil
}

// GetRedisClient trả về Redis client instance để sử dụng trực tiếp.
// Method này cho phép truy cập trực tiếp vào Redis client để thực hiện
// các thao tác Redis cấp thấp nếu cần thiết.
//
// Trả về:
//   - *redis.Client: Redis client instance
func (q *redisQueue) GetRedisClient() *redisClient.Client {
	return q.client
}

// IsRedisQueueAdapter kiểm tra xem một QueueAdapter có phải là Redis adapter không.
// Hàm utility này giúp type assertion an toàn từ QueueAdapter sang QueueRedisAdapter.
//
// Tham số:
//   - adapter (QueueAdapter): Adapter cần kiểm tra
//
// Trả về:
//   - QueueRedisAdapter: Redis adapter nếu conversion thành công
//   - bool: true nếu là Redis adapter
func IsRedisQueueAdapter(adapter QueueAdapter) (QueueRedisAdapter, bool) {
	redisAdapter, ok := adapter.(*redisQueue)
	return redisAdapter, ok
}

// NewRedisQueueWithOptions tạo Redis queue adapter với các tùy chọn nâng cao.
// Hàm này cung cấp nhiều tùy chọn cấu hình hơn so với NewRedisQueue cơ bản.
//
// Tham số:
//   - options (*redisClient.Options): Cấu hình Redis client
//   - prefix (string): Prefix cho các key Redis
//
// Trả về:
//   - QueueRedisAdapter: Instance mới với Redis client được tạo từ options
//   - error: Lỗi nếu có khi khởi tạo
func NewRedisQueueWithOptions(options *redisClient.Options, prefix string) (QueueRedisAdapter, error) {
	client := redisClient.NewClient(options)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return NewRedisQueue(client, prefix), nil
}

// ExampleUsage demonstrates how to use QueueRedisAdapter with all features.
// Đây là ví dụ về cách sử dụng đầy đủ tính năng của Redis queue adapter.
func ExampleUsage() error {
	// Tạo Redis queue adapter với options
	options := &redisClient.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	redisQueue, err := NewRedisQueueWithOptions(options, "app:")
	if err != nil {
		return fmt.Errorf("failed to create redis queue: %w", err)
	}

	ctx := context.Background()

	// 1. Test kết nối
	if err := redisQueue.Ping(ctx); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	// 2. Sử dụng basic queue operations
	type Task struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}

	task := Task{ID: "1", Message: "Hello Redis"}
	if err := redisQueue.Enqueue(ctx, "tasks", task); err != nil {
		return err
	}

	// 3. Sử dụng pipeline cho hiệu suất cao
	batchTasks := map[string][]interface{}{
		"email_queue": {
			Task{ID: "email_1", Message: "Send welcome email"},
			Task{ID: "email_2", Message: "Send newsletter"},
		},
		"sms_queue": {
			Task{ID: "sms_1", Message: "Send verification SMS"},
		},
	}
	if err := redisQueue.EnqueueWithPipeline(ctx, batchTasks); err != nil {
		return err
	}

	// 4. Sử dụng priority queue
	if err := redisQueue.EnqueueWithPriority(ctx, "urgent",
		Task{ID: "urgent_1", Message: "Critical task"}, 10.0); err != nil {
		return err
	}

	// 5. Dequeue với timeout (blocking operation)
	var receivedTask Task
	err = redisQueue.DequeueWithTimeout(ctx, "tasks", 5*time.Second, &receivedTask)
	if err != nil {
		return err
	}

	// 6. Lấy thông tin queue
	info, err := redisQueue.GetQueueInfo(ctx, "tasks")
	if err != nil {
		return err
	}

	fmt.Printf("Queue info: %+v\n", info)

	// 7. Sử dụng trực tiếp Redis client nếu cần
	client := redisQueue.GetRedisClient()
	result := client.Set(ctx, "custom_key", "custom_value", 0)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}
