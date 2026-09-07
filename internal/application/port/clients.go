package port

import (
	"context"

	appdto "github.com/keepguard/bff-auth/internal/application/dto"
)

type AuthClient interface {
	Login(ctx context.Context, req appdto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (appdto.AuthResponseDTO, error)
	RefreshToken(ctx context.Context, req appdto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (appdto.RefreshTokenResponseDTO, error)
	Logout(ctx context.Context, token, tenantId, correlationID string) error
	ValidateToken(ctx context.Context, token, tenantId, correlationID string) error
	ChangePassword(ctx context.Context, req appdto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error
	ResetPassword(ctx context.Context, req appdto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error
	GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (appdto.GenerateResetTokenMSResponseDTO, error)
	SendDeviceChallenge(ctx context.Context, req appdto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error)
	VerifyDeviceChallenge(ctx context.Context, req appdto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (appdto.AuthResponseDTO, error)
	ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]appdto.DeviceSessionDTO, error)
	RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error
	RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error
	QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error)
	ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]appdto.DeviceBlacklistDTO, error)
	AddDeviceToBlacklist(ctx context.Context, req appdto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error
	RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error
	SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (appdto.PaginatedDeviceBlacklistResponseDTO, error)
	AdminAddDeviceToBlacklist(ctx context.Context, req appdto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error
	AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error
	ListTenantUserSessions(ctx context.Context, userId, token, tenantId, correlationID string) ([]appdto.DeviceSessionDTO, error)
	RevokeTenantUserSession(ctx context.Context, userId, deviceId, token, tenantId, correlationID string) error
	ListTenantUserBlacklist(ctx context.Context, userId, token, tenantId, correlationID string) ([]appdto.AdminDeviceBlacklistEntryDTO, error)
	SearchTenantSessions(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (appdto.PaginatedDeviceSessionResponseDTO, error)
	GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (appdto.UserByCodeResponseDTO, error)
	BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error
	DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error
}

type UserClient interface {
	GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (appdto.UserByEmailResponseDTO, error)
}

type CompanyClient interface {
	GetByTenantId(ctx context.Context, tenantId, correlationID string) (appdto.CompanySimpleResponseDTO, error)
}

type CommunicationClient interface {
	SendMessage(ctx context.Context, req appdto.SendMessageRequestDTO, tenantId, correlationID string) (appdto.SendMessageResponseDTO, error)
}

// Alias para call sites que ainda usam o nome antigo no pacote port.
type CompanySimpleResponseDTO = appdto.CompanySimpleResponseDTO
