package dto

import "context"

// LoginCommand representa o comando de login no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de login
type LoginCommand struct {
	Username      string
	Password      string
	TenantId  string
	CorrelationID string
	ClientId      string
	Context       context.Context
}

// NewLoginCommand cria uma nova instância do LoginCommand
func NewLoginCommand(username, password, tenantId, correlationID, clientId string, ctx context.Context) LoginCommand {
	return LoginCommand{
		Username:      username,
		Password:      password,
		TenantId:  tenantId,
		CorrelationID: correlationID,
		ClientId:      clientId,
		Context:       ctx,
	}
}

// Validate valida os dados do comando de login
func (c *LoginCommand) Validate() error {
	if c.Username == "" {
		return &ValidationError{Field: "username", Message: "username é obrigatório"}
	}
	if c.Password == "" {
		return &ValidationError{Field: "password", Message: "password é obrigatório"}
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
