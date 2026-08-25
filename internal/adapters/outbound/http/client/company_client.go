package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"go.uber.org/zap"
)

// companyClient implementa CompanyClient usando HTTP
type companyClient struct {
	httpClient *resty.Client
	config     *config.Config
	logger     *zap.Logger
}

// NewCompanyClient cria uma nova instância do CompanyClient
func NewCompanyClient(config *config.Config, logger *zap.Logger) client.CompanyClient {
	httpClient := resty.New()
	httpClient.SetTimeout(30 * time.Second)
	httpClient.SetRetryCount(3)
	httpClient.SetRetryWaitTime(1 * time.Second)

	return &companyClient{
		httpClient: httpClient,
		config:     config,
		logger:     logger,
	}
}

// NewCompanyClientWithLogger cria uma nova instância do CompanyClient usando logger.Logger
func NewCompanyClientWithLogger(config *config.Config, log logger.Logger) client.CompanyClient {
	zapLogger, _ := zap.NewDevelopment()
	return NewCompanyClient(config, zapLogger)
}

// GetByTenantId busca uma empresa pelo X-Tenant-Id
func (c *companyClient) GetByTenantId(ctx context.Context, tenantId, correlationID string) (client.CompanySimpleResponseDTO, error) {
	url := fmt.Sprintf("%s/api/v1/companies/x-tenant-id/%s", c.config.Services.Company.BaseURL, tenantId)

	resp, err := c.httpClient.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("Content-Type", "application/json").
		Get(url)

	if err != nil {
		return client.CompanySimpleResponseDTO{}, fmt.Errorf("erro ao comunicar com company service: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		var errorMsg string
		var msError appdto.MSAuthErrorResponse
		if jsonErr := json.Unmarshal(resp.Body(), &msError); jsonErr == nil && (msError.Detail != "" || msError.Title != "") {
			errorMsg = msError.Detail
			if errorMsg == "" {
				errorMsg = msError.Title
			}
		} else {
			errorMsg = "company service retornou erro"
		}

		return client.CompanySimpleResponseDTO{}, &appdto.HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    errorMsg,
			ErrorCode:  "COMPANY_NOT_FOUND",
			Details:    map[string]interface{}{"body": string(resp.Body())},
		}
	}

	var company client.CompanySimpleResponseDTO
	if err := json.Unmarshal(resp.Body(), &company); err != nil {
		return client.CompanySimpleResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return company, nil
}
