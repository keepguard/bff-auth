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
func (d *circuitBreakerDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, xApplication, correlationID string) (inboundDto.AuthResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.Login(ctx, req, xApplication, correlationID)
	})

	if err != nil {
		return inboundDto.AuthResponseDTO{}, err
	}

	return result.(inboundDto.AuthResponseDTO), nil
}

// RefreshToken implementa o método RefreshToken com circuit breaker
func (d *circuitBreakerDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, xApplication, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.RefreshToken(ctx, req, xApplication, correlationID)
	})

	if err != nil {
		return inboundDto.RefreshTokenResponseDTO{}, err
	}

	return result.(inboundDto.RefreshTokenResponseDTO), nil
}

// Logout implementa o método Logout com circuit breaker
func (d *circuitBreakerDecorator) Logout(ctx context.Context, token, xApplication, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.Logout(ctx, token, xApplication, correlationID)
	})

	return err
}

// ValidateToken implementa o método ValidateToken com circuit breaker
func (d *circuitBreakerDecorator) ValidateToken(ctx context.Context, token, xApplication, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ValidateToken(ctx, token, xApplication, correlationID)
	})

	return err
}

// ChangePassword implementa o método ChangePassword com circuit breaker
func (d *circuitBreakerDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, xApplication, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ChangePassword(ctx, req, xApplication, correlationID)
	})

	return err
}

// ResetPassword implementa o método ResetPassword com circuit breaker
func (d *circuitBreakerDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, xApplication, correlationID string) error {
	_, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return nil, d.inner.ResetPassword(ctx, req, xApplication, correlationID)
	})

	return err
}

// GenerateResetToken implementa o método GenerateResetToken com circuit breaker
func (d *circuitBreakerDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, xApplication, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	result, err := d.circuitBreaker.Execute(ctx, d.serviceName, func() (interface{}, error) {
		return d.inner.GenerateResetToken(ctx, req, xApplication, correlationID)
	})

	if err != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, err
	}

	return result.(outboundDto.GenerateResetTokenMSResponseDTO), nil
}
