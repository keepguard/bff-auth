package dto

import "context"

// LogoutCommand representa o comando de logout no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de logout
type LogoutCommand struct {
	Token         string
	XApplication  string
	CorrelationID string
	Context       context.Context
}

// NewLogoutCommand cria uma nova instância do LogoutCommand
func NewLogoutCommand(token, xApplication, correlationID string, ctx context.Context) LogoutCommand {
	return LogoutCommand{
		Token:         token,
		XApplication:  xApplication,
		CorrelationID: correlationID,
		Context:       ctx,
	}
}

// Validate valida os dados do comando de logout
func (c *LogoutCommand) Validate() error {
	if c.Token == "" {
		return &ValidationError{Field: "token", Message: "token é obrigatório"}
	}
	if c.XApplication == "" {
		return &ValidationError{Field: "xApplication", Message: "xApplication é obrigatório"}
	}
	if c.CorrelationID == "" {
		return &ValidationError{Field: "correlationID", Message: "correlationID é obrigatório"}
	}
	if c.Context == nil {
		return &ValidationError{Field: "context", Message: "context é obrigatório"}
	}
	return nil
}
