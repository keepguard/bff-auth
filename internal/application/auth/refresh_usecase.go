package auth

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
)

type refreshUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

func NewRefreshUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) RefreshUseCase {
	return &refreshUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

func (uc *refreshUseCaseImpl) Execute(ctx context.Context, command appdto.RefreshTokenCommand) (appdto.RefreshTokenResponseDTO, error) {
	_, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return appdto.RefreshTokenResponseDTO{}, err
	}

	req := appdto.RefreshTokenRequestDTO{
		Token: command.RefreshToken,
	}

	response, err := uc.authClient.RefreshToken(ctx, req, command.TenantId, command.CorrelationID, command.ClientId)
	if err != nil {
		return appdto.RefreshTokenResponseDTO{}, err
	}

	return response, nil
}
