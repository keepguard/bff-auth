package dto

type LogoutCommand struct {
	Token         string
	TenantId      string
	CorrelationID string
}

func NewLogoutCommand(token, tenantId, correlationID string) LogoutCommand {
	return LogoutCommand{
		Token:         token,
		TenantId:      tenantId,
		CorrelationID: correlationID,
	}
}

func (c *LogoutCommand) Validate() error {
	if c.Token == "" {
		return &ValidationError{Field: "token", Message: "Token é obrigatório"}
	}
	if c.TenantId == "" {
		return &ValidationError{Field: "tenantId", Message: "Tenant é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "Identificador de correlação é obrigatório"}
	}
	return nil
}
