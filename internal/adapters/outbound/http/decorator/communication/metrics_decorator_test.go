package communication

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	// Instância compartilhada de métricas para todos os testes
	testMetrics = metrics.New()
)

func TestNewCommunicationMetricsDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	serviceName := "test-service"

	// Act
	decorator := NewCommunicationMetricsDecorator(mockInner, testMetrics, serviceName)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &communicationMetricsDecorator{}, decorator)
}

func TestCommunicationMetricsDecorator_SendMessage_Success(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	serviceName := "ms-communication"

	decorator := NewCommunicationMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "user@example.com",
		TemplateType: "RECUPERACAO_SENHA",
	}
	tenantId := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.SendMessageResponseDTO{
		Success: true,
		Message: "Message sent",
	}

	mockInner.On("SendMessage", ctx, req, tenantId, correlationID).Return(expectedResponse, nil)

	// Act
	result, err := decorator.SendMessage(ctx, req, tenantId, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
}

func TestCommunicationMetricsDecorator_SendMessage_Error(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	serviceName := "ms-communication"

	decorator := NewCommunicationMetricsDecorator(mockInner, testMetrics, serviceName)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "user@example.com",
		TemplateType: "INVALID",
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
}
