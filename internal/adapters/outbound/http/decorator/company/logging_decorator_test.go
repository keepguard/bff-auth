package company

import (
	"context"
	"errors"
	"testing"

	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// MockCompanyClient é um mock para CompanyClient
type MockCompanyClient struct {
	mock.Mock
}

func (m *MockCompanyClient) GetByXApplication(ctx context.Context, xApplication, correlationID string) (portsclient.CompanySimpleResponseDTO, error) {
	args := m.Called(ctx, xApplication, correlationID)
	return args.Get(0).(portsclient.CompanySimpleResponseDTO), args.Error(1)
}

func TestNewCompanyLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewCompanyLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &companyLoggingDecorator{}, decorator)
}

func TestCompanyLoggingDecorator_GetByXApplication_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-company"

	decorator := NewCompanyLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := portsclient.CompanySimpleResponseDTO{
		ID:           "company-123",
		XApplication: xApplication,
		Name:         "Test Company",
		LegalName:    "Test Company LTDA",
		CNPJ:         "12345678000199",
		Status:       "ACTIVE",
	}

	mockInner.On("GetByXApplication", ctx, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Requisição concluída com sucesso", logs.All()[1].Message)
}

func TestCompanyLoggingDecorator_GetByXApplication_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCompanyClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-company"

	decorator := NewCompanyLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	xApplication := "invalid-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("company not found")

	mockInner.On("GetByXApplication", ctx, xApplication, correlationID).Return(portsclient.CompanySimpleResponseDTO{}, expectedError)

	// Act
	result, err := decorator.GetByXApplication(ctx, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, portsclient.CompanySimpleResponseDTO{}, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Erro na requisição", logs.All()[1].Message)
}
