package company

import (
	"context"
	"errors"
	"testing"

	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	// Instância compartilhada de métricas para todos os testes
	testMetrics = metrics.New()
)

func TestNewCompanyMetricsDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	serviceName := "test-service"

	// Act
	decorator := NewCompanyMetricsDecorator(mockInner, testMetrics, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &companyMetricsDecorator{}, decorator)
}

func TestCompanyMetricsDecorator_GetByTenantId_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:   "company-123",
		Name: "Test Company",
	}

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCompanyMetricsDecorator_GetByTenantId_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	serviceName := "ms-company"

	decorator := NewCompanyMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	tenantId := "invalid-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("company not found")

	mockInner.On("GetByTenantId", ctx, tenantId, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, expectedError)

	// Act
	_, err := decorator.GetByTenantId(ctx, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)
}
