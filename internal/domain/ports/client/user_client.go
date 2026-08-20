package client

import (
	"context"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
)

// UserClient define a interface para o cliente de usuário do ms-auth
type UserClient interface {
	GetByEmail(ctx context.Context, email, tenantId, correlationID string) (outboundDto.UserByEmailResponseDTO, error)
}
