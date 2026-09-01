package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
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
	return buildHTTPErrorFromResponse(resp, operation)
}

// SendMessage envia uma mensagem através do ms-communication
func (c *communicationClient) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/messages/send", c.config.Services.Communication.BaseURL)

	var response outboundDto.SendMessageResponseDTO
	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
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
