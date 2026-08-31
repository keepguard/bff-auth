package auth

import (
	"context"
	"encoding/json"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"go.uber.org/zap"
)

// loggingDecorator implementa o padrão decorator para logging estruturado
type loggingDecorator struct {
	inner       authclient.AuthClient
	logger      *zap.Logger
	serviceName string
}

// NewLoggingDecorator cria um decorator de logging para AuthClient
func NewLoggingDecorator(
	inner authclient.AuthClient,
	logger *zap.Logger,
	serviceName string,
) authclient.AuthClient {
	return &loggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// Login implementa o método Login com logging estruturado
func (d *loggingDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "Login"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("username", req.Username),
		zap.String("deviceId", deviceId),
	)

	response, err := d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "Login"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "Login"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Duration("duration", duration),
		zap.Int64("expiresIn", response.ExpiresIn),
	)

	return response, nil
}

// RefreshToken implementa o método RefreshToken com logging estruturado
func (d *loggingDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "RefreshToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	response, err := d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "RefreshToken"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "RefreshToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Duration("duration", duration),
		zap.Int64("expiresIn", response.ExpiresIn),
	)

	return response, nil
}

// Logout implementa o método Logout com logging estruturado
func (d *loggingDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "Logout"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	err := d.inner.Logout(ctx, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "Logout"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "Logout"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Duration("duration", duration),
	)

	return nil
}

// ValidateToken implementa o método ValidateToken com logging estruturado
func (d *loggingDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	start := time.Now()

	d.logger.Debug("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ValidateToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	err := d.inner.ValidateToken(ctx, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Warn("Token inválido",
			zap.String("service", d.serviceName),
			zap.String("operation", "ValidateToken"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Debug("Token validado com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ValidateToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Duration("duration", duration),
	)

	return nil
}

// ChangePassword implementa o método ChangePassword com logging estruturado
func (d *loggingDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ChangePassword"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", req.CodeUser),
	)

	err := d.inner.ChangePassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ChangePassword"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("codeUser", req.CodeUser),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Senha alterada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ChangePassword"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", req.CodeUser),
		zap.Duration("duration", duration),
	)

	return nil
}

// ResetPassword implementa o método ResetPassword com logging estruturado
func (d *loggingDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	start := time.Now()

	// Converte request para JSON para logging (sem senha)
	reqForLog := map[string]interface{}{
		"codeUser":   req.CodeUser,
		"resetToken": "***", // Oculta token
	}
	reqJSON, _ := json.Marshal(reqForLog)

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ResetPassword"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("request", string(reqJSON)),
	)

	err := d.inner.ResetPassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ResetPassword"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Senha resetada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ResetPassword"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Duration("duration", duration),
	)

	return nil
}

// GenerateResetToken implementa o método GenerateResetToken com logging estruturado
func (d *loggingDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GenerateResetToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", getStringFromMap(req, "codeUser")),
	)

	response, err := d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GenerateResetToken"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("codeUser", getStringFromMap(req, "codeUser")),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Token de reset gerado com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GenerateResetToken"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("codeUser", getStringFromMap(req, "codeUser")),
		zap.Int64("expiresInSeconds", response.ExpiresInSeconds),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// SendDeviceChallenge implementa o método SendDeviceChallenge com logging estruturado
