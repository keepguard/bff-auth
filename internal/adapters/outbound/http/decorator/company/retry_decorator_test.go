package company

import (
	"context"
	"net/http"
	"testing"
	"time"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := DefaultRetryConfig()

	// Act
	decorator := NewRetryDecorator(mockInner, config)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &retryDecorator{}, decorator)
}

func TestRetryDecorator_GetByTenantId_SuccessFirstAttempt(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 1)
}

func TestRetryDecorator_GetByTenantId_SuccessAfterRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	// Primeira tentativa: 503 (retryable)
	retryableError := &appdto.HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service temporarily unavailable",
		ErrorCode:  "SERVICE_UNAVAILABLE",
	}
	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, retryableError).Once()

	// Segunda tentativa: sucesso
	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 2) // 2 chamadas (retry)
}

func TestRetryDecorator_GetByTenantId_NonRetryableError(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := DefaultRetryConfig()

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"

	// 404 não é retryable
	nonRetryableError := &appdto.HTTPError{
		StatusCode: http.StatusNotFound,
		Message:    "Company not found",
		ErrorCode:  "COMPANY_NOT_FOUND",
	}

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, nonRetryableError).Once()

	// Act
	_, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, nonRetryableError, err)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 1) // Apenas 1 tentativa (não retenta)
}
