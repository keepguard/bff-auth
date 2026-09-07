package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSendResetPasswordMessageCommand(t *testing.T) {
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"

	command := NewSendResetPasswordMessageCommand(email, tenantId, correlationID)

	assert.Equal(t, email, command.Email)
	assert.Equal(t, tenantId, command.TenantId)
	assert.Equal(t, correlationID, command.CorrelationID)
}

func TestSendResetPasswordMessageCommand_Validate_Success(t *testing.T) {
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:      "test-app-uuid",
		CorrelationID: "test-correlation-id",
	}

	err := command.Validate()

	assert.NoError(t, err)
}

func TestSendResetPasswordMessageCommand_Validate_EmptyEmail(t *testing.T) {
	command := SendResetPasswordMessageCommand{
		Email:         "",
		TenantId:      "test-app-uuid",
		CorrelationID: "test-correlation-id",
	}

	err := command.Validate()

	assert.Error(t, err)
	assert.Equal(t, "email é obrigatório", err.Error())
}

func TestSendResetPasswordMessageCommand_Validate_InvalidEmailFormat(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"sem @", "testexample.com"},
		{"sem domínio", "test@"},
		{"sem local", "@example.com"},
		{"sem TLD", "test@example"},
		{"com espaços", "test @example.com"},
		{"inválido", "not-an-email"},
		{"duplo @", "test@@example.com"},
		{"sem ponto", "test@examplecom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := SendResetPasswordMessageCommand{
				Email:         tt.email,
				TenantId:      "test-app-uuid",
				CorrelationID: "test-correlation-id",
			}

			err := command.Validate()

			assert.Error(t, err)
			assert.Equal(t, "formato de email inválido", err.Error())
		})
	}
}

func TestSendResetPasswordMessageCommand_Validate_ValidEmailFormats(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"email simples", "test@example.com"},
		{"com pontos", "test.user@example.com"},
		{"com underline", "test_user@example.com"},
		{"com hífen", "test-user@example.com"},
		{"com números", "test123@example.com"},
		{"com plus", "test+tag@example.com"},
		{"domínio com hífen", "test@example-domain.com"},
		{"TLD longo", "test@example.technology"},
		{"subdomain", "test@mail.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command := SendResetPasswordMessageCommand{
				Email:         tt.email,
				TenantId:      "test-app-uuid",
				CorrelationID: "test-correlation-id",
			}

			err := command.Validate()

			assert.NoError(t, err)
		})
	}
}

func TestSendResetPasswordMessageCommand_Validate_EmptyTenantId(t *testing.T) {
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:      "",
		CorrelationID: "test-correlation-id",
	}

	err := command.Validate()

	assert.Error(t, err)
	assert.Equal(t, "tenantId é obrigatório", err.Error())
}

func TestSendResetPasswordMessageCommand_Validate_EmptyCorrelationID(t *testing.T) {
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:      "test-app-uuid",
		CorrelationID: "",
	}

	err := command.Validate()

	assert.Error(t, err)
	assert.Equal(t, "correlationID é obrigatório", err.Error())
}
