package dto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewSendResetPasswordMessageCommand testa a criação do comando
func TestNewSendResetPasswordMessageCommand(t *testing.T) {
	// Arrange
	email := "test@example.com"
	tenantId := "test-app-uuid"
	correlationID := "test-correlation-id"
	ctx := context.Background()

	// Act
	command := NewSendResetPasswordMessageCommand(email, tenantId, correlationID, ctx)

	// Assert
	assert.Equal(t, email, command.Email)
	assert.Equal(t, tenantId, command.TenantId)
	assert.Equal(t, correlationID, command.CorrelationID)
	assert.Equal(t, ctx, command.Context)
}

// TestSendResetPasswordMessageCommand_Validate_Success testa validação com sucesso
func TestSendResetPasswordMessageCommand_Validate_Success(t *testing.T) {
	// Arrange
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:  "test-app-uuid",
		CorrelationID: "test-correlation-id",
		Context:       context.Background(),
	}

	// Act
	err := command.Validate()

	// Assert
	assert.NoError(t, err)
}

// TestSendResetPasswordMessageCommand_Validate_EmptyEmail testa validação com email vazio
func TestSendResetPasswordMessageCommand_Validate_EmptyEmail(t *testing.T) {
	// Arrange
	command := SendResetPasswordMessageCommand{
		Email:         "",
		TenantId:  "test-app-uuid",
		CorrelationID: "test-correlation-id",
		Context:       context.Background(),
	}

	// Act
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "email é obrigatório", err.Error())
}

// TestSendResetPasswordMessageCommand_Validate_InvalidEmailFormat testa validação com formato de email inválido
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
			// Arrange
			command := SendResetPasswordMessageCommand{
				Email:         tt.email,
				TenantId:  "test-app-uuid",
				CorrelationID: "test-correlation-id",
				Context:       context.Background(),
			}

			// Act
			err := command.Validate()

			// Assert
			assert.Error(t, err)
			assert.Equal(t, "formato de email inválido", err.Error())
		})
	}
}

// TestSendResetPasswordMessageCommand_Validate_ValidEmailFormats testa validação com vários formatos válidos de email
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
			// Arrange
			command := SendResetPasswordMessageCommand{
				Email:         tt.email,
				TenantId:  "test-app-uuid",
				CorrelationID: "test-correlation-id",
				Context:       context.Background(),
			}

			// Act
			err := command.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

// TestSendResetPasswordMessageCommand_Validate_EmptyTenantId testa validação com tenantId vazio
func TestSendResetPasswordMessageCommand_Validate_EmptyTenantId(t *testing.T) {
	// Arrange
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:  "",
		CorrelationID: "test-correlation-id",
		Context:       context.Background(),
	}

	// Act
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "tenantId é obrigatório", err.Error())
}

// TestSendResetPasswordMessageCommand_Validate_EmptyCorrelationID testa validação com correlationID vazio
func TestSendResetPasswordMessageCommand_Validate_EmptyCorrelationID(t *testing.T) {
	// Arrange
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:  "test-app-uuid",
		CorrelationID: "",
		Context:       context.Background(),
	}

	// Act
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "correlationID é obrigatório", err.Error())
}

// TestSendResetPasswordMessageCommand_Validate_NilContext testa validação com contexto nil
func TestSendResetPasswordMessageCommand_Validate_NilContext(t *testing.T) {
	// Arrange
	command := SendResetPasswordMessageCommand{
		Email:         "test@example.com",
		TenantId:  "test-app-uuid",
		CorrelationID: "test-correlation-id",
		Context:       nil,
	}

	// Act
	err := command.Validate()

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "contexto é obrigatório", err.Error())
}