func (d *loggingDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	start := time.Now()
	d.logger.Info("Iniciando envio de desafio de dispositivo",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendDeviceChallenge"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("channel", req.Channel),
	)

	response, err := d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao enviar desafio de dispositivo",
			zap.String("service", d.serviceName),
			zap.String("operation", "SendDeviceChallenge"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	return response, nil
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge com logging estruturado
func (d *loggingDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()
	d.logger.Info("Iniciando verificação de desafio de dispositivo",
		zap.String("service", d.serviceName),
		zap.String("operation", "VerifyDeviceChallenge"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	response, err := d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao verificar desafio de dispositivo",
			zap.String("service", d.serviceName),
			zap.String("operation", "VerifyDeviceChallenge"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	return response, nil
}

// ListUserSessions implementa o método ListUserSessions com logging estruturado
func (d *loggingDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	start := time.Now()
	d.logger.Info("Iniciando listagem de sessões do usuário",
		zap.String("service", d.serviceName),
		zap.String("operation", "ListUserSessions"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	response, err := d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao listar sessões do usuário",
			zap.String("service", d.serviceName),
			zap.String("operation", "ListUserSessions"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	return response, nil
}

// RevokeSession implementa o método RevokeSession com logging estruturado
func (d *loggingDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	start := time.Now()
	d.logger.Info("Iniciando revogação de sessão",
		zap.String("service", d.serviceName),
		zap.String("operation", "RevokeSession"),
		zap.String("correlationID", correlationID),
		zap.String("deviceIdToRevoke", deviceIdToRevoke),
	)

	err := d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao revogar sessão",
			zap.String("service", d.serviceName),
			zap.String("operation", "RevokeSession"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions com logging estruturado
func (d *loggingDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	start := time.Now()
	d.logger.Info("Iniciando revogação de todas as outras sessões",
		zap.String("service", d.serviceName),
		zap.String("operation", "RevokeAllOtherSessions"),
		zap.String("correlationID", correlationID),
		zap.String("currentDeviceId", currentDeviceId),
	)

	err := d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao revogar outras sessões",
			zap.String("service", d.serviceName),
			zap.String("operation", "RevokeAllOtherSessions"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// QuickRevoke implementa o método QuickRevoke com logging estruturado
func (d *loggingDecorator) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	start := time.Now()
	d.logger.Info("Iniciando quick revoke de dispositivo",
		zap.String("service", d.serviceName),
		zap.String("operation", "QuickRevoke"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.Bool("blacklist", blacklist),
	)

	response, err := d.inner.QuickRevoke(ctx, token, blacklist, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro no quick revoke",
			zap.String("service", d.serviceName),
			zap.String("operation", "QuickRevoke"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	return response, nil
}

// ListDeviceBlacklist implementa o método ListDeviceBlacklist com logging estruturado
func (d *loggingDecorator) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	start := time.Now()
	d.logger.Info("Iniciando listagem da blacklist de dispositivos",
		zap.String("service", d.serviceName),
		zap.String("operation", "ListDeviceBlacklist"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
	)

	response, err := d.inner.ListDeviceBlacklist(ctx, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao listar blacklist de dispositivos",
			zap.String("service", d.serviceName),
			zap.String("operation", "ListDeviceBlacklist"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	return response, nil
}

// AddDeviceToBlacklist implementa o método AddDeviceToBlacklist com logging estruturado
func (d *loggingDecorator) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	start := time.Now()
	d.logger.Info("Iniciando adição de dispositivo à blacklist",
		zap.String("service", d.serviceName),
		zap.String("operation", "AddDeviceToBlacklist"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("deviceId", req.DeviceID),
	)

	err := d.inner.AddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao adicionar dispositivo à blacklist",
			zap.String("service", d.serviceName),
			zap.String("operation", "AddDeviceToBlacklist"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// RemoveDeviceFromBlacklist implementa o método RemoveDeviceFromBlacklist com logging estruturado
func (d *loggingDecorator) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	start := time.Now()
	d.logger.Info("Iniciando remoção de dispositivo da blacklist",
		zap.String("service", d.serviceName),
		zap.String("operation", "RemoveDeviceFromBlacklist"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("deviceId", deviceId),
	)

	err := d.inner.RemoveDeviceFromBlacklist(ctx, deviceId, token, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro ao remover dispositivo da blacklist",
			zap.String("service", d.serviceName),
			zap.String("operation", "RemoveDeviceFromBlacklist"),
			zap.String("correlationID", correlationID),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// getStringFromMap extrai uma string de um map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (d *loggingDecorator) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error) {
	return d.inner.SearchAdminDeviceBlacklist(ctx, queryParams, token, tenantId, correlationID)
}

func (d *loggingDecorator) AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.inner.AdminAddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

func (d *loggingDecorator) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	return d.inner.AdminRemoveDeviceFromBlacklist(ctx, deviceId, userId, token, tenantId, correlationID)
}

func (d *loggingDecorator) ListTenantUserSessions(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	return d.inner.ListTenantUserSessions(ctx, userId, token, tenantId, correlationID)
}

func (d *loggingDecorator) RevokeTenantUserSession(ctx context.Context, userId, deviceId, token, tenantId, correlationID string) error {
	return d.inner.RevokeTenantUserSession(ctx, userId, deviceId, token, tenantId, correlationID)
}

func (d *loggingDecorator) ListTenantUserBlacklist(ctx context.Context, userId, token, tenantId, correlationID string) ([]inboundDto.AdminDeviceBlacklistEntryDTO, error) {
	return d.inner.ListTenantUserBlacklist(ctx, userId, token, tenantId, correlationID)
}

func (d *loggingDecorator) SearchTenantSessions(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceSessionResponseDTO, error) {
	return d.inner.SearchTenantSessions(ctx, queryParams, token, tenantId, correlationID)
}

func (d *loggingDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error) {
	return d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
}

func (d *loggingDecorator) BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	return d.inner.BlockUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
}

func (d *loggingDecorator) DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	return d.inner.DeleteUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
}
