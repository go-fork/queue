# Cấu hình hệ thống - Go Queue

Go Queue cung cấp hệ thống cấu hình linh hoạt và mạnh mẽ thông qua struct `Config` và file cấu hình YAML. Tài liệu này mô tả chi tiết về tất cả các tùy chọn cấu hình có sẵn.

## Cấu trúc cấu hình

### Config struct

**Vị trí**: [`config.go`](../config.go)

```go
type Config struct {
    Adapter AdapterConfig `mapstructure:"adapter"`
    Server  ServerConfig  `mapstructure:"server"`
    Client  ClientConfig  `mapstructure:"client"`
}
```

## Cấu hình Adapter

### AdapterConfig

Quản lý cấu hình cho các queue adapters:

```go
type AdapterConfig struct {
    Default string       `mapstructure:"default"`      // "memory" hoặc "redis"
    Memory  MemoryConfig `mapstructure:"memory"`       // Cấu hình Memory adapter
    Redis   RedisConfig  `mapstructure:"redis"`        // Cấu hình Redis adapter
}
```

### Memory Adapter

**Cấu hình cho in-memory queue adapter:**

```go
type MemoryConfig struct {
    Prefix string `mapstructure:"prefix"`   // Tiền tố cho queue names
}
```

**Ví dụ cấu hình:**
```yaml
queue:
  adapter:
    default: "memory"
    memory:
      prefix: "queue:"
```

**Đặc điểm:**
- ✅ **Hiệu suất cao**: Truy cập trực tiếp bộ nhớ
- ✅ **Không dependencies**: Không cần external services
- ❌ **Không persistent**: Mất dữ liệu khi restart
- ❌ **Không distributed**: Chỉ trong 1 process

### Redis Adapter

**Cấu hình cho Redis-based queue adapter:**

```go
type RedisConfig struct {
    Prefix      string `mapstructure:"prefix"`        // Tiền tố cho Redis keys
    ProviderKey string `mapstructure:"provider_key"`  // Key của Redis provider trong DI
}
```

**Ví dụ cấu hình:**
```yaml
queue:
  adapter:
    default: "redis"
    redis:
      prefix: "queue:"
      provider_key: "default"  # Tham chiếu đến redis.default
```

**Đặc điểm:**
- ✅ **Persistent**: Dữ liệu được lưu trữ bền vững
- ✅ **Distributed**: Hỗ trợ multiple instances
- ✅ **Scalable**: Có thể mở rộng theo horizontal
- ❌ **Network latency**: Phụ thuộc vào kết nối mạng

## Cấu hình Server

### ServerConfig

Quản lý cấu hình cho queue processing server:

```go
type ServerConfig struct {
    Concurrency      int      `mapstructure:"concurrency"`       // Số worker đồng thời
    PollingInterval  int      `mapstructure:"pollingInterval"`   // Khoảng thời gian polling (ms)
    DefaultQueue     string   `mapstructure:"defaultQueue"`      // Queue mặc định
    StrictPriority   bool     `mapstructure:"strictPriority"`    // Ưu tiên nghiêm ngặt
    Queues           []string `mapstructure:"queues"`            // Danh sách queues theo ưu tiên
    ShutdownTimeout  int      `mapstructure:"shutdownTimeout"`   // Timeout khi shutdown (s)
    LogLevel         int      `mapstructure:"logLevel"`          // Mức độ log
    RetryLimit       int      `mapstructure:"retryLimit"`        // Số lần retry tối đa
}
```

### Chi tiết các tham số Server

#### Concurrency
```yaml
server:
  concurrency: 10  # Số worker xử lý song song
```
- **Mặc định**: 10
- **Mô tả**: Số lượng goroutines xử lý tác vụ đồng thời
- **Gợi ý**: 
  - CPU-bound tasks: Số CPU cores
  - I/O-bound tasks: 2-4 lần số CPU cores
  - Memory-bound: Tùy thuộc vào RAM available

#### PollingInterval
```yaml
server:
  pollingInterval: 1000  # Milliseconds
```
- **Mặc định**: 1000ms (1 giây)
- **Mô tả**: Khoảng thời gian giữa các lần kiểm tra queue
- **Trade-offs**:
  - Giá trị thấp: Responsive hơn, CPU usage cao hơn
  - Giá trị cao: Tiết kiệm CPU, latency cao hơn

#### Queue Priority System
```yaml
server:
  defaultQueue: "default"
  strictPriority: true
  queues:
    - "critical"    # Ưu tiên cao nhất
    - "high"        
    - "default"     
    - "low"         # Ưu tiên thấp nhất
```

**StrictPriority**:
- `true`: Xử lý hết queue ưu tiên cao trước khi chuyển sang queue thấp hơn
- `false`: Round-robin giữa các queues có task

#### Graceful Shutdown
```yaml
server:
  shutdownTimeout: 30  # Seconds
```
- **Mặc định**: 30 giây
- **Mô tả**: Thời gian chờ workers hoàn thành task khi shutdown
- **Hành vi**: 
  - Ngừng nhận task mới
  - Chờ current tasks hoàn thành
  - Force stop sau timeout

#### Logging
```yaml
server:
  logLevel: 1  # 0=debug, 1=info, 2=warning, 3=error, 4=fatal
```

