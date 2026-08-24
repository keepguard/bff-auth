package mapper

import (
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/enums"
	"github.com/stretchr/testify/assert"
)

// TestNewMessageMapper testa a criação do mapper
func TestNewMessageMapper(t *testing.T) {
	// Act
	mapper := NewMessageMapper()

	// Assert
	assert.NotNil(t, mapper)
}

// TestMessageMapper_ToSendMessageRequest testa o mapeamento para SendMessageRequest
func TestMessageMapper_ToSendMessageRequest(t *testing.T) {
	// Arrange
	mapper := NewMessageMapper()
	messageType := enums.MessageTypeEmail
	communicationType := enums.CommunicationTypeEmail
	templateType := enums.TemplateTypeRecuperacaoSenha
	recipient := "test@example.com"
	codeUser := "123e4567-e89b-12d3-a456-426614174000"
	variables := map[string]string{
		"userName": "Test User",
		"token":    "123456",
	}

	// Act
	result := mapper.ToSendMessageRequest(
		messageType,
		communicationType,
		templateType,
		recipient,
		variables,
		codeUser,
	)

	// Assert
	assert.Equal(t, "EMAIL", result.MessageType)
	assert.Equal(t, "EMAIL", result.CommunicationType)
	assert.Equal(t, "RECUPERACAO_SENHA", result.TemplateType)
	assert.Equal(t, recipient, result.Recipient)
	assert.Equal(t, codeUser, result.CodeUser)
	expectedMap := make(map[string]interface{})
	for k, v := range variables {
		expectedMap[k] = v
	}
	assert.Equal(t, expectedMap, result.Variables)
}

// TestMessageMapper_ToSendMessageRequest_DifferentTypes testa com diferentes tipos de enums
func TestMessageMapper_ToSendMessageRequest_DifferentTypes(t *testing.T) {
	tests := []struct {
		name              string
		messageType       enums.MessageType
		communicationType enums.CommunicationType
		templateType      enums.TemplateType
		expectedMsg       string
		expectedComm      string
		expectedTemplate  string
	}{
		{
			name:              "Email tipos",
			messageType:       enums.MessageTypeEmail,
			communicationType: enums.CommunicationTypeEmail,
			templateType:      enums.TemplateTypeRecuperacaoSenha,
			expectedMsg:       "EMAIL",
			expectedComm:      "EMAIL",
			expectedTemplate:  "RECUPERACAO_SENHA",
		},
		{
			name:              "SMS tipos",
			messageType:       enums.MessageTypeSMS,
			communicationType: enums.CommunicationTypeSMS,
			templateType:      enums.TemplateTypeAutenticacaoSMSToken,
			expectedMsg:       "SMS",
			expectedComm:      "SMS",
			expectedTemplate:  "AUTENTICACAO_SMS_TOKEN",
		},
		{
			name:              "WhatsApp tipos",
			messageType:       enums.MessageTypeWhatsApp,
			communicationType: enums.CommunicationTypeWhatsApp,
			templateType:      enums.TemplateTypeAutenticacaoWhatsAppToken,
			expectedMsg:       "WHATSAPP",
			expectedComm:      "WHATSAPP",
			expectedTemplate:  "AUTENTICACAO_WHATSAPP_TOKEN",
		},
		{
			name:              "Push tipos",
			messageType:       enums.MessageTypePush,
			communicationType: enums.CommunicationTypePush,
			templateType:      enums.TemplateTypeNotificacaoGeral,
			expectedMsg:       "PUSH",
			expectedComm:      "PUSH",
			expectedTemplate:  "NOTIFICACAO_GERAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mapper := NewMessageMapper()
			recipient := "test@example.com"
			variables := map[string]string{"key": "value"}

			// Act
			result := mapper.ToSendMessageRequest(
				tt.messageType,
				tt.communicationType,
				tt.templateType,
				recipient,
				variables,
				"test-code-user",
			)

			// Assert
			assert.Equal(t, tt.expectedMsg, result.MessageType)
			assert.Equal(t, tt.expectedComm, result.CommunicationType)
			assert.Equal(t, tt.expectedTemplate, result.TemplateType)
			assert.Equal(t, recipient, result.Recipient)
			expectedMap := make(map[string]interface{})
			for k, v := range variables {
				expectedMap[k] = v
			}
			assert.Equal(t, expectedMap, result.Variables)
		})
	}
}

// TestMessageMapper_ToSendMessageRequest_EmptyVariables testa com variáveis vazias
func TestMessageMapper_ToSendMessageRequest_EmptyVariables(t *testing.T) {
	// Arrange
	mapper := NewMessageMapper()
	messageType := enums.MessageTypeEmail
	communicationType := enums.CommunicationTypeEmail
	templateType := enums.TemplateTypeCadastroSucesso
	recipient := "test@example.com"
	variables := map[string]string{}

	// Act
	result := mapper.ToSendMessageRequest(
		messageType,
		communicationType,
		templateType,
		recipient,
		variables,
		"test-code-user",
	)

	// Assert
	assert.Equal(t, "EMAIL", result.MessageType)
	assert.Equal(t, "EMAIL", result.CommunicationType)
	assert.Equal(t, "CADASTRO_SUCESSO", result.TemplateType)
	assert.Equal(t, recipient, result.Recipient)
	assert.Equal(t, "test-code-user", result.CodeUser)
	assert.NotNil(t, result.Variables)
	assert.Empty(t, result.Variables)
}

// TestMessageMapper_ToSendMessageRequest_MultipleVariables testa com múltiplas variáveis
func TestMessageMapper_ToSendMessageRequest_MultipleVariables(t *testing.T) {
	// Arrange
	mapper := NewMessageMapper()
	messageType := enums.MessageTypeEmail
	communicationType := enums.CommunicationTypeEmail
	templateType := enums.TemplateTypeAlertaSeguranca
	recipient := "security@example.com"
	codeUser := "security-user-123"
	variables := map[string]string{
		"userName":  "John Doe",
		"ipAddress": "192.168.1.1",
		"location":  "São Paulo",
		"timestamp": "2023-01-01T00:00:00Z",
	}

	// Act
	result := mapper.ToSendMessageRequest(
		messageType,
		communicationType,
		templateType,
		recipient,
		variables,
		codeUser,
	)

	// Assert
	assert.Equal(t, "EMAIL", result.MessageType)
	assert.Equal(t, "EMAIL", result.CommunicationType)
	assert.Equal(t, "ALERTA_SEGURANCA", result.TemplateType)
	assert.Equal(t, recipient, result.Recipient)
	assert.Equal(t, codeUser, result.CodeUser)
	assert.Len(t, result.Variables, 4)
	assert.Equal(t, "John Doe", result.Variables["userName"])
	assert.Equal(t, "192.168.1.1", result.Variables["ipAddress"])
	assert.Equal(t, "São Paulo", result.Variables["location"])
	assert.Equal(t, "2023-01-01T00:00:00Z", result.Variables["timestamp"])
}
