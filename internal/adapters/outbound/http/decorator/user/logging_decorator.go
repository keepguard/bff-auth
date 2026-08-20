package user

import (
	"context"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"go.uber.org/zap"
)

// userLoggingDecorator implementa logging para UserClient
type userLoggingDecorator struct {
	inner       portsclient.UserClient
	logger      *zap.Logger
	serviceName string
}

// NewUserLoggingDecorator cria um decorator de logging para UserClient
func NewUserLoggingDecorator(
	inner portsclient.UserClient,
	logger *zap.Logger,
	serviceName string,
) portsclient.UserClient {
	return &userLoggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// GetByEmail implementa GetByEmail com logging
func (d *userLoggingDecorator) GetByEmail(ctx context.Context, email, tenantId, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByEmail"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
	)

	response, err := d.inner.GetByEmail(ctx, email, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "GetByEmail"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("email", email),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Requisição concluída com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "GetByEmail"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("email", email),
		zap.String("userID", response.ID),
		zap.String("codeUser", response.CodeUser),
		zap.Duration("duration", duration),
	)

	return response, nil
}
