package user

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	// Instância compartilhada de métricas para todos os testes
	testMetrics = metrics.New()
)

func TestNewUserMetricsDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	serviceName := "test-service"

	// Act
	decorator := NewUserMetricsDecorator(mockInner, testMetrics, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &userMetricsDecorator{}, decorator)
}

func TestUserMetricsDecorator_GetByEmail_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	serviceName := "ms-auth-users"

	decorator := NewUserMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	email := "user@example.com"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.UserByEmailResponseDTO{
		ID:       "user-123",
		Email:    email,
		Username: "testuser",
		Status:   "ACTIVE",
	}

	mockInner.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByEmail(ctx, email, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestUserMetricsDecorator_GetByEmail_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	serviceName := "ms-auth-users"

	decorator := NewUserMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	email := "notfound@example.com"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("user not found")

	mockInner.On("GetByEmail", ctx, email, xApplication, correlationID).Return(outboundDto.UserByEmailResponseDTO{}, expectedError)

	// Act
	_, err := decorator.GetByEmail(ctx, email, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}
