// Package adapter cung cấp các triển khai khác nhau cho queue backend.
package adapter

import (
	"context"
)

// QueueAdapter định nghĩa các hoạt động có sẵn cho hàng đợi.
// Interface này tách biệt các hoạt động hàng đợi khỏi implementation cụ thể,
// cho phép thay đổi backend mà không ảnh hưởng đến code sử dụng.
type QueueAdapter interface {
	// Enqueue thêm một item vào cuối hàng đợi.
	Enqueue(ctx context.Context, queueName string, item interface{}) error

	// Dequeue lấy và xóa item ở đầu hàng đợi.
	Dequeue(ctx context.Context, queueName string, dest interface{}) error

	// EnqueueBatch thêm nhiều item vào cuối hàng đợi.
	EnqueueBatch(ctx context.Context, queueName string, items []interface{}) error

	// Size trả về số lượng item trong hàng đợi.
	Size(ctx context.Context, queueName string) (int64, error)

	// IsEmpty kiểm tra xem hàng đợi có rỗng không.
	IsEmpty(ctx context.Context, queueName string) (bool, error)

	// Clear xóa tất cả các item trong hàng đợi.
	Clear(ctx context.Context, queueName string) error
}
