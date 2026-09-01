package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	userclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// authUserClient implementa UserClient usando HTTP para acessar usuários do ms-auth
type authUserClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewAuthUserClient cria uma nova instância do AuthUserClient
func NewAuthUserClient(config *config.Config, logger *zap.Logger) userclient.UserClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(3)
	httpClient.SetRetryWaitTime(1 * time.Second)

	return &authUserClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewAuthUserClientWithLogger cria uma nova instância do AuthUserClient usando logger.Logger
func NewAuthUserClientWithLogger(config *config.Config, log logger.Logger) userclient.UserClient {
	zapLogger, _ := zap.NewDevelopment()
	return NewAuthUserClient(config, zapLogger)
}

// handleHTTPError trata erros HTTP e extrai informações do ms-auth
func (c *authUserClient) handleHTTPError(resp *resty.Response, operation string) error {
	return buildHTTPErrorFromResponse(resp, operation)
}

// GetByEmail busca um usuário por email no ms-auth
func (c *authUserClient) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/users/email/%s", c.config.Services.Auth.BaseURL, email)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Company-Id", companyHeader(ctx)).
		SetHeader("X-Company-Id", companyId).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return outboundDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao comunicar com ms-auth user service: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "get user by email"); httpErr != nil {
		return outboundDto.UserByEmailResponseDTO{}, httpErr
	}

	var user outboundDto.UserByEmailResponseDTO
	if err := json.Unmarshal(resp.Body(), &user); err != nil {
		return outboundDto.UserByEmailResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return user, nil
}
