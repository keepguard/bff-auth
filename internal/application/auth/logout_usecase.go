package auth

import (
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// logoutUseCaseImpl implementa o caso de uso de logout
type logoutUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

// NewLogoutUseCase cria um novo caso de uso de logout
func NewLogoutUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) LogoutUseCase {
	return &logoutUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

// Execute executa o caso de uso de logout
func (uc *logoutUseCaseImpl) Execute(command appdto.LogoutCommand) error {
	// Primeiro, verifica se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	// Chama o cliente de autenticação
	err = uc.authClient.Logout(command.Context, command.Token, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	return nil
}
