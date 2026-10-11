package lifecycle

import (
	"context"
	"net/http"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
)

type LifecyclePort interface {
	BlockMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error
	DeleteMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error
	// DeleteUser: o ms-auth confere o ator (mesma company, ADMIN/MANAGER com
	// user:delete; alvo ADMIN/SYSTEM protegido).
	DeleteUser(ctx context.Context, cmd appdto.AdminDeleteUserCommand) error
	// DeleteUserByCode: resolve o idUserExternal a partir do codeUser dentro da
	// company do tenant e exclui o usuário com as mesmas garantias de DeleteUser.
	DeleteUserByCode(ctx context.Context, cmd appdto.AdminDeleteUserByCodeCommand) error
}

type service struct {
	auth    port.AuthClient
	company port.CompanyClient
}

func NewLifecyclePort(authClient port.AuthClient, companyClient port.CompanyClient) LifecyclePort {
	return &service{auth: authClient, company: companyClient}
}

func (s *service) unavailable() error {
	return pkg.NewAppError("SERVICE_UNAVAILABLE", "Ciclo de vida da conta indisponível", http.StatusServiceUnavailable)
}

func (s *service) BlockMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error {
	idUserExternal, err := s.resolveSelfUserExternalID(ctx, cmd)
	if err != nil {
		return err
	}
	return s.auth.BlockUser(ctx, idUserExternal, cmd.Reason, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) DeleteMe(ctx context.Context, cmd appdto.AccountLifecycleCommand) error {
	idUserExternal, err := s.resolveSelfUserExternalID(ctx, cmd)
	if err != nil {
		return err
	}
	return s.auth.DeleteUser(ctx, idUserExternal, cmd.Reason, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) DeleteUser(ctx context.Context, cmd appdto.AdminDeleteUserCommand) error {
	if s == nil || s.auth == nil || s.company == nil {
		return s.unavailable()
	}
	if cmd.Token == "" {
		return pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}
	if cmd.IDUserExternal == "" {
		return pkg.NewAppError("VALIDATION_ERROR", "Usuário alvo não informado", http.StatusBadRequest)
	}
	company, err := s.company.GetByTenantId(ctx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return err
	}
	if company.ID == "" {
		return pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant informado", http.StatusNotFound)
	}
	return s.auth.DeleteUser(port.WithCompanyID(ctx, company.ID), cmd.IDUserExternal, cmd.Reason, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) DeleteUserByCode(ctx context.Context, cmd appdto.AdminDeleteUserByCodeCommand) error {
	if s == nil || s.auth == nil || s.company == nil {
		return s.unavailable()
	}
	if cmd.Token == "" {
		return pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}
	if cmd.CodeUser == "" {
		return pkg.NewAppError("VALIDATION_ERROR", "Código do usuário alvo não informado", http.StatusBadRequest)
	}
	company, err := s.company.GetByTenantId(ctx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return err
	}
	if company.ID == "" {
		return pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant informado", http.StatusNotFound)
	}
	user, err := s.auth.GetUserByCodeUser(port.WithCompanyID(ctx, company.ID), cmd.CodeUser, cmd.Token, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return err
	}
	idUserExternal := user.ExternalID()
	if idUserExternal == "" {
		return pkg.NewAppError("USER_NOT_FOUND", "Usuário alvo não encontrado", http.StatusNotFound)
	}
	return s.auth.DeleteUser(port.WithCompanyID(ctx, company.ID), idUserExternal, cmd.Reason, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) resolveSelfUserExternalID(ctx context.Context, cmd appdto.AccountLifecycleCommand) (string, error) {
	if s == nil || s.auth == nil || s.company == nil {
		return "", s.unavailable()
	}
	if cmd.Token == "" {
		return "", pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}
	if cmd.CodeUser == "" {
		return "", pkg.NewAppError("UNAUTHORIZED", "Token JWT sem identificador de usuário", http.StatusUnauthorized)
	}

	company, err := s.company.GetByTenantId(ctx, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return "", err
	}
	if company.ID == "" {
		return "", pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant informado", http.StatusNotFound)
	}

	user, err := s.auth.GetUserByCodeUser(port.WithCompanyID(ctx, company.ID), cmd.CodeUser, cmd.Token, cmd.TenantID, cmd.CorrelationID)
	if err != nil {
		return "", err
	}

	idUserExternal := user.ExternalID()
	if idUserExternal == "" {
		return "", pkg.NewAppError("USER_NOT_FOUND", "Usuário autenticado não encontrado", http.StatusNotFound)
	}
	return idUserExternal, nil
}
