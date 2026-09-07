package dto

type ValidateTokenCommand struct {
	Token         string
	TenantId      string
	CorrelationID string
}

func NewValidateTokenCommand(token, tenantId, correlationID string) ValidateTokenCommand {
	return ValidateTokenCommand{
		Token:         token,
		TenantId:      tenantId,
		CorrelationID: correlationID,
	}
}

func (c *ValidateTokenCommand) Validate() error {
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
