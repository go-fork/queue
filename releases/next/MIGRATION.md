# Hướng dẫn Migration - v0.1.1

## Tổng quan
Hướng dẫn này giúp bạn migrate từ v0.1.0 lên v0.1.1 của gói go-fork/queue. Phiên bản này giới thiệu mẫu API đơn giản hóa để loại bỏ mẫu `*.manager` và ưu tiên truy cập dịch vụ dựa trên package trực tiếp.

## Điều kiện tiên quyết
- Go 1.23 trở lên
- Đã cài đặt go-fork/queue v0.1.0

## Checklist Migration nhanh
- [ ] Cập nhật các mẫu truy cập dịch vụ để sử dụng giao diện manager
- [ ] Loại bỏ đăng ký container trực tiếp `queue.client` và `queue.server`
- [ ] Cập nhật code dependency injection để sử dụng mẫu mới
- [ ] Chạy test để đảm bảo tương thích
- [ ] Cập nhật tham chiếu tài liệu

## Thay đổi phá vỡ tương thích

### Đơn giản hóa API
Thay đổi phá vỡ tương thích chính trong v0.1.1 là việc loại bỏ mẫu `*.manager` và đơn giản hóa truy cập dịch vụ.

#### Thay đổi mẫu truy cập dịch vụ

**Mẫu cũ (v0.1.0)**:
```go
// Lấy queue client
client := container.MustMake("queue.client").(queue.Client)

// Lấy queue server  
server := container.MustMake("queue.server").(queue.Server)
```

**Mẫu mới (v0.1.1)**:
```go
// Lấy queue manager trước, sau đó truy cập dịch vụ
manager := container.MustMake("queue").(queue.Manager)

// Lấy queue client thông qua manager
client := manager.Client()

// Lấy queue server thông qua manager
server := manager.Server()
```

#### Đăng ký đã loại bỏ
Các đăng ký dịch vụ sau đã được loại bỏ khỏi ServiceProvider:
- `queue.manager` - Không còn được đăng ký như một dịch vụ riêng biệt
- Truy cập trực tiếp đến `queue.client` và `queue.server` giờ được thực hiện thông qua giao diện manager

### Thay đổi Service Provider
Nếu bạn đang đăng ký trực tiếp queue service provider, giao diện vẫn giữ nguyên nhưng các đăng ký nội bộ đã được đơn giản hóa:

```go
// Điều này tiếp tục hoạt động theo cách tương tự
provider := queue.NewServiceProvider()
container.Register(provider)

// Nhưng nội bộ, chỉ "queue" được đăng ký, không phải "queue.manager"
```

## Migration từng bước

### Bước 1: Cập nhật Dependencies
```bash
go get go.fork.vn/queue@v0.1.1
go mod tidy
```

### Bước 2: Cập nhật code truy cập dịch vụ
Thay thế tất cả các instance truy cập dịch vụ trực tiếp bằng truy cập dựa trên manager:

```go
// Trước: Truy cập client trực tiếp
client := container.MustMake("queue.client").(queue.Client)
client.Publish("my-job", map[string]interface{}{"data": "value"})

// Sau: Truy cập client dựa trên manager
manager := container.MustMake("queue").(queue.Manager)
client := manager.Client()
client.Publish("my-job", map[string]interface{}{"data": "value"})
```

```go
// Trước: Truy cập server trực tiếp
server := container.MustMake("queue.server").(queue.Server)
server.Start()

// Sau: Truy cập server dựa trên manager
manager := container.MustMake("queue").(queue.Manager)
server := manager.Server()
server.Start()
```

### Bước 3: Loại bỏ đăng ký Container cũ
Nếu bạn có code tùy chỉnh nào đó cố gắng truy cập `queue.manager`, `queue.client`, hoặc `queue.server` trực tiếp, hãy cập nhật nó để sử dụng mẫu mới.

### Bước 4: Chạy Tests
```bash
go test ./...
```

## Các vấn đề thường gặp và Giải pháp

### Vấn đề 1: Không tìm thấy dịch vụ
**Vấn đề**: `service 'queue.client' not found in container`  
**Giải pháp**: Sử dụng truy cập dựa trên manager: `container.MustMake("queue").(queue.Manager).Client()`

### Vấn đề 2: Không tìm thấy dịch vụ - queue.server
**Vấn đề**: `service 'queue.server' not found in container`  
**Giải pháp**: Sử dụng truy cập dựa trên manager: `container.MustMake("queue").(queue.Manager).Server()`

### Vấn đề 3: Không tìm thấy dịch vụ - queue.manager
**Vấn đề**: `service 'queue.manager' not found in container`  
**Giải pháp**: Sử dụng truy cập queue trực tiếp: `container.MustMake("queue").(queue.Manager)`

## Lợi ích của mẫu mới

1. **API đơn giản**: Điểm truy cập duy nhất thông qua giao diện manager
2. **Đóng gói tốt hơn**: Các dịch vụ được truy cập thông qua manager của chúng thay vì trực tiếp từ container
3. **Container sạch hơn**: Ít dịch vụ được đăng ký trong dependency injection container
4. **Mẫu nhất quán**: Phù hợp với best practices của Go cho quản lý dịch vụ

## Nhận trợ giúp
- Kiểm tra [tài liệu](https://pkg.go.dev/go.fork.vn/queue@v0.1.1)
- Tìm kiếm [các issues hiện có](https://github.com/go-fork/queue/issues)
- Tạo [issue mới](https://github.com/go-fork/queue/issues/new) nếu cần

## Hướng dẫn Rollback
Nếu bạn cần rollback:

```bash
go get go.fork.vn/queue@previous-version
go mod tidy
```

Thay thế `previous-version` bằng tag phiên bản trước đó của bạn.

---
**Cần trợ giúp?** Đừng ngần ngại mở issue hoặc discussion trên GitHub.
