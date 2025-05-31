package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MemoryQueue triển khai interface QueueAdapter bằng cách sử dụng bộ nhớ trong.
// Struct này cung cấp một triển khai đơn giản của queue cho các trường hợp
// không yêu cầu persistence hoặc cho môi trường kiểm thử.
type memoryQueue struct {
	queues map[string][][]byte
	prefix string
	mutex  sync.RWMutex
}

// NewMemoryQueue tạo một instance mới của memoryQueue.
// Hàm này khởi tạo một map để lưu trữ các hàng đợi và thiết lập prefix.
//
// Trả về:
//   - QueueAdapter: Instance mới của memoryQueue
func NewMemoryQueue(prefix string) QueueAdapter {
	if prefix == "" {
		prefix = "queue:"
	}
	return &memoryQueue{
		queues: make(map[string][][]byte),
		prefix: prefix,
	}
}

// prefixKey thêm prefix đã cấu hình vào tên hàng đợi.
// Hàm này đảm bảo tất cả các key queue có cùng prefix
// để dễ dàng quản lý và tránh xung đột tên.
//
// Tham số:
//   - queueName (string): Tên gốc của hàng đợi
//
// Trả về:
//   - string: Tên hàng đợi có prefix
func (q *memoryQueue) prefixKey(queueName string) string {
	return q.prefix + queueName
}

// Enqueue thêm một item vào cuối hàng đợi.
// Hàm này serialize item thành JSON và thêm vào slice
// tương ứng với hàng đợi, sử dụng khóa mutex để đảm bảo an toàn
// trong môi trường đa luồng.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - item (interface{}): Đối tượng cần đưa vào hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi thêm item vào hàng đợi
func (q *memoryQueue) Enqueue(ctx context.Context, queueName string, item interface{}) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("error marshaling queue item: %w", err)
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	key := q.prefixKey(queueName)
	if _, exists := q.queues[key]; !exists {
		q.queues[key] = make([][]byte, 0)
	}
	q.queues[key] = append(q.queues[key], data)
	return nil
}

// Dequeue lấy và xóa item ở đầu hàng đợi.
// Hàm này lấy item đầu tiên từ slice tương ứng với hàng đợi,
// sử dụng khóa mutex để đảm bảo an toàn trong môi trường đa luồng,
// sau đó deserialize thành đối tượng định sẵn.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - dest (interface{}): Con trỏ đến đối tượng sẽ nhận dữ liệu
//
// Trả về:
//   - error: Lỗi nếu có khi lấy item từ hàng đợi hoặc khi hàng đợi rỗng
func (q *memoryQueue) Dequeue(ctx context.Context, queueName string, dest interface{}) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	key := q.prefixKey(queueName)
	queue, exists := q.queues[key]
	if !exists || len(queue) == 0 {
		return fmt.Errorf("queue is empty: %s", queueName)
	}

	// Get the first item
	data := queue[0]
	// Remove it from the queue
	q.queues[key] = queue[1:]

	return json.Unmarshal(data, dest)
}

// EnqueueBatch thêm nhiều item vào cuối hàng đợi.
// Hàm này serialize mỗi item thành JSON và thêm tất cả vào slice
// tương ứng với hàng đợi, sử dụng khóa mutex để đảm bảo an toàn
// trong môi trường đa luồng.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//   - items ([]interface{}): Slice các đối tượng cần đưa vào hàng đợi
//
// Trả về:
//   - error: Lỗi nếu có khi thêm items vào hàng đợi
func (q *memoryQueue) EnqueueBatch(ctx context.Context, queueName string, items []interface{}) error {
	if len(items) == 0 {
		return nil
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	key := q.prefixKey(queueName)
	if _, exists := q.queues[key]; !exists {
		q.queues[key] = make([][]byte, 0, len(items))
	}

	for i, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("error marshaling queue item at index %d: %w", i, err)
		}
		q.queues[key] = append(q.queues[key], data)
	}

	return nil
}

// Size trả về số lượng item trong hàng đợi.
// Hàm này đếm số lượng phần tử trong slice tương ứng với hàng đợi,
// sử dụng khóa đọc mutex để đảm bảo an toàn trong môi trường đa luồng.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - int64: Số lượng item trong hàng đợi
//   - error: Luôn là nil cho implementation này
func (q *memoryQueue) Size(ctx context.Context, queueName string) (int64, error) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	key := q.prefixKey(queueName)
	queue, exists := q.queues[key]
	if !exists {
		return 0, nil
	}
	return int64(len(queue)), nil
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
func (q *memoryQueue) IsEmpty(ctx context.Context, queueName string) (bool, error) {
	size, err := q.Size(ctx, queueName)
	if err != nil {
		return false, err
	}
	return size == 0, nil
}

// Clear xóa tất cả các item trong hàng đợi.
// Hàm này xóa hoàn toàn slice tương ứng với hàng đợi khỏi map,
// sử dụng khóa mutex để đảm bảo an toàn trong môi trường đa luồng.
//
// Tham số:
//   - ctx (context.Context): Context cho request
//   - queueName (string): Tên của hàng đợi
//
// Trả về:
//   - error: Luôn là nil cho implementation này
func (q *memoryQueue) Clear(ctx context.Context, queueName string) error {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	key := q.prefixKey(queueName)
	delete(q.queues, key)
	return nil
}
