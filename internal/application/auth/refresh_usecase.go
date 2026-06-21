package auth

import (
	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// refreshUseCaseImpl implementa o caso de uso de refresh
type refreshUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

// NewRefreshUseCase cria um novo caso de uso de refresh
func NewRefreshUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) RefreshUseCase {
	return &refreshUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

// Execute executa o caso de uso de refresh
func (uc *refreshUseCaseImpl) Execute(command appdto.RefreshTokenCommand) (dto.RefreshTokenResponseDTO, error) {

	// Primeiro, verifica se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByXApplication(command.Context, command.XApplication, command.CorrelationID)
	if err != nil {
		return dto.RefreshTokenResponseDTO{}, err
	}

	// Criar DTO de requisição para o cliente
	req := dto.RefreshTokenRequestDTO{
		Token: command.RefreshToken,
	}

	// Chama o cliente de autenticação
	response, err := uc.authClient.RefreshToken(command.Context, req, command.XApplication, command.CorrelationID)
	if err != nil {
		return dto.RefreshTokenResponseDTO{}, err
	}

	return response, nil
}
