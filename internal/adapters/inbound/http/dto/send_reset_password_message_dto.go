package dto

// SendResetPasswordMessageRequestDTO representa a requisição de envio de mensagem de reset de senha
type SendResetPasswordMessageRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// SendResetPasswordMessageResponseDTO representa a resposta do envio de mensagem
type SendResetPasswordMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
