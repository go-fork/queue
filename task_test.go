package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

func TestNewTask(t *testing.T) {
	// Chuẩn bị dữ liệu
	payload := []byte(`{"name":"test","message":"hello world"}`)

	// Thực thi
	task := NewTask("email:send", payload)

	// Kiểm tra
	assert.Equal(t, "email:send", task.Name)
	assert.Equal(t, payload, task.Payload)
	assert.NotEmpty(t, task.CreatedAt)
	assert.NotEmpty(t, task.ProcessAt)
	assert.Empty(t, task.ID)
	assert.Empty(t, task.Queue)
	assert.Equal(t, 0, task.MaxRetry)
	assert.Equal(t, 0, task.RetryCount)
}

func TestTaskUnmarshal(t *testing.T) {
	// Chuẩn bị dữ liệu
	expected := testPayload{
		Name:    "test",
		Message: "hello world",
	}
	jsonData, err := json.Marshal(expected)
	require.NoError(t, err)

	task := NewTask("test", jsonData)

	// Thực thi
	var actual testPayload
	err = task.Unmarshal(&actual)

	// Kiểm tra
	assert.NoError(t, err)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Message, actual.Message)
}

func TestTaskUnmarshalError(t *testing.T) {
	// Chuẩn bị dữ liệu với json không hợp lệ
	invalidJSON := []byte(`{"name":"test", invalid json}`)
	task := NewTask("test", invalidJSON)

	// Thực thi
	var result testPayload
	err := task.Unmarshal(&result)

	// Kiểm tra
	assert.Error(t, err)
}

func TestTaskInfoString(t *testing.T) {
	// Chuẩn bị dữ liệu
	now := time.Now()
	info := TaskInfo{
		ID:        "task-123",
		Name:      "email:send",
		Queue:     "emails",
		MaxRetry:  3,
		State:     "pending",
		CreatedAt: now,
		ProcessAt: now.Add(5 * time.Minute),
	}

	// Thực thi
	result := info.String()

	// Kiểm tra
	assert.Contains(t, result, "task-123")
	assert.Contains(t, result, "email:send")
	assert.Contains(t, result, "emails")
	assert.Contains(t, result, "pending")
}

func TestTaskOptionsWithQueue(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	options := ApplyOptions(WithQueue("high-priority"))

	// Kiểm tra
	assert.Equal(t, "high-priority", options.Queue)
}

func TestTaskOptionsWithMaxRetry(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	options := ApplyOptions(WithMaxRetry(5))

	// Kiểm tra
	assert.Equal(t, 5, options.MaxRetry)
}

func TestTaskOptionsWithTimeout(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	duration := 10 * time.Minute
	options := ApplyOptions(WithTimeout(duration))

	// Kiểm tra
	assert.Equal(t, duration, options.Timeout)
}

func TestTaskOptionsWithDeadline(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	deadline := time.Now().Add(1 * time.Hour)
	options := ApplyOptions(WithDeadline(deadline))

	// Kiểm tra
	assert.Equal(t, deadline, options.Deadline)
}

func TestTaskOptionsWithDelay(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	delay := 5 * time.Minute
	options := ApplyOptions(WithDelay(delay))

	// Kiểm tra
	assert.Equal(t, delay, options.Delay)
}

func TestTaskOptionsWithProcessAt(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	processAt := time.Now().Add(30 * time.Minute)
	options := ApplyOptions(WithProcessAt(processAt))

	// Kiểm tra
	assert.Equal(t, processAt, options.ProcessAt)
}

func TestTaskOptionsWithTaskID(t *testing.T) {
	// Chuẩn bị dữ liệu và thực thi
	options := ApplyOptions(WithTaskID("custom-id-123"))

	// Kiểm tra
	assert.Equal(t, "custom-id-123", options.TaskID)
}

func TestGetDefaultOptions(t *testing.T) {
	// Thực thi
	options := GetDefaultOptions()

	// Kiểm tra
	assert.Equal(t, "default", options.Queue)
	assert.Equal(t, 3, options.MaxRetry)
	assert.Equal(t, 30*time.Minute, options.Timeout)
	assert.True(t, options.Deadline.IsZero())
	assert.Zero(t, options.Delay)
	assert.True(t, options.ProcessAt.IsZero())
	assert.Empty(t, options.TaskID)
}

func TestApplyOptions(t *testing.T) {
	// Chuẩn bị dữ liệu
	processAt := time.Now().Add(1 * time.Hour)

	// Thực thi
	options := ApplyOptions(
		WithQueue("critical"),
		WithMaxRetry(5),
		WithTimeout(15*time.Minute),
		WithProcessAt(processAt),
		WithTaskID("task-abc-123"),
	)

	// Kiểm tra
	assert.Equal(t, "critical", options.Queue)
	assert.Equal(t, 5, options.MaxRetry)
	assert.Equal(t, 15*time.Minute, options.Timeout)
	assert.Equal(t, processAt, options.ProcessAt)
	assert.Equal(t, "task-abc-123", options.TaskID)
}

// TestWithTimeout tests the WithTimeout option
func TestWithTimeout(t *testing.T) {
	timeout := 10 * time.Minute
	options := ApplyOptions(WithTimeout(timeout))

	assert.Equal(t, timeout, options.Timeout, "Timeout should be set correctly")
}

// TestWithDeadline tests the WithDeadline option
func TestWithDeadline(t *testing.T) {
	deadline := time.Now().Add(2 * time.Hour)
	options := ApplyOptions(WithDeadline(deadline))

	assert.Equal(t, deadline, options.Deadline, "Deadline should be set correctly")
}

// TestWithDelay tests the WithDelay option
func TestWithDelay(t *testing.T) {
	delay := 5 * time.Minute
	options := ApplyOptions(WithDelay(delay))

	assert.Equal(t, delay, options.Delay, "Delay should be set correctly")
}

// TestWithTaskID tests the WithTaskID option
func TestWithTaskID(t *testing.T) {
	taskID := "custom-task-id-123"
	options := ApplyOptions(WithTaskID(taskID))

	assert.Equal(t, taskID, options.TaskID, "TaskID should be set correctly")
}

// TestTaskUnmarshalInvalidJSON tests Unmarshal with invalid JSON
func TestTaskUnmarshalInvalidJSON(t *testing.T) {
	task := &Task{
		Payload: []byte(`invalid json`),
	}

	var result testPayload
	err := task.Unmarshal(&result)

	assert.Error(t, err, "Should return error for invalid JSON")
}

// TestTaskUnmarshalNilTarget tests Unmarshal with nil target
func TestTaskUnmarshalNilTarget(t *testing.T) {
	task := &Task{
		Payload: []byte(`{"name":"test"}`),
	}

	err := task.Unmarshal(nil)

	assert.Error(t, err, "Should return error for nil target")
}

// TestNewTaskWithComplexPayload tests NewTask with complex payload
func TestNewTaskWithComplexPayload(t *testing.T) {
	complexPayload := map[string]interface{}{
		"user_id": 123,
		"email":   "test@example.com",
		"data": map[string]string{
			"template": "welcome",
			"language": "en",
		},
		"timestamps": []int64{1234567890, 1234567891},
	}

	jsonData, err := json.Marshal(complexPayload)
	require.NoError(t, err)

	task := NewTask("email:complex", jsonData)

	assert.Equal(t, "email:complex", task.Name)
	assert.Equal(t, jsonData, task.Payload)
	assert.NotZero(t, task.CreatedAt)
	assert.NotZero(t, task.ProcessAt)
	// ID is empty by default in NewTask - it gets generated when enqueued
	assert.Empty(t, task.ID)
}
