package mapper

import (
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	"github.com/keepguard/bff-auth/internal/domain/enums"
)

// MessageMapper é responsável por mapear dados de mensagens entre camadas
type MessageMapper struct{}

// NewMessageMapper cria uma nova instância do MessageMapper
func NewMessageMapper() *MessageMapper {
	return &MessageMapper{}
}

// ToSendMessageRequest mapeia os parâmetros de domínio para o DTO de requisição do ms-communication
func (m *MessageMapper) ToSendMessageRequest(
	messageType enums.MessageType,
	communicationType enums.CommunicationType,
	templateType enums.TemplateType,
	recipient string,
	variables map[string]string,
	codeUser string,
) outboundDto.SendMessageRequestDTO {
	// Converte map[string]string para map[string]interface{}
	interfaceVariables := make(map[string]interface{})
	for k, v := range variables {
		interfaceVariables[k] = v
	}

	return outboundDto.SendMessageRequestDTO{
		MessageType:       messageType.String(),
		CommunicationType: communicationType.String(),
		TemplateType:      templateType.String(),
		Recipient:         recipient,
		CodeUser:          codeUser,
		Variables:         interfaceVariables,
	}
}

// ToGenerateResetTokenRequest mapeia os parâmetros de domínio para o DTO de requisição do ms-auth
func (m *MessageMapper) ToGenerateResetTokenRequest(
	codeUser string,
	messageType enums.MessageType,
	communicationType enums.CommunicationType,
	templateType enums.TemplateType,
) outboundDto.GenerateResetTokenMSRequestDTO {
	return outboundDto.GenerateResetTokenMSRequestDTO{
		CodeUser:          codeUser,
		MessageType:       messageType.String(),
		CommunicationType: communicationType.String(),
		TemplateType:      templateType.String(),
	}
}
