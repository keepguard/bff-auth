package auth

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

type changePasswordUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
	logger        *zap.Logger
}

func NewChangePasswordUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient, logger *zap.Logger) ChangePasswordUseCase {
	return &changePasswordUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
		logger:        logger,
	}
}

func (uc *changePasswordUseCaseImpl) Execute(ctx context.Context, command appdto.ChangePasswordCommand) error {
	_, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	codeUser, err := pkg.ExtractCodeUserFromToken(command.Token)
	if err != nil {
		return err
	}

	req := appdto.ChangePasswordMSRequestDTO{
		CodeUser:           codeUser,
		CurrentPassword:    command.CurrentPassword,
		NewPassword:        command.NewPassword,
		ConfirmNewPassword: command.ConfirmNewPassword,
	}

	return uc.authClient.ChangePassword(
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
