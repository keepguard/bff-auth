package communication

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
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

// DefaultRetryConfig retorna configuração padrão para CommunicationClient
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  4,                      // 4 tentativas (APIs de e-mail são instáveis)
		InitialDelay: 200 * time.Millisecond, // Mais lento (API externa)
		MaxDelay:     10 * time.Second,       // Aguenta esperar mais
		Multiplier:   2.5,                    // Backoff agressivo
		Jitter:       true,
	}
}

// retryDecorator implementa retry para CommunicationClient
type retryDecorator struct {
	inner  portsclient.CommunicationClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para CommunicationClient
func NewRetryDecorator(
	inner portsclient.CommunicationClient,
	config RetryConfig,
) portsclient.CommunicationClient {
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
func (d *retryDecorator) retry(ctx context.Context, operation func() (outboundDto.SendMessageResponseDTO, error)) (outboundDto.SendMessageResponseDTO, error) {
	var lastErr error
	var lastResult outboundDto.SendMessageResponseDTO

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return outboundDto.SendMessageResponseDTO{}, ctx.Err()
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
			return outboundDto.SendMessageResponseDTO{}, ctx.Err()
		}
	}

	return lastResult, lastErr
}

// SendMessage implementa SendMessage com retry
func (d *retryDecorator) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	return d.retry(ctx, func() (outboundDto.SendMessageResponseDTO, error) {
		return d.inner.SendMessage(ctx, req, tenantId, correlationID)
	})
}
