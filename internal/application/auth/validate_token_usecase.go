package auth

import (
	"context"
	"net/http"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/application/port"
	"go.uber.org/zap"
)

type validateTokenUseCaseImpl struct {
	authClient    authclient.AuthClient
	companyClient authclient.CompanyClient
	tokenChecker  authclient.LoginTokenChecker
	logger        *zap.Logger
}

func NewValidateTokenUseCase(
	authClient authclient.AuthClient,
	companyClient authclient.CompanyClient,
	tokenChecker authclient.LoginTokenChecker,
	logger *zap.Logger,
) ValidateTokenUseCase {
	return &validateTokenUseCaseImpl{
		authClient:    authClient,
		companyClient: companyClient,
		tokenChecker:  tokenChecker,
		logger:        logger,
	}
}

func (uc *validateTokenUseCaseImpl) Execute(ctx context.Context, command appdto.ValidateTokenCommand) error {
	_, err := uc.companyClient.GetByTenantId(ctx, command.TenantId, command.CorrelationID)
	if err != nil {
		return err
	}

	if uc.tokenChecker != nil {
		presence, codeUser, checkErr := uc.tokenChecker.Check(ctx, command.Token)
		if checkErr != nil {
			uc.logger.Warn("Falha ao consultar tokenlogin no Redis; validando no ms-auth",
				zap.String("correlationId", command.CorrelationID),
				zap.String("codeUser", codeUser),
				zap.Error(checkErr),
			)
		} else if presence == authclient.LoginTokenAbsent {
			uc.logger.Warn("Token ausente no Redis",
				zap.String("correlationId", command.CorrelationID),
				zap.String("codeUser", codeUser),
			)
			return &appdto.HTTPError{
				StatusCode: http.StatusUnauthorized,
				Message:    "Sessão revogada ou expirada. Por favor, realize login novamente.",
				ErrorCode:  "TOKEN_REVOKED",
			}
		}
	}

	return uc.authClient.ValidateToken(ctx, command.Token, command.TenantId, command.CorrelationID)
}
