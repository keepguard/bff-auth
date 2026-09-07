package auth

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"go.uber.org/zap"
)

type loginUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

func NewLoginUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) LoginUseCase {
	return &loginUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

func (uc *loginUseCaseImpl) Execute(ctx context.Context, command appdto.LoginCommand) (appdto.AuthResponseDTO, error) {
	company, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.AuthResponseDTO{}, err
	}

	ctx = authclient.WithCompanyID(ctx, company.ID)

	req := appdto.AuthRequestDTO{
		Username:  command.Username,
		Password:  command.Password,
		CompanyID: company.ID,
	}

	response, err := uc.authClient.Login(
		ctx,
		req,
		command.TenantId,
		command.CorrelationID,
		command.ClientId,
		command.DeviceId,
		command.DeviceName,
		command.DeviceType,
		command.IPAddress,
		command.UserAgent,
	)
	if err != nil {
		return appdto.AuthResponseDTO{}, err
	}

	return response, nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type ServiceUnavailableError struct {
	Service string
	Message string
	Details map[string]interface{}
}

func (e *ServiceUnavailableError) Error() string {
	return e.Message
}
