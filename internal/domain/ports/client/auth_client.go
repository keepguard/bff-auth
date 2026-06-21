package client

import (
	"context"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
)

// AuthClient define a interface para o cliente de autenticação
type AuthClient interface {
	Login(ctx context.Context, req inboundDto.AuthRequestDTO, xApplication, correlationID string) (inboundDto.AuthResponseDTO, error)
	RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, xApplication, correlationID string) (inboundDto.RefreshTokenResponseDTO, error)
	Logout(ctx context.Context, token, xApplication, correlationID string) error
	ValidateToken(ctx context.Context, token, xApplication, correlationID string) error
	ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, xApplication, correlationID string) error
	ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, xApplication, correlationID string) error
	GenerateResetToken(ctx context.Context, req map[string]interface{}, xApplication, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error)
}
