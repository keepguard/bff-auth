package company

import (
	"context"
	"errors"
	"testing"
	"time"

	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
)

func TestNewCacheDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := DefaultCacheConfig()

	// Act
	decorator := NewCacheDecorator(mockInner, config, nil)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &cacheDecorator{}, decorator)
}

func TestCacheDecorator_GetByXApplication_CacheHit(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:           "company-123",
		XApplication: xApplication,
		Name:         "Test Company",
		Status:       "ACTIVE",
	}

	// Primeira chamada: cache miss
	mockInner.On("GetByXApplication", ctx, xApplication, correlationID).Return(expectedResponse, nil).Once()

	// Act - Primeira chamada
	result1, err1 := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Act - Segunda chamada (deve vir do cache)
	result2, err2 := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result1)
	assert.Equal(t, expectedResponse, result2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByXApplication", 1) // Apenas 1 chamada (cache hit na 2ª)
}

func TestCacheDecorator_GetByXApplication_CacheMiss(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := CacheConfig{
		TTL:             50 * time.Millisecond, // TTL curto para teste
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	// Duas chamadas com espera entre elas
	mockInner.On("GetByXApplication", ctx, xApplication, correlationID).Return(expectedResponse, nil).Twice()

	// Act - Primeira chamada
	result1, err1 := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Aguarda cache expirar
	time.Sleep(60 * time.Millisecond)

	// Act - Segunda chamada (cache expirado, deve chamar novamente)
	result2, err2 := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result1)
	assert.Equal(t, expectedResponse, result2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByXApplication", 2) // 2 chamadas (cache expirou)
}

func TestCacheDecorator_GetByXApplication_ErrorNotCached(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := DefaultCacheConfig()

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	xApplication := "invalid-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("company not found")

	// Erro não deve ser cacheado, então ambas as chamadas vão ao ms-company
	mockInner.On("GetByXApplication", ctx, xApplication, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, expectedError).Twice()

	// Act
	_, err1 := decorator.GetByXApplication(ctx, xApplication, correlationID)
	_, err2 := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Assert
	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, expectedError, err1)
	assert.Equal(t, expectedError, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByXApplication", 2) // 2 chamadas (erro não é cacheado)
}
