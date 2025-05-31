package queue

import (
	"time"

	redisClient "github.com/redis/go-redis/v9"
	"go.fork.vn/di"
	"go.fork.vn/queue/adapter"
	"go.fork.vn/redis"
	"go.fork.vn/scheduler"
)

// Manager định nghĩa interface cho việc quản lý các thành phần queue.
type Manager interface {
	// RedisClient trả về Redis client.
	RedisClient() redisClient.UniversalClient

	// MemoryAdapter trả về memory queue adapter.
	MemoryAdapter() adapter.QueueAdapter

	// RedisAdapter trả về redis queue adapter.
	RedisAdapter() adapter.QueueAdapter

	// Adapter trả về queue adapter dựa trên cấu hình.
	Adapter(name string) adapter.QueueAdapter

	// Client trả về Client.
	Client() Client

	// Server trả về Server.
	Server() Server

	// Scheduler trả về Scheduler manager để lên lịch tasks.
	Scheduler() scheduler.Manager

	// SetScheduler thiết lập scheduler manager từ bên ngoài.
	SetScheduler(scheduler scheduler.Manager)
}

// manager quản lý các thành phần trong queue.
type manager struct {
	config      Config
	container   *di.Container
	client      Client
	server      Server
	scheduler   scheduler.Manager
	redisClient redisClient.UniversalClient
	memoryQueue adapter.QueueAdapter
	redisQueue  adapter.QueueAdapter
}

// NewManager tạo một manager mới với cấu hình mặc định.
func NewManager(config Config) Manager {
	return &manager{
		config: config,
	}
}

// NewManagerWithContainer tạo một manager mới với container DI để truy cập Redis provider.
func NewManagerWithContainer(config Config, container *di.Container) Manager {
	return &manager{
		config:    config,
		container: container,
	}
}

// RedisClient trả về Redis client.
func (m *manager) RedisClient() redisClient.UniversalClient {
	if m.redisClient == nil {
		// Lấy Redis provider từ container
		if m.container != nil {
			// Lấy provider key từ config, mặc định là "redis"
			providerKey := m.config.Adapter.Redis.ProviderKey
			if providerKey == "" {
				providerKey = "redis"
			}

			if redisService, err := m.container.Make(providerKey); err == nil {
				if redisManager, ok := redisService.(redis.Manager); ok {
					// Thử lấy universal client trước
					if universalClient, err := redisManager.UniversalClient(); err == nil {
						m.redisClient = universalClient
					} else if client, err := redisManager.Client(); err == nil {
						// Fallback sang standard client
						m.redisClient = client
					}
				}
			}
		}

		// Fallback: tạo client mặc định nếu không có provider
		if m.redisClient == nil {
			m.redisClient = redisClient.NewClient(&redisClient.Options{
				Addr: "localhost:6379",
			})
		}
	}
	return m.redisClient
}

// MemoryAdapter trả về memory queue adapter.
func (m *manager) MemoryAdapter() adapter.QueueAdapter {
	if m.memoryQueue == nil {
		m.memoryQueue = adapter.NewMemoryQueue(m.config.Adapter.Memory.Prefix)
	}
	return m.memoryQueue
}

// RedisAdapter trả về redis queue adapter.
func (m *manager) RedisAdapter() adapter.QueueAdapter {
	if m.redisQueue == nil {
		// Lấy Redis client từ provider
		universalClient := m.RedisClient()

		// Chuyển đổi universal client thành standard client nếu cần
		var stdClient *redisClient.Client
		if client, ok := universalClient.(*redisClient.Client); ok {
			stdClient = client
		} else {
			// Fallback: tạo client mới từ provider
			if m.container != nil {
				providerKey := m.config.Adapter.Redis.ProviderKey
				if providerKey == "" {
					providerKey = "redis"
				}

				if redisService, err := m.container.Make(providerKey); err == nil {
					if redisManager, ok := redisService.(redis.Manager); ok {
						if client, err := redisManager.Client(); err == nil {
							stdClient = client
						}
					}
				}
			}

			// Final fallback
			if stdClient == nil {
				stdClient = redisClient.NewClient(&redisClient.Options{
					Addr: "localhost:6379",
				})
			}
		}

		m.redisQueue = adapter.NewRedisQueue(stdClient, m.config.Adapter.Redis.Prefix)
	}
	return m.redisQueue
}

// Adapter trả về queue adapter dựa trên cấu hình.
func (m *manager) Adapter(name string) adapter.QueueAdapter {
	if name == "" {
		name = m.config.Adapter.Default
	}

	switch name {
	case "redis":
		return m.RedisAdapter()
	case "memory":
		return m.MemoryAdapter()
	default:
		// Mặc định sử dụng memory adapter
		return m.MemoryAdapter()
	}
}

// Client trả về Client.
func (m *manager) Client() Client {
	if m.client == nil {
		if m.config.Adapter.Default == "redis" {
			m.client = NewClientWithUniversalClient(m.RedisClient())
		} else {
			m.client = NewClientWithAdapter(m.Adapter(m.config.Adapter.Default))
		}
	}
	return m.client
}

// Server trả về Server.
func (m *manager) Server() Server {
	if m.server == nil {
		serverOpts := ServerOptions{
			Concurrency:     m.config.Server.Concurrency,
			PollingInterval: m.config.Server.PollingInterval,
			DefaultQueue:    m.config.Server.DefaultQueue,
			StrictPriority:  m.config.Server.StrictPriority,
			Queues:          m.config.Server.Queues,
			ShutdownTimeout: time.Duration(m.config.Server.ShutdownTimeout) * time.Second,
			LogLevel:        m.config.Server.LogLevel,
			RetryLimit:      m.config.Server.RetryLimit,
		}

		if m.config.Adapter.Default == "redis" {
			m.server = NewServer(m.RedisClient(), serverOpts)
		} else {
			m.server = NewServerWithAdapter(m.Adapter(m.config.Adapter.Default), serverOpts)
		}
	}
	return m.server
}

// Scheduler trả về Scheduler manager để lên lịch tasks.
func (m *manager) Scheduler() scheduler.Manager {
	if m.scheduler == nil {
		// Tạo scheduler mới nếu chưa có
		m.scheduler = scheduler.NewScheduler()
	}
	return m.scheduler
}

// SetScheduler thiết lập scheduler manager từ bên ngoài.
func (m *manager) SetScheduler(sched scheduler.Manager) {
	m.scheduler = sched
}
