package decorator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

// MockLoggingMessagePublisher é um mock do MessagePublisher para logging
type MockLoggingMessagePublisher struct {
	mock.Mock
}

func (m *MockLoggingMessagePublisher) PublishMessage(ctx context.Context, message messaging.MessageDTO) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockLoggingMessagePublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewLoggingDecorator(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	decorator := NewLoggingDecorator(mockPublisher, logger)

	assert.NotNil(t, decorator)
	assert.IsType(t, &loggingDecorator{}, decorator)
}

func TestLoggingDecorator_PublishMessage_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

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

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(context.Background(), message)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_PublishMessage_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

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

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(context.Background(), message)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_PublishMessage_WithAllFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	message := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test@example.com",
		Subject:           "Test Subject",
		Content:           "Test Content",
		CodeUser:          "user123",
		Variables: map[string]interface{}{
			"userName": "Test User",
			"token":    "test-token",
		},
	}

	// Configurar mock para retornar sucesso
	mockPublisher.On("PublishMessage", mock.Anything, message).Return(nil)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(context.Background(), message)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_PublishMessage_EmptyFields(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	message := messaging.MessageDTO{
		XApplication:      "",
		XCorrelationID:    "",
		MessageType:       "",
		CommunicationType: "",
		TemplateType:      "",
		Recipient:         "",
		CodeUser:          "",
		Variables:         nil,
	}

	// Configurar mock para retornar sucesso
	mockPublisher.On("PublishMessage", mock.Anything, message).Return(nil)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(context.Background(), message)

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	// Configurar mock para retornar sucesso
	mockPublisher.On("Close").Return(nil)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.Close()

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_Close_Error(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	expectedError := errors.New("close failed")

	// Configurar mock para retornar erro
	mockPublisher.On("Close").Return(expectedError)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.Close()

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_ContextCancellation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

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

	// Criar contexto cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Configurar mock para retornar erro de contexto cancelado
	mockPublisher.On("PublishMessage", ctx, message).Return(context.Canceled)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(ctx, message)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_ContextTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

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

	// Criar contexto com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Aguardar timeout
	time.Sleep(1 * time.Millisecond)

	// Configurar mock para retornar erro de timeout
	mockPublisher.On("PublishMessage", ctx, message).Return(context.DeadlineExceeded)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err := decorator.PublishMessage(ctx, message)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	mockPublisher.AssertExpectations(t)
}

func TestLoggingDecorator_MultipleMessages(t *testing.T) {
	logger := zaptest.NewLogger(t)
	mockPublisher := &MockLoggingMessagePublisher{}

	message1 := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id-1",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test1@example.com",
		CodeUser:          "user123",
		Variables: map[string]interface{}{
			"userName": "Test User 1",
		},
	}

	message2 := messaging.MessageDTO{
		XApplication:      "test-app",
		XCorrelationID:    "test-correlation-id-2",
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RESET_PASSWORD",
		Recipient:         "test2@example.com",
		CodeUser:          "user456",
		Variables: map[string]interface{}{
			"userName": "Test User 2",
		},
	}

	// Configurar mocks para retornar sucesso
	mockPublisher.On("PublishMessage", mock.Anything, message1).Return(nil)
	mockPublisher.On("PublishMessage", mock.Anything, message2).Return(nil)

	decorator := NewLoggingDecorator(mockPublisher, logger)

	err1 := decorator.PublishMessage(context.Background(), message1)
	err2 := decorator.PublishMessage(context.Background(), message2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	mockPublisher.AssertExpectations(t)
}
