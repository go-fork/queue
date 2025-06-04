# Ghi chú phiên bản - v0.1.1

## Tổng quan
Phiên bản này tập trung vào việc đơn giản hóa API và cải thiện trải nghiệm phát triển bằng cách hợp lý hóa các mẫu truy cập dịch vụ trong gói queue. Điểm nổi bật chính là việc loại bỏ mẫu `*.manager` để thay thế bằng cách tiếp cận thống nhất dựa trên manager.

## Có gì mới
### 🚀 Tính năng
- **Truy cập API đơn giản**: Hợp lý hóa truy cập dịch vụ thông qua giao diện manager
  - Điểm truy cập duy nhất qua `container.MustMake("queue")` trả về giao diện Manager
  - Manager cung cấp các phương thức `.Client()` và `.Server()` để truy cập dịch vụ
  - Loại bỏ nhu cầu đăng ký nhiều dịch vụ (`queue.client`, `queue.server`, v.v.)
  - Cải thiện trải nghiệm phát triển với thiết kế API trực quan hơn

### 🐛 Sửa lỗi
- **Tính nhất quán API**: Cải thiện tính nhất quán truy cập dịch vụ trong toàn bộ gói queue
  - Loại bỏ nhiều mẫu đăng ký dịch vụ để ưu tiên truy cập dựa trên manager duy nhất
  - Sửa lỗi không nhất quán trong tài liệu giữa các phương thức truy cập dịch vụ khác nhau
  - Chuẩn hóa tất cả ví dụ để sử dụng mẫu truy cập thống nhất mới

### 🔧 Cải tiến
- **Chất lượng code**: Codebase sạch hơn và dễ bảo trì hơn
  - Loại bỏ các đăng ký dịch vụ dư thừa
  - Đơn giản hóa các mẫu dependency injection
  - Tách biệt mối quan tâm tốt hơn trong service provider

### 📚 Tài liệu
- Cập nhật README.md với các mẫu truy cập dịch vụ mới
- Cập nhật tất cả file tài liệu (docs/overview.md, docs/provider.md) để nhất quán
- Thêm ví dụ migration hiển thị mẫu trước/sau
- Tài liệu CHANGELOG.md toàn diện

## Thay đổi phá vỡ tương thích
### ⚠️ Ghi chú quan trọng
- **Thay đổi mẫu truy cập dịch vụ**: Cách bạn truy cập các dịch vụ queue đã thay đổi
  - **Cũ**: `container.MustMake("queue.client").(queue.Client)`
  - **Mới**: `manager := container.MustMake("queue").(queue.Manager); client := manager.Client()`
  - **Cũ**: `container.MustMake("queue.server").(queue.Server)`
  - **Mới**: `manager := container.MustMake("queue").(queue.Manager); server := manager.Server()`

## Hướng dẫn Migration
Xem [MIGRATION.md](./MIGRATION.md) để biết hướng dẫn migration chi tiết.

## Phụ thuộc
### Đã cập nhật
- `go.fork.vn/config`: v0.1.0 → v0.1.3 (Thêm ServiceProvider với các phương thức Boot, Requires, Providers)
- `go.fork.vn/di`: v0.1.0 → v0.1.3 (Nâng cao giao diện ServiceProvider với các phương thức mới)
- `go.fork.vn/redis`: v0.1.0 → v0.1.2 (Cập nhật tương thích giao diện ServiceProvider)
- `go.fork.vn/scheduler`: v0.1.0 → v0.1.1 (Cách tiếp cận dựa trên cấu hình với hỗ trợ auto-start)

### Đã thêm
- Không có phụ thuộc mới nào được thêm

### Removed
- No dependencies removed

## Performance
- API simplification reduces service lookup overhead
- Cleaner code paths for better maintainability

## Security
- Updated dependencies with latest security patches
- No specific security vulnerabilities addressed in this release

## Testing
- All existing tests updated to use new API pattern
- Maintained 100% test coverage
- Added consistency verification tests

## Contributors
Thanks to all contributors who made this release possible:
- Development team for API design improvements
- Documentation reviewers for comprehensive updates

## Download
- Source code: [go.fork.vn/queue@v0.1.1]
- Documentation: [pkg.go.dev/go.fork.vn/queue@v0.1.1]

---
Release Date: 2025-06-05
