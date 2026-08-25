package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// AuthClient implementa o cliente HTTP para autenticação
type AuthClient struct {
	client  *resty.Client
	baseURL string
}

// NewAuthClient cria uma nova instância do AuthClient
func NewAuthClient(baseURL string, timeout time.Duration) authclient.AuthClient {
	restyClient := resty.New()
	restyClient.SetTimeout(timeout)
	restyClient.SetRetryCount(1) // Reduzido para 1 retry apenas

	return &AuthClient{
		client:  restyClient,
		baseURL: baseURL,
	}
}

// handleHTTPError trata erros HTTP e extrai informações do ms-auth
func (c *AuthClient) handleHTTPError(resp *resty.Response, operation string) error {
	// Se não é um erro HTTP, retorna o erro original
	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		return nil
	}

	// Tenta extrair erro do ms-auth (ProblemDetail)
	var msError appdto.MSAuthErrorResponse
	if err := json.Unmarshal(resp.Body(), &msError); err == nil && (msError.Title != "" || msError.Detail != "") {
		message := msError.Detail
		if message == "" {
			message = msError.Title
		}
		return &appdto.HTTPError{
			StatusCode: msError.Status,
			Message:    message,
			ErrorCode:  getErrorCodeFromProperties(msError.Properties),
			Details:    msError.Properties,
		}
	}

	// Fallback para erro genérico
	return &appdto.HTTPError{
		StatusCode: resp.StatusCode(),
		Message:    fmt.Sprintf("erro de %s: status %d", operation, resp.StatusCode()),
		ErrorCode:  "HTTP_ERROR",
		Details:    map[string]interface{}{"body": string(resp.Body())},
	}
}

// getErrorCodeFromProperties extrai o código de erro das propriedades
func getErrorCodeFromProperties(properties map[string]interface{}) string {
	if errorCode, ok := properties["errorCode"].(string); ok {
		return errorCode
	}
	return "UNKNOWN_ERROR"
}

// Login realiza login no ms-auth
func (c *AuthClient) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	var response inboundDto.AuthResponseDTO

	reqBuilder := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("X-Client-ID", clientId)

	if deviceId != "" {
		reqBuilder.SetHeader("X-Device-Id", deviceId)
	}
	if deviceName != "" {
		reqBuilder.SetHeader("X-Device-Name", deviceName)
	}
	if deviceType != "" {
		reqBuilder.SetHeader("X-Device-Type", deviceType)
	}
	if ipAddress != "" {
		reqBuilder.SetHeader("X-Forwarded-For", ipAddress)
	}
	if userAgent != "" {
		reqBuilder.SetHeader("User-Agent", userAgent)
	}

	resp, err := reqBuilder.Post(c.baseURL + "/api/v1/auth/login")

	if err != nil {
		return inboundDto.AuthResponseDTO{}, fmt.Errorf("erro ao fazer requisição de login: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "login"); httpErr != nil {
		return inboundDto.AuthResponseDTO{}, httpErr
	}

	return response, nil
}

// SendDeviceChallenge envia código de desafio para novo dispositivo
func (c *AuthClient) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	var response map[string]interface{}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/device/challenge/send")

	if err != nil {
		return nil, fmt.Errorf("erro ao solicitar envio de código de dispositivo: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "device_challenge_send"); httpErr != nil {
		return nil, httpErr
	}

	return response, nil
}

// VerifyDeviceChallenge valida código de desafio e emite JWT
func (c *AuthClient) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	var response inboundDto.AuthResponseDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/device/challenge/verify")

	if err != nil {
		return inboundDto.AuthResponseDTO{}, fmt.Errorf("erro ao validar código de dispositivo: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "device_challenge_verify"); httpErr != nil {
		return inboundDto.AuthResponseDTO{}, httpErr
	}

	return response, nil
}

// ListUserSessions lista sessões ativas do usuário
func (c *AuthClient) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	var response []inboundDto.DeviceSessionDTO

	reqBuilder := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token)

	if deviceId != "" {
		reqBuilder.SetHeader("X-Device-Id", deviceId)
	}

	resp, err := reqBuilder.Get(c.baseURL + "/api/v1/users/me/sessions")

	if err != nil {
		return nil, fmt.Errorf("erro ao listar sessões de usuário: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "list_user_sessions"); httpErr != nil {
		return nil, httpErr
	}

	return response, nil
}

// RevokeSession revoga sessão de dispositivo específico
func (c *AuthClient) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token).
		Delete(fmt.Sprintf("%s/api/v1/users/me/sessions/%s", c.baseURL, deviceIdToRevoke))

	if err != nil {
		return fmt.Errorf("erro ao revogar sessão de dispositivo: %w", err)
	}

	return c.handleHTTPError(resp, "revoke_session")
}

// RevokeAllOtherSessions revoga todas as outras sessões exceto atual
func (c *AuthClient) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	reqBuilder := c.client.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token)

	if currentDeviceId != "" {
		reqBuilder.SetHeader("X-Device-Id", currentDeviceId)
	}

	resp, err := reqBuilder.Delete(c.baseURL + "/api/v1/users/me/sessions")

	if err != nil {
		return fmt.Errorf("erro ao revogar outras sessões: %w", err)
	}

	return c.handleHTTPError(resp, "revoke_all_other_sessions")
}

