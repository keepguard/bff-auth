package auth

import (
	"context"
	"net/http"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// metricsDecorator implementa o padrão decorator para coleta de métricas
type metricsDecorator struct {
	inner       authclient.AuthClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewMetricsDecorator cria um decorator de métricas para AuthClient
func NewMetricsDecorator(
	inner authclient.AuthClient,
	metrics *metrics.Metrics,
	serviceName string,
) authclient.AuthClient {
	return &metricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro HTTP
func (d *metricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Tenta converter para HTTPError
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.StatusCode
	}

	// Default para 500 se não conseguir identificar
	return http.StatusInternalServerError
}

// Login implementa o método Login com coleta de métricas
func (d *metricsDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, xApplication, correlationID string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.Login(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/login", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/login", errorType)
	}

	return response, err
}

// RefreshToken implementa o método RefreshToken com coleta de métricas
func (d *metricsDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, xApplication, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.RefreshToken(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/refresh", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/refresh", errorType)
	}

	return response, err
}

// Logout implementa o método Logout com coleta de métricas
func (d *metricsDecorator) Logout(ctx context.Context, token, xApplication, correlationID string) error {
	start := time.Now()

	err := d.inner.Logout(ctx, token, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/logout", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/logout", errorType)
	}

	return err
}

// ValidateToken implementa o método ValidateToken com coleta de métricas
func (d *metricsDecorator) ValidateToken(ctx context.Context, token, xApplication, correlationID string) error {
	start := time.Now()

	err := d.inner.ValidateToken(ctx, token, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/validate", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/validate", errorType)
	}

	return err
}

// ChangePassword implementa o método ChangePassword com coleta de métricas
func (d *metricsDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, xApplication, correlationID string) error {
	start := time.Now()

	err := d.inner.ChangePassword(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/change-password", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/change-password", errorType)
	}

	return err
}

// ResetPassword implementa o método ResetPassword com coleta de métricas
func (d *metricsDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, xApplication, correlationID string) error {
	start := time.Now()

	err := d.inner.ResetPassword(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/reset-password", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/reset-password", errorType)
	}

	return err
}

// GenerateResetToken implementa o método GenerateResetToken com coleta de métricas
func (d *metricsDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, xApplication, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GenerateResetToken(ctx, req, xApplication, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/generate-reset-token", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/generate-reset-token", errorType)
	}

	return response, err
}
