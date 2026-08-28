package dto

// ChangePasswordMSRequestDTO representa a requisição de alteração de senha para o ms-auth
type ChangePasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	CurrentPassword    string `json:"currentPassword"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
}

// ResetPasswordMSRequestDTO representa a requisição de reset de senha para o ms-auth
type ResetPasswordMSRequestDTO struct {
	CodeUser           string `json:"codeUser"`
	ResetToken         string `json:"resetToken"`
	NewPassword        string `json:"newPassword"`
	ConfirmNewPassword string `json:"confirmNewPassword"`
	MessageType        string `json:"messageType"`
	TemplateType       string `json:"templateType"`
}

// UserByEmailResponseDTO representa a resposta de busca de usuário por email do ms-auth
type UserByEmailResponseDTO struct {
	ID            string `json:"id"`
	CodeUser      string `json:"codeUser"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Status        string `json:"status"`
	EmailVerified bool   `json:"emailVerified"`
}

// UserByCodeResponseDTO representa a busca de usuário por codeUser no ms-auth.
type UserByCodeResponseDTO struct {
	ID                string `json:"id"`
	IDUserExternal    string `json:"idUserExternal"`
	IDUserExternalAlt string `json:"id_user_external"`
	CodeUser          string `json:"codeUser"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Status            string `json:"status"`
}

// ExternalID retorna o idUserExternal independentemente do naming do JSON.
func (u UserByCodeResponseDTO) ExternalID() string {
	if u.IDUserExternal != "" {
		return u.IDUserExternal
	}
	return u.IDUserExternalAlt
}

// GenerateResetTokenMSRequestDTO representa a requisição de geração de token de reset para o ms-auth
type GenerateResetTokenMSRequestDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
}

// GenerateResetTokenMSResponseDTO representa a resposta de geração de token de reset do ms-auth
type GenerateResetTokenMSResponseDTO struct {
	CodeUser          string `json:"codeUser"`
	MessageType       string `json:"messageType"`
	CommunicationType string `json:"communicationType"`
	TemplateType      string `json:"templateType"`
	Token             string `json:"token"`
	ExpiresInSeconds  int64  `json:"expiresInSeconds"`
}
