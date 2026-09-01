package dto

import "context"

// LogoutCommand representa o comando de logout no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de logout
type LogoutCommand struct {
	Token         string
	TenantId  string
	CorrelationID string
	Context       context.Context
}

// NewLogoutCommand cria uma nova instância do LogoutCommand
func NewLogoutCommand(token, tenantId, correlationID string, ctx context.Context) LogoutCommand {
	return LogoutCommand{
		Token:         token,
		TenantId:  tenantId,
		CorrelationID: correlationID,
		Context:       ctx,
	}
}

// Validate valida os dados do comando de logout
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
	if c.Context == nil {
		return &ValidationError{Field: "context", Message: "Contexto da requisição é obrigatório"}
	}
	return nil
}
