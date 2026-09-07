package blacklist

import (
	"context"
	"net/http"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/port"
	"github.com/keepguard/bff-auth/internal/pkg"
)

type BlacklistPort interface {
	ListMe(ctx context.Context, query appdto.ListDeviceBlacklistQuery) ([]appdto.DeviceBlacklistDTO, error)
	AddMe(ctx context.Context, cmd appdto.AddDeviceBlacklistCommand) error
	RemoveMe(ctx context.Context, cmd appdto.RemoveDeviceBlacklistCommand) error
	SearchAdmin(ctx context.Context, query appdto.SearchAdminDeviceBlacklistQuery) (appdto.PaginatedDeviceBlacklistResponseDTO, error)
	AdminAdd(ctx context.Context, cmd appdto.AdminAddDeviceBlacklistCommand) error
	AdminRemove(ctx context.Context, cmd appdto.AdminRemoveDeviceBlacklistCommand) error
	ListTenantUser(ctx context.Context, query appdto.ListTenantUserBlacklistQuery) ([]appdto.AdminDeviceBlacklistEntryDTO, error)
}

type service struct {
	auth port.AuthClient
}

func NewBlacklistPort(authClient port.AuthClient) BlacklistPort {
	return &service{auth: authClient}
}

func (s *service) unavailable() error {
	return pkg.NewAppError("SERVICE_UNAVAILABLE", "Blacklist indisponível", http.StatusServiceUnavailable)
}

func (s *service) ListMe(ctx context.Context, query appdto.ListDeviceBlacklistQuery) ([]appdto.DeviceBlacklistDTO, error) {
	if s == nil || s.auth == nil {
		return nil, s.unavailable()
	}
	return s.auth.ListDeviceBlacklist(ctx, query.Token, query.TenantID, query.CorrelationID)
}

func (s *service) AddMe(ctx context.Context, cmd appdto.AddDeviceBlacklistCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.AddDeviceToBlacklist(ctx, appdto.AddDeviceBlacklistRequestDTO{
		DeviceID:   cmd.DeviceID,
		DeviceName: cmd.DeviceName,
		Reason:     cmd.Reason,
	}, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) RemoveMe(ctx context.Context, cmd appdto.RemoveDeviceBlacklistCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.RemoveDeviceFromBlacklist(ctx, cmd.DeviceID, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) SearchAdmin(ctx context.Context, query appdto.SearchAdminDeviceBlacklistQuery) (appdto.PaginatedDeviceBlacklistResponseDTO, error) {
	if s == nil || s.auth == nil {
		return appdto.PaginatedDeviceBlacklistResponseDTO{}, s.unavailable()
	}
	return s.auth.SearchAdminDeviceBlacklist(ctx, query.QueryParams, query.Token, query.TenantID, query.CorrelationID)
}

func (s *service) AdminAdd(ctx context.Context, cmd appdto.AdminAddDeviceBlacklistCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.AdminAddDeviceToBlacklist(ctx, appdto.AdminAddDeviceBlacklistRequestDTO{
		UserID:     cmd.UserID,
		DeviceID:   cmd.DeviceID,
		DeviceName: cmd.DeviceName,
		Reason:     cmd.Reason,
		ExpiresAt:  cmd.ExpiresAt,
	}, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) AdminRemove(ctx context.Context, cmd appdto.AdminRemoveDeviceBlacklistCommand) error {
	if s == nil || s.auth == nil {
		return s.unavailable()
	}
	return s.auth.AdminRemoveDeviceFromBlacklist(ctx, cmd.DeviceID, cmd.UserID, cmd.Token, cmd.TenantID, cmd.CorrelationID)
}

func (s *service) ListTenantUser(ctx context.Context, query appdto.ListTenantUserBlacklistQuery) ([]appdto.AdminDeviceBlacklistEntryDTO, error) {
	if s == nil || s.auth == nil {
		return nil, s.unavailable()
	}
	return s.auth.ListTenantUserBlacklist(ctx, query.UserID, query.Token, query.TenantID, query.CorrelationID)
}
