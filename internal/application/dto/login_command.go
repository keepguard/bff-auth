package dto

type LoginCommand struct {
	Username      string
	Password      string
	TenantId      string
	CorrelationID string
	ClientId      string
	DeviceId      string
	DeviceName    string
	DeviceType    string
	IPAddress     string
	UserAgent     string
}

func NewLoginCommand(username, password, tenantId, correlationID, clientId string) LoginCommand {
	return LoginCommand{
		Username:      username,
		Password:      password,
		TenantId:      tenantId,
		CorrelationID: correlationID,
		ClientId:      clientId,
	}
}

func NewLoginCommandWithDevice(username, password, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) LoginCommand {
	return LoginCommand{
		Username:      username,
		Password:      password,
		TenantId:      tenantId,
		CorrelationID: correlationID,
		ClientId:      clientId,
		DeviceId:      deviceId,
		DeviceName:    deviceName,
		DeviceType:    deviceType,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}
}

func (c *LoginCommand) Validate() error {
	if c.Username == "" {
		return &ValidationError{Field: "username", Message: "Nome de usuário é obrigatório"}
	}
	if c.Password == "" {
		return &ValidationError{Field: "password", Message: "Senha é obrigatória"}
	}
	if c.TenantId == "" {
		return &ValidationError{Field: "tenantId", Message: "Tenant é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "Identificador de correlação é obrigatório"}
	}
	return nil
}
