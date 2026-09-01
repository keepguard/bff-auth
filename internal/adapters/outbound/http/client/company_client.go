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
	"github.com/keepguard/bff-auth/internal/pkg"
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
		return client.CompanySimpleResponseDTO{}, fmt.Errorf("erro ao comunicar com o serviço de empresas: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		var rawMessage string
		errorCode := companyErrorCodeForStatus(resp.StatusCode())

		var msError appdto.MSAuthErrorResponse
		if jsonErr := json.Unmarshal(resp.Body(), &msError); jsonErr == nil && (msError.Detail != "" || msError.Title != "") {
			rawMessage = msError.Detail
			if rawMessage == "" {
				rawMessage = msError.Title
			}
			if code, ok := msError.Properties["errorCode"].(string); ok && code != "" {
				errorCode = code
			}
		} else {
			var simple struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			if jsonErr := json.Unmarshal(resp.Body(), &simple); jsonErr == nil {
				rawMessage = simple.Message
				if rawMessage == "" {
					rawMessage = simple.Error
				}
			}
		}

		errorMsg := pkg.ResolveUserMessage(errorCode, rawMessage)
		if errorMsg == "" {
			errorMsg = companyDefaultMessageForStatus(resp.StatusCode())
		}

		return client.CompanySimpleResponseDTO{}, &appdto.HTTPError{
			StatusCode: resp.StatusCode(),
			Message:    errorMsg,
			ErrorCode:  errorCode,
			Details:    map[string]interface{}{"body": string(resp.Body())},
		}
	}

	var company client.CompanySimpleResponseDTO
	if err := json.Unmarshal(resp.Body(), &company); err != nil {
		return client.CompanySimpleResponseDTO{}, fmt.Errorf("erro ao fazer parse da resposta: %w", err)
	}

	return company, nil
}

func companyErrorCodeForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "COMPANY_NOT_FOUND"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		return "HTTP_ERROR"
	}
}

func companyDefaultMessageForStatus(statusCode int) string {
	switch statusCode {
	case http.StatusNotFound:
		return "Empresa não encontrada"
	case http.StatusInternalServerError:
		return "Erro interno do servidor"
	default:
		return "Não foi possível consultar a empresa"
	}
}
