package session

import (
	"context"
	"net/http"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
)

type SessionPort interface {
	ListMe(ctx context.Context, query appdto.ListUserSessionsQuery) ([]appdto.DeviceSessionDTO, error)
	Revoke(ctx context.Context, cmd appdto.RevokeSessionCommand) error
	RevokeOthers(ctx context.Context, cmd appdto.RevokeAllOtherSessionsCommand) error
	ListTenantUser(ctx context.Context, query appdto.ListTenantUserSessionsQuery) ([]appdto.DeviceSessionDTO, error)
	RevokeTenantUser(ctx context.Context, cmd appdto.RevokeTenantUserSessionCommand) error
	SearchTenant(ctx context.Context, query appdto.SearchTenantSessionsQuery) (appdto.PaginatedDeviceSessionResponseDTO, error)
}

type service struct {
	auth port.AuthClient
}

func NewSessionPort(authClient port.AuthClient) SessionPort {
	return &service{auth: authClient}
}

func (s *service) unavailable() error {
	return pkg.NewAppError("SERVICE_UNAVAILABLE", "Sessões indisponíveis", http.StatusServiceUnavailable)
}

func (s *service) ListMe(ctx context.Context, query appdto.ListUserSessionsQuery) ([]appdto.DeviceSessionDTO, error) {
	if s == nil || s.auth == nil {
		return nil, s.unavailable()
	}
	return s.auth.ListUserSessions(ctx, query.Token, query.DeviceID, query.TenantID, query.CorrelationID)
}

func (s *service) Revoke(ctx context.Context, cmd appdto.RevokeSessionCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.RevokeSession(ctx, cmd.DeviceID, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) RevokeOthers(ctx context.Context, cmd appdto.RevokeAllOtherSessionsCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.RevokeAllOtherSessions(ctx, cmd.Token, cmd.CurrentDeviceID, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) ListTenantUser(ctx context.Context, query appdto.ListTenantUserSessionsQuery) ([]appdto.DeviceSessionDTO, error) {
	if s == nil || s.auth == nil {
		return nil, s.unavailable()
	}
	return s.auth.ListTenantUserSessions(ctx, query.UserID, query.Token, query.TenantID, query.CorrelationID)
}

func (s *service) RevokeTenantUser(ctx context.Context, cmd appdto.RevokeTenantUserSessionCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.RevokeTenantUserSession(ctx, cmd.UserID, cmd.DeviceID, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) SearchTenant(ctx context.Context, query appdto.SearchTenantSessionsQuery) (appdto.PaginatedDeviceSessionResponseDTO, error) {
	if s == nil || s.auth == nil {
		return appdto.PaginatedDeviceSessionResponseDTO{}, s.unavailable()
	}
	return s.auth.SearchTenantSessions(ctx, query.QueryParams, query.Token, query.TenantID, query.CorrelationID)
}
