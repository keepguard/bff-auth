package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	commclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// communicationClient implementa CommunicationClient usando HTTP para acessar o ms-communication
type communicationClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewCommunicationClient cria uma nova instância do CommunicationClient
func NewCommunicationClient(config *config.Config, logger *zap.Logger) commclient.CommunicationClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(3)
	httpClient.SetRetryWaitTime(1 * time.Second)

	return &communicationClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewCommunicationClientWithLogger cria uma nova instância do CommunicationClient usando logger.Logger
func NewCommunicationClientWithLogger(config *config.Config, log logger.Logger) commclient.CommunicationClient {
	zapLogger, _ := zap.NewDevelopment()
	return NewCommunicationClient(config, zapLogger)
}

// handleHTTPError trata erros HTTP e extrai informações do ms-communication
func (c *communicationClient) handleHTTPError(resp *resty.Response, operation string) error {
	// Se não é um erro HTTP, retorna o erro original
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}

	// Tenta extrair erro do ms-communication (ProblemDetail)
	var msError appdto.MSAuthErrorResponse
	if err := json.Unmarshal(resp.Body(), &msError); err == nil && msError.Title != "" {
		errorCode := "UNKNOWN_ERROR"
		if code, ok := msError.Properties["errorCode"].(string); ok {
			errorCode = code
		}
		// Erro estruturado do ms-communication
		return &appdto.HTTPError{
			StatusCode: msError.Status,
			Message:    msError.Detail,
			ErrorCode:  errorCode,
			Details:    msError.Properties,
		}
	}

	// Fallback para erro genérico
	return &appdto.HTTPError{
		StatusCode: resp.StatusCode(),
		Message:    fmt.Sprintf("erro de %s: status %d", operation, resp.StatusCode()),
		ErrorCode:  "HTTP_ERROR",
		Details:    map[string]interface{}{"body": string(resp.Body())},
	}
}

// SendMessage envia uma mensagem através do ms-communication
func (c *communicationClient) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/messages/send", c.config.Services.Communication.BaseURL)

	var response outboundDto.SendMessageResponseDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		SetResult(&response).
		Post(url)

	if err != nil {
		return outboundDto.SendMessageResponseDTO{}, fmt.Errorf("erro ao comunicar com ms-communication: %w", err)
	}

	// Trata erros HTTP do ms-communication
	if httpErr := c.handleHTTPError(resp, "send message"); httpErr != nil {
		return outboundDto.SendMessageResponseDTO{}, httpErr
	}

	return response, nil
}
