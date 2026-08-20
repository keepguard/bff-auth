package auth

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultRetryConfig()

	// Act
	decorator := NewRetryDecorator(mockInner, config)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &retryDecorator{}, decorator)
}

func TestRetryDecorator_Login_SuccessFirstAttempt(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

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

	// Sucesso na primeira tentativa
	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 1) // Apenas uma chamada
}

func TestRetryDecorator_Login_SuccessAfterRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

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

	// Primeira tentativa: 503 (retryable)
	retryableError := &appdto.HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service temporarily unavailable",
		ErrorCode:  "SERVICE_UNAVAILABLE",
	}
	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, retryableError).Once()

	// Segunda tentativa: sucesso
	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	start := time.Now()
	result, err := decorator.Login(ctx, req, tenantId, correlationID)
	elapsed := time.Since(start)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 2)           // Duas chamadas
	assert.GreaterOrEqual(t, elapsed, 10*time.Millisecond) // Pelo menos o delay inicial
}

func TestRetryDecorator_Login_NonRetryableError(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "wrongpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// 401 não é retryable (credenciais inválidas)
	nonRetryableError := &appdto.HTTPError{
		StatusCode: http.StatusUnauthorized,
		Message:    "Invalid credentials",
		ErrorCode:  "INVALID_CREDENTIALS",
	}

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, nonRetryableError).Once()

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, nonRetryableError, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 1) // Apenas uma tentativa (não retenta)
}

func TestRetryDecorator_Login_MaxAttemptsReached(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Sempre retorna 503 (retryable)
	retryableError := &appdto.HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service temporarily unavailable",
		ErrorCode:  "SERVICE_UNAVAILABLE",
	}

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(inboundDto.AuthResponseDTO{}, retryableError).Times(3)

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, retryableError, err)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 3) // Todas as 3 tentativas
}

func TestRetryDecorator_Logout_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultRetryConfig()

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("Logout", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Act
	err := decorator.Logout(ctx, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestRetryDecorator_ChangePassword_SuccessAfterRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		CurrentPassword:    "old_pass",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Primeira tentativa: timeout (retryable)
	timeoutError := &appdto.HTTPError{
		StatusCode: http.StatusRequestTimeout,
		Message:    "Request timeout",
		ErrorCode:  "TIMEOUT",
	}
	mockInner.On("ChangePassword", ctx, req, tenantId, correlationID).Return(timeoutError).Once()

	// Segunda tentativa: sucesso
	mockInner.On("ChangePassword", ctx, req, tenantId, correlationID).Return(nil).Once()

	// Act
	err := decorator.ChangePassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ChangePassword", 2)
}

func TestRetryDecorator_CalculateDelay_ExponentialBackoff(t *testing.T) {
	// Arrange
	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       false, // Sem jitter para teste determinístico
	}

	decorator := &retryDecorator{config: config}

	// Act & Assert
	delay0 := decorator.calculateDelay(0)
	assert.Equal(t, 100*time.Millisecond, delay0) // 100 * 2^0 = 100ms

	delay1 := decorator.calculateDelay(1)
	assert.Equal(t, 200*time.Millisecond, delay1) // 100 * 2^1 = 200ms

	delay2 := decorator.calculateDelay(2)
	assert.Equal(t, 400*time.Millisecond, delay2) // 100 * 2^2 = 400ms

	delay3 := decorator.calculateDelay(3)
	assert.Equal(t, 800*time.Millisecond, delay3) // 100 * 2^3 = 800ms
}

func TestRetryDecorator_CalculateDelay_WithMaxDelay(t *testing.T) {
	// Arrange
	config := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	decorator := &retryDecorator{config: config}

	// Act
	delay5 := decorator.calculateDelay(5) // 100 * 2^5 = 3200ms, mas max é 500ms

	// Assert
	assert.Equal(t, 500*time.Millisecond, delay5)
}

func TestRetryDecorator_IsRetryableError(t *testing.T) {
	decorator := &retryDecorator{config: DefaultRetryConfig()}

	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "Nil error",
			err:       nil,
			retryable: false,
		},
		{
			name: "503 Service Unavailable",
			err: &appdto.HTTPError{
				StatusCode: http.StatusServiceUnavailable,
				Message:    "Service unavailable",
			},
			retryable: true,
		},
		{
			name: "429 Too Many Requests",
			err: &appdto.HTTPError{
				StatusCode: http.StatusTooManyRequests,
				Message:    "Rate limit exceeded",
			},
			retryable: true,
		},
		{
			name: "504 Gateway Timeout",
			err: &appdto.HTTPError{
				StatusCode: http.StatusGatewayTimeout,
				Message:    "Gateway timeout",
			},
			retryable: true,
		},
		{
			name: "401 Unauthorized",
			err: &appdto.HTTPError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Unauthorized",
			},
			retryable: false,
		},
		{
			name: "404 Not Found",
			err: &appdto.HTTPError{
				StatusCode: http.StatusNotFound,
				Message:    "Not found",
			},
			retryable: false,
		},
		{
			name:      "Generic network error",
			err:       errors.New("network error"),
			retryable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decorator.isRetryableError(tt.err)
			assert.Equal(t, tt.retryable, result)
		})
	}
}
