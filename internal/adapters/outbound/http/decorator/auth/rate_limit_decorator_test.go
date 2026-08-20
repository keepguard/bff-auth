package auth

import (
	"context"
	"testing"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/stretchr/testify/assert"
)

func TestNewRateLimitDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultRateLimitConfig()

	// Act
	decorator := NewRateLimitDecorator(mockInner, config)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &rateLimitDecorator{}, decorator)
}

func TestRateLimitDecorator_Login_WithinLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             5,
	}

	decorator := NewRateLimitDecorator(mockInner, config)

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

	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestRateLimitDecorator_Login_ExceedsLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RateLimitConfig{
		RequestsPerSecond: 2,
		Burst:             2, // Permite apenas 2 requests em burst
	}

	decorator := NewRateLimitDecorator(mockInner, config)

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

	// Permite as primeiras 2 requisições
	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Twice()

	// Act - Primeira e segunda requisição (dentro do limite)
	result1, err1 := decorator.Login(ctx, req, tenantId, correlationID)
	result2, err2 := decorator.Login(ctx, req, tenantId, correlationID)

	// Terceira requisição (excede o limite)
	result3, err3 := decorator.Login(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.Equal(t, expectedResponse, result1)

	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result2)

	// Terceira requisição deve falhar com rate limit error
	assert.Error(t, err3)
	assert.IsType(t, &RateLimitError{}, err3)
	assert.Equal(t, inboundDto.AuthResponseDTO{}, result3)

	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 2) // Apenas 2 chamadas (a terceira foi bloqueada)
}

func TestRateLimitDecorator_Logout_WithinLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		Burst:             5,
	}

	decorator := NewRateLimitDecorator(mockInner, config)

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

func TestRateLimitDecorator_ValidateToken_ExceedsLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := RateLimitConfig{
		RequestsPerSecond: 1,
		Burst:             1,
	}

	decorator := NewRateLimitDecorator(mockInner, config)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Act
	err1 := decorator.ValidateToken(ctx, token, tenantId, correlationID) // Primeira: OK
	err2 := decorator.ValidateToken(ctx, token, tenantId, correlationID) // Segunda: Rate limited

	// Assert
	assert.NoError(t, err1)

	assert.Error(t, err2)
	assert.IsType(t, &RateLimitError{}, err2)

	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 1)
}

func TestRateLimitDecorator_ChangePassword_WithinLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultRateLimitConfig()

	decorator := NewRateLimitDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		CurrentPassword:    "old_pass",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ChangePassword", ctx, req, tenantId, correlationID).Return(nil).Once()

	// Act
	err := decorator.ChangePassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestRateLimitDecorator_ResetPassword_WithinLimit(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultRateLimitConfig()

	decorator := NewRateLimitDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		ResetToken:         "reset_token",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	mockInner.On("ResetPassword", ctx, req, tenantId, correlationID).Return(nil).Once()

	// Act
	err := decorator.ResetPassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	mockInner.AssertExpectations(t)
}

func TestTokenBucket_Refill(t *testing.T) {
	// Arrange
	bucket := newTokenBucket(10, 10) // 10 req/s, burst de 10

	// Consome todos os tokens
	for i := 0; i < 10; i++ {
		allowed, _ := bucket.tryConsume()
		assert.True(t, allowed)
	}

	// Próxima requisição deve falhar
	allowed, _ := bucket.tryConsume()
	assert.False(t, allowed)

	// Aguarda 200ms (deve reabilitar ~2 tokens a 10 req/s)
	time.Sleep(200 * time.Millisecond)

	// Deve permitir pelo menos 1 requisição após refill
	allowed, _ = bucket.tryConsume()
	assert.True(t, allowed)
}

func TestTokenBucket_BurstCapacity(t *testing.T) {
	// Arrange
	bucket := newTokenBucket(5, 10) // 5 req/s, mas burst de 10

	// Deve permitir 10 requisições imediatamente (burst)
	for i := 0; i < 10; i++ {
		allowed, _ := bucket.tryConsume()
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// 11ª requisição deve falhar
	allowed, _ := bucket.tryConsume()
	assert.False(t, allowed)
}

func TestTokenBucket_MaxTokensLimit(t *testing.T) {
	// Arrange
	bucket := newTokenBucket(10, 5) // 10 req/s, burst de 10 (max será 10 porque maxTokens = max(burst, requestsPerSecond))

	// Aguarda bastante tempo (mais que necessário para reabastecer)
	time.Sleep(2 * time.Second)

	// Mesmo após muito tempo, não deve ter mais que maxTokens (10)
	successCount := 0
	for i := 0; i < 15; i++ {
		allowed, _ := bucket.tryConsume()
		if allowed {
			successCount++
		}
	}

	// Deve permitir no máximo 10 requisições (maxTokens = max(5, 10) = 10)
	assert.LessOrEqual(t, successCount, 10)
	assert.GreaterOrEqual(t, successCount, 10) // Pelo menos 10 após aguardar
}

func TestRateLimitError_ErrorMessage(t *testing.T) {
	// Arrange
	err := &RateLimitError{
		Message:    "Too many requests",
		RetryAfter: 500 * time.Millisecond,
	}

	// Act
	msg := err.Error()

	// Assert
	assert.Contains(t, msg, "rate limit exceeded")
	assert.Contains(t, msg, "Too many requests")
	assert.Contains(t, msg, "500ms")
}
