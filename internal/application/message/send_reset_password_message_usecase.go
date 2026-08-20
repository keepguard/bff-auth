package message

import (
	"fmt"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/domain/enums"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

// SendResetPasswordMessageUseCase define a interface do caso de uso
type SendResetPasswordMessageUseCase interface {
	Execute(command appdto.SendResetPasswordMessageCommand) (inboundDto.SendResetPasswordMessageResponseDTO, error)
}

// sendResetPasswordMessageUseCaseImpl implementa o caso de uso de envio de mensagem de reset de senha
type sendResetPasswordMessageUseCaseImpl struct {
	authClient       authclient.AuthClient
	userClient       authclient.UserClient
	companyClient    authclient.CompanyClient
	messagePublisher messaging.MessagePublisher
	logger           *zap.Logger
}

// NewSendResetPasswordMessageUseCase cria um novo caso de uso de envio de mensagem de reset
func NewSendResetPasswordMessageUseCase(
	authClient authclient.AuthClient,
	userClient authclient.UserClient,
	companyClient authclient.CompanyClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
) SendResetPasswordMessageUseCase {
	return &sendResetPasswordMessageUseCaseImpl{
		authClient:       authClient,
		userClient:       userClient,
		companyClient:    companyClient,
		messagePublisher: messagePublisher,
		logger:           logger,
	}
}

// Execute executa o caso de uso de envio de mensagem de reset de senha
func (uc *sendResetPasswordMessageUseCaseImpl) Execute(command appdto.SendResetPasswordMessageCommand) (inboundDto.SendResetPasswordMessageResponseDTO, error) {
	// Passo 1: Verificar se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return inboundDto.SendResetPasswordMessageResponseDTO{}, err
	}

	// Passo 2: Buscar usuário por email no User Service
	user, err := uc.userClient.GetByEmail(command.Context, command.Email, command.TenantId, command.CorrelationID)
	if err != nil {
		return inboundDto.SendResetPasswordMessageResponseDTO{}, err
	}

	// Passo 3: Verificar se o usuário está ACTIVE
	if user.Status != "ACTIVE" {
		return inboundDto.SendResetPasswordMessageResponseDTO{}, &pkg.AppError{
			StatusCode: 400,
			Code:       "USER_NOT_ACTIVE",
			Message:    fmt.Sprintf("Usuário não está ativo. Status atual: %s", user.Status),
		}
	}

	// Passo 4: Gerar token de reset através do Auth Service
	generateTokenReq := map[string]interface{}{
		"codeUser":          user.CodeUser,
		"messageType":       enums.MessageTypeEmail.String(),
		"communicationType": enums.CommunicationTypeEmail.String(),
		"templateType":      enums.TemplateTypeRecuperacaoSenha.String(),
	}

	tokenResponse, err := uc.authClient.GenerateResetToken(command.Context, generateTokenReq, command.TenantId, command.CorrelationID)
	if err != nil {
		return inboundDto.SendResetPasswordMessageResponseDTO{}, err
	}

	// Passo 5: Preparar mensagem para publicação na fila
	variables := map[string]interface{}{
		"userName": user.Username,
		"token":    tokenResponse.Token,
	}

	messageReq := messaging.MessageDTO{
		TenantId:      command.TenantId,
		XCorrelationID:    command.CorrelationID,
		MessageType:       enums.MessageTypeEmail.String(),
		CommunicationType: enums.CommunicationTypeEmail.String(),
		TemplateType:      enums.TemplateTypeRecuperacaoSenha.String(),
		Recipient:         command.Email,
		CodeUser:          user.CodeUser,
		Variables:         variables,
	}

	// Passo 6: Publicar mensagem na fila RabbitMQ (com fallback HTTP automático)
	err = uc.messagePublisher.PublishMessage(command.Context, messageReq)
	if err != nil {
		// Não falha se o email não for enviado (melhor UX)
		uc.logger.Error("Erro ao publicar mensagem de reset de senha",
			zap.String("email", command.Email),
			zap.String("correlation_id", command.CorrelationID),
			zap.Error(err))
	} else {
		uc.logger.Info("Mensagem de reset de senha publicada com sucesso",
			zap.String("email", command.Email),
			zap.String("correlation_id", command.CorrelationID))
	}

	return inboundDto.SendResetPasswordMessageResponseDTO{
		Success: true,
		Message: "Token de reset de senha gerado e mensagem enviada com sucesso",
	}, nil
}
