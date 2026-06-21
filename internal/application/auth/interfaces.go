package auth

import (
	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
)

// LoginUseCase define a interface para o caso de uso de login
type LoginUseCase interface {
	Execute(command appdto.LoginCommand) (dto.AuthResponseDTO, error)
}

// RefreshUseCase define a interface para o caso de uso de refresh
type RefreshUseCase interface {
	Execute(command appdto.RefreshTokenCommand) (dto.RefreshTokenResponseDTO, error)
}

// LogoutUseCase define a interface para o caso de uso de logout
type LogoutUseCase interface {
	Execute(command appdto.LogoutCommand) error
}

// ValidateTokenUseCase define a interface para o caso de uso de validação de token
type ValidateTokenUseCase interface {
	Execute(command appdto.ValidateTokenCommand) error
}

// ChangePasswordUseCase define a interface para o caso de uso de alteração de senha
type ChangePasswordUseCase interface {
	Execute(command appdto.ChangePasswordCommand) error
}

// ResetPasswordUseCase define a interface para o caso de uso de reset de senha
type ResetPasswordUseCase interface {
	Execute(command appdto.ResetPasswordCommand) error
}
