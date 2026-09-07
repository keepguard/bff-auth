package auth

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
)

type LoginUseCase interface {
	Execute(ctx context.Context, command appdto.LoginCommand) (appdto.AuthResponseDTO, error)
}

type RefreshUseCase interface {
	Execute(ctx context.Context, command appdto.RefreshTokenCommand) (appdto.RefreshTokenResponseDTO, error)
}

type LogoutUseCase interface {
	Execute(ctx context.Context, command appdto.LogoutCommand) error
}

type ValidateTokenUseCase interface {
	Execute(ctx context.Context, command appdto.ValidateTokenCommand) error
}

type ChangePasswordUseCase interface {
	Execute(ctx context.Context, command appdto.ChangePasswordCommand) error
}

type ResetPasswordUseCase interface {
	Execute(ctx context.Context, command appdto.ResetPasswordCommand) error
}
