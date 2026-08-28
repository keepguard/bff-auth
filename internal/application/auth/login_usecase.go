package auth

import (
	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"go.uber.org/zap"
)

// loginUseCaseImpl implementa o caso de uso de login
type loginUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
}

// Logger interface para logging
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
}

// NewLoginUseCase cria um novo caso de uso de login
func NewLoginUseCase(authClient authclient.AuthClient, companyClient authclient.CompanyClient) LoginUseCase {
	return &loginUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
	}
}

// Execute executa o caso de uso de login
func (uc *loginUseCaseImpl) Execute(command appdto.LoginCommand) (dto.AuthResponseDTO, error) {
	// Primeiro, verifica se a empresa existe consultando o Company Service
	company, err := uc.companyClient.GetByTenantId(command.Context, command.TenantId, command.CorrelationID)
	if err != nil {
		return dto.AuthResponseDTO{}, err
	}

	command.Context = authclient.WithCompanyID(command.Context, company.ID)

	// Criar DTO de requisição para o cliente
	req := dto.AuthRequestDTO{
		Username:  command.Username,
		Password:  command.Password,
		CompanyID: company.ID,
	}

	// Chama o cliente de autenticação com metadados de dispositivo
	response, err := uc.authClient.Login(
		command.Context,
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
		return dto.AuthResponseDTO{}, err
	}

	return response, nil
}

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ServiceUnavailableError representa um erro de serviço indisponível
type ServiceUnavailableError struct {
	Service string
	Message string
	Details map[string]interface{}
}

func (e *ServiceUnavailableError) Error() string {
	return e.Message
}
