package auth

import (
	"context"
	"net/http"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// metricsDecorator implementa o padrão decorator para coleta de métricas
type metricsDecorator struct {
	inner       authclient.AuthClient
	metrics     *metrics.Metrics
	serviceName string
}

// NewMetricsDecorator cria um decorator de métricas para AuthClient
func NewMetricsDecorator(
	inner authclient.AuthClient,
	metrics *metrics.Metrics,
	serviceName string,
) authclient.AuthClient {
	return &metricsDecorator{
		inner:       inner,
		metrics:     metrics,
		serviceName: serviceName,
	}
}

// getStatusCodeFromError extrai o status code de um erro HTTP
func (d *metricsDecorator) getStatusCodeFromError(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Tenta converter para HTTPError
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return httpErr.StatusCode
	}

	// Default para 500 se não conseguir identificar
	return http.StatusInternalServerError
}

// Login implementa o método Login com coleta de métricas
func (d *metricsDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/login", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/login", errorType)
	}

	return response, err
}

// RefreshToken implementa o método RefreshToken com coleta de métricas
func (d *metricsDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/refresh", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/refresh", errorType)
	}

	return response, err
}

// Logout implementa o método Logout com coleta de métricas
func (d *metricsDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	start := time.Now()

	err := d.inner.Logout(ctx, token, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/logout", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/logout", errorType)
	}

	return err
}

// ValidateToken implementa o método ValidateToken com coleta de métricas
func (d *metricsDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	start := time.Now()

	err := d.inner.ValidateToken(ctx, token, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/validate", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/validate", errorType)
	}

	return err
}

// ChangePassword implementa o método ChangePassword com coleta de métricas
func (d *metricsDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	start := time.Now()

	err := d.inner.ChangePassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/change-password", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/change-password", errorType)
	}

	return err
}

// ResetPassword implementa o método ResetPassword com coleta de métricas
func (d *metricsDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	start := time.Now()

	err := d.inner.ResetPassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/reset-password", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/reset-password", errorType)
	}

	return err
}

// GenerateResetToken implementa o método GenerateResetToken com coleta de métricas
func (d *metricsDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	start := time.Now()

	response, err := d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)

	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)

	// Registra métricas
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/generate-reset-token", statusCode, duration)

	if err != nil {
		errorType := "unknown"
		if httpErr, ok := err.(*appdto.HTTPError); ok {
			errorType = httpErr.ErrorCode
		}
		d.metrics.RecordUpstreamError(d.serviceName, "POST", "/generate-reset-token", errorType)
	}

	return response, err
}

// SendDeviceChallenge implementa o método SendDeviceChallenge com coleta de métricas
func (d *metricsDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	start := time.Now()
	response, err := d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/auth/device/challenge/send", statusCode, duration)
	return response, err
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge com coleta de métricas
func (d *metricsDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()
	response, err := d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/auth/device/challenge/verify", statusCode, duration)
	return response, err
}

// ListUserSessions implementa o método ListUserSessions com coleta de métricas
func (d *metricsDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	start := time.Now()
	response, err := d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/me/sessions", statusCode, duration)
	return response, err
}

// RevokeSession implementa o método RevokeSession com coleta de métricas
func (d *metricsDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users/me/sessions/:id", statusCode, duration)
	return err
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions com coleta de métricas
func (d *metricsDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users/me/sessions", statusCode, duration)
	return err
}

// QuickRevoke implementa o método QuickRevoke com coleta de métricas
func (d *metricsDecorator) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	start := time.Now()
	response, err := d.inner.QuickRevoke(ctx, token, blacklist, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/auth/device/quick-revoke", statusCode, duration)
	return response, err
}

// ListDeviceBlacklist implementa o método ListDeviceBlacklist com coleta de métricas
func (d *metricsDecorator) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	start := time.Now()
	response, err := d.inner.ListDeviceBlacklist(ctx, token, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/me/devices/blacklist", statusCode, duration)
	return response, err
}

// AddDeviceToBlacklist implementa o método AddDeviceToBlacklist com coleta de métricas
func (d *metricsDecorator) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.AddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/users/me/devices/blacklist", statusCode, duration)
	return err
}

// RemoveDeviceFromBlacklist implementa o método RemoveDeviceFromBlacklist com coleta de métricas
func (d *metricsDecorator) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.RemoveDeviceFromBlacklist(ctx, deviceId, token, tenantId, correlationID)
	duration := time.Since(start)
	statusCode := d.getStatusCodeFromError(err)
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users/me/devices/blacklist/:id", statusCode, duration)
	return err
}

func (d *metricsDecorator) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error) {
	return d.inner.SearchAdminDeviceBlacklist(ctx, queryParams, token, tenantId, correlationID)
}

func (d *metricsDecorator) AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.inner.AdminAddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

func (d *metricsDecorator) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	return d.inner.AdminRemoveDeviceFromBlacklist(ctx, deviceId, userId, token, tenantId, correlationID)
}

func (d *metricsDecorator) ListTenantUserSessions(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	start := time.Now()
	response, err := d.inner.ListTenantUserSessions(ctx, userId, token, tenantId, correlationID)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/:id/sessions", d.getStatusCodeFromError(err), time.Since(start))
	return response, err
}

func (d *metricsDecorator) RevokeTenantUserSession(ctx context.Context, userId, deviceId, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.RevokeTenantUserSession(ctx, userId, deviceId, token, tenantId, correlationID)
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users/:id/sessions/:deviceId", d.getStatusCodeFromError(err), time.Since(start))
	return err
}

func (d *metricsDecorator) ListTenantUserBlacklist(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.AdminDeviceBlacklistEntryDTO, error) {
	start := time.Now()
	response, err := d.inner.ListTenantUserBlacklist(ctx, userId, token, tenantId, correlationID)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/:id/devices/blacklist", d.getStatusCodeFromError(err), time.Since(start))
	return response, err
}

func (d *metricsDecorator) SearchTenantSessions(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceSessionResponseDTO, error) {
	start := time.Now()
	response, err := d.inner.SearchTenantSessions(ctx, queryParams, token, tenantId, correlationID)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/sessions", d.getStatusCodeFromError(err), time.Since(start))
	return response, err
}

func (d *metricsDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error) {
	start := time.Now()
	response, err := d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	duration := time.Since(start)
	d.metrics.RecordUpstreamRequest(d.serviceName, "GET", "/users/code-user/:id", d.getStatusCodeFromError(err), duration)
	return response, err
}

func (d *metricsDecorator) BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.BlockUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
	duration := time.Since(start)
	d.metrics.RecordUpstreamRequest(d.serviceName, "POST", "/users/block/:id", d.getStatusCodeFromError(err), duration)
	return err
}

func (d *metricsDecorator) DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	start := time.Now()
	err := d.inner.DeleteUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
	duration := time.Since(start)
	d.metrics.RecordUpstreamRequest(d.serviceName, "DELETE", "/users/delete/:id", d.getStatusCodeFromError(err), duration)
	return err
}
