package message

import (
	"context"
	"fmt"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/domain/enums"
	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

type SendResetPasswordMessageUseCase interface {
	Execute(ctx context.Context, command appdto.SendResetPasswordMessageCommand) (appdto.SendResetPasswordMessageViewDTO, error)
}

type sendResetPasswordMessageUseCaseImpl struct {
	authClient       authclient.AuthClient
	userClient       authclient.UserClient
	companyClient    authclient.CompanyClient
	messagePublisher messaging.MessagePublisher
	logger           *zap.Logger
}

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

func (uc *sendResetPasswordMessageUseCaseImpl) Execute(ctx context.Context, command appdto.SendResetPasswordMessageCommand) (appdto.SendResetPasswordMessageViewDTO, error) {
	company, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.SendResetPasswordMessageViewDTO{}, err
	}

	user, err := uc.userClient.GetByEmail(ctx, command.Email, command.TenantId, company.ID, command.CorrelationID)
	if err != nil {
		return appdto.SendResetPasswordMessageViewDTO{}, err
	}

	if user.Status != "ACTIVE" {
		return appdto.SendResetPasswordMessageViewDTO{}, &pkg.AppError{
			StatusCode: 400,
			Code:       "USER_NOT_ACTIVE",
			Message:    fmt.Sprintf("Usuário não está ativo. Status atual: %s", user.Status),
		}
	}

	generateTokenReq := map[string]interface{}{
		"codeUser":          user.CodeUser,
		"messageType":       enums.MessageTypeEmail.String(),
		"communicationType": enums.CommunicationTypeEmail.String(),
		"templateType":      enums.TemplateTypeRecuperacaoSenha.String(),
	}

	tokenResponse, err := uc.authClient.GenerateResetToken(ctx, generateTokenReq, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.SendResetPasswordMessageViewDTO{}, err
	}

	variables := map[string]interface{}{
		"userName": user.Username,
		"token":    tokenResponse.Token,
	}

	messageReq := messaging.MessageDTO{
		TenantId:          command.TenantId,
		CorrelationID:     command.CorrelationID,
		XCorrelationID:    command.CorrelationID,
		MessageType:       enums.MessageTypeEmail.String(),
		CommunicationType: enums.CommunicationTypeEmail.String(),
		TemplateType:      enums.TemplateTypeRecuperacaoSenha.String(),
		Recipient:         command.Email,
		CodeUser:          user.CodeUser,
		Variables:         variables,
	}

	err = uc.messagePublisher.PublishMessage(ctx, messageReq)
	if err != nil {
		uc.logger.Error("Erro ao publicar mensagem de reset de senha",
			zap.String("email", command.Email),
			zap.String("correlation_id", command.CorrelationID),
			zap.Error(err))
	} else {
		uc.logger.Info("Mensagem de reset de senha publicada com sucesso",
			zap.String("email", command.Email),
			zap.String("correlation_id", command.CorrelationID))
	}

	return appdto.SendResetPasswordMessageViewDTO{
		Success: true,
		Message: "Token de reset de senha gerado e mensagem enviada com sucesso",
	}, nil
}
