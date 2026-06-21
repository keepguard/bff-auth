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
func (d *loggingDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, xApplication, correlationID string) (inboundDto.AuthResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "Login"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("username", req.Username),
	)

	response, err := d.inner.Login(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "Login"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "Login"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.Duration("duration", duration),
		zap.Int64("expiresIn", response.ExpiresIn),
	)

	return response, nil
}

// RefreshToken implementa o método RefreshToken com logging estruturado
func (d *loggingDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, xApplication, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "RefreshToken"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
	)

	response, err := d.inner.RefreshToken(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "RefreshToken"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "RefreshToken"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.Duration("duration", duration),
		zap.Int64("expiresIn", response.ExpiresIn),
	)

	return response, nil
}

// Logout implementa o método Logout com logging estruturado
func (d *loggingDecorator) Logout(ctx context.Context, token, xApplication, correlationID string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "Logout"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
	)

	err := d.inner.Logout(ctx, token, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "Logout"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "Logout"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.Duration("duration", duration),
	)

	return nil
}

// ValidateToken implementa o método ValidateToken com logging estruturado
func (d *loggingDecorator) ValidateToken(ctx context.Context, token, xApplication, correlationID string) error {
	start := time.Now()

	d.logger.Debug("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ValidateToken"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
	)

	err := d.inner.ValidateToken(ctx, token, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Warn("Token inválido",
			zap.String("service", d.serviceName),
			zap.String("operation", "ValidateToken"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Debug("Token validado com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ValidateToken"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.Duration("duration", duration),
	)

	return nil
}

// ChangePassword implementa o método ChangePassword com logging estruturado
func (d *loggingDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, xApplication, correlationID string) error {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "ChangePassword"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("codeUser", req.CodeUser),
	)

	err := d.inner.ChangePassword(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ChangePassword"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
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
		zap.String("xApplication", xApplication),
		zap.String("codeUser", req.CodeUser),
		zap.Duration("duration", duration),
	)

	return nil
}

// ResetPassword implementa o método ResetPassword com logging estruturado
func (d *loggingDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, xApplication, correlationID string) error {
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
		zap.String("xApplication", xApplication),
		zap.String("request", string(reqJSON)),
	)

	err := d.inner.ResetPassword(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "ResetPassword"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return err
	}

	d.logger.Info("Senha resetada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "ResetPassword"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.Duration("duration", duration),
	)

	return nil
}

// GenerateResetToken implementa o método GenerateResetToken com logging estruturado
func (d *loggingDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, xApplication, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GenerateResetToken"),
		zap.String("correlationID", correlationID),
		zap.String("xApplication", xApplication),
		zap.String("codeUser", getStringFromMap(req, "codeUser")),
	)

	response, err := d.inner.GenerateResetToken(ctx, req, xApplication, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GenerateResetToken"),
			zap.String("correlationID", correlationID),
			zap.String("xApplication", xApplication),
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
		zap.String("xApplication", xApplication),
		zap.String("codeUser", getStringFromMap(req, "codeUser")),
		zap.Int64("expiresInSeconds", response.ExpiresInSeconds),
		zap.Duration("duration", duration),
	)

	return response, nil
}

// getStringFromMap extrai uma string de um map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
