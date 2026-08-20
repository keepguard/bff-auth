package http

import (
	"net/http"
	"strings"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-auth/internal/application/auth"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// AuthHandlers implementa os handlers HTTP para autenticação
type AuthHandlers struct {
	loginUseCase          auth.LoginUseCase
	refreshTokenUseCase   auth.RefreshUseCase
	logoutUseCase         auth.LogoutUseCase
	validateTokenUseCase  auth.ValidateTokenUseCase
	changePasswordUseCase auth.ChangePasswordUseCase
	resetPasswordUseCase  auth.ResetPasswordUseCase
	logger                *zap.Logger
}

// NewAuthHandlers cria uma nova instância dos AuthHandlers
func NewAuthHandlers(
	loginUseCase auth.LoginUseCase,
	refreshTokenUseCase auth.RefreshUseCase,
	logoutUseCase auth.LogoutUseCase,
	validateTokenUseCase auth.ValidateTokenUseCase,
	changePasswordUseCase auth.ChangePasswordUseCase,
	resetPasswordUseCase auth.ResetPasswordUseCase,
	logger *zap.Logger,
) *AuthHandlers {
	return &AuthHandlers{
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		validateTokenUseCase:  validateTokenUseCase,
		changePasswordUseCase: changePasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
		logger:                logger,
	}
}

// NewAuthHandlersWithLogger cria uma nova instância dos AuthHandlers usando a interface logger.Logger
func NewAuthHandlersWithLogger(
	loginUseCase auth.LoginUseCase,
	refreshTokenUseCase auth.RefreshUseCase,
	logoutUseCase auth.LogoutUseCase,
	validateTokenUseCase auth.ValidateTokenUseCase,
	changePasswordUseCase auth.ChangePasswordUseCase,
	resetPasswordUseCase auth.ResetPasswordUseCase,
	log logger.Logger,
) *AuthHandlers {
	// Converter logger.Logger para *zap.Logger se necessário
	// Por enquanto, vamos criar um logger zap básico
	zapLogger, _ := zap.NewDevelopment()
	return &AuthHandlers{
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		validateTokenUseCase:  validateTokenUseCase,
		changePasswordUseCase: changePasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
		logger:                zapLogger,
	}
}

// LoginHandler trata requisições de login
// @Summary Login
// @Description Realiza login do usuário usando credenciais de username e password. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.AuthRequestDTO true "Credenciais de login"
// @Success 200 {object} dto.AuthResponseDTO "Login realizado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 401 {object} pkg.ErrorResponse "Credenciais inválidas"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/login [post]
func (h *AuthHandlers) LoginHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, clientId, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	var req dto.AuthRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de login",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewLoginCommand(
		req.Username,
		req.Password,
		tenantId,
		correlationID,
		clientId,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de login",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	response, err := h.loginUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de login",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Login realizado com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
		zap.String("username", req.Username),
	)

	return c.JSON(http.StatusOK, response)
}

// RefreshHandler trata requisições de refresh de token
// @Summary Refresh token
// @Description Renova o token de acesso usando o refresh token. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.RefreshTokenRequestDTO true "Token de refresh para renovação"
// @Success 200 {object} dto.RefreshTokenResponseDTO "Token renovado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 401 {object} pkg.ErrorResponse "Refresh token inválido ou expirado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/refresh [post]
func (h *AuthHandlers) RefreshHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, clientId, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	var req dto.RefreshTokenRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de refresh",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewRefreshTokenCommand(
		req.Token,
		tenantId,
		correlationID,
		clientId,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de refresh",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	response, err := h.refreshTokenUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de refresh",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Refresh realizado com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, response)
}

// LogoutHandler trata requisições de logout
// @Summary Logout
// @Description Realiza logout do usuário invalidando o token de acesso. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id, e token no header Authorization.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param Authorization header string true "Token Bearer para autenticação"
// @Success 200 {object} map[string]string "Logout realizado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes)"
// @Failure 401 {object} pkg.ErrorResponse "Token de autorização inválido ou ausente"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/logout [post]
func (h *AuthHandlers) LogoutHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Extrair token do header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Token de autorização não fornecido",
			TraceID: correlationID,
		})
	}

	// Remover "Bearer " do token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Token inválido",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de logout",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	err = h.logoutUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de logout",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Logout realizado com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Logout realizado com sucesso"})
}

// HeaderError representa um erro de header obrigatório
type HeaderError struct {
	Message string
}

func (e *HeaderError) Error() string {
	return e.Message
}

