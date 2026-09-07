package dto

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
}

func NewResetPasswordCommand(
	email, resetToken, newPassword, confirmNewPassword, tenantId, correlationID,
	deviceId, deviceName, deviceType, ipAddress, userAgent string,
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
	}
}

func (c *ResetPasswordCommand) Validate() error {
	if c.Email == "" {
		return &ValidationError{Field: "email", Message: "E-mail é obrigatório"}
	}
	if c.ResetToken == "" {
		return &ValidationError{Field: "resetToken", Message: "Token de reset é obrigatório"}
	}
	if c.NewPassword == "" {
		return &ValidationError{Field: "newPassword", Message: "Nova senha é obrigatória"}
	}
	if c.ConfirmNewPassword == "" {
		return &ValidationError{Field: "confirmNewPassword", Message: "Confirmação da nova senha é obrigatória"}
	}
	if c.NewPassword != c.ConfirmNewPassword {
		return &ValidationError{Field: "confirmNewPassword", Message: "Nova senha e confirmação devem ser iguais"}
	}
	if c.TenantId == "" {
		return &ValidationError{Field: "tenantId", Message: "Tenant é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "Identificador de correlação é obrigatório"}
	}
	return nil
}
