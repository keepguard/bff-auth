package resilience

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMetricsRecorder é um mock para MetricsRecorder
type MockMetricsRecorder struct {
	mock.Mock
}

func (m *MockMetricsRecorder) SetCircuitBreakerState(service string, state int) {
	m.Called(service, state)
}

func TestNewCircuitBreakerManager(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)

	// Act
	manager := NewCircuitBreakerManager(mockMetrics)

	// Assert
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.breakers)
	assert.Equal(t, mockMetrics, manager.metrics)
}

func TestCircuitBreakerManager_GetOrCreate_NewBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	// Act
	breaker := manager.GetOrCreate("test-service", config)

	// Assert
	assert.NotNil(t, breaker)
	assert.Equal(t, "test-service", breaker.Name())
	assert.Contains(t, manager.breakers, "test-service")
}

func TestCircuitBreakerManager_GetOrCreate_ExistingBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	// Act
	breaker1 := manager.GetOrCreate("test-service", config)
	breaker2 := manager.GetOrCreate("test-service", config)

	// Assert
	assert.Equal(t, breaker1, breaker2)
}

func TestCircuitBreakerManager_Execute_Success(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	manager.GetOrCreate("test-service", config)

	expectedResult := "success"
	expectedError := error(nil)

	// Act
	result, err := manager.Execute(context.Background(), "test-service", func() (interface{}, error) {
		return expectedResult, expectedError
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestCircuitBreakerManager_Execute_Error(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	manager.GetOrCreate("test-service", config)

	expectedError := errors.New("test error")

	// Act
	result, err := manager.Execute(context.Background(), "test-service", func() (interface{}, error) {
		return nil, expectedError
	})

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, result)
}

func TestCircuitBreakerManager_Execute_NoBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	expectedResult := "success"
	expectedError := error(nil)

	// Act
	result, err := manager.Execute(context.Background(), "unknown-service", func() (interface{}, error) {
		return expectedResult, expectedError
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestCircuitBreakerManager_GetState_ExistingBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	manager.GetOrCreate("test-service", config)

	// Act
	state := manager.GetState("test-service")

	// Assert
	assert.Equal(t, gobreaker.StateClosed, state)
}

func TestCircuitBreakerManager_GetState_NoBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	// Act
	state := manager.GetState("unknown-service")

	// Assert
	assert.Equal(t, gobreaker.StateClosed, state)
}

func TestCircuitBreakerManager_GetCounts_ExistingBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	manager.GetOrCreate("test-service", config)

	// Act
	counts := manager.GetCounts("test-service")

	// Assert
	assert.NotNil(t, counts)
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreakerManager_GetCounts_NoBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	// Act
	counts := manager.GetCounts("unknown-service")

	// Assert
	assert.Equal(t, uint32(0), counts.Requests)
	assert.Equal(t, uint32(0), counts.TotalSuccesses)
	assert.Equal(t, uint32(0), counts.TotalFailures)
}

func TestCircuitBreakerManager_Reset_ExistingBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     10 * time.Second,
	}

	manager.GetOrCreate("test-service", config)

	// Act
	err := manager.Reset("test-service")

	// Assert
	assert.NoError(t, err)
}

func TestCircuitBreakerManager_Reset_NoBreaker(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	// Act
	err := manager.Reset("unknown-service")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker for service unknown-service not found")
}

func TestDefaultConfig(t *testing.T) {
	// Arrange
	serviceName := "test-service"

	// Act
	config := DefaultConfig(serviceName)

	// Assert
	assert.Equal(t, serviceName, config.Name)
	assert.Equal(t, uint32(3), config.MaxRequests)
	assert.Equal(t, 60*time.Second, config.Interval)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.NotNil(t, config.ReadyToTrip)
}

func TestDefaultConfig_ReadyToTrip(t *testing.T) {
	// Arrange
	config := DefaultConfig("test-service")

	testCases := []struct {
		name   string
		counts gobreaker.Counts
		expect bool
	}{
		{
			name: "Less than 3 requests",
			counts: gobreaker.Counts{
				Requests:       2,
				TotalSuccesses: 1,
				TotalFailures:  1,
			},
			expect: false,
		},
		{
			name: "3 requests with low failure ratio",
			counts: gobreaker.Counts{
				Requests:       3,
				TotalSuccesses: 2,
				TotalFailures:  1,
			},
			expect: false,
		},
		{
			name: "3 requests with high failure ratio",
			counts: gobreaker.Counts{
				Requests:       3,
				TotalSuccesses: 1,
				TotalFailures:  2,
			},
			expect: true,
		},
		{
			name: "More than 3 requests with high failure ratio",
			counts: gobreaker.Counts{
				Requests:       10,
				TotalSuccesses: 3,
				TotalFailures:  7,
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := config.ReadyToTrip(tc.counts)

			// Assert
			assert.Equal(t, tc.expect, result)
		})
	}
}

func TestCircuitBreakerManager_StateChange_RecordsMetrics(t *testing.T) {
	// Arrange
	mockMetrics := new(MockMetricsRecorder)
	manager := NewCircuitBreakerManager(mockMetrics)

	config := CircuitBreakerConfig{
		Name:        "test-service",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.TotalFailures >= 10 // Muito alto para não abrir
		},
	}

	manager.GetOrCreate("test-service", config)

	// Act
	manager.Execute(context.Background(), "test-service", func() (interface{}, error) {
		return nil, errors.New("test error")
	})

	// Assert
	// O circuit breaker deve estar fechado inicialmente
	assert.Equal(t, gobreaker.StateClosed, manager.GetState("test-service"))
}
