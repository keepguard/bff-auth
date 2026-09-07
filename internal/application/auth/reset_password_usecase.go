package auth

import (
	"context"
	"fmt"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/domain/ports/messaging"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

type resetPasswordUseCaseImpl struct {
	authClient       authclient.AuthClient
	userClient       authclient.UserClient
	companyClient    authclient.CompanyClient
	messagePublisher messaging.MessagePublisher
	logger           *zap.Logger
}

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

func (uc *resetPasswordUseCaseImpl) Execute(ctx context.Context, command appdto.ResetPasswordCommand) error {
	company, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	user, err := uc.userClient.GetByEmail(ctx, command.Email, command.TenantId, company.ID, command.CorrelationID)
	if err != nil {
		return err
	}

	if user.Status != "ACTIVE" {
		return &pkg.AppError{
			StatusCode: 400,
			Code:       "USER_NOT_ACTIVE",
			Message:    fmt.Sprintf("Usuário não está ativo. Status atual: %s", user.Status),
		}
	}

	req := appdto.ResetPasswordMSRequestDTO{
		CodeUser:           user.CodeUser,
		ResetToken:         command.ResetToken,
		NewPassword:        command.NewPassword,
		ConfirmNewPassword: command.ConfirmNewPassword,
		MessageType:        "EMAIL",
		TemplateType:       "RECUPERACAO_SENHA",
	}

	return uc.authClient.ResetPassword(
		ctx,
		req,
		command.TenantId,
		command.CorrelationID,
		command.DeviceId,
		command.DeviceName,
		command.DeviceType,
		command.IpAddress,
		command.UserAgent,
	)
}
