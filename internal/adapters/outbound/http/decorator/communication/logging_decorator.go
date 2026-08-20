package communication

import (
	"context"
	"time"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"go.uber.org/zap"
)

// communicationLoggingDecorator implementa logging para CommunicationClient
type communicationLoggingDecorator struct {
	inner       portsclient.CommunicationClient
	logger      *zap.Logger
	serviceName string
}

// NewCommunicationLoggingDecorator cria um decorator de logging para CommunicationClient
func NewCommunicationLoggingDecorator(
	inner portsclient.CommunicationClient,
	logger *zap.Logger,
	serviceName string,
) portsclient.CommunicationClient {
	return &communicationLoggingDecorator{
		inner:       inner,
		logger:      logger,
		serviceName: serviceName,
	}
}

// SendMessage implementa SendMessage com logging
func (d *communicationLoggingDecorator) SendMessage(ctx context.Context, req outboundDto.SendMessageRequestDTO, tenantId, correlationID string) (outboundDto.SendMessageResponseDTO, error) {
	start := time.Now()

	d.logger.Info("Iniciando requisição",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendMessage"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("recipient", req.Recipient),
		zap.String("messageType", req.MessageType),
		zap.String("templateType", req.TemplateType),
	)

	response, err := d.inner.SendMessage(ctx, req, tenantId, correlationID)
	duration := time.Since(start)

	if err != nil {
		d.logger.Error("Erro na requisição",
			zap.String("service", d.serviceName),
			zap.String("operation", "SendMessage"),
			zap.String("correlationID", correlationID),
			zap.String("tenantId", tenantId),
			zap.String("recipient", req.Recipient),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return response, err
	}

	d.logger.Info("Mensagem enviada com sucesso",
		zap.String("service", d.serviceName),
		zap.String("operation", "SendMessage"),
		zap.String("correlationID", correlationID),
		zap.String("tenantId", tenantId),
		zap.String("recipient", req.Recipient),
		zap.Bool("success", response.Success),
		zap.Duration("duration", duration),
	)

	return response, nil
}
