package auth

import (
	"context"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/resilience"
)

// circuitBreakerDecorator implementa o padrão decorator para circuit breaker
type circuitBreakerDecorator struct {
	inner          authclient.AuthClient
	circuitBreaker *resilience.CircuitBreakerManager
	serviceName    string
}

// NewCircuitBreakerDecorator cria um decorator de circuit breaker para AuthClient
func NewCircuitBreakerDecorator(
	inner authclient.AuthClient,
	cbManager *resilience.CircuitBreakerManager,
	serviceName string,
) authclient.AuthClient {
	return &circuitBreakerDecorator{
		inner:          inner,
		circuitBreaker: cbManager,
		serviceName:    serviceName,
	}
}

// Login implementa o método Login com circuit breaker
func (d *circuitBreakerDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)
	})

	if err != nil {
		return inboundDto.AuthResponseDTO{}, err
	}

	return result.(inboundDto.AuthResponseDTO), nil
}

// RefreshToken implementa o método RefreshToken com circuit breaker
func (d *circuitBreakerDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)
	})

	if err != nil {
		return inboundDto.RefreshTokenResponseDTO{}, err
	}

	return result.(inboundDto.RefreshTokenResponseDTO), nil
}

// Logout implementa o método Logout com circuit breaker
func (d *circuitBreakerDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.Logout(ctx, token, tenantId, correlationID)
	})

	return err
}

// ValidateToken implementa o método ValidateToken com circuit breaker
func (d *circuitBreakerDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ValidateToken(ctx, token, tenantId, correlationID)
	})

	return err
}

// ChangePassword implementa o método ChangePassword com circuit breaker
func (d *circuitBreakerDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ChangePassword(ctx, req, tenantId, correlationID)
	})

	return err
}

// ResetPassword implementa o método ResetPassword com circuit breaker
func (d *circuitBreakerDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ResetPassword(ctx, req, tenantId, correlationID)
	})

	return err
}

// GenerateResetToken implementa o método GenerateResetToken com circuit breaker
func (d *circuitBreakerDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, err
	}

	return result.(outboundDto.GenerateResetTokenMSResponseDTO), nil
}

// SendDeviceChallenge implementa o método SendDeviceChallenge com circuit breaker
func (d *circuitBreakerDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return nil, err
	}

	return result.(map[string]interface{}), nil
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge com circuit breaker
func (d *circuitBreakerDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
	})

	if err != nil {
		return inboundDto.AuthResponseDTO{}, err
	}

	return result.(inboundDto.AuthResponseDTO), nil
}

// ListUserSessions implementa o método ListUserSessions com circuit breaker
func (d *circuitBreakerDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
	})

	if err != nil {
		return nil, err
	}

	return result.([]inboundDto.DeviceSessionDTO), nil
}

// RevokeSession implementa o método RevokeSession com circuit breaker
func (d *circuitBreakerDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
	})
	return err
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions com circuit breaker
func (d *circuitBreakerDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
	})
	return err
}
