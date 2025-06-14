// Code generated by mockery. DO NOT EDIT.

package queue_mocks

import (
	mock "github.com/stretchr/testify/mock"
	adapter "go.fork.vn/queue/adapter"

	queue "go.fork.vn/queue"

	redis "github.com/redis/go-redis/v9"

	scheduler "go.fork.vn/scheduler"
)

// MockManager is an autogenerated mock type for the Manager type
type MockManager struct {
	mock.Mock
}

type MockManager_Expecter struct {
	mock *mock.Mock
}

func (_m *MockManager) EXPECT() *MockManager_Expecter {
	return &MockManager_Expecter{mock: &_m.Mock}
}

// Adapter provides a mock function with given fields: name
func (_m *MockManager) Adapter(name string) adapter.QueueAdapter {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Adapter")
	}

	var r0 adapter.QueueAdapter
	if rf, ok := ret.Get(0).(func(string) adapter.QueueAdapter); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(adapter.QueueAdapter)
		}
	}

	return r0
}

// MockManager_Adapter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Adapter'
type MockManager_Adapter_Call struct {
	*mock.Call
}

// Adapter is a helper method to define mock.On call
//   - name string
func (_e *MockManager_Expecter) Adapter(name interface{}) *MockManager_Adapter_Call {
	return &MockManager_Adapter_Call{Call: _e.mock.On("Adapter", name)}
}

func (_c *MockManager_Adapter_Call) Run(run func(name string)) *MockManager_Adapter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockManager_Adapter_Call) Return(_a0 adapter.QueueAdapter) *MockManager_Adapter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_Adapter_Call) RunAndReturn(run func(string) adapter.QueueAdapter) *MockManager_Adapter_Call {
	_c.Call.Return(run)
	return _c
}

// Client provides a mock function with no fields
func (_m *MockManager) Client() queue.Client {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Client")
	}

	var r0 queue.Client
	if rf, ok := ret.Get(0).(func() queue.Client); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Client)
		}
	}

	return r0
}

// MockManager_Client_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Client'
type MockManager_Client_Call struct {
	*mock.Call
}

// Client is a helper method to define mock.On call
func (_e *MockManager_Expecter) Client() *MockManager_Client_Call {
	return &MockManager_Client_Call{Call: _e.mock.On("Client")}
}

func (_c *MockManager_Client_Call) Run(run func()) *MockManager_Client_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_Client_Call) Return(_a0 queue.Client) *MockManager_Client_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_Client_Call) RunAndReturn(run func() queue.Client) *MockManager_Client_Call {
	_c.Call.Return(run)
	return _c
}

// MemoryAdapter provides a mock function with no fields
func (_m *MockManager) MemoryAdapter() adapter.QueueAdapter {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for MemoryAdapter")
	}

	var r0 adapter.QueueAdapter
	if rf, ok := ret.Get(0).(func() adapter.QueueAdapter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(adapter.QueueAdapter)
		}
	}

	return r0
}

// MockManager_MemoryAdapter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'MemoryAdapter'
type MockManager_MemoryAdapter_Call struct {
	*mock.Call
}

// MemoryAdapter is a helper method to define mock.On call
func (_e *MockManager_Expecter) MemoryAdapter() *MockManager_MemoryAdapter_Call {
	return &MockManager_MemoryAdapter_Call{Call: _e.mock.On("MemoryAdapter")}
}

func (_c *MockManager_MemoryAdapter_Call) Run(run func()) *MockManager_MemoryAdapter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_MemoryAdapter_Call) Return(_a0 adapter.QueueAdapter) *MockManager_MemoryAdapter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_MemoryAdapter_Call) RunAndReturn(run func() adapter.QueueAdapter) *MockManager_MemoryAdapter_Call {
	_c.Call.Return(run)
	return _c
}

// RedisAdapter provides a mock function with no fields
func (_m *MockManager) RedisAdapter() adapter.QueueAdapter {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RedisAdapter")
	}

	var r0 adapter.QueueAdapter
	if rf, ok := ret.Get(0).(func() adapter.QueueAdapter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(adapter.QueueAdapter)
		}
	}

	return r0
}

// MockManager_RedisAdapter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RedisAdapter'
type MockManager_RedisAdapter_Call struct {
	*mock.Call
}

// RedisAdapter is a helper method to define mock.On call
func (_e *MockManager_Expecter) RedisAdapter() *MockManager_RedisAdapter_Call {
	return &MockManager_RedisAdapter_Call{Call: _e.mock.On("RedisAdapter")}
}

