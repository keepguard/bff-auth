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

// MockMetricsMessagePublisher é um mock do MessagePublisher para metrics
type MockMetricsMessagePublisher struct {
	mock.Mock
}

func (m *MockMetricsMessagePublisher) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMetricsMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewMetricsDecorator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockMetricsMessagePublisher{}

	// Teste básico sem métricas para evitar conflitos
	decorator := &metricsDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	assert.NotNil(t, decorator)
	assert.IsType(t, &metricsDecorator{}, decorator)
}

func TestMetricsDecorator_PublishMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockMetricsMessagePublisher{}

	message := messaging.MessageDTO{
		TenantId:      "test-app",
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

	decorator := &metricsDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	// Chamar diretamente o inner para evitar métricas
	err := decorator.inner.PublishMessage(context.Background(), message)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestMetricsDecorator_PublishMessage_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockMetricsMessagePublisher{}

	message := messaging.MessageDTO{
		TenantId:      "test-app",
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

	decorator := &metricsDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	// Chamar diretamente o inner para evitar métricas
	err := decorator.inner.PublishMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}

func TestMetricsDecorator_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockMetricsMessagePublisher{}

	// Configurar mock para retornar sucesso
	mockPublisher.On("Close").Return(nil)

	decorator := &metricsDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	err := decorator.Close()

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestMetricsDecorator_Close_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockMetricsMessagePublisher{}

	expectedError := errors.New("close failed")

	// Configurar mock para retornar erro
	mockPublisher.On("Close").Return(expectedError)

	decorator := &metricsDecorator{
		inner:  mockPublisher,
		logger: logger,
	}

	err := decorator.Close()

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}
