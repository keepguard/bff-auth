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

func TestCacheDecorator_GetByTenantId_CacheHit(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := CacheConfig{
		TTL:             1 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:           "company-123",
		TenantId: tenantId,
		Name:         "Test Company",
		Status:       "ACTIVE",
	}

	// Primeira chamada: cache miss
	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil).Once()

	// Act - Primeira chamada
	result1, err1 := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Act - Segunda chamada (deve vir do cache)
	result2, err2 := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result1)
	assert.Equal(t, expectedResponse, result2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 1) // Apenas 1 chamada (cache hit na 2ª)
}

func TestCacheDecorator_GetByTenantId_CacheMiss(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := CacheConfig{
		TTL:             50 * time.Millisecond, // TTL curto para teste
		MaxSize:         100,
		CleanupInterval: 0,
	}

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	// Duas chamadas com espera entre elas
	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil).Twice()

	// Act - Primeira chamada
	result1, err1 := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Aguarda cache expirar
	time.Sleep(60 * time.Millisecond)

	// Act - Segunda chamada (cache expirado, deve chamar novamente)
	result2, err2 := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, expectedResponse, result1)
	assert.Equal(t, expectedResponse, result2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 2) // 2 chamadas (cache expirou)
}

func TestCacheDecorator_GetByTenantId_ErrorNotCached(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	config := DefaultCacheConfig()

	decorator := NewCacheDecorator(mockInner, config, nil)

	ctx := context.Background()
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("company not found")

	// Erro não deve ser cacheado, então ambas as chamadas vão ao ms-company
	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, expectedError).Twice()

	// Act
	_, err1 := decorator.GetByTenantId(ctx, tenantId, correlationID)
	_, err2 := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, expectedError, err1)
	assert.Equal(t, expectedError, err2)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "GetByTenantId", 2) // 2 chamadas (erro não é cacheado)
}
