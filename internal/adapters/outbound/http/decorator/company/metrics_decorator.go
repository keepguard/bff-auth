package company

import (
	"context"
	"net/http"
	"time"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// companyMetricsDecorator implementa métricas para CompanyClient
type companyMetricsDecorator struct {
	inner       portsclient.CompanyClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewCompanyMetricsDecorator cria um decorator de métricas para CompanyClient
func NewCompanyMetricsDecorator(
	inner portsclient.CompanyClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.CompanyClient {
	return &companyMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *companyMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.StatusCode
	}

	return http.StatusInternalServerError
}

// GetByXApplication implementa GetByXApplication com métricas
func (d *companyMetricsDecorator) GetByXApplication(ctx context.Context, xApplication, correlationID string) (portsclient.CompanySimpleResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GetByXApplication(ctx, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/companies/x-application", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "GET", "/companies/x-application", errorType)
	}

	return response, err
}
