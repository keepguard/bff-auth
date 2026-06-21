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

func TestNewSmartCacheDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultSmartCacheConfig()

	// Act
	decorator := NewSmartCacheDecorator(mockInner, config, nil)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &smartCacheDecorator{}, decorator)
}

func TestSmartCacheDecorator_Login_StoresTokenMetadata(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := SmartCacheConfig{
		MaxTTL:          10 * time.Minute,
		MinTTL:          30 * time.Second,
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	req := inboundDto.AuthRequestDTO{
		Username: "testuser",
		Password: "testpass",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := inboundDto.AuthResponseDTO{
		Token:     "abc123token",
		ExpiresIn: 3600, // 1 hora
	}

	mockInner.On("Login", ctx, req, xApplication, correlationID).Return(expectedResponse, nil).Once()

	// Act
	response, err := decorator.Login(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, response)

	// Verifica se metadados do token foram armazenados
	metadata, found := decorator.cache.getTokenMetadata(response.Token, xApplication)
	assert.True(t, found)
	assert.NotNil(t, metadata)
	assert.Equal(t, response.Token, metadata.token)
	assert.Equal(t, xApplication, metadata.xApplication)

	mockInner.AssertExpectations(t)
}

func TestSmartCacheDecorator_ValidateToken_UsesDynamicTTL(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := SmartCacheConfig{
		MaxTTL:          10 * time.Minute,
		MinTTL:          30 * time.Second,
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	token := "abc123token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// 1. Simula Login que armazena metadados
	loginReq := inboundDto.AuthRequestDTO{Username: "user", Password: "pass"}
	loginResponse := inboundDto.AuthResponseDTO{
		Token:     token,
		ExpiresIn: 3600, // 1 hora
	}
	mockInner.On("Login", ctx, loginReq, xApplication, correlationID).Return(loginResponse, nil).Once()
	decorator.Login(ctx, loginReq, xApplication, correlationID)

	// 2. Primeira validação: cache miss
	mockInner.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil).Once()
	err1 := decorator.ValidateToken(ctx, token, xApplication, correlationID)
	assert.NoError(t, err1)

	// 3. Segunda validação: cache hit (dentro do TTL do token)
	err2 := decorator.ValidateToken(ctx, token, xApplication, correlationID)
	assert.NoError(t, err2)

	// 4. Terceira validação: ainda cache hit
	err3 := decorator.ValidateToken(ctx, token, xApplication, correlationID)
	assert.NoError(t, err3)

	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 1) // Apenas 1 chamada, resto veio do cache
}

func TestSmartCacheDecorator_ValidateToken_RespectsMaxTTL(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := SmartCacheConfig{
		MaxTTL:          1 * time.Minute, // MaxTTL de apenas 1 minuto
		MinTTL:          10 * time.Second,
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	token := "long-lived-token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// Simula token com ExpiresIn muito alto (10 horas)
	loginReq := inboundDto.AuthRequestDTO{Username: "user", Password: "pass"}
	loginResponse := inboundDto.AuthResponseDTO{
		Token:     token,
		ExpiresIn: 36000, // 10 horas (mas MaxTTL é 1 minuto)
	}
	mockInner.On("Login", ctx, loginReq, xApplication, correlationID).Return(loginResponse, nil).Once()
	decorator.Login(ctx, loginReq, xApplication, correlationID)

	// Verifica se TTL foi limitado ao MaxTTL
	metadata, found := decorator.cache.getTokenMetadata(token, xApplication)
	assert.True(t, found)

	remainingTTL := metadata.getRemainingTTL()
	assert.LessOrEqual(t, remainingTTL, config.MaxTTL)
	assert.Greater(t, remainingTTL, 50*time.Second) // Deve estar próximo de 1 minuto

	mockInner.AssertExpectations(t)
}

func TestSmartCacheDecorator_ValidateToken_RespectsMinTTL(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := SmartCacheConfig{
		MaxTTL:          10 * time.Minute,
		MinTTL:          1 * time.Minute, // MinTTL de 1 minuto
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	token := "short-lived-token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// Simula token com ExpiresIn muito baixo (10 segundos)
	loginReq := inboundDto.AuthRequestDTO{Username: "user", Password: "pass"}
	loginResponse := inboundDto.AuthResponseDTO{
		Token:     token,
		ExpiresIn: 10, // 10 segundos (mas MinTTL é 1 minuto)
	}
	mockInner.On("Login", ctx, loginReq, xApplication, correlationID).Return(loginResponse, nil).Once()
	decorator.Login(ctx, loginReq, xApplication, correlationID)

	// Verifica se TTL foi ajustado ao MinTTL
	metadata, found := decorator.cache.getTokenMetadata(token, xApplication)
	assert.True(t, found)

	remainingTTL := metadata.getRemainingTTL()
	assert.GreaterOrEqual(t, remainingTTL, 50*time.Second) // Deve estar próximo de 1 minuto

	mockInner.AssertExpectations(t)
}

func TestSmartCacheDecorator_Logout_InvalidatesToken(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultSmartCacheConfig()

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	token := "token-to-invalidate"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// 1. Simula Login que armazena metadados
	loginReq := inboundDto.AuthRequestDTO{Username: "user", Password: "pass"}
	loginResponse := inboundDto.AuthResponseDTO{Token: token, ExpiresIn: 3600}
	mockInner.On("Login", ctx, loginReq, xApplication, correlationID).Return(loginResponse, nil).Once()
	decorator.Login(ctx, loginReq, xApplication, correlationID)

	// 2. Valida token (popula cache)
	mockInner.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil).Once()
	decorator.ValidateToken(ctx, token, xApplication, correlationID)

	// 3. Verifica que token está no cache
	_, found := decorator.cache.getTokenMetadata(token, xApplication)
	assert.True(t, found)

	// 4. Faz logout
	mockInner.On("Logout", ctx, token, xApplication, correlationID).Return(nil).Once()
	err := decorator.Logout(ctx, token, xApplication, correlationID)
	assert.NoError(t, err)

	// 5. Verifica que token foi invalidado
	_, found = decorator.cache.getTokenMetadata(token, xApplication)
	assert.False(t, found)

	// 6. Próxima validação deve chamar ms-auth novamente (cache foi invalidado)
	mockInner.On("ValidateToken", ctx, token, xApplication, correlationID).Return(nil).Once()
	decorator.ValidateToken(ctx, token, xApplication, correlationID)

	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 2) // 2 chamadas (cache foi invalidado)
}

func TestSmartCacheDecorator_ValidateToken_CachesErrors(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultSmartCacheConfig()

	decorator := NewSmartCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	token := "invalid-token"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("token inválido")

	// Primeira validação: retorna erro
	mockInner.On("ValidateToken", ctx, token, xApplication, correlationID).Return(expectedError).Once()

	// Act
	err1 := decorator.ValidateToken(ctx, token, xApplication, correlationID)
	err2 := decorator.ValidateToken(ctx, token, xApplication, correlationID) // Deve vir do cache

	// Assert
	assert.Error(t, err1)
	assert.Equal(t, expectedError, err1)
	assert.Error(t, err2)
	assert.Equal(t, expectedError, err2)

	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ValidateToken", 1) // Apenas 1 chamada (erro foi cacheado)
}

func TestSmartCacheDecorator_RefreshToken_UpdatesMetadata(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultSmartCacheConfig()

	decorator := NewSmartCacheDecorator(mockInner, config, nil).(*smartCacheDecorator)

	ctx := context.Background()
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// 1. Login inicial
	loginReq := inboundDto.AuthRequestDTO{Username: "user", Password: "pass"}
	loginResponse := inboundDto.AuthResponseDTO{Token: "old-token", ExpiresIn: 3600}
	mockInner.On("Login", ctx, loginReq, xApplication, correlationID).Return(loginResponse, nil).Once()
	decorator.Login(ctx, loginReq, xApplication, correlationID)

	// 2. Refresh gera novo token
	refreshReq := inboundDto.RefreshTokenRequestDTO{Token: "old-token"}
	refreshResponse := inboundDto.RefreshTokenResponseDTO{Token: "new-token", ExpiresIn: 7200}
	mockInner.On("RefreshToken", ctx, refreshReq, xApplication, correlationID).Return(refreshResponse, nil).Once()

	// Act
	response, err := decorator.RefreshToken(ctx, refreshReq, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "new-token", response.Token)

	// Verifica que novo token foi armazenado
	metadata, found := decorator.cache.getTokenMetadata("new-token", xApplication)
	assert.True(t, found)
	assert.Equal(t, "new-token", metadata.token)

	mockInner.AssertExpectations(t)
}

func TestSmartCacheDecorator_ChangePassword_NoCaching(t *testing.T) {
	// Arrange
	mockInner := new(MockAuthClient)
	config := DefaultSmartCacheConfig()

	decorator := NewSmartCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           "user-123",
		CurrentPassword:    "old",
		NewPassword:        "new",
		ConfirmNewPassword: "new",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// Operações de escrita não devem ser cacheadas
	mockInner.On("ChangePassword", ctx, req, xApplication, correlationID).Return(nil).Twice()

	// Act
	err1 := decorator.ChangePassword(ctx, req, xApplication, correlationID)
	err2 := decorator.ChangePassword(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "ChangePassword", 2) // Duas chamadas (sem cache)
}
