package dto

import "context"

// RefreshTokenCommand representa o comando de refresh de token no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de refresh
type RefreshTokenCommand struct {
	RefreshToken  string
	TenantId  string
	CorrelationID string
	Context       context.Context
}

// NewRefreshTokenCommand cria uma nova instância do RefreshTokenCommand
func NewRefreshTokenCommand(refreshToken, tenantId, correlationID string, ctx context.Context) RefreshTokenCommand {
	return RefreshTokenCommand{
		RefreshToken:  refreshToken,
		TenantId:  tenantId,
		CorrelationID: correlationID,
		Context:       ctx,
	}
}

// Validate valida os dados do comando de refresh token
func (c *RefreshTokenCommand) Validate() error {
	if c.RefreshToken == "" {
		return &ValidationError{Field: "refreshToken", Message: "refreshToken é obrigatório"}
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
