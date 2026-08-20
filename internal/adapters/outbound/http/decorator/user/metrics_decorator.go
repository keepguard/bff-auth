package user

import (
	"context"
	"net/http"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// userMetricsDecorator implementa métricas para UserClient
type userMetricsDecorator struct {
	inner       portsclient.UserClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewUserMetricsDecorator cria um decorator de métricas para UserClient
func NewUserMetricsDecorator(
	inner portsclient.UserClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.UserClient {
	return &userMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *userMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.StatusCode
	}

	return http.StatusInternalServerError
}

// GetByEmail implementa GetByEmail com métricas
func (d *userMetricsDecorator) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GetByEmail(ctx, email, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/email", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "GET", "/users/email", errorType)
	}

	return response, err
}
