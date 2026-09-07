package dto

import appdto "github.com/keepguard/bff-auth/internal/application/dto"

type AuthRequestDTO = appdto.AuthRequestDTO
type RefreshTokenRequestDTO = appdto.RefreshTokenRequestDTO

type LogoutRequestDTO struct {
	Token string `json:"token,omitempty"`
}

type ValidateTokenRequestDTO struct {
	Token string `json:"token" validate:"required,min=1"`
}

type ChangePasswordRequestDTO struct {
	CurrentPassword    string `json:"currentPassword" validate:"required,min=6"`
	NewPassword        string `json:"newPassword" validate:"required,min=6"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,min=6"`
}

type ResetPasswordRequestDTO struct {
	Email              string `json:"email" validate:"required,email"`
	ResetToken         string `json:"resetToken" validate:"required,min=1"`
	NewPassword        string `json:"newPassword" validate:"required,min=6"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,min=6"`
}
