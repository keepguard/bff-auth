package dto

// AuthRequestDTO representa a requisição de autenticação
type AuthRequestDTO struct {
	Username  string `json:"username" validate:"required,username"`
	Password  string `json:"password" validate:"required,password"`
	CompanyID string `json:"companyId,omitempty"`
}

// RefreshTokenRequestDTO representa a requisição de refresh de token
type RefreshTokenRequestDTO struct {
	Token string `json:"token" validate:"required,min=1"`
}

// LogoutRequestDTO representa a requisição de logout
type LogoutRequestDTO struct {
	Token string `json:"token,omitempty"`
}

// ValidateTokenRequestDTO representa a requisição de validação de token
type ValidateTokenRequestDTO struct {
	Token string `json:"token" validate:"required,min=1"`
}

// ChangePasswordRequestDTO representa a requisição de alteração de senha
type ChangePasswordRequestDTO struct {
	CurrentPassword    string `json:"currentPassword" validate:"required,min=6"`
	NewPassword        string `json:"newPassword" validate:"required,min=6"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,min=6"`
}

// ResetPasswordRequestDTO representa a requisição de reset de senha
type ResetPasswordRequestDTO struct {
	Email              string `json:"email" validate:"required,email"`
	ResetToken         string `json:"resetToken" validate:"required,min=1"`
	NewPassword        string `json:"newPassword" validate:"required,min=6"`
	ConfirmNewPassword string `json:"confirmNewPassword" validate:"required,min=6"`
}
