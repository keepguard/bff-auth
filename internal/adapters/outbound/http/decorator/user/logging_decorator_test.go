package user

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// MockUserClient é um mock para UserClient
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetByEmail(ctx context.Context, email, xApplication, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	args := m.Called(ctx, email, xApplication, correlationID)
	return args.Get(0).(outboundDto.UserByEmailResponseDTO), args.Error(1)
}

func TestNewUserLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &userLoggingDecorator{}, decorator)
}

func TestUserLoggingDecorator_GetByEmail_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth-users"

	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	email := "user@example.com"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.UserByEmailResponseDTO{
		ID:            "user-123",
		CodeUser:      "code-123",
		Username:      "testuser",
		Email:         email,
		Status:        "ACTIVE",
		EmailVerified: true,
	}

	mockInner.On("GetByEmail", ctx, email, xApplication, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.GetByEmail(ctx, email, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Requisição concluída com sucesso", logs.All()[1].Message)
}

func TestUserLoggingDecorator_GetByEmail_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockUserClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-auth-users"

	decorator := NewUserLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	email := "notfound@example.com"
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("user not found")

	mockInner.On("GetByEmail", ctx, email, xApplication, correlationID).Return(outboundDto.UserByEmailResponseDTO{}, expectedError)

	// Act
	_, err := decorator.GetByEmail(ctx, email, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Erro na requisição", logs.All()[1].Message)
}
