package client

import (
	"context"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
)

// AuthClient define a interface para o cliente de autenticação
type AuthClient interface {
	Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId string) (inboundDto.AuthResponseDTO, error)
	RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error)
	Logout(ctx context.Context, token, tenantId, correlationID string) error
	ValidateToken(ctx context.Context, token, tenantId, correlationID string) error
	ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID string) error
	ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID string) error
	GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error)
}
