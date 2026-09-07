package dto

type ChangePasswordCommand struct {
	Token              string
	CurrentPassword    string
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

func NewChangePasswordCommand(
	token, currentPassword, newPassword, confirmNewPassword, tenantId, correlationID,
	deviceId, deviceName, deviceType, ipAddress, userAgent string,
) ChangePasswordCommand {
	return ChangePasswordCommand{
		Token:              token,
		CurrentPassword:    currentPassword,
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

func (c *ChangePasswordCommand) Validate() error {
	if c.Token == "" {
		return &ValidationError{Field: "token", Message: "Token é obrigatório"}
	}
	if c.CurrentPassword == "" {
		return &ValidationError{Field: "currentPassword", Message: "Senha atual é obrigatória"}
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
