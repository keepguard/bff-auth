package communication

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

// MockCommunicationClient é um mock para CommunicationClient
type MockCommunicationClient struct {
	mock.Mock
}

func (m *MockCommunicationClient) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	args := m.Called(ctx, req, tenantId, correlationID)
	return args.Get(0).(outboundDto.SendMessageResponseDTO), args.Error(1)
}

func TestNewCommunicationLoggingDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	logger, _ := zap.NewDevelopment()
	serviceName := "test-service"

	// Act
	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &communicationLoggingDecorator{}, decorator)
}

func TestCommunicationLoggingDecorator_SendMessage_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-communication"

	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:       "EMAIL",
		CommunicationType: "EMAIL",
		TemplateType:      "RECUPERACAO_SENHA",
		Recipient:         "user@example.com",
		Variables:         map[string]interface{}{"token": "123456"},
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.SendMessageResponseDTO{
		Success: true,
		Message: "Message sent successfully",
	}

	mockInner.On("SendMessage", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.SendMessage(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Mensagem enviada com sucesso", logs.All()[1].Message)
}

func TestCommunicationLoggingDecorator_SendMessage_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	serviceName := "ms-communication"

	decorator := NewCommunicationLoggingDecorator(mockInner, logger, serviceName)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "invalid@example.com",
		TemplateType: "INVALID_TEMPLATE",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedError := errors.New("template not found")

	mockInner.On("SendMessage", ctx, req, tenantId, correlationID).Return(outboundDto.SendMessageResponseDTO{}, expectedError)

	// Act
	_, err := decorator.SendMessage(ctx, req, tenantId, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockInner.AssertExpectations(t)

	// Verifica logs
	assert.Equal(t, 2, logs.Len())
	assert.Equal(t, "Iniciando requisição", logs.All()[0].Message)
	assert.Equal(t, "Erro na requisição", logs.All()[1].Message)
}
