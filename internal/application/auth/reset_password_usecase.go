package auth

import (
	"fmt"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"go.uber.org/zap"
)

// resetPasswordUseCaseImpl implementa o caso de uso de reset de senha
type resetPasswordUseCaseImpl struct {
	authClient    authclient.AuthClient
	userClient    authclient.UserClient
	companyClient authclient.CompanyClient
	logger        *zap.Logger
}

// NewResetPasswordUseCase cria um novo caso de uso de reset de senha
func NewResetPasswordUseCase(
	authClient authclient.AuthClient,
	userClient authclient.UserClient,
	companyClient authclient.CompanyClient,
	logger *zap.Logger,
) ResetPasswordUseCase {
	return &resetPasswordUseCaseImpl{
		authClient:    authClient,
		userClient:    userClient,
		companyClient: companyClient,
		logger:        logger,
	}
}

// Execute executa o caso de uso de reset de senha
func (uc *resetPasswordUseCaseImpl) Execute(command appdto.ResetPasswordCommand) error {
	// Passo 1: Verificar se a empresa existe consultando o Company Service
	_, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
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

	return nil
}
