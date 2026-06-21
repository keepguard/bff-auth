package communication

import (
	"context"
	"net/http"
	"testing"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/stretchr/testify/assert"
)

func TestNewRetryDecorator(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	config := DefaultRetryConfig()

	// Act
	decorator := NewRetryDecorator(mockInner, config)

	// Assert
	assert.NotNil(t, decorator)
	assert.IsType(t, &retryDecorator{}, decorator)
}

func TestRetryDecorator_SendMessage_SuccessFirstAttempt(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	config := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.5,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "user@example.com",
		TemplateType: "RECUPERACAO_SENHA",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.SendMessageResponseDTO{
		Success: true,
		Message: "Message sent",
	}

	mockInner.On("SendMessage", ctx, req, xApplication, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "SendMessage", 1)
}

func TestRetryDecorator_SendMessage_SuccessAfterRetry(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	config := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.5,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "user@example.com",
		TemplateType: "RECUPERACAO_SENHA",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	expectedResponse := outboundDto.SendMessageResponseDTO{
		Success: true,
		Message: "Message sent",
	}

	// Primeira tentativa: 503 (SendGrid instável)
	retryableError := &appdto.HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service temporarily unavailable",
		ErrorCode:  "SERVICE_UNAVAILABLE",
	}
	mockInner.On("SendMessage", ctx, req, xApplication, correlationID).Return(outboundDto.SendMessageResponseDTO{}, retryableError).Once()

	// Segunda tentativa: sucesso
	mockInner.On("SendMessage", ctx, req, xApplication, correlationID).Return(expectedResponse, nil).Once()

	// Act
	result, err := decorator.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse, result)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "SendMessage", 2) // 2 chamadas (retry salvou!)
}

func TestRetryDecorator_SendMessage_MaxAttemptsReached(t *testing.T) {
	// Arrange
	mockInner := new(MockCommunicationClient)
	config := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.5,
		Jitter:       false,
	}

	decorator := NewRetryDecorator(mockInner, config)

	ctx := context.Background()
	req := outboundDto.SendMessageRequestDTO{
		MessageType:  "EMAIL",
		Recipient:    "user@example.com",
		TemplateType: "RECUPERACAO_SENHA",
	}
	xApplication := "test-app-id"
	correlationID := "test-correlation-id"

	// Sempre retorna 503
	retryableError := &appdto.HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service unavailable",
		ErrorCode:  "SERVICE_UNAVAILABLE",
	}

	mockInner.On("SendMessage", ctx, req, xApplication, correlationID).Return(outboundDto.SendMessageResponseDTO{}, retryableError).Times(4)

	// Act
	_, err := decorator.SendMessage(ctx, req, xApplication, correlationID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, retryableError, err)
	mockInner.AssertExpectations(t)
	mockInner.AssertNumberOfCalls(t, "SendMessage", 4) // Todas as 4 tentativas
}
