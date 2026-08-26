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
func (d *retryDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	var result inboundDto.AuthResponseDTO
	var err error

	retryErr := d.retry(ctx, func() error {
		result, err = d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)
		return err
	})

	return result, retryErr
}

// RefreshToken implementa o método RefreshToken com retry
func (d *retryDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	var result inboundDto.RefreshTokenResponseDTO
	var err error

	retryErr := d.retry(ctx, func() error {
		result, err = d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)
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
func (d *retryDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	return d.retry(ctx, func() error {
		return d.inner.ChangePassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
	})
}

// ResetPassword implementa o método ResetPassword com retry
func (d *retryDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	return d.retry(ctx, func() error {
		return d.inner.ResetPassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
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

// SendDeviceChallenge implementa o método SendDeviceChallenge com retry
func (d *retryDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
		return innerErr
	})
	return response, err
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge com retry
func (d *retryDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	var response inboundDto.AuthResponseDTO
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
		return innerErr
	})
	return response, err
}

// ListUserSessions implementa o método ListUserSessions com retry
func (d *retryDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	var response []inboundDto.DeviceSessionDTO
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
		return innerErr
	})
	return response, err
}

// RevokeSession implementa o método RevokeSession com retry
func (d *retryDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
	})
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions com retry
func (d *retryDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
	})
}

// QuickRevoke implementa o método QuickRevoke com retry
func (d *retryDecorator) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.QuickRevoke(ctx, token, blacklist, tenantId, correlationID)
		return innerErr
	})
	return response, err
}

// ListDeviceBlacklist implementa o método ListDeviceBlacklist com retry
func (d *retryDecorator) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	var response []inboundDto.DeviceBlacklistDTO
	err := d.retry(ctx, func() error {
		var innerErr error
		response, innerErr = d.inner.ListDeviceBlacklist(ctx, token, tenantId, correlationID)
		return innerErr
	})
	return response, err
}

// AddDeviceToBlacklist implementa o método AddDeviceToBlacklist com retry
func (d *retryDecorator) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.AddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
	})
}

// RemoveDeviceFromBlacklist implementa o método RemoveDeviceFromBlacklist com retry
func (d *retryDecorator) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	return d.retry(ctx, func() error {
		return d.inner.RemoveDeviceFromBlacklist(ctx, deviceId, token, tenantId, correlationID)
	})
}

func (d *retryDecorator) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error) {
	return d.inner.SearchAdminDeviceBlacklist(ctx, queryParams, token, tenantId, correlationID)
}

func (d *retryDecorator) AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.inner.AdminAddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

func (d *retryDecorator) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	return d.inner.AdminRemoveDeviceFromBlacklist(ctx, deviceId, userId, token, tenantId, correlationID)
}
