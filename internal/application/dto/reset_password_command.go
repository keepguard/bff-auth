package dto

import "context"

// ResetPasswordCommand representa o comando de reset de senha no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de reset de senha
type ResetPasswordCommand struct {
	Email              string
	ResetToken         string
	NewPassword        string
	ConfirmNewPassword string
	TenantId           string
	CorrelationID      string
	DeviceId           string
	DeviceName         string
	DeviceType         string
	IpAddress          string
	UserAgent          string
	Context            context.Context
}

// NewResetPasswordCommand cria uma nova instância do ResetPasswordCommand
func NewResetPasswordCommand(
	email, resetToken, newPassword, confirmNewPassword, tenantId, correlationID,
	deviceId, deviceName, deviceType, ipAddress, userAgent string,
	ctx context.Context,
) ResetPasswordCommand {
	return ResetPasswordCommand{
		Email:              email,
		ResetToken:         resetToken,
		NewPassword:        newPassword,
		ConfirmNewPassword: confirmNewPassword,
		TenantId:           tenantId,
		CorrelationID:      correlationID,
		DeviceId:           deviceId,
		DeviceName:         deviceName,
		DeviceType:         deviceType,
		IpAddress:          ipAddress,
		UserAgent:          userAgent,
		Context:            ctx,
	}
}

// Validate valida os dados do comando de reset de senha
func (c *ResetPasswordCommand) Validate() error {
	if c.Email == "" {
		return &ValidationError{Field: "email", Message: "email é obrigatório"}
	}
	if c.ResetToken == "" {
		return &ValidationError{Field: "resetToken", Message: "resetToken é obrigatório"}
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
