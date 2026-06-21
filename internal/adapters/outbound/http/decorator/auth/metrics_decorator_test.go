package auth

import (
	"context"
	"errors"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	// Instância compartilhada de métricas para todos os testes
	testMetrics = metrics.New()
)

func TestNewMetricsDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "test-service"

	// Act
	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &metricsDecorator{}, decorator)
}

func TestMetricsDecorator_Login_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	mockInner.On("Login", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.Login(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Métricas são coletadas (testadas via Prometheus, não asserções diretas aqui)
}

func TestMetricsDecorator_Login_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("auth service error")

	mockInner.On("Login", ctx, req, xApplication, correlationID).Return(inboundDto.AuthResponseDTO{}, expectedError)

	// Act
	result, err := decorator.Login(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockInner.AssertExpectations(t)

	// Métricas de erro são coletadas
}

func TestMetricsDecorator_RefreshToken_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := inboundDto.RefreshTokenRequestDTO{
		Token: "valid_refresh_token",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.RefreshTokenResponseDTO{
		Token:     "new_access_token",
		ExpiresIn: 3600,
	}

	mockInner.On("RefreshToken", ctx, req, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.RefreshToken(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestMetricsDecorator_Logout_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	token := "valid_token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("Logout", ctx, token, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.Logout(ctx, token, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestMetricsDecorator_ValidateToken_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	token := "valid_token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.ValidateToken(ctx, token, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestMetricsDecorator_ChangePassword_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		CurrentPassword:    "old_pass",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ChangePassword", ctx, req, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.ChangePassword(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestMetricsDecorator_ResetPassword_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	serviceName := "ms-auth"

	decorator := NewMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		ResetToken:         "reset_token",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ResetPassword", ctx, req, xApplication, correlationID).Return(nil)

	// Act
	err := decorator.ResetPassword(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}
