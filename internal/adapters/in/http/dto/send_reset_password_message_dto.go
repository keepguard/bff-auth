package dto

// SendResetPasswordMessageRequestDTO representa a requisição de envio de mensagem de reset de senha
type SendResetPasswordMessageRequestDTO struct {
	Email string `json:"email" validate:"required,email"`
	// Locale is the language the person is using on the screen (optional, e.g. "en-US"). It picks the e-mail
	// language when the account has none stored; it is normalized to pt-BR/en-US/es-ES before use.
	Locale string `json:"locale,omitempty"`
}

// SendResetPasswordMessageResponseDTO representa a resposta do envio de mensagem
type SendResetPasswordMessageResponseDTO struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
