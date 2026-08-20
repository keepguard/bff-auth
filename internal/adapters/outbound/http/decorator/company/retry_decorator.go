package company

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// RetryConfig configuração para retry
type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryConfig retorna configuração padrão para CompanyClient
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  2,                     // 2 tentativas (rápido!)
		InitialDelay: 50 * time.Millisecond, // Muito rápido
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// retryDecorator implementa retry para CompanyClient
type retryDecorator struct {
	inner  portsclient.CompanyClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para CompanyClient
func NewRetryDecorator(
	inner portsclient.CompanyClient,
	config RetryConfig,
) portsclient.CompanyClient {
	return &retryDecorator{
		inner:  inner,
		config: config,
	}
}

// isRetryableError verifica se um erro deve ser retentado
func (d *retryDecorator) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		switch httpErr.StatusCode {
		case http.StatusTooManyRequests,
			http.StatusRequestTimeout,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
			http.StatusBadGateway:
			return true
		}
		return false
	}

	return true // Erros de rede são retryable
}

// calculateDelay calcula o delay com backoff exponencial
func (d *retryDecorator) calculateDelay(attempt int) time.Duration {
	delay := float64(d.config.InitialDelay) * math.Pow(d.config.Multiplier, float64(attempt))

	if delay > float64(d.config.MaxDelay) {
		delay = float64(d.config.MaxDelay)
	}

	if d.config.Jitter {
		jitterRange := delay * 0.25
		jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		delay += jitter

		if delay < 0 {
			delay = float64(d.config.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// retry executa uma função com retry
func (d *retryDecorator) retry(ctx context.Context, operation func() (portsclient.CompanySimpleResponseDTO, error)) (portsclient.CompanySimpleResponseDTO, error) {
	var lastErr error
	var lastResult portsclient.CompanySimpleResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return portsclient.CompanySimpleResponseDTO{}, ctx.Err()
		default:
		}

		result, err := operation()

		if err == nil {
			return result, nil
		}

		lastErr = err
		lastResult = result

		if !d.isRetryableError(err) || attempt == d.config.MaxAttempts-1 {
			return lastResult, err
		}

		delay := d.calculateDelay(attempt)

		select {
		case <-time.After(delay):
			// Continua para próxima tentativa
		case <-ctx.Done():
			return portsclient.CompanySimpleResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// GetByTenantId implementa GetByTenantId com retry
func (d *retryDecorator) GetByTenantId(ctx context.Context, tenantId, correlationID string) (portsclient.CompanySimpleResponseDTO, error) {
	return d.retry(ctx, func() (portsclient.CompanySimpleResponseDTO, error) {
		return d.inner.GetByTenantId(ctx, tenantId, correlationID)
	})
}
