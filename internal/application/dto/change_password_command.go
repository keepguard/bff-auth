package dto

import "context"

// ChangePasswordCommand representa o comando de alteração de senha no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de alteração de senha
type ChangePasswordCommand struct {
	Token              string
	CurrentPassword    string
	NewPassword        string
	ConfirmNewPassword string
	TenantId       string
	CorrelationID      string
	Context            context.Context
}

// NewChangePasswordCommand cria uma nova instância do ChangePasswordCommand
func NewChangePasswordCommand(token, currentPassword, newPassword, confirmNewPassword, tenantId, correlationID string, ctx context.Context) ChangePasswordCommand {
	return ChangePasswordCommand{
		Token:              token,
		CurrentPassword:    currentPassword,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		TenantId:       tenantId,
		CorrelationID:      correlationID,
		Context:            ctx,
	}
}

// Validate valida os dados do comando de alteração de senha
func (c *ChangePasswordCommand) Validate() error {
	if c.Token == "" {
		return &ValidationError{Field: "token", Message: "token é obrigatório"}
	}
	if c.CurrentPassword == "" {
		return &ValidationError{Field: "currentPassword", Message: "currentPassword é obrigatório"}
	}
	if c.NewPassword == "" {
		return &ValidationError{Field: "newPassword", Message: "newPassword é obrigatório"}
	}
	if c.ConfirmNewPassword == "" {
		return &ValidationError{Field: "confirmNewPassword", Message: "confirmNewPassword é obrigatório"}
	}
	if c.NewPassword != c.ConfirmNewPassword {
		return &ValidationError{Field: "confirmNewPassword", Message: "newPassword e confirmNewPassword devem ser iguais"}
	}
	if c.TenantId == "" {
		return &ValidationError{Field: "tenantId", Message: "tenantId é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "correlationID é obrigatório"}
	}
	if c.Context == nil {
		return &ValidationError{Field: "context", Message: "context é obrigatório"}
	}
	return nil
}