func (_c *MockManager_RedisAdapter_Call) Run(run func()) *MockManager_RedisAdapter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_RedisAdapter_Call) Return(_a0 adapter.QueueAdapter) *MockManager_RedisAdapter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_RedisAdapter_Call) RunAndReturn(run func() adapter.QueueAdapter) *MockManager_RedisAdapter_Call {
	_c.Call.Return(run)
	return _c
}

// RedisClient provides a mock function with no fields
func (_m *MockManager) RedisClient() redis.UniversalClient {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RedisClient")
	}

	var r0 redis.UniversalClient
	if rf, ok := ret.Get(0).(func() redis.UniversalClient); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(redis.UniversalClient)
		}
	}

	return r0
}

// MockManager_RedisClient_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RedisClient'
type MockManager_RedisClient_Call struct {
	*mock.Call
}

// RedisClient is a helper method to define mock.On call
func (_e *MockManager_Expecter) RedisClient() *MockManager_RedisClient_Call {
	return &MockManager_RedisClient_Call{Call: _e.mock.On("RedisClient")}
}

func (_c *MockManager_RedisClient_Call) Run(run func()) *MockManager_RedisClient_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_RedisClient_Call) Return(_a0 redis.UniversalClient) *MockManager_RedisClient_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_RedisClient_Call) RunAndReturn(run func() redis.UniversalClient) *MockManager_RedisClient_Call {
	_c.Call.Return(run)
	return _c
}

// Scheduler provides a mock function with no fields
func (_m *MockManager) Scheduler() scheduler.Manager {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Scheduler")
	}

	var r0 scheduler.Manager
	if rf, ok := ret.Get(0).(func() scheduler.Manager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(scheduler.Manager)
		}
	}

	return r0
}

// MockManager_Scheduler_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Scheduler'
type MockManager_Scheduler_Call struct {
	*mock.Call
}

// Scheduler is a helper method to define mock.On call
func (_e *MockManager_Expecter) Scheduler() *MockManager_Scheduler_Call {
	return &MockManager_Scheduler_Call{Call: _e.mock.On("Scheduler")}
}

func (_c *MockManager_Scheduler_Call) Run(run func()) *MockManager_Scheduler_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_Scheduler_Call) Return(_a0 scheduler.Manager) *MockManager_Scheduler_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_Scheduler_Call) RunAndReturn(run func() scheduler.Manager) *MockManager_Scheduler_Call {
	_c.Call.Return(run)
	return _c
}

// Server provides a mock function with no fields
func (_m *MockManager) Server() queue.Server {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Server")
	}

	var r0 queue.Server
	if rf, ok := ret.Get(0).(func() queue.Server); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Server)
		}
	}

	return r0
}

// MockManager_Server_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Server'
type MockManager_Server_Call struct {
	*mock.Call
}

// Server is a helper method to define mock.On call
func (_e *MockManager_Expecter) Server() *MockManager_Server_Call {
	return &MockManager_Server_Call{Call: _e.mock.On("Server")}
}

func (_c *MockManager_Server_Call) Run(run func()) *MockManager_Server_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockManager_Server_Call) Return(_a0 queue.Server) *MockManager_Server_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockManager_Server_Call) RunAndReturn(run func() queue.Server) *MockManager_Server_Call {
	_c.Call.Return(run)
	return _c
}

// SetScheduler provides a mock function with given fields: _a0
func (_m *MockManager) SetScheduler(_a0 scheduler.Manager) {
	_m.Called(_a0)
}

// MockManager_SetScheduler_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetScheduler'
type MockManager_SetScheduler_Call struct {
	*mock.Call
}

// SetScheduler is a helper method to define mock.On call
//   - _a0 scheduler.Manager
func (_e *MockManager_Expecter) SetScheduler(_a0 interface{}) *MockManager_SetScheduler_Call {
	return &MockManager_SetScheduler_Call{Call: _e.mock.On("SetScheduler", _a0)}
}

func (_c *MockManager_SetScheduler_Call) Run(run func(_a0 scheduler.Manager)) *MockManager_SetScheduler_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(scheduler.Manager))
	})
	return _c
}

func (_c *MockManager_SetScheduler_Call) Return() *MockManager_SetScheduler_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockManager_SetScheduler_Call) RunAndReturn(run func(scheduler.Manager)) *MockManager_SetScheduler_Call {
	_c.Run(run)
	return _c
}

// NewMockManager creates a new instance of MockManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockManager {
	mock := &MockManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
