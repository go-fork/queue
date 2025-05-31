package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.fork.vn/scheduler/mocks"
)

// TestManagerScheduler tests the scheduler integration in manager
func TestManagerScheduler(t *testing.T) {
	config := Config{
		Server: ServerConfig{
			DefaultQueue: "default",
			Concurrency:  5,
		},
	}

	manager := NewManager(config)

	// Test scheduler được tạo tự động
	scheduler := manager.Scheduler()
	assert.NotNil(t, scheduler, "Scheduler should not be nil")

	// Test scheduler được cache
	scheduler2 := manager.Scheduler()
	assert.Same(t, scheduler, scheduler2, "Scheduler should be cached")
}

// TestManagerSetScheduler tests setting external scheduler
func TestManagerSetScheduler(t *testing.T) {
	config := Config{
		Server: ServerConfig{
			DefaultQueue: "default",
			Concurrency:  5,
		},
	}

	manager := NewManager(config)

	// Tạo mock scheduler
	mockScheduler := mocks.NewMockManager(t)

	// Set scheduler từ bên ngoài
	manager.SetScheduler(mockScheduler)

	// Verify scheduler đã được set
	scheduler := manager.Scheduler()
	assert.Same(t, mockScheduler, scheduler, "External scheduler should be used")
}

// TestManagerSchedulerIntegration tests integration between manager and scheduler
func TestManagerSchedulerIntegration(t *testing.T) {
	config := Config{
		Server: ServerConfig{
			DefaultQueue: "default",
			Concurrency:  5,
		},
	}

	manager := NewManager(config)

	// Tạo mock scheduler với expectations
	mockScheduler := mocks.NewMockManager(t)

	// Set up expectations for fluent interface
	mockScheduler.EXPECT().Every(5).Return(mockScheduler)
	mockScheduler.EXPECT().Minutes().Return(mockScheduler)
	mockScheduler.EXPECT().Do(mock.AnythingOfType("func()")).Return(nil, nil)

	// Set mock scheduler
	manager.SetScheduler(mockScheduler)

	// Test sử dụng scheduler trong queue operations
	scheduler := manager.Scheduler()
	_, err := scheduler.Every(5).Minutes().Do(func() {
		// Mock cleanup task
	})

	assert.NoError(t, err, "Scheduling task should not return error")

	// Verify all expectations were met
	mockScheduler.AssertExpectations(t)
}

// TestManagerRedisClient tests the RedisClient method
func TestManagerRedisClient(t *testing.T) {
	t.Run("creates fallback redis client", func(t *testing.T) {
		config := Config{
			Adapter: AdapterConfig{
				Redis: RedisConfig{
					Prefix:      "test:",
					ProviderKey: "redis",
				},
			},
		}
		manager := NewManager(config)

		client := manager.RedisClient()
		assert.NotNil(t, client, "RedisClient should not be nil")

		// Call again to test caching
		client2 := manager.RedisClient()
		assert.Same(t, client, client2, "RedisClient should return the same instance")
	})
}

// TestManagerRedisAdapter tests the RedisAdapter method
func TestManagerRedisAdapter(t *testing.T) {
	config := Config{
		Adapter: AdapterConfig{
			Redis: RedisConfig{
				Prefix:      "test:",
				ProviderKey: "redis",
			},
		},
	}
	manager := NewManager(config)

	adapter := manager.RedisAdapter()
	assert.NotNil(t, adapter, "RedisAdapter should not be nil")

	// Call again to test caching
	adapter2 := manager.RedisAdapter()
	assert.Same(t, adapter, adapter2, "RedisAdapter should return the same instance")
}

// TestManagerClient tests the Client method
func TestManagerClient(t *testing.T) {
	t.Run("creates redis client when default adapter is redis", func(t *testing.T) {
		config := Config{
			Adapter: AdapterConfig{
				Default: "redis",
				Redis: RedisConfig{
					Prefix:      "test:",
					ProviderKey: "redis",
				},
			},
		}
		manager := NewManager(config)

		client := manager.Client()
		assert.NotNil(t, client, "Client should not be nil")

		// Call again to test caching
		client2 := manager.Client()
		assert.Same(t, client, client2, "Client should return the same instance")
	})

	t.Run("creates adapter client when default adapter is not redis", func(t *testing.T) {
		config := Config{
			Adapter: AdapterConfig{
				Default: "memory",
				Memory: MemoryConfig{
					Prefix: "test:",
				},
			},
		}
		manager := NewManager(config)

		client := manager.Client()
		assert.NotNil(t, client, "Client should not be nil")

		// Call again to test caching
		client2 := manager.Client()
		assert.Same(t, client, client2, "Client should return the same instance")
	})
}
