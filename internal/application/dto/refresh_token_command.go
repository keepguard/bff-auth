package dto

type RefreshTokenCommand struct {
	// LegacyToken é o JWT usado como credencial de rotação no fluxo antigo
	// (pré refresh-token opaco). Mantido por compatibilidade.
	LegacyToken string
	// OpaqueRefreshToken é o refresh token opaco do novo fluxo, sempre lido
	// do cookie HttpOnly — nunca do corpo da requisição.
	OpaqueRefreshToken string
	TenantId           string
	CorrelationID      string
	ClientId           string
}

func NewRefreshTokenCommand(legacyToken, opaqueRefreshToken, tenantId, correlationID, clientId string) RefreshTokenCommand {
	return RefreshTokenCommand{
		LegacyToken:        legacyToken,
		OpaqueRefreshToken: opaqueRefreshToken,
		TenantId:           tenantId,
		CorrelationID:      correlationID,
		ClientId:           clientId,
	}
}

func (c *RefreshTokenCommand) Validate() error {
	if c.LegacyToken == "" && c.OpaqueRefreshToken == "" {
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
