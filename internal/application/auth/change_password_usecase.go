package auth

import (
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

// changePasswordUseCaseImpl implementa o caso de uso de alteração de senha
type changePasswordUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
	logger        *zap.Logger
}

// NewChangePasswordUseCase cria um novo caso de uso de alteração de senha
func NewChangePasswordUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient, logger *zap.Logger) ChangePasswordUseCase {
	return &changePasswordUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
		logger:        logger,
	}
}

// Execute executa o caso de uso de alteração de senha
func (uc *changePasswordUseCaseImpl) Execute(command appdto.ChangePasswordCommand) error {
	// Primeiro, verifica se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByXApplication(command.Context, command.XApplication, command.CorrelationID)
	if err != nil {
		return err
	}

	// Extrai o codeUser do token
	codeUser, err := pkg.ExtractCodeUserFromToken(command.Token)
	if err != nil {
		return err
	}

	// Criar DTO de requisição para o cliente (para enviar ao ms-auth)
	req := outboundDto.ChangePasswordMSRequestDTO{
		CodeUser:           codeUser,
		CurrentPassword:    command.CurrentPassword,
		NewPassword:        command.NewPassword,
		ConfirmNewPassword: command.ConfirmNewPassword,
	}

	// Chama o cliente de autenticação para alterar a senha
	err = uc.authClient.ChangePassword(command.Context, req, command.XApplication, command.CorrelationID)
	if err != nil {
		return err
	}

	return nil
}
