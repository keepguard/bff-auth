package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/stretchr/testify/assert"
)

func TestNewCacheDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultCacheConfig()

	// Act
	decorator := NewCacheDecorator(mockInner, config, nil)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &cacheDecorator{}, decorator)
}

func TestCacheDecorator_Login_NoCaching(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0, // Desabilita cleanup para teste
	}
	decorator := NewCacheDecorator(mockInner, config, nil)

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

	// Login deve ser chamado sempre (sem cache)
	mockInner.On("Login", ctx, req, tenantId, correlationID).Return(expectedResponse, nil).Twice()

	// Act - Primeira chamada
	result1, err1 := decorator.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Act - Segunda chamada
	result2, err2 := decorator.Login(ctx, req, tenantId, correlationID, "keepguard-default-client", "", "", "", "", "")

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result1)
	assert.Equal(t, expectedResponse, result2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "Login", 2) // Duas chamadas (sem cache)
}

func TestCacheDecorator_ValidateToken_WithCaching(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0,
	}
	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Primeira chamada: cache miss, chama o inner
	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Act - Primeira chamada
	err1 := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Act - Segunda chamada (deve vir do cache)
	err2 := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Act - Terceira chamada (ainda do cache)
	err3 := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 1) // Apenas uma chamada (as outras vieram do cache)
}

func TestCacheDecorator_ValidateToken_CachesErrors(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0,
	}
	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	token := "invalid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("invalid token")

	// Primeira chamada: cache miss
	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(expectedError).Once()

	// Act - Primeira chamada
	err1 := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Act - Segunda chamada (erro vem do cache)
	err2 := decorator.ValidateToken(ctx, token, tenantId, correlationID)

	// Assert
	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, expectedError, err1)
	assert.Equal(t, expectedError, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 1) // Apenas uma chamada
}

func TestCacheDecorator_Logout_InvalidatesCache(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0,
	}
	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	token := "valid_token"
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Primeira ValidateToken: popula o cache
	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Logout: deve invalidar o cache
	mockInner.On("Logout", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Segunda ValidateToken: cache foi invalidado, deve chamar inner novamente
	mockInner.On("ValidateToken", ctx, token, tenantId, correlationID).Return(nil).Once()

	// Act
	err1 := decorator.ValidateToken(ctx, token, tenantId, correlationID) // Popula cache
	assert.NoError(t, err1)

	err2 := decorator.Logout(ctx, token, tenantId, correlationID) // Invalida cache
	assert.NoError(t, err2)

	err3 := decorator.ValidateToken(ctx, token, tenantId, correlationID) // Cache foi invalidado
	assert.NoError(t, err3)

	// Assert
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 2) // Duas chamadas (cache foi invalidado)
	mockInner.AssertNumberOfCalls(t, "Logout", 1)
}

func TestCacheDecorator_ChangePassword_NoCaching(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultCacheConfig()
	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		CurrentPassword:    "old_pass",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Operações de escrita não devem ser cacheadas
	mockInner.On("ChangePassword", ctx, req, tenantId, correlationID).Return(nil).Twice()

	// Act
	err1 := decorator.ChangePassword(ctx, req, tenantId, correlationID)
	err2 := decorator.ChangePassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ChangePassword", 2) // Duas chamadas (sem cache)
}

func TestCacheDecorator_ResetPassword_NoCaching(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultCacheConfig()
	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	req := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           "user-code-123",
		ResetToken:         "reset_token",
		NewPassword:        "new_pass",
		ConfirmNewPassword: "new_pass",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	// Operações de escrita não devem ser cacheadas
	mockInner.On("ResetPassword", ctx, req, tenantId, correlationID).Return(nil).Twice()

	// Act
	err1 := decorator.ResetPassword(ctx, req, tenantId, correlationID)
	err2 := decorator.ResetPassword(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ResetPassword", 2) // Duas chamadas (sem cache)
}

func TestCache_Expiration(t *testing.T) {
	// Arrange
	config := CacheConfig{
		TTL:             50 * time.Millisecond, // TTL muito curto para teste
		MaxSize:         100,
		CleanupInterval: 0,
	}
	cache := newCache(config, nil)

	// Act
	cache.set("test-key", "test-value")

	// Imediatamente após set, deve estar no cache
	value1, found1 := cache.get("test-key")

	// Aguarda expiração
	time.Sleep(60 * time.Millisecond)

	// Após expiração, não deve estar no cache
	value2, found2 := cache.get("test-key")

	// Assert
	assert.True(t, found1)
	assert.Equal(t, "test-value", value1)

	assert.False(t, found2)
	assert.Nil(t, value2)
}

func TestCache_MaxSize(t *testing.T) {
	// Arrange
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         3, // Tamanho pequeno para teste
		CleanupInterval: 0,
	}
	cache := newCache(config, nil)

	// Act - Adiciona 4 entradas (excede MaxSize)
	cache.set("key1", "value1")
	cache.set("key2", "value2")
	cache.set("key3", "value3")
	cache.set("key4", "value4") // Deve remover uma entrada antiga

	// Assert
	cache.mu.RLock()
	size := len(cache.entries)
	cache.mu.RUnlock()

	assert.Equal(t, 3, size) // Tamanho máximo respeitado
}
