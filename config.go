package queue

// Config chứa cấu hình cho queue package.
type Config struct {
	// Adapter chứa cấu hình cho các queue adapter.
	Adapter AdapterConfig `mapstructure:"adapter"`

	// Server chứa cấu hình cho queue server.
	Server ServerConfig `mapstructure:"server"`

	// Client chứa cấu hình cho queue client.
	Client ClientConfig `mapstructure:"client"`
}

// AdapterConfig chứa cấu hình cho các adapter.
type AdapterConfig struct {
	// Default xác định adapter mặc định sẽ được sử dụng.
	// Các giá trị hợp lệ: "memory", "redis"
	Default string `mapstructure:"default"`

	// Memory chứa cấu hình cho memory adapter.
	Memory MemoryConfig `mapstructure:"memory"`

	// Redis chứa cấu hình cho redis adapter.
	Redis RedisConfig `mapstructure:"redis"`
}

// MemoryConfig chứa cấu hình cho memory adapter.
type MemoryConfig struct {
	// Prefix là tiền tố cho tên của các queue trong bộ nhớ.
	Prefix string `mapstructure:"prefix"`
}

// RedisConfig chứa cấu hình cho redis adapter.
type RedisConfig struct {
	// Prefix là tiền tố cho tên của các queue trong Redis.
	Prefix string `mapstructure:"prefix"`

	// ProviderKey là khóa để lấy Redis provider từ DI container.
	// Mặc định là "redis" nếu không được cấu hình.
	ProviderKey string `mapstructure:"provider_key"`
}

// ServerConfig chứa cấu hình cho queue server.
type ServerConfig struct {
	// Concurrency là số lượng worker xử lý tác vụ cùng một lúc.
	Concurrency int `mapstructure:"concurrency"`

	// PollingInterval là khoảng thời gian giữa các lần kiểm tra tác vụ mới (tính bằng mili giây).
	PollingInterval int `mapstructure:"pollingInterval"`

	// DefaultQueue là tên queue mặc định nếu không có queue nào được chỉ định.
	DefaultQueue string `mapstructure:"defaultQueue"`

	// StrictPriority xác định liệu có ưu tiên nghiêm ngặt các queue ưu tiên cao hay không.
	StrictPriority bool `mapstructure:"strictPriority"`

	// Queues là danh sách các queue cần lắng nghe, theo thứ tự ưu tiên.
	Queues []string `mapstructure:"queues"`

	// ShutdownTimeout là thời gian chờ để các worker hoàn tất tác vụ khi dừng server (tính bằng giây).
	ShutdownTimeout int `mapstructure:"shutdownTimeout"`

	// LogLevel xác định mức độ log.
	LogLevel int `mapstructure:"logLevel"`

	// RetryLimit xác định số lần thử lại tối đa cho tác vụ bị lỗi.
	RetryLimit int `mapstructure:"retryLimit"`
}

// ClientConfig chứa cấu hình cho queue client.
type ClientConfig struct {
	// DefaultOptions chứa các tùy chọn mặc định cho tác vụ.
	DefaultOptions ClientDefaultOptions `mapstructure:"defaultOptions"`
}

// ClientDefaultOptions chứa các tùy chọn mặc định cho tác vụ.
type ClientDefaultOptions struct {
	// Queue là tên queue mặc định cho các tác vụ.
	Queue string `mapstructure:"queue"`

	// MaxRetry là số lần thử lại tối đa cho tác vụ bị lỗi.
	MaxRetry int `mapstructure:"maxRetry"`

	// Timeout là thời gian tối đa để tác vụ hoàn thành (tính bằng phút).
	Timeout int `mapstructure:"timeout"`
}

// DefaultConfig trả về cấu hình mặc định cho queue.
func DefaultConfig() Config {
	return Config{
		Adapter: AdapterConfig{
			Default: "memory",
			Memory: MemoryConfig{
				Prefix: "queue:",
			},
			Redis: RedisConfig{
				Prefix:      "queue:",
				ProviderKey: "redis",
			},
		},
		Server: ServerConfig{
			Concurrency:     10,
			PollingInterval: 1000,
			DefaultQueue:    "default",
			StrictPriority:  true,
			Queues:          []string{"critical", "high", "default", "low"},
			ShutdownTimeout: 30,
			LogLevel:        1,
			RetryLimit:      3,
		},
		Client: ClientConfig{
			DefaultOptions: ClientDefaultOptions{
				Queue:    "default",
				MaxRetry: 3,
				Timeout:  30,
			},
		},
	}
}
