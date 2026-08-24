package auth

import (
	"fmt"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/domain/enums"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

// resetPasswordUseCaseImpl implementa o caso de uso de reset de senha
type resetPasswordUseCaseImpl struct {
	authClient       authclient.AuthClient
	userClient       authclient.UserClient
	companyClient    authclient.CompanyClient
	messagePublisher messaging.MessagePublisher
	logger           *zap.Logger
}

// NewResetPasswordUseCase cria um novo caso de uso de reset de senha
func NewResetPasswordUseCase(
	authClient authclient.AuthClient,
	userClient authclient.UserClient,
	companyClient authclient.CompanyClient,
	messagePublisher messaging.MessagePublisher,
	logger *zap.Logger,
) ResetPasswordUseCase {
	return &resetPasswordUseCaseImpl{
		authClient:       authClient,
		userClient:       userClient,
		companyClient:    companyClient,
		messagePublisher: messagePublisher,
		logger:           logger,
	}
}

// Execute executa o caso de uso de reset de senha
func (uc *resetPasswordUseCaseImpl) Execute(command appdto.ResetPasswordCommand) error {
	// Passo 1: Verificar se a empresa existe consultando o Company Service
	company, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	// Passo 2: Buscar usuário por email no User Service
	user, err := uc.userClient.GetByEmail(command.Context, command.Email, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	// Verificar se o usuário está ACTIVE
	if user.Status != "ACTIVE" {
		return &pkg.AppError{
			StatusCode: 400,
			Code:       "USER_NOT_ACTIVE",
			Message:    fmt.Sprintf("Usuário não está ativo. Status atual: %s", user.Status),
		}
	}

	// Passo 3: Criar DTO de requisição para o cliente (para enviar ao ms-auth)
	req := outboundDto.ResetPasswordMSRequestDTO{
		CodeUser:           user.CodeUser,
		ResetToken:         command.ResetToken,
		NewPassword:        command.NewPassword,
		ConfirmNewPassword: command.ConfirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	// Chamar o cliente de autenticação para resetar a senha
	err = uc.authClient.ResetPassword(command.Context, req, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	// Passo 4: Notificar o usuário por email informando que a senha foi alterada com sucesso
	if uc.messagePublisher != nil {
		appName := "KeepGuard"
		if company.Name != "" {
			appName = company.Name
		}

		variables := map[string]interface{}{
			"userName":  user.Username,
			"appName":   appName,
			"updatedAt": time.Now().Format("02/01/2006 15:04:05"),
		}

		messageReq := messaging.MessageDTO{
			TenantId:          command.TenantId,
			XCorrelationID:    command.CorrelationID,
			MessageType:       enums.MessageTypeEmail.String(),
			CommunicationType: enums.CommunicationTypeEmail.String(),
			TemplateType:      enums.TemplateTypeSenhaAlteradaSucesso.String(),
			Recipient:         command.Email,
			CodeUser:          user.CodeUser,
			Variables:         variables,
		}

		pubErr := uc.messagePublisher.PublishMessage(command.Context, messageReq)
		if pubErr != nil {
			uc.logger.Error("Erro ao publicar notificação de senha resetada com sucesso",
				zap.String("email", command.Email),
				zap.String("code_user", user.CodeUser),
				zap.Error(pubErr))
		} else {
			uc.logger.Info("Notificação de senha resetada publicada com sucesso",
				zap.String("email", command.Email),
				zap.String("code_user", user.CodeUser))
		}
	}

	return nil
}
