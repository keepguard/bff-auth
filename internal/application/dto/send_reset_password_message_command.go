package dto

import (
	"context"
	"errors"
	"regexp"
)

// SendResetPasswordMessageCommand representa o comando de envio de mensagem de reset de senha
type SendResetPasswordMessageCommand struct {
	Email         string
	TenantId  string
	CorrelationID string
	Context       context.Context
}

// NewSendResetPasswordMessageCommand cria um novo comando de envio de mensagem de reset
func NewSendResetPasswordMessageCommand(
	email string,
	tenantId string,
	correlationID string,
	ctx context.Context,
) SendResetPasswordMessageCommand {
	return SendResetPasswordMessageCommand{
		Email:         email,
		TenantId:  tenantId,
		CorrelationID: correlationID,
		Context:       ctx,
	}
}

// Validate valida o comando
func (c SendResetPasswordMessageCommand) Validate() error {
	if c.Email == "" {
		return errors.New("email é obrigatório")
	}

	// Validação de formato de email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(c.Email) {
		return errors.New("formato de email inválido")
	}

	if c.TenantId == "" {
		return errors.New("tenantId é obrigatório")
	}

	if c.CorrelationID == "" {
		return errors.New("correlationID é obrigatório")
	}

	if c.Context == nil {
		return errors.New("contexto é obrigatório")
	}

	return nil
}
