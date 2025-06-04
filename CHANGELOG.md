# Changelog

## [Chưa phát hành]

### Đã thêm
- **Truy cập API đơn giản**: Hợp lý hóa truy cập dịch vụ thông qua giao diện manager
  - Điểm truy cập duy nhất qua `container.MustMake("queue")` trả về giao diện Manager
  - Manager cung cấp các phương thức `.Client()` và `.Server()` để truy cập dịch vụ
  - Loại bỏ nhu cầu đăng ký nhiều dịch vụ (`queue.client`, `queue.server`, v.v.)
  - Cải thiện trải nghiệm phát triển với thiết kế API trực quan hơn
- **GitHub Workflows**: Pipeline CI/CD hoàn chỉnh với automated testing, release management và dependency updates
  - CI workflow cho automated testing và build verification
  - Release workflow cho quản lý phiên bản tự động và GitHub releases
  - Dependency update workflow với tự động tạo PR
- **Cấu hình GitHub**: File CODEOWNERS cho phân công code review
- **Issue Templates**: 
  - Template báo cáo bug với thu thập thông tin có cấu trúc
  - Template yêu cầu tính năng cho đề xuất cải tiến
- **Pull Request Template**: Format mô tả PR chuẩn hóa
- **Cấu hình Funding**: Thông tin GitHub sponsors và funding
- **Dependabot**: Cập nhật dependency tự động với giám sát bảo mật
- **Scripts tự động hóa Release**: 
  - Script archive release cho quản lý phiên bản
  - Script tạo template release cho phiên bản nhất quán
  - Tài liệu cho việc sử dụng automation scripts

### Đã thay đổi
- **Đơn giản hóa API**: Loại bỏ mẫu `*.manager` từ gói queue để truy cập dịch vụ sạch hơn
  - Loại bỏ đăng ký instance `queue.manager` từ ServiceProvider
  - Cập nhật mẫu truy cập dịch vụ để sử dụng các phương thức `manager.Client()` và `manager.Server()`
  - Thay thế `container.MustMake("queue.client")` bằng `manager := container.MustMake("queue"); client := manager.Client()`
  - Thay thế `container.MustMake("queue.server")` bằng `manager := container.MustMake("queue"); server := manager.Server()`
  - Cập nhật tất cả tài liệu và ví dụ để phản ánh mẫu truy cập mới
- **Nâng cấp Dependencies**: Nâng cấp tất cả direct dependencies lên phiên bản mới nhất với tương thích breaking changes
  - `go.fork.vn/config`: v0.1.0 → v0.1.3 (Thêm ServiceProvider với các phương thức Boot, Requires, Providers)
  - `go.fork.vn/di`: v0.1.0 → v0.1.3 (Nâng cao giao diện ServiceProvider với các phương thức mới)
  - `go.fork.vn/redis`: v0.1.0 → v0.1.2 (Cập nhật tương thích giao diện ServiceProvider)
  - `go.fork.vn/scheduler`: v0.1.0 → v0.1.1 (Cách tiếp cận dựa trên cấu hình với hỗ trợ auto-start)

### Đã sửa
- **Tính nhất quán API**: Cải thiện tính nhất quán truy cập dịch vụ trong toàn bộ gói queue
  - Loại bỏ nhiều mẫu đăng ký dịch vụ để ưu tiên truy cập dựa trên manager duy nhất
  - Sửa lỗi không nhất quán trong tài liệu giữa các phương thức truy cập dịch vụ khác nhau
  - Chuẩn hóa tất cả ví dụ để sử dụng mẫu truy cập thống nhất mới
- **Tương thích Breaking Changes**: Cập nhật code để hoạt động với các giao diện ServiceProvider mới
  - Sửa các implementation ServiceProvider để bao gồm các phương thức bắt buộc mới (Boot, Requires, Providers)
  - Cập nhật việc sử dụng DI container để phù hợp với thông số giao diện mới
  - Duy trì backward compatibility khi có thể

### Dependencies
- **Core Dependencies**:
  - `github.com/go-redis/redismock/v9` v9.2.0 - Redis mocking for testing
  - `github.com/redis/go-redis/v9` v9.9.0 - Redis Go client
  - `github.com/stretchr/testify` v1.10.0 - Testing framework
  - `go.fork.vn/config` v0.1.3 - Configuration management (upgraded)
  - `go.fork.vn/di` v0.1.3 - Dependency injection (upgraded)
  - `go.fork.vn/redis` v0.1.2 - Redis provider integration (upgraded)
  - `go.fork.vn/scheduler` v0.1.1 - Task scheduling (upgraded)

