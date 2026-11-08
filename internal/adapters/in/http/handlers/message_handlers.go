package http

import (
	"errors"
	"net/http"

	"github.com/keepguard/bff-auth/internal/adapters/in/http/dto"
	"github.com/keepguard/bff-auth/internal/adapters/in/http/mapper"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/message"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// MessageHandlers implementa os handlers HTTP para mensagens
type MessageHandlers struct {
	sendResetPasswordMessageUseCase message.SendResetPasswordMessageUseCase
	logger                          *zap.Logger
}

// NewMessageHandlers cria uma nova instância dos MessageHandlers
func NewMessageHandlers(
	sendResetPasswordMessageUseCase message.SendResetPasswordMessageUseCase,
	logger *zap.Logger,
) *MessageHandlers {
	return &MessageHandlers{
		sendResetPasswordMessageUseCase: sendResetPasswordMessageUseCase,
		logger:                          logger,
	}
}

// NewMessageHandlersWithLogger cria uma nova instância dos MessageHandlers usando a interface logger.Logger
func NewMessageHandlersWithLogger(
	sendResetPasswordMessageUseCase message.SendResetPasswordMessageUseCase,
	log logger.Logger,
) *MessageHandlers {
	// Converter logger.Logger para *zap.Logger se necessário
	zapLogger, _ := zap.NewDevelopment()
	return &MessageHandlers{
		sendResetPasswordMessageUseCase: sendResetPasswordMessageUseCase,
		logger:                          zapLogger,
	}
}

// SendResetPasswordMessageHandler trata requisições de envio de mensagem de reset de senha
// @Summary Send forgot password message
// @Description Envia uma mensagem com token de recuperação de senha para o email do usuário. Não requer autenticação. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.SendResetPasswordMessageRequestDTO true "Email do usuário"
// @Success 200 {object} dto.SendResetPasswordMessageResponseDTO "Mensagem enviada com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes, email inválido ou usuário não ativo)"
// @Failure 404 {object} pkg.ErrorResponse "Usuário não encontrado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/forgot-password [post]
func (h *MessageHandlers) SendResetPasswordMessageHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	var req dto.SendResetPasswordMessageRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de envio de mensagem de reset",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewSendResetPasswordMessageCommand(
		req.Email,
		tenantId,
		correlationID,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de envio de mensagem de reset",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "VALIDATION_ERROR",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	response, err := h.sendResetPasswordMessageUseCase.Execute(c.Request().Context(), command)
	if err != nil && contaSemRecuperacao(err) {
		// Conta inexistente ou inativa: mesma resposta do sucesso, para ninguém descobrir quais e-mails têm conta
		// (antes: 404 USER_NOT_FOUND e 400 USER_NOT_ACTIVE). Nada é enviado.
		h.logger.Info("Esqueci a senha sem conta ativa para o e-mail (resposta genérica)",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
		)
		return c.JSON(http.StatusOK, mapper.ToSendResetPasswordMessageResponse(appdto.SendResetPasswordMessageViewDTO{
			Success: true, Message: appdto.ResetPasswordGenericMessage,
		}))
	}
	if err != nil {
		h.logger.Error("Erro no caso de uso de envio de mensagem de reset",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Mensagem de reset de senha enviada com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, mapper.ToSendResetPasswordMessageResponse(response))
}

// contaSemRecuperacao: o e-mail não tem conta (404) ou a conta não está ativa (USER_NOT_ACTIVE).
func contaSemRecuperacao(err error) bool {
	var httpErr *appdto.HTTPError
	if errors.As(err, &httpErr) && httpErr.StatusCode == http.StatusNotFound {
		return true
	}
	var appErr *pkg.AppError
	if errors.As(err, &appErr) && (appErr.StatusCode == http.StatusNotFound || string(appErr.Code) == "USER_NOT_ACTIVE") {
		return true
	}
	return false
}
