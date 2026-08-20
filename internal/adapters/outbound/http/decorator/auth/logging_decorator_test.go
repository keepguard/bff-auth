package auth

import (
	"context"
	"errors"
	"testing"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &loggingDecorator{}, decorator)
}

func TestLoggingDecorator_Login_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth"

	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "access_token",
		ExpiresIn: 3600,
	}

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len()) // Inicio e fim
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Requisição concluída com sucesso", logs.All()[1].Message)
}

func TestLoggingDecorator_Login_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth"

	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("auth service error")

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, expectedError)

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len()) // Inicio e erro
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Erro na requisição", logs.All()[1].Message)
}

func TestLoggingDecorator_ValidateToken_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	serviceName := "ms-auth"

	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil)

	// Act
	err := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)

	// Verifica logs (debug level)
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Token validado com sucesso", logs.All()[1].Message)
}

func TestLoggingDecorator_ChangePassword_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth"

	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		CurrentPassword:    "old_pass",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ChangePassword", ctx, req, tenantId, correlationID).Return(nil)

	// Act
	err := decorator.ChangePassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Senha alterada com sucesso", logs.All()[1].Message)
}

func TestLoggingDecorator_ResetPassword_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth"

	decorator := NewLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		ResetToken:         "reset_token",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ResetPassword", ctx, req, tenantId, correlationID).Return(nil)

	// Act
	err := decorator.ResetPassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Senha resetada com sucesso", logs.All()[1].Message)

	// Verifica que o token foi ocultado no log
	logEntry := logs.All()[0]
	requestField := logEntry.ContextMap()["request"]
	assert.Contains(t, requestField, "***")
}
