package company

import (
	"context"
	"errors"
	"testing"

	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/resilience"
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

func TestNewCircuitBreakerDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	// Act
	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &circuitBreakerDecorator{}, decorator)
}

func TestCircuitBreakerDecorator_GetByTenantId_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_GetByTenantId_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	mockMetrics := new(MockMetricsRecorder)
	cbManager := resilience.NewCircuitBreakerManager(mockMetrics)
	serviceName := "test-service"

	decorator := NewCircuitBreakerDecorator(mockInner, cbManager, serviceName)

	ctx := context.Background()
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("company service error")

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, expectedError)

	// Act
	result, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, portsclient.CompanySimpleResponseDTO{}, result)
	mockInner.AssertExpectations(t)
}