// GetCorrelationID obtém o correlation ID do header (obrigatório)
func GetCorrelationID(c echo.Context) (string, error) {
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	if correlationID == "" {
		return "", &HeaderError{
			Message: "Header X-Correlation-ID é obrigatório",
		}
	}

	// Define o header na resposta para o frontend
	c.Response().Header().Set("X-Correlation-ID", correlationID)

	return correlationID, nil
}

// GetTenantAndClientId obtém o application ID e client ID do header
func GetTenantAndClientId(c echo.Context) (string, string, error) {
	tenantId := c.Request().Header.Get("X-Tenant-Id")
	clientId := c.Request().Header.Get("X-Client-ID")
	if clientId == "" {
		clientId = "keepguard-default-client"
	}
	if tenantId == "" {
		return "", "", &HeaderError{
			Message: "Header X-Tenant-Id é obrigatório",
		}
	}

	// Define o header na resposta para o frontend
	c.Response().Header().Set("X-Tenant-Id", tenantId)

	return tenantId, clientId, nil
}

// ValidateTokenHandler trata requisições de validação de token
// @Summary Validate token
// @Description Valida se um token JWT é válido e não expirou. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.ValidateTokenRequestDTO true "Token para validação"
// @Success 200 {object} map[string]string "Token válido"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 401 {object} pkg.ErrorResponse "Token inválido ou expirado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/validate [post]
func (h *AuthHandlers) ValidateTokenHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	var req dto.ValidateTokenRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de validação de token",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewValidateTokenCommand(
		req.Token,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de validação de token",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	err = h.validateTokenUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de validação de token",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Token validado com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Token válido"})
}

// ChangePasswordHandler trata requisições de alteração de senha
// @Summary Change password
// @Description Altera a senha do usuário autenticado. Requer headers obrigatórios X-Correlation-ID, X-Tenant-Id e Authorization.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param Authorization header string true "Token Bearer para autenticação"
// @Param request body dto.ChangePasswordRequestDTO true "Dados para alteração de senha"
// @Success 200 {object} map[string]string "Senha alterada com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes, dados inválidos ou senha atual incorreta)"
// @Failure 401 {object} pkg.ErrorResponse "Token de autorização inválido ou ausente"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/change-password [post]
func (h *AuthHandlers) ChangePasswordHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Extrair token do header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Token de autorização não fornecido",
			TraceID: correlationID,
		})
	}

	// Remover "Bearer " do token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:   "UNAUTHORIZED",
			Message: "Token inválido",
			TraceID: correlationID,
		})
	}

	var req dto.ChangePasswordRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de alteração de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewChangePasswordCommand(
		token,
		req.CurrentPassword,
		req.NewPassword,
		req.ConfirmNewPassword,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de alteração de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	err = h.changePasswordUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de alteração de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Senha alterada com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Senha alterada com sucesso"})
}

// ResetPasswordHandler trata requisições de reset de senha
// @Summary Reset password
// @Description Reseta a senha do usuário usando um token de reset válido. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string true "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.ResetPasswordRequestDTO true "Dados para reset de senha"
// @Success 200 {object} map[string]string "Senha resetada com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes, dados inválidos, token inválido ou usuário não ativo)"
// @Failure 404 {object} pkg.ErrorResponse "Usuário não encontrado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/reset-password [post]
func (h *AuthHandlers) ResetPasswordHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID, err := GetCorrelationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: "",
		})
	}

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_HEADER",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	var req dto.ResetPasswordRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de reset de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewResetPasswordCommand(
		req.Email,
		req.ResetToken,
		req.NewPassword,
		req.ConfirmNewPassword,
		tenantId,
		correlationID,
		c.Request().Context(),
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de reset de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: err.Error(),
			TraceID: correlationID,
		})
	}

	// Executar caso de uso com comando encapsulado
	err = h.resetPasswordUseCase.Execute(command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de reset de senha",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	h.logger.Info("Senha resetada com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, map[string]string{"message": "Senha resetada com sucesso"})
}

// handleError trata erros de forma padronizada
func handleError(c echo.Context, err error, correlationID string) error {
	// Trata erros do ms-auth (HTTPError)
	if httpErr, ok := err.(*appdto.HTTPError); ok {
		return c.JSON(httpErr.StatusCode, pkg.ErrorResponse{
			Error:   httpErr.ErrorCode,
			Message: httpErr.Message,
			TraceID: correlationID,
		})
	}

	// Trata erros da aplicação (AppError)
	if appErr, ok := err.(*pkg.AppError); ok {
		return c.JSON(appErr.StatusCode, appErr.WithTraceID(correlationID).ToResponse())
	}

	// Erro genérico
	return c.JSON(http.StatusInternalServerError, pkg.ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "Erro interno do servidor",
		TraceID: correlationID,
	})
}
