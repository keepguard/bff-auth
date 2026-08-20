package auth

import (
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"go.uber.org/zap"
)

// validateTokenUseCaseImpl implementa o caso de uso de validação de token
type validateTokenUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
	logger        *zap.Logger
}

// NewValidateTokenUseCase cria um novo caso de uso de validação de token
func NewValidateTokenUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient, logger *zap.Logger) ValidateTokenUseCase {
	return &validateTokenUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
		logger:        logger,
	}
}

// Execute executa o caso de uso de validação de token
func (uc *validateTokenUseCaseImpl) Execute(command appdto.ValidateTokenCommand) error {
	// Primeiro, verifica se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	// Chama o cliente de autenticação para validar o token
	err = uc.authClient.ValidateToken(command.Context, command.Token, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	return nil
}
