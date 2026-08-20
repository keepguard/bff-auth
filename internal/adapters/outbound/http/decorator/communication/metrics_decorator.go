package communication

import (
	"context"
	"net/http"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// communicationMetricsDecorator implementa métricas para CommunicationClient
type communicationMetricsDecorator struct {
	inner       portsclient.CommunicationClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewCommunicationMetricsDecorator cria um decorator de métricas para CommunicationClient
func NewCommunicationMetricsDecorator(
	inner portsclient.CommunicationClient,
	metrics *metrics.Metrics,
	serviceName string,
) portsclient.CommunicationClient {
	return &communicationMetricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro
func (d *communicationMetricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.StatusCode
	}

	return http.StatusInternalServerError
}

// SendMessage implementa SendMessage com métricas
func (d *communicationMetricsDecorator) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.SendMessage(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/messages/send", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/messages/send", errorType)
	}

	return response, err
}
