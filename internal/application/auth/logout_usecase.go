package auth

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
)

type logoutUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

func NewLogoutUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) LogoutUseCase {
	return &logoutUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

func (uc *logoutUseCaseImpl) Execute(ctx context.Context, command appdto.LogoutCommand) error {
	_, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	return uc.authClient.Logout(ctx, command.Token, command.TenantId, command.CorrelationID)
}
