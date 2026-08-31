package client

import (
	"context"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
)

// AuthClient define a interface para o cliente de autenticação
type AuthClient interface {
	Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error)
	RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error)
	Logout(ctx context.Context, token, tenantId, correlationID string) error
	ValidateToken(ctx context.Context, token, tenantId, correlationID string) error
	ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error
	ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error
	GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error)
	SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error)
	VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error)
	ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error)
	RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error
	RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error
	QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error)
	ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error)
	AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error
	RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error
	SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error)
	AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error
	AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error
	ListTenantUserSessions(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error)
	RevokeTenantUserSession(ctx context.Context, userId, deviceId, token, tenantId, correlationID string) error
	ListTenantUserBlacklist(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.AdminDeviceBlacklistEntryDTO, error)
	SearchTenantSessions(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceSessionResponseDTO, error)
	GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error)
	BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error
	DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error
}
