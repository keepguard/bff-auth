package primary

import (
	"context"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
)

// LoginUseCase define o port primário para login
type LoginUseCase interface {
	Execute(ctx context.Context, req dto.AuthRequestDTO, correlationID string) (dto.AuthResponseDTO, error)
}

// RefreshTokenUseCase define o port primário para refresh de token
type RefreshTokenUseCase interface {
	Execute(ctx context.Context, req dto.RefreshTokenRequestDTO, correlationID string) (dto.RefreshTokenResponseDTO, error)
}

// LogoutUseCase define o port primário para logout
type LogoutUseCase interface {
	Execute(ctx context.Context, token string, correlationID string) error
}
