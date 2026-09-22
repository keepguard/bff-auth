package dto

import (
	"errors"
	"testing"
)

func TestNewValidateTokenCommand(t *testing.T) {
	cmd := NewValidateTokenCommand("tok", "tenant-1", "corr-1")

	if cmd.Token != "tok" {
		t.Errorf("Token = %q", cmd.Token)
	}
	if cmd.TenantId != "tenant-1" {
		t.Errorf("TenantId = %q", cmd.TenantId)
	}
	if cmd.CorrelationID != "corr-1" {
		t.Errorf("CorrelationID = %q", cmd.CorrelationID)
	}
}

func TestValidateTokenCommand_Validate(t *testing.T) {
	t.Run("comando completo passa", func(t *testing.T) {
		cmd := NewValidateTokenCommand("tok", "tenant-1", "corr-1")
		if err := cmd.Validate(); err != nil {
			t.Errorf("recusou comando válido: %v", err)
		}
	})

	// O erro tem que dizer QUAL campo faltou: é o que vira a mensagem de
	// 400 para quem chamou a API.
	casos := []struct {
		nome      string
		cmd       ValidateTokenCommand
		querCampo string
	}{
		{
			nome:      "sem token",
			cmd:       NewValidateTokenCommand("", "tenant-1", "corr-1"),
			querCampo: "token",
		},
		{
			nome:      "sem tenant",
			cmd:       NewValidateTokenCommand("tok", "", "corr-1"),
			querCampo: "tenantId",
		},
		{
			nome:      "sem correlationID",
			cmd:       NewValidateTokenCommand("tok", "tenant-1", ""),
			querCampo: "correlationID",
		},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			err := c.cmd.Validate()
			if err == nil {
				t.Fatal("aceitou comando incompleto")
			}
			if !IsValidationError(err) {
				t.Fatalf("erro = %T, queria *ValidationError", err)
			}

			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatal("errors.As não reconheceu o ValidationError")
			}
			if ve.Field != c.querCampo {
				t.Errorf("Field = %q, quer %q", ve.Field, c.querCampo)
			}
			if ve.Message == "" {
				t.Error("Message vazia; vira resposta de erro sem explicação")
			}
		})
	}
}

// Token é validado antes do tenant: com os dois faltando, o erro tem que
// apontar o primeiro, não o último.
func TestValidateTokenCommand_OrdemDaValidacao(t *testing.T) {
	cmd := NewValidateTokenCommand("", "", "")

	var ve *ValidationError
	if !errors.As(cmd.Validate(), &ve) {
		t.Fatal("esperava ValidationError")
	}
	if ve.Field != "token" {
		t.Errorf("Field = %q, quer 'token' (primeiro campo validado)", ve.Field)
	}
}

func TestValidationError_Error(t *testing.T) {
	ve := &ValidationError{Field: "email", Message: "Email é obrigatório"}

	if ve.Error() != "Email é obrigatório" {
		t.Errorf("Error() = %q", ve.Error())
	}
}

func TestIsValidationError(t *testing.T) {
	casos := []struct {
		nome string
		err  error
		quer bool
	}{
		{"ValidationError", &ValidationError{Field: "x", Message: "y"}, true},
		{"erro comum", errors.New("falha qualquer"), false},
		{"nil", nil, false},
	}

	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			if got := IsValidationError(c.err); got != c.quer {
				t.Errorf("IsValidationError(%v) = %v, quer %v", c.err, got, c.quer)
			}
		})
	}
}

func TestQuickRevokeFromPayload(t *testing.T) {
	t.Run("extrai a mensagem", func(t *testing.T) {
		view := QuickRevokeFromPayload(map[string]any{
			"message": "Sessão revogada",
			"outro":   123,
		})

		if view.Message != "Sessão revogada" {
			t.Errorf("Message = %q", view.Message)
		}
		if view.Payload == nil {
			t.Error("Payload não foi preservado")
		}
	})

	// Payload de serviço externo pode vir sem a chave, com tipo errado ou
	// nulo — nenhum desses casos pode derrubar o handler.
	t.Run("entradas fora do esperado não quebram", func(t *testing.T) {
		casos := map[string]map[string]any{
			"sem a chave message":  {"outro": "x"},
			"message não é string": {"message": 42},
			"message nula":         {"message": nil},
			"payload vazio":        {},
		}

		for nome, payload := range casos {
			t.Run(nome, func(t *testing.T) {
				view := QuickRevokeFromPayload(payload)
				if view.Message != "" {
					t.Errorf("Message = %q, esperava vazia", view.Message)
				}
			})
		}
	})

	t.Run("payload nil", func(t *testing.T) {
		view := QuickRevokeFromPayload(nil)
		if view.Message != "" {
			t.Errorf("Message = %q, esperava vazia", view.Message)
		}
		if view.Payload != nil {
			t.Error("Payload deveria seguir nil")
		}
	})
}