| Level | Giá trị | Mô tả |
|-------|---------|-------|
| Debug | 0 | Chi tiết debug, performance metrics |
| Info | 1 | Thông tin hoạt động bình thường |
| Warning | 2 | Cảnh báo, không critical |
| Error | 3 | Lỗi xử lý, retry attempts |
| Fatal | 4 | Lỗi nghiêm trọng, system crash |

#### Retry Logic
```yaml
server:
  retryLimit: 3  # Số lần retry tối đa
```
- **Mặc định**: 3 lần
- **Hành vi**: Exponential backoff giữa các lần retry
- **Dead Letter Queue**: Tasks vượt quá retry limit sẽ được move vào DLQ

## Cấu hình Client

### ClientConfig

Cấu hình mặc định cho queue client:

```go
type ClientConfig struct {
    DefaultOptions ClientDefaultOptions `mapstructure:"defaultOptions"`
}

type ClientDefaultOptions struct {
    Queue    string `mapstructure:"queue"`      // Queue mặc định
    MaxRetry int    `mapstructure:"maxRetry"`   // Retry mặc định
    Timeout  int    `mapstructure:"timeout"`    // Timeout mặc định (phút)
}
```

**Ví dụ cấu hình:**
```yaml
queue:
  client:
    defaultOptions:
      queue: "default"
      maxRetry: 3
      timeout: 30  # minutes
```

## File cấu hình mẫu

### Development Config

**Vị trí**: [`configs/app.sample.yaml`](../configs/app.sample.yaml)

```yaml
# Development - sử dụng Memory adapter
queue:
  adapter:
    default: "memory"
    memory:
      prefix: "dev_queue:"
      
  server:
    concurrency: 2
    pollingInterval: 500
    defaultQueue: "default"
    strictPriority: false
    queues: ["default"]
    shutdownTimeout: 10
    logLevel: 0  # Debug
    retryLimit: 1
    
  client:
    defaultOptions:
      queue: "default"
      maxRetry: 1
      timeout: 10
```

### Production Config

**Vị trí**: [`configs/production.sample.yaml`](../configs/production.sample.yaml)

```yaml
# Production - sử dụng Redis adapter
queue:
  adapter:
    default: "redis"
    redis:
      prefix: "prod_queue:"
      provider_key: "default"
      
  server:
    concurrency: 20
    pollingInterval: 1000
    defaultQueue: "default"
    strictPriority: true
    queues:
      - "critical"
      - "high" 
      - "default"
      - "low"
      - "batch"
    shutdownTimeout: 60
    logLevel: 1  # Info
    retryLimit: 5
    
  client:
    defaultOptions:
      queue: "default"
      maxRetry: 3
      timeout: 30

# Redis configuration
redis:
  default:
    host: "redis-cluster.internal"
    port: 6379
    password: "${REDIS_PASSWORD}"
    db: 2
    pool_size: 50
    min_idle_conns: 10
    max_conn_age: 3600
    
    dial_timeout: 10
    read_timeout: 5
    write_timeout: 5
    
    max_retries: 3
    min_retry_backoff: 8
    max_retry_backoff: 512
    
    cluster:
      enabled: true
      hosts:
        - "redis-1.internal:7000"
        - "redis-2.internal:7000" 
        - "redis-3.internal:7000"
```

## Cấu hình động

### Environment Variables

```yaml
redis:
  default:
    host: "${REDIS_HOST:localhost}"
    port: ${REDIS_PORT:6379}
    password: "${REDIS_PASSWORD:}"
```

### Config Loading

```go
import (
    "go.fork.vn/config"
    "go.fork.vn/queue"
)

func loadConfig() queue.Config {
    // Load từ file
    cfg := config.NewConfig()
    cfg.AddConfigPath("./configs")
    cfg.SetConfigName("app")
    cfg.SetConfigType("yaml")
    
    // Load từ environment
    cfg.AutomaticEnv()
    cfg.SetEnvPrefix("QUEUE")
    
    var queueConfig queue.Config
    cfg.UnmarshalKey("queue", &queueConfig)
    
    return queueConfig
}
```

## Best Practices

### 1. Chọn Adapter phù hợp

**Memory Adapter - sử dụng khi:**
- Development và testing
- Low latency requirements
- Single instance application
- Task không critical (có thể mất khi restart)

**Redis Adapter - sử dụng khi:**
- Production environment
- Multi-instance deployment
- High availability requirements
- Task persistence cần thiết

### 2. Tuning Performance

**High Throughput:**
```yaml
server:
  concurrency: 50
  pollingInterval: 100
  strictPriority: false
```

**Low Latency:**
```yaml
server:
  concurrency: 10
  pollingInterval: 50
  strictPriority: true
```

**Resource Conservative:**
```yaml
server:
  concurrency: 5
  pollingInterval: 2000
  strictPriority: true
```

### 3. Monitoring Config

```yaml
server:
  logLevel: 1
  shutdownTimeout: 120  # Đủ thời gian cho long-running tasks
  retryLimit: 5         # Cân bằng giữa resilience và performance
```

### 4. Security và Reliability

```yaml
redis:
  default:
    password: "${REDIS_PASSWORD}"  # Luôn dùng env vars cho credentials
    tls:
      enabled: true
      cert_file: "/etc/ssl/redis-client.crt"
      key_file: "/etc/ssl/redis-client.key"
      ca_file: "/etc/ssl/ca.crt"
```
