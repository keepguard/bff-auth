package dto

import "context"

// ValidateTokenCommand representa o comando de validação de token no domínio
// Encapsula todos os parâmetros necessários para executar o caso de uso de validação de token
type ValidateTokenCommand struct {
	Token         string
	XApplication  string
	CorrelationID string
	Context       context.Context
}

// NewValidateTokenCommand cria uma nova instância do ValidateTokenCommand
func NewValidateTokenCommand(token, xApplication, correlationID string, ctx context.Context) ValidateTokenCommand {
	return ValidateTokenCommand{
		Token:         token,
		XApplication:  xApplication,
		CorrelationID: correlationID,
		Context:       ctx,
	}
}

// Validate valida os dados do comando de validação de token
func (c *ValidateTokenCommand) Validate() error {
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

