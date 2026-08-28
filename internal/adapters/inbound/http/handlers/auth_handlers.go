package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-auth/internal/application/auth"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/clientip"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/infrastructure/requestmeta"
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
	authClient            authclient.AuthClient
	companyClient         authclient.CompanyClient
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
	authClient authclient.AuthClient,
	companyClient authclient.CompanyClient,
	logger *zap.Logger,
) *AuthHandlers {
	return &AuthHandlers{
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		validateTokenUseCase:  validateTokenUseCase,
		changePasswordUseCase: changePasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
		authClient:            authClient,
		companyClient:         companyClient,
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
	authClient authclient.AuthClient,
	companyClient authclient.CompanyClient,
	log logger.Logger,
) *AuthHandlers {
	zapLogger, _ := zap.NewDevelopment()
	return &AuthHandlers{
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		validateTokenUseCase:  validateTokenUseCase,
		changePasswordUseCase: changePasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
		authClient:            authClient,
		companyClient:         companyClient,
		logger:                zapLogger,
	}
}

func resolveClientIP(c echo.Context) string {
	if ip := clientip.FromRequest(c.Request()); ip != "" {
		return ip
	}
	return c.RealIP()
}

func withClientNetwork(c echo.Context) context.Context {
	ctx := requestmeta.WithClientIP(c.Request().Context(), resolveClientIP(c))
	return requestmeta.WithClientLocation(ctx, c.Request().Header.Get("X-Public-Location"))
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
// @Failure 429 {object} pkg.ErrorResponse "Muitas tentativas (Rate limit excedido)"
// @Failure 503 {object} pkg.ErrorResponse "Serviço temporariamente indisponível (Circuit breaker)"
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

	// Captura headers opcionais de dispositivo
	deviceId := c.Request().Header.Get("X-Device-Id")
	deviceName := c.Request().Header.Get("X-Device-Name")
	deviceType := c.Request().Header.Get("X-Device-Type")
	ipAddress := resolveClientIP(c)
	userAgent := c.Request().UserAgent()

	// Criar comando de domínio encapsulado com device
	command := appdto.NewLoginCommandWithDevice(
		req.Username,
		req.Password,
		tenantId,
		correlationID,
		clientId,
		deviceId,
		deviceName,
		deviceType,
		ipAddress,
		userAgent,
		withClientNetwork(c),
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

	deviceId := c.Request().Header.Get("X-Device-Id")
	deviceName := c.Request().Header.Get("X-Device-Name")
	deviceType := c.Request().Header.Get("X-Device-Type")
	ipAddress := resolveClientIP(c)
	userAgent := c.Request().Header.Get("User-Agent")

	// Criar comando de domínio encapsulado
	command := appdto.NewChangePasswordCommand(
		token,
		req.CurrentPassword,
		req.NewPassword,
		req.ConfirmNewPassword,
		tenantId,
		correlationID,
		deviceId,
		deviceName,
		deviceType,
		ipAddress,
		userAgent,
		withClientNetwork(c),
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

	deviceId := c.Request().Header.Get("X-Device-Id")
	deviceName := c.Request().Header.Get("X-Device-Name")
	deviceType := c.Request().Header.Get("X-Device-Type")
	ipAddress := resolveClientIP(c)
	userAgent := c.Request().Header.Get("User-Agent")

	// Criar comando de domínio encapsulado
	command := appdto.NewResetPasswordCommand(
		req.Email,
		req.ResetToken,
		req.NewPassword,
		req.ConfirmNewPassword,
		tenantId,
		correlationID,
		deviceId,
		deviceName,
		deviceType,
		ipAddress,
		userAgent,
		withClientNetwork(c),
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

// SendDeviceChallengeHandler dispara OTP para canal selecionado no dispositivo
func (h *AuthHandlers) SendDeviceChallengeHandler(c echo.Context) error {
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

	var req dto.DeviceChallengeSendRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	res, err := h.authClient.SendDeviceChallenge(c.Request().Context(), req, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, res)
}

// VerifyDeviceChallengeHandler valida OTP de dispositivo e emite JWT
func (h *AuthHandlers) VerifyDeviceChallengeHandler(c echo.Context) error {
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

	var req dto.DeviceChallengeVerifyRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	res, err := h.authClient.VerifyDeviceChallenge(c.Request().Context(), req, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, res)
}

// ListUserSessionsHandler lista sessões do usuário autenticado
func (h *AuthHandlers) ListUserSessionsHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Request().Header.Get("X-Device-Id")
	ctx := withClientNetwork(c)

	sessions, err := h.authClient.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, sessions)
}

func (h *AuthHandlers) bindAccountLifecycleReason(c echo.Context) (string, error) {
	var body dto.AccountLifecycleReasonDTO
	if err := c.Bind(&body); err != nil {
		return "", pkg.NewAppError("INVALID_REQUEST", "Requisição inválida", http.StatusBadRequest)
	}
	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		return "", pkg.NewAppError("INVALID_REQUEST", "Motivo é obrigatório", http.StatusBadRequest)
	}
	return reason, nil
}

func (h *AuthHandlers) resolveSelfUserExternalID(c echo.Context, token, tenantId, correlationID string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}

	codeUser, err := pkg.ExtractCodeUserFromToken(token)
	if err != nil || codeUser == "" {
		return "", pkg.NewAppError("UNAUTHORIZED", "Token JWT sem identificador de usuário", http.StatusUnauthorized)
	}

	company, err := h.companyClient.GetByTenantId(c.Request().Context(), tenantId, correlationID)
	if err != nil {
		return "", err
	}
	if company.ID == "" {
		return "", pkg.NewAppError("COMPANY_NOT_FOUND", "Empresa não encontrada para o tenant informado", http.StatusNotFound)
	}

	ctx := authclient.WithCompanyID(c.Request().Context(), company.ID)
	user, err := h.authClient.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
	if err != nil {
		return "", err
	}

	idUserExternal := user.ExternalID()
	if idUserExternal == "" {
		return "", pkg.NewAppError("USER_NOT_FOUND", "Usuário autenticado não encontrado", http.StatusNotFound)
	}
	return idUserExternal, nil
}

func (h *AuthHandlers) accountLifecycleHeaders(c echo.Context) (correlationID, tenantId, token string, err error) {
	correlationID, err = GetCorrelationID(c)
	if err != nil {
		return "", "", "", pkg.NewAppError("MISSING_HEADER", err.Error(), http.StatusBadRequest)
	}

	tenantId, _, err = GetTenantAndClientId(c)
	if err != nil {
		return "", "", "", pkg.NewAppError("MISSING_HEADER", err.Error(), http.StatusBadRequest)
	}

	token = strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
	return correlationID, tenantId, token, nil
}

// BlockMeHandler bloqueia a própria conta do usuário autenticado.
func (h *AuthHandlers) BlockMeHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.accountLifecycleHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	reason, err := h.bindAccountLifecycleReason(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	idUserExternal, err := h.resolveSelfUserExternalID(c, token, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	if err := h.authClient.BlockUser(c.Request().Context(), idUserExternal, reason, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteMeHandler exclui a própria conta do usuário autenticado.
func (h *AuthHandlers) DeleteMeHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.accountLifecycleHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	reason, err := h.bindAccountLifecycleReason(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	idUserExternal, err := h.resolveSelfUserExternalID(c, token, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	if err := h.authClient.DeleteUser(c.Request().Context(), idUserExternal, reason, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RevokeSessionHandler revoga sessão de dispositivo específico
func (h *AuthHandlers) RevokeSessionHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceIdToRevoke := c.Param("deviceId")

	if err := h.authClient.RevokeSession(c.Request().Context(), deviceIdToRevoke, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RevokeAllOtherSessionsHandler revoga todas as outras sessões exceto atual
func (h *AuthHandlers) RevokeAllOtherSessionsHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	currentDeviceId := c.Request().Header.Get("X-Device-Id")

	if err := h.authClient.RevokeAllOtherSessions(c.Request().Context(), token, currentDeviceId, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// QuickRevokeHandler revoga dispositivo através do link de e-mail (com opção de blacklist)
func (h *AuthHandlers) QuickRevokeHandler(c echo.Context) error {
	correlationID := c.Request().Header.Get("X-Correlation-ID")
	if correlationID == "" {
		correlationID = "rev_" + c.Request().Header.Get("CF-Ray")
	}

	tenantId := c.Request().Header.Get("X-Tenant-Id")
	if tenantId == "" {
		tenantId = "keepguard-default"
	}

	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_PARAM",
			Message: "Parâmetro 'token' é obrigatório",
			TraceID: correlationID,
		})
	}

	blacklistParam := c.QueryParam("blacklist")
	blacklist := blacklistParam != "false"

	res, err := h.authClient.QuickRevoke(c.Request().Context(), token, blacklist, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		msg, _ := res["message"].(string)
		if msg == "" {
			msg = "Sessão revogada com sucesso."
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KeepGuard - Dispositivo Revogado</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 25px -5px rgba(0,0,0,0.5); max-width: 480px; text-align: center; border: 1px solid #334155; }
        .icon { font-size: 3.5rem; margin-bottom: 1rem; color: #ef4444; }
        h1 { font-size: 1.5rem; margin-bottom: 0.75rem; color: #ffffff; }
        p { color: #94a3b8; line-height: 1.5; font-size: 0.95rem; margin-bottom: 1.5rem; }
        .badge { display: inline-block; padding: 0.35rem 0.75rem; background: #334155; color: #38bdf8; border-radius: 9999px; font-size: 0.85rem; font-weight: 600; margin-bottom: 1.5rem; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">&#128737;</div>
        <h1>Acesso Revogado com Sucesso</h1>
        <div class="badge">Dispositivo Bloqueado</div>
        <p>%s O acesso desta sessão foi encerrado e este dispositivo foi adicionado à sua lista de bloqueios.</p>
        <p style="font-size:0.85rem; color:#64748b;">Sua conta permanece protegida.</p>
    </div>
</body>
</html>`, msg)
		return c.HTML(http.StatusOK, html)
	}

	return c.JSON(http.StatusOK, res)
}

// ListDeviceBlacklistHandler lista dispositivos bloqueados na blacklist
func (h *AuthHandlers) ListDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	blacklist, err := h.authClient.ListDeviceBlacklist(c.Request().Context(), token, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, blacklist)
}

// AddDeviceBlacklistHandler adiciona um dispositivo à blacklist
func (h *AuthHandlers) AddDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	var req dto.AddDeviceBlacklistRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	if err := h.authClient.AddDeviceToBlacklist(c.Request().Context(), req, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RemoveDeviceBlacklistHandler remove um dispositivo da blacklist
func (h *AuthHandlers) RemoveDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Param("deviceId")

	if err := h.authClient.RemoveDeviceFromBlacklist(c.Request().Context(), deviceId, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchAdminDeviceBlacklistHandler consulta blacklist com filtros paginados (Admin)
func (h *AuthHandlers) SearchAdminDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	queryParams := map[string]string{
		"userId":     c.QueryParam("userId"),
		"deviceId":   c.QueryParam("deviceId"),
		"deviceName": c.QueryParam("deviceName"),
		"ipAddress":  c.QueryParam("ipAddress"),
		"startDate":  c.QueryParam("startDate"),
		"endDate":    c.QueryParam("endDate"),
		"page":       c.QueryParam("page"),
		"size":       c.QueryParam("size"),
		"sort":       c.QueryParam("sort"),
	}

	blacklist, err := h.authClient.SearchAdminDeviceBlacklist(c.Request().Context(), queryParams, token, tenantId, correlationID)
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, blacklist)
}

// AdminAddDeviceBlacklistHandler adiciona dispositivo à blacklist (Admin)
func (h *AuthHandlers) AdminAddDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	var req dto.AdminAddDeviceBlacklistRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Requisição inválida",
			TraceID: correlationID,
		})
	}

	if err := h.authClient.AdminAddDeviceToBlacklist(c.Request().Context(), req, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// AdminRemoveDeviceBlacklistHandler remove dispositivo da blacklist (Admin)
func (h *AuthHandlers) AdminRemoveDeviceBlacklistHandler(c echo.Context) error {
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

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Param("deviceId")
	userId := c.QueryParam("userId")

	if userId == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:   "MISSING_PARAM",
			Message: "Parâmetro 'userId' é obrigatório",
			TraceID: correlationID,
		})
	}

	if err := h.authClient.AdminRemoveDeviceFromBlacklist(c.Request().Context(), deviceId, userId, token, tenantId, correlationID); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// handleError trata erros de forma padronizada
func handleError(c echo.Context, err error, correlationID string) error {
	statusCode := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	msg := "Erro interno do servidor"

	if err != nil && (strings.Contains(err.Error(), "circuit breaker is open") || strings.Contains(err.Error(), "circuit breaker is half-open")) {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:   "SERVICE_TEMPORARILY_UNAVAILABLE",
			Message: "O serviço está temporariamente indisponível. Por favor, tente novamente em instantes.",
			TraceID: correlationID,
		})
	}

	if httpErr, ok := err.(*appdto.HTTPError); ok {
		statusCode = httpErr.StatusCode
		errorCode = httpErr.ErrorCode
		if httpErr.Message != "" {
			msg = httpErr.Message
		} else {
			msg = "Erro no serviço de autenticação"
		}
	} else if appErr, ok := err.(*pkg.AppError); ok {
		statusCode = appErr.StatusCode
		errorCode = string(appErr.Code)
		if appErr.Message != "" {
			msg = appErr.Message
		}
	}

	if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KeepGuard - Erro</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .card { background: #1e293b; padding: 2.5rem; border-radius: 12px; box-shadow: 0 10px 25px -5px rgba(0,0,0,0.5); max-width: 480px; text-align: center; border: 1px solid #334155; }
        .icon { font-size: 3.5rem; margin-bottom: 1rem; color: #f59e0b; }
        h1 { font-size: 1.5rem; margin-bottom: 0.75rem; color: #ffffff; }
        p { color: #94a3b8; line-height: 1.5; font-size: 0.95rem; margin-bottom: 1.5rem; }
        .badge { display: inline-block; padding: 0.35rem 0.75rem; background: #334155; color: #f87171; border-radius: 9999px; font-size: 0.85rem; font-weight: 600; margin-bottom: 1.5rem; }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon">&#9888;</div>
        <h1>Não foi possível processar</h1>
        <div class="badge">%s</div>
        <p>%s</p>
    </div>
</body>
</html>`, errorCode, msg)
		return c.HTML(statusCode, html)
	}

	return c.JSON(statusCode, pkg.ErrorResponse{
		Error:   errorCode,
		Message: msg,
		TraceID: correlationID,
	})
}
