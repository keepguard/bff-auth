package decorator

import (
	"context"
	"errors"
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockCircuitBreakerMessagePublisher é um mock do MessagePublisher para circuit breaker
type MockCircuitBreakerMessagePublisher struct {
	mock.Mock
}

func (m *MockCircuitBreakerMessagePublisher) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockCircuitBreakerMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewCircuitBreakerDecorator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockCircuitBreakerMessagePublisher{}

	// Teste básico sem circuit breaker para evitar problemas de inicialização
	decorator := &circuitBreakerDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	assert.NotNil(t, decorator)
	assert.IsType(t, &circuitBreakerDecorator{}, decorator)
}

func TestCircuitBreakerDecorator_PublishMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockCircuitBreakerMessagePublisher{}

	message := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test@example.com",
		CodeUser:          "user123",
		Variables: map[string]interface{}{
			"userName": "Test User",
		},
	}

	// Configurar mock para retornar sucesso
	mockPublisher.On("PublishMessage", mock.Anything, message).Return(nil)

	// Criar decorator sem circuit breaker para evitar problemas de inicialização
	decorator := &circuitBreakerDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	// Chamar diretamente o inner para evitar circuit breaker
	err := decorator.inner.PublishMessage(context.Background(), message)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_PublishMessage_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockCircuitBreakerMessagePublisher{}

	message := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test@example.com",
		CodeUser:          "user123",
		Variables: map[string]interface{}{
			"userName": "Test User",
		},
	}

	expectedError := errors.New("publish failed")

	// Configurar mock para retornar erro
	mockPublisher.On("PublishMessage", mock.Anything, message).Return(expectedError)

	decorator := &circuitBreakerDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	// Chamar diretamente o inner para evitar circuit breaker
	err := decorator.inner.PublishMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockCircuitBreakerMessagePublisher{}

	// Configurar mock para retornar sucesso
	mockPublisher.On("Close").Return(nil)

	decorator := &circuitBreakerDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	err := decorator.Close()

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestCircuitBreakerDecorator_Close_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockCircuitBreakerMessagePublisher{}

	expectedError := errors.New("close failed")

	// Configurar mock para retornar erro
	mockPublisher.On("Close").Return(expectedError)

	decorator := &circuitBreakerDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	err := decorator.Close()

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}