// RefreshToken renova o token no ms-auth
func (c *AuthClient) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	var response inboundDto.RefreshTokenResponseDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("X-Client-ID", clientId).
		Post(c.baseURL + "/api/v1/auth/refresh")

	if err != nil {
		return inboundDto.RefreshTokenResponseDTO{}, fmt.Errorf("erro ao fazer requisição de refresh: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "refresh"); httpErr != nil {
		return inboundDto.RefreshTokenResponseDTO{}, httpErr
	}

	return response, nil
}

// Logout realiza logout no ms-auth
func (c *AuthClient) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/logout")

	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de logout: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "logout"); httpErr != nil {
		return httpErr
	}

	return nil
}

// ValidateToken valida um token no ms-auth
func (c *AuthClient) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	req := map[string]string{
		"token": token,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/validate")

	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de validação de token: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "validate token"); httpErr != nil {
		return httpErr
	}

	return nil
}

// ChangePassword altera a senha do usuário no ms-auth
func (c *AuthClient) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	reqBuilder := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId)

	if deviceId != "" {
		reqBuilder.SetHeader("X-Device-Id", deviceId)
	}
	if deviceName != "" {
		reqBuilder.SetHeader("X-Device-Name", deviceName)
	}
	if deviceType != "" {
		reqBuilder.SetHeader("X-Device-Type", deviceType)
	}
	if ipAddress != "" {
		reqBuilder.SetHeader("X-Forwarded-For", ipAddress)
	}
	if userAgent != "" {
		reqBuilder.SetHeader("User-Agent", userAgent)
	}

	resp, err := reqBuilder.Post(c.baseURL + "/api/v1/auth/change-password")

	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de alteração de senha: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "change password"); httpErr != nil {
		return httpErr
	}

	return nil
}

// ResetPassword reseta a senha do usuário no ms-auth
func (c *AuthClient) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	reqBuilder := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId)

	if deviceId != "" {
		reqBuilder.SetHeader("X-Device-Id", deviceId)
	}
	if deviceName != "" {
		reqBuilder.SetHeader("X-Device-Name", deviceName)
	}
	if deviceType != "" {
		reqBuilder.SetHeader("X-Device-Type", deviceType)
	}
	if ipAddress != "" {
		reqBuilder.SetHeader("X-Forwarded-For", ipAddress)
	}
	if userAgent != "" {
		reqBuilder.SetHeader("User-Agent", userAgent)
	}

	resp, err := reqBuilder.Post(c.baseURL + "/api/v1/auth/reset-password")

	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de reset de senha: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "reset password"); httpErr != nil {
		return httpErr
	}

	return nil
}

// GenerateResetToken gera um token de reset no ms-auth
func (c *AuthClient) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	var response outboundDto.GenerateResetTokenMSResponseDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/generate-reset-token")

	if err != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, fmt.Errorf("erro ao fazer requisição de gerar reset token: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "generate_reset_token"); httpErr != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, httpErr
	}

	return response, nil
}

// QuickRevoke realiza revogação rápida via token de e-mail
func (c *AuthClient) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	var response map[string]interface{}

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetQueryParam("token", token).
		SetQueryParam("blacklist", fmt.Sprintf("%t", blacklist)).
		Get(c.baseURL + "/api/v1/auth/device/quick-revoke")

	if err != nil {
		return nil, fmt.Errorf("erro ao fazer quick revoke: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "quick_revoke"); httpErr != nil {
		return nil, httpErr
	}

	return response, nil
}

// ListDeviceBlacklist lista dispositivos na blacklist do usuário
func (c *AuthClient) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	var response []inboundDto.DeviceBlacklistDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token).
		Get(c.baseURL + "/api/v1/users/me/devices/blacklist")

	if err != nil {
		return nil, fmt.Errorf("erro ao listar blacklist de dispositivos: %w", err)
	}

	if httpErr := c.handleHTTPError(resp, "list_device_blacklist"); httpErr != nil {
		return nil, httpErr
	}

	return response, nil
}

// AddDeviceToBlacklist adiciona dispositivo à blacklist
func (c *AuthClient) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token).
		Post(c.baseURL + "/api/v1/users/me/devices/blacklist")

	if err != nil {
		return fmt.Errorf("erro ao adicionar dispositivo à blacklist: %w", err)
	}

	return c.handleHTTPError(resp, "add_device_blacklist")
}

// RemoveDeviceFromBlacklist remove dispositivo da blacklist
func (c *AuthClient) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetAuthToken(token).
		Delete(fmt.Sprintf("%s/api/v1/users/me/devices/blacklist/%s", c.baseURL, deviceId))

	if err != nil {
		return fmt.Errorf("erro ao remover dispositivo da blacklist: %w", err)
	}

	return c.handleHTTPError(resp, "remove_device_blacklist")
}
