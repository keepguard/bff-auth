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
	if err := json.Unmarshal(resp.Body(), &msError); err == nil && msError.Title != "" {
		// Erro estruturado do ms-auth
		return &appdto.HTTPError{
			StatusCode: msError.Status,
			Message:    msError.Detail,
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
func (c *AuthClient) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId string) (inboundDto.AuthResponseDTO, error) {
	var response inboundDto.AuthResponseDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		SetHeader("X-Client-ID", clientId).
		Post(c.baseURL + "/api/v1/auth/login")

	if err != nil {
		return inboundDto.AuthResponseDTO{}, fmt.Errorf("erro ao fazer requisição de login: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "login"); httpErr != nil {
		return inboundDto.AuthResponseDTO{}, httpErr
	}

	return response, nil
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
func (c *AuthClient) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/change-password")

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
func (c *AuthClient) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID string) error {
	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(req).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/reset-password")

	if err != nil {
		return fmt.Errorf("erro ao fazer requisição de reset de senha: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "reset password"); httpErr != nil {
		return httpErr
	}

	return nil
}

// GenerateResetToken gera um token de recuperação de senha no ms-auth
func (c *AuthClient) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	// Converter map para DTO
	reqDTO := outboundDto.GenerateResetTokenMSRequestDTO{}
	if codeUser, ok := req["codeUser"].(string); ok {
		reqDTO.CodeUser = codeUser
	}
	if messageType, ok := req["messageType"].(string); ok {
		reqDTO.MessageType = messageType
	}
	if communicationType, ok := req["communicationType"].(string); ok {
		reqDTO.CommunicationType = communicationType
	}
	if templateType, ok := req["templateType"].(string); ok {
		reqDTO.TemplateType = templateType
	}

	var response outboundDto.GenerateResetTokenMSResponseDTO

	resp, err := c.client.R().
		SetContext(ctx).
		SetBody(reqDTO).
		SetResult(&response).
		SetHeader("X-Correlation-ID", correlationID).
		SetHeader("X-Tenant-Id", tenantId).
		Post(c.baseURL + "/api/v1/auth/generate-reset-token")

	if err != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, fmt.Errorf("erro ao fazer requisição de geração de token de reset: %w", err)
	}

	// Trata erros HTTP do ms-auth
	if httpErr := c.handleHTTPError(resp, "generate reset token"); httpErr != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, httpErr
	}

	return response, nil
}