- **Indirect Dependencies**:
  - `github.com/cespare/xxhash/v2` v2.3.0 - Hàm hash
  - `github.com/fsnotify/fsnotify` v1.9.0 - Thông báo file system
  - `github.com/go-co-op/gocron` v1.37.0 - Lập lịch cron job
  - `github.com/google/uuid` v1.6.0 - Tạo UUID
  - `github.com/spf13/viper` v1.20.1 - Quản lý cấu hình
  - `golang.org/x/sys` v0.33.0 - System calls
  - `golang.org/x/text` v0.25.0 - Xử lý văn bản

## v0.1.0 - 2025-05-31

### Đã thêm
- **Hệ thống quản lý Queue**: Hệ thống quản lý queue và xử lý task toàn diện cho ứng dụng Go
- **Nhiều Queue Adapters**: Hỗ trợ cho Memory, Redis, và các queue dựa trên Redis Provider
- **Tích hợp Redis**: Các tính năng Redis nâng cao với priority queues, hỗ trợ TTL, và batch operations
- **Priority Queues**: Implementation Redis Sorted Sets cho việc ưu tiên task
- **Batch Operations**: Hỗ trợ pipeline cho các tình huống throughput cao
- **Hỗ trợ TTL**: Các task tạm thời với tự động hết hạn
- **Giám sát Queue**: Thống kê queue theo thời gian thực và giám sát health
- **Tích hợp DI**: Tích hợp liền mạch với Dependency Injection container
- **Dựa trên cấu hình**: Cấu hình linh hoạt thông qua quản lý config
- **Thread-Safe**: Các operations queue an toàn đồng thời và xử lý task
- **Quản lý tài nguyên**: Dọn dẹp tự động và xử lý tài nguyên thích hợp
- **Khả năng phục hồi lỗi**: Xử lý lỗi mạnh mẽ và cơ chế khôi phục
- **Health Checks**: Giám sát health tích hợp và kiểm tra kết nối
- **Development Tools**: Tiện ích để queue flushing và debugging
- **Tối ưu hóa bộ nhớ**: Sử dụng bộ nhớ hiệu quả cho các tình huống khối lượng lớn
- **Sẵn sàng sản xuất**: Templates cấu hình nâng cao và best practices
- **Worker System**: Implementation worker hoàn chỉnh với tích hợp scheduler
- **Delayed Tasks**: Hỗ trợ cho việc thực thi task được trì hoãn và lập lịch
- **Batch Processing**: Khả năng batch processing nâng cao cho các tình huống throughput cao
- **Message Filtering**: Khả năng lọc message nâng cao
- **Dead Letter Queue**: Hỗ trợ xử lý message thất bại
- **Task Status Tracking**: Quản lý trạng thái task toàn diện trong môi trường phân tán
- **Retry Logic**: Các chiến lược retry có thể cấu hình với cơ chế backoff
- **Asynchronous Processing**: Xử lý message bất đồng bộ hiệu quả
- **Task Serialization**: Tuần tự hóa và giải tuần tự hóa payload mạnh mẽ
- **Server Component**: Component server chuyên dụng để xử lý queue task

### Tính năng Queue
- **Task Enqueuing**: Queue task chuẩn với JSON serialization
- **Hỗ trợ Priority**: EnqueueWithPriority cho việc ưu tiên task
- **Quản lý TTL**: EnqueueWithTTL cho các task tạm thời
- **Batch Processing**: Các operations EnqueueWithPipeline và MultiDequeue
- **Thông tin Queue**: GetQueueInfo cho giám sát và thống kê
- **Development Utilities**: FlushQueues cho testing và development
- **Immediate Execution**: Hỗ trợ thực thi task ngay lập tức
- **Scheduled Tasks**: Lập lịch task dựa trên thời gian và khoảng thời gian
- **Time-specific Scheduling**: Thực thi task tại các thời điểm cụ thể

### Chi tiết kỹ thuật
- Phiên bản đầu tiên như module standalone `go.fork.vn/queue`
- Repository tại `github.com/Fork/queue`
- Được xây dựng với Go 1.23.9
- Test coverage đầy đủ với tài liệu toàn diện
- Tích hợp Redis Provider cho quản lý Redis tập trung
- Tách biệt sạch giữa logic queue và kết nối Redis
- Giao diện QueueRedisAdapter cho chức năng cụ thể của Redis
- Hướng dẫn migration hoàn chỉnh và mẫu cấu hình sản xuất
- Memory adapter cho môi trường development
- Redis adapter được tối ưu cho môi trường sản xuất
- Client API với giao diện task enqueueing đơn giản
- Mô hình Worker với các chiến lược xử lý có thể cấu hình

### Dependencies
- `go.fork.vn/di`: Tích hợp dependency injection
- `go.fork.vn/config`: Quản lý cấu hình
- `go.fork.vn/redis`: Kết nối Redis và tích hợp provider
- `go.fork.vn/scheduler`: Tích hợp scheduler cho xử lý queue

[Chưa phát hành]: github.com/go-fork/queue/compare/v0.1.0...HEAD
[v0.1.0]: github.com/go-fork/queue/releases/tag/v0.1.0
