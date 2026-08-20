package auth

import (
	"context"
	"math"
	"math/rand"
	"net/http"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// RetryConfig configuração para retry
type RetryConfig struct {
	MaxAttempts  int           // Número máximo de tentativas
	InitialDelay time.Duration // Delay inicial
	MaxDelay     time.Duration // Delay máximo
	Multiplier   float64       // Multiplicador para backoff exponencial
	Jitter       bool          // Se deve adicionar jitter aleatório
}

// DefaultRetryConfig retorna uma configuração padrão de retry
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// retryDecorator implementa o padrão decorator para retry inteligente
type retryDecorator struct {
	inner  authclient.AuthClient
	config RetryConfig
}

// NewRetryDecorator cria um decorator de retry para AuthClient
func NewRetryDecorator(
	inner authclient.AuthClient,
	config RetryConfig,
) authclient.AuthClient {
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

	// Verifica se é HTTPError
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		// Retry apenas para erros temporários
		switch httpErr.StatusCode {
		case http.StatusTooManyRequests, // 429
			http.StatusRequestTimeout,     // 408
			http.StatusServiceUnavailable, // 503
			http.StatusGatewayTimeout,     // 504
			http.StatusBadGateway:         // 502
			return true
		}
		return false
	}

	// Erros de rede/timeout são retryable
	return true
}

// calculateDelay calcula o delay para a próxima tentativa com backoff exponencial
func (d *retryDecorator) calculateDelay(attempt int) time.Duration {
	// Backoff exponencial: initialDelay * (multiplier ^ attempt)
	delay := float64(d.config.InitialDelay) * math.Pow(d.config.Multiplier, float64(attempt))

	// Aplica o limite máximo
	if delay > float64(d.config.MaxDelay) {
		delay = float64(d.config.MaxDelay)
	}

	// Adiciona jitter se configurado (±25%)
	if d.config.Jitter {
		jitterRange := delay * 0.25
		jitter := (rand.Float64() * 2 * jitterRange) - jitterRange
		delay += jitter

		// Garante que não seja negativo
		if delay < 0 {
			delay = float64(d.config.InitialDelay)
		}
	}

	return time.Duration(delay)
}

// retry executa uma função com retry
func (d *retryDecorator) retry(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt < d.config.MaxAttempts; attempt++ {
		// Verifica se o contexto foi cancelado
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Executa a operação
		err := operation()

		// Se sucesso, retorna
		if err == nil {
			return nil
		}

		lastErr = err

		// Se não é retryable ou é a última tentativa, retorna o erro
		if !d.isRetryableError(err) || attempt == d.config.MaxAttempts-1 {
			return err
		}

		// Calcula o delay e aguarda
		delay := d.calculateDelay(attempt)

		select {
		case <-time.After(delay):
			// Continua para próxima tentativa
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// Login implementa o método Login com retry
func (d *retryDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	var result inboundDto.AuthResponseDTO
	var err error

	retryErr := d.retry(ctx, func() error {
		result, err = d.inner.Login(ctx, req, tenantId, correlationID)
		return err
	})

	return result, retryErr
}

// RefreshToken implementa o método RefreshToken com retry
func (d *retryDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
	var result inboundDto.RefreshTokenResponseDTO
	var err error

	retryErr := d.retry(ctx, func() error {
		result, err = d.inner.RefreshToken(ctx, req, tenantId, correlationID)
		return err
	})

	return result, retryErr
}

// Logout implementa o método Logout com retry
func (d *retryDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.Logout(ctx, token, tenantId, correlationID)
	})
}

// ValidateToken implementa o método ValidateToken com retry
func (d *retryDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.ValidateToken(ctx, token, tenantId, correlationID)
	})
}

// ChangePassword implementa o método ChangePassword com retry
func (d *retryDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.ChangePassword(ctx, req, tenantId, correlationID)
	})
}

// ResetPassword implementa o método ResetPassword com retry
func (d *retryDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.ResetPassword(ctx, req, tenantId, correlationID)
	})
}

// GenerateResetToken implementa o método GenerateResetToken com retry
func (d *retryDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	var response outboundDto.GenerateResetTokenMSResponseDTO
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)
		return innerErr
	})
	return response, err
}
