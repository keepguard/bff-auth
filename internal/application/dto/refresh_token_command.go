package dto

type RefreshTokenCommand struct {
	RefreshToken  string
	TenantId      string
	CorrelationID string
	ClientId      string
}

func NewRefreshTokenCommand(refreshToken, tenantId, correlationID, clientId string) RefreshTokenCommand {
	return RefreshTokenCommand{
		RefreshToken:  refreshToken,
		TenantId:      tenantId,
		CorrelationID: correlationID,
		ClientId:      clientId,
	}
}

func (c *RefreshTokenCommand) Validate() error {
	if c.RefreshToken == "" {
		return &ValidationError{Field: "refreshToken", Message: "Token de refresh é obrigatório"}
	}
	if c.TenantId == "" {
		return &ValidationError{Field: "tenantId", Message: "Tenant é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "Identificador de correlação é obrigatório"}
	}
	return nil
}
