package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	"github.com/keepguard/bff-auth/internal/adapters/inbound/http/mapper"
	"github.com/keepguard/bff-auth/internal/application/auth"
	"github.com/keepguard/bff-auth/internal/application/blacklist"
	"github.com/keepguard/bff-auth/internal/application/device"
	appdto "github.com/keepguard/bff-auth/internal/application/dto"
	"github.com/keepguard/bff-auth/internal/application/lifecycle"
	"github.com/keepguard/bff-auth/internal/application/session"
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
	devicePort            device.DevicePort
	sessionPort           session.SessionPort
	blacklistPort         blacklist.BlacklistPort
	lifecyclePort         lifecycle.LifecyclePort
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
	devicePort device.DevicePort,
	sessionPort session.SessionPort,
	blacklistPort blacklist.BlacklistPort,
	lifecyclePort lifecycle.LifecyclePort,
	logger *zap.Logger,
) *AuthHandlers {
	return &AuthHandlers{
		loginUseCase:          loginUseCase,
		refreshTokenUseCase:   refreshTokenUseCase,
		logoutUseCase:         logoutUseCase,
		validateTokenUseCase:  validateTokenUseCase,
		changePasswordUseCase: changePasswordUseCase,
		resetPasswordUseCase:  resetPasswordUseCase,
		devicePort:            devicePort,
		sessionPort:           sessionPort,
		blacklistPort:         blacklistPort,
		lifecyclePort:         lifecyclePort,
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
	devicePort device.DevicePort,
	sessionPort session.SessionPort,
	blacklistPort blacklist.BlacklistPort,
	lifecyclePort lifecycle.LifecyclePort,
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
		devicePort:            devicePort,
		sessionPort:           sessionPort,
		blacklistPort:         blacklistPort,
		lifecyclePort:         lifecyclePort,
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

const (
	RefreshTokenCookieName = "keepguard_refresh_token"
	DefaultCookieMaxAge    = 7 * 24 * 3600 // 7 dias em segundos
)

func (h *AuthHandlers) getCookieConfig(c echo.Context) (domain string, sameSite http.SameSite, secure bool, path string) {
	path = "/api/v1/auth"
	if p := os.Getenv("BFF_AUTH_COOKIE_PATH"); p != "" {
		path = p
	}

	// 1. Resolver Domain
	if d := os.Getenv("BFF_AUTH_COOKIE_DOMAIN"); d != "" {
		domain = d
	} else {
		host := c.Request().Host
		origin := c.Request().Header.Get("Origin")
		if strings.Contains(host, "keepguard.com.br") || strings.Contains(origin, "keepguard.com.br") {
			domain = ".keepguard.com.br"
		}
	}

	// 2. Resolver Secure
	secure = c.IsTLS() || c.Request().Header.Get("X-Forwarded-Proto") == "https" || os.Getenv("BFF_AUTH_COOKIE_SECURE") == "true"
	if os.Getenv("BFF_AUTH_COOKIE_SECURE") == "false" {
		secure = false
	}

	// 3. Resolver SameSite
	sameSiteStr := strings.ToLower(strings.TrimSpace(os.Getenv("BFF_AUTH_COOKIE_SAMESITE")))
	switch sameSiteStr {
	case "none":
		sameSite = http.SameSiteNoneMode
		secure = true // Browsers exigem Secure=true para SameSite=None
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "lax":
		sameSite = http.SameSiteLaxMode
	default:
		// Em produção HTTPS keepguard.com.br, SameSite=None é obrigatório para que
		// requisições cross-subdomain (ex: app-core.keepguard.com.br -> api.keepguard.com.br)
		// enviem cookies HttpOnly com credentials: 'include'.
		if secure && (strings.Contains(c.Request().Host, "keepguard.com.br") || strings.Contains(c.Request().Header.Get("Origin"), "keepguard.com.br") || domain == ".keepguard.com.br") {
			sameSite = http.SameSiteNoneMode
			secure = true
		} else {
			sameSite = http.SameSiteLaxMode
		}
	}

	return domain, sameSite, secure, path
}

func (h *AuthHandlers) setRefreshTokenCookie(c echo.Context, token string) {
	if token == "" {
		return
	}
	domain, sameSite, secure, path := h.getCookieConfig(c)

	cookie := &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     path,
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   DefaultCookieMaxAge,
	}

	c.SetCookie(cookie)
}

func (h *AuthHandlers) clearRefreshTokenCookie(c echo.Context) {
	domain, sameSite, secure, path := h.getCookieConfig(c)

	cookie := &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     path,
		Domain:   domain,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}

	c.SetCookie(cookie)
}

// LoginHandler trata requisições de login
// @Summary Login
// @Description Realiza login do usuário usando credenciais de username e password. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
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
	correlationID := GetCorrelationID(c)

	tenantId, clientId, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
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
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
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
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de login",
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
	response, err := h.loginUseCase.Execute(withClientNetwork(c), command)
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

	if response.Token != "" {
		h.setRefreshTokenCookie(c, response.Token)
	}

	return c.JSON(http.StatusOK, mapper.ToAuthResponse(response))
}

// RefreshHandler trata requisições de refresh de token
// @Summary Refresh token
// @Description Renova o token de acesso usando o refresh token. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.RefreshTokenRequestDTO true "Token de refresh para renovação"
// @Success 200 {object} dto.RefreshTokenResponseDTO "Token renovado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 401 {object} pkg.ErrorResponse "Refresh token inválido ou expirado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/refresh [post]
func (h *AuthHandlers) RefreshHandler(c echo.Context) error {
	// Obter correlation ID e application ID (obrigatórios)
	correlationID := GetCorrelationID(c)

	tenantId, clientId, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	var req dto.RefreshTokenRequestDTO
	_ = c.Bind(&req)

	tokenToUse := strings.TrimSpace(req.Token)
	if tokenToUse == "" {
		if cookie, err := c.Cookie(RefreshTokenCookieName); err == nil && cookie != nil && cookie.Value != "" {
			tokenToUse = strings.TrimSpace(cookie.Value)
		}
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewRefreshTokenCommand(
		tokenToUse,
		tenantId,
		correlationID,
		clientId,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de refresh",
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
	response, err := h.refreshTokenUseCase.Execute(c.Request().Context(), command)
	if err != nil {
		h.logger.Error("Erro no caso de uso de refresh",
			zap.String("correlationId", correlationID),
			zap.String("applicationId", tenantId),
			zap.Error(err),
		)
		return handleError(c, err, correlationID)
	}

	// Rotacionar o cookie com novo token gerado
	if response.Token != "" {
		h.setRefreshTokenCookie(c, response.Token)
	}

	h.logger.Info("Refresh realizado com sucesso",
		zap.String("correlationId", correlationID),
		zap.String("applicationId", tenantId),
	)

	return c.JSON(http.StatusOK, mapper.ToRefreshTokenResponse(response))
}

// LogoutHandler trata requisições de logout
// @Summary Logout
// @Description Realiza logout do usuário invalidando o token de acesso. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id, e token no header Authorization.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param Authorization header string true "Token Bearer para autenticação"
// @Success 200 {object} map[string]string "Logout realizado com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes)"
// @Failure 401 {object} pkg.ErrorResponse "Token de autorização inválido ou ausente"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/logout [post]
func (h *AuthHandlers) LogoutHandler(c echo.Context) error {
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

	// Limpa o cookie HttpOnly no logout
	h.clearRefreshTokenCookie(c)

	// Extrair token do header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token de autorização não fornecido",
			CorrelationID: correlationID,
		})
	}

	// Remover "Bearer " do token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token inválido",
			CorrelationID: correlationID,
		})
	}

	// Criar comando de domínio encapsulado
	command := appdto.NewLogoutCommand(
		token,
		tenantId,
		correlationID,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de logout",
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
	err = h.logoutUseCase.Execute(c.Request().Context(), command)
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

// GetCorrelationID lê ou gera UUID e ecoa no header X-Correlation-ID.
func GetCorrelationID(c echo.Context) string {
	correlationID := strings.TrimSpace(c.Request().Header.Get("X-Correlation-ID"))
	if correlationID == "" {
		correlationID = strings.TrimSpace(c.Response().Header().Get("X-Correlation-ID"))
	}
	if correlationID == "" {
		correlationID = generateCorrelationUUID()
	}
	c.Request().Header.Set("X-Correlation-ID", correlationID)
	c.Response().Header().Set("X-Correlation-ID", correlationID)
	return correlationID
}

func generateCorrelationUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// GetTenantAndClientId obtém o tenant (JWT autenticado ou header) e o client ID.
func GetTenantAndClientId(c echo.Context) (string, string, error) {
	headerTenant := strings.TrimSpace(c.Request().Header.Get("X-Tenant-Id"))
	clientId := c.Request().Header.Get("X-Client-ID")
	if clientId == "" {
		clientId = "keepguard-default-client"
	}

	jwtTenant := tenantIdFromAuthorization(c)
	if jwtTenant != "" && headerTenant != "" && jwtTenant != headerTenant {
		return "", "", &HeaderError{
			Message: "X-Tenant-Id não corresponde ao tenant_id do token",
		}
	}

	tenantId := jwtTenant
	if tenantId == "" {
		tenantId = headerTenant
	}
	if tenantId == "" {
		return "", "", &HeaderError{
			Message: "Header X-Tenant-Id é obrigatório",
		}
	}

	c.Response().Header().Set("X-Tenant-Id", tenantId)

	return tenantId, clientId, nil
}

func tenantIdFromAuthorization(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	tenantId, err := pkg.ExtractTenantIdFromToken(authHeader)
	if err != nil {
		return ""
	}
	return tenantId
}

// ValidateTokenHandler trata requisições de validação de token
// @Summary Validate token
// @Description Valida se um token JWT está ativo no Redis (tokenlogin) e no ms-auth. Requer headers obrigatórios X-Correlation-ID e X-Tenant-Id.
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.ValidateTokenRequestDTO true "Token para validação"
// @Success 200 {object} map[string]string "Token válido"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes ou dados inválidos)"
// @Failure 401 {object} pkg.ErrorResponse "Token inválido ou expirado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/validate [post]
func (h *AuthHandlers) ValidateTokenHandler(c echo.Context) error {
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

	var req dto.ValidateTokenRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de validação de token",
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
	command := appdto.NewValidateTokenCommand(
		req.Token,
		tenantId,
		correlationID,
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de validação de token",
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
	err = h.validateTokenUseCase.Execute(c.Request().Context(), command)
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
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
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
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	// Extrair token do header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token de autorização não fornecido",
			CorrelationID: correlationID,
		})
	}

	// Remover "Bearer " do token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
			Error:         "UNAUTHORIZED",
			Message:       "Token inválido",
			CorrelationID: correlationID,
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
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
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
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de alteração de senha",
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
	err = h.changePasswordUseCase.Execute(withClientNetwork(c), command)
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
// @Param X-Correlation-ID header string false "ID de correlação para rastreamento da requisição"
// @Param X-Tenant-Id header string true "ID da aplicação cliente (UUID)"
// @Param request body dto.ResetPasswordRequestDTO true "Dados para reset de senha"
// @Success 200 {object} map[string]string "Senha resetada com sucesso"
// @Failure 400 {object} pkg.ErrorResponse "Erro de validação (headers ausentes, dados inválidos, token inválido ou usuário não ativo)"
// @Failure 404 {object} pkg.ErrorResponse "Usuário não encontrado"
// @Failure 500 {object} pkg.ErrorResponse "Erro interno do servidor"
// @Router /auth/reset-password [post]
func (h *AuthHandlers) ResetPasswordHandler(c echo.Context) error {
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

	var req dto.ResetPasswordRequestDTO
	if err := c.Bind(&req); err != nil {
		h.logger.Error("Erro ao fazer bind da requisição de reset de senha",
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
	)

	// Validar comando
	if err := command.Validate(); err != nil {
		h.logger.Error("Erro de validação no comando de reset de senha",
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
	err = h.resetPasswordUseCase.Execute(withClientNetwork(c), command)
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
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	var req dto.DeviceChallengeSendRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	res, err := h.devicePort.SendChallenge(c.Request().Context(), mapper.ToSendDeviceChallengeCommand(req, tenantId, correlationID))
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, res)
}

// VerifyDeviceChallengeHandler valida OTP de dispositivo e emite JWT
func (h *AuthHandlers) VerifyDeviceChallengeHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	var req dto.DeviceChallengeVerifyRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	res, err := h.devicePort.VerifyChallenge(c.Request().Context(), mapper.ToVerifyDeviceChallengeCommand(req, tenantId, correlationID))
	if err != nil {
		return handleError(c, err, correlationID)
	}

	if res.Token != "" {
		h.setRefreshTokenCookie(c, res.Token)
	}

	return c.JSON(http.StatusOK, mapper.ToAuthResponse(res))
}

// ListUserSessionsHandler lista sessões do usuário autenticado
func (h *AuthHandlers) ListUserSessionsHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Request().Header.Get("X-Device-Id")
	ctx := withClientNetwork(c)

	sessions, err := h.sessionPort.ListMe(ctx, appdto.ListUserSessionsQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      deviceId,
	})
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

func (h *AuthHandlers) accountLifecycleCommand(c echo.Context) (appdto.AccountLifecycleCommand, error) {
	correlationID, tenantId, token, err := h.accountLifecycleHeaders(c)
	if err != nil {
		return appdto.AccountLifecycleCommand{CorrelationID: correlationID}, err
	}

	reason, err := h.bindAccountLifecycleReason(c)
	if err != nil {
		return appdto.AccountLifecycleCommand{CorrelationID: correlationID}, err
	}

	if strings.TrimSpace(token) == "" {
		return appdto.AccountLifecycleCommand{CorrelationID: correlationID}, pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}

	codeUser, err := pkg.ExtractCodeUserFromToken(token)
	if err != nil || codeUser == "" {
		return appdto.AccountLifecycleCommand{CorrelationID: correlationID}, pkg.NewAppError("UNAUTHORIZED", "Token JWT sem identificador de usuário", http.StatusUnauthorized)
	}

	return appdto.AccountLifecycleCommand{
		TenantID:      tenantId,
		Token:         token,
		CodeUser:      codeUser,
		Reason:        reason,
		CorrelationID: correlationID,
	}, nil
}

func (h *AuthHandlers) accountLifecycleHeaders(c echo.Context) (correlationID, tenantId, token string, err error) {
	correlationID = GetCorrelationID(c)

	tenantId, _, err = GetTenantAndClientId(c)
	if err != nil {
		return "", "", "", pkg.NewAppError("MISSING_HEADER", err.Error(), http.StatusBadRequest)
	}

	token = strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
	return correlationID, tenantId, token, nil
}

// BlockMeHandler bloqueia a própria conta do usuário autenticado.
func (h *AuthHandlers) BlockMeHandler(c echo.Context) error {
	cmd, err := h.accountLifecycleCommand(c)
	if err != nil {
		return handleError(c, err, cmd.CorrelationID)
	}

	if err := h.lifecyclePort.BlockMe(c.Request().Context(), cmd); err != nil {
		return handleError(c, err, cmd.CorrelationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeleteMeHandler exclui a própria conta do usuário autenticado.
func (h *AuthHandlers) DeleteMeHandler(c echo.Context) error {
	cmd, err := h.accountLifecycleCommand(c)
	if err != nil {
		return handleError(c, err, cmd.CorrelationID)
	}

	if err := h.lifecyclePort.DeleteMe(c.Request().Context(), cmd); err != nil {
		return handleError(c, err, cmd.CorrelationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RevokeSessionHandler revoga sessão de dispositivo específico
func (h *AuthHandlers) RevokeSessionHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceIdToRevoke := c.Param("deviceId")

	if err := h.sessionPort.Revoke(c.Request().Context(), appdto.RevokeSessionCommand{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      deviceIdToRevoke,
	}); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RevokeAllOtherSessionsHandler revoga todas as outras sessões exceto atual
func (h *AuthHandlers) RevokeAllOtherSessionsHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	currentDeviceId := c.Request().Header.Get("X-Device-Id")

	if err := h.sessionPort.RevokeOthers(c.Request().Context(), appdto.RevokeAllOtherSessionsCommand{
		TenantID:        tenantId,
		CorrelationID:   correlationID,
		Token:           token,
		CurrentDeviceID: currentDeviceId,
	}); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// QuickRevokeHandler revoga dispositivo através do link de e-mail (com opção de blacklist)
func (h *AuthHandlers) QuickRevokeHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId := c.Request().Header.Get("X-Tenant-Id")
	if tenantId == "" {
		tenantId = "keepguard-default"
	}

	token := c.QueryParam("token")
	if token == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_PARAM",
			Message:       "Parâmetro 'token' é obrigatório",
			CorrelationID: correlationID,
		})
	}

	blacklistParam := c.QueryParam("blacklist")
	blacklist := blacklistParam != "false"

	view, err := h.devicePort.QuickRevoke(c.Request().Context(), mapper.ToQuickRevokeCommand(token, tenantId, correlationID, blacklist))
	if err != nil {
		return handleError(c, err, correlationID)
	}

	if strings.Contains(c.Request().Header.Get("Accept"), "text/html") {
		return c.HTML(http.StatusOK, renderQuickRevokeHTML(view.Message))
	}

	return c.JSON(http.StatusOK, mapper.QuickRevokeJSON(view))
}

// ListDeviceBlacklistHandler lista dispositivos bloqueados na blacklist
func (h *AuthHandlers) ListDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	blacklist, err := h.blacklistPort.ListMe(c.Request().Context(), appdto.ListDeviceBlacklistQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, blacklist)
}

// AddDeviceBlacklistHandler adiciona um dispositivo à blacklist
func (h *AuthHandlers) AddDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	var req dto.AddDeviceBlacklistRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	if err := h.blacklistPort.AddMe(c.Request().Context(), mapper.ToAddDeviceBlacklistCommand(req, token, tenantId, correlationID)); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// RemoveDeviceBlacklistHandler remove um dispositivo da blacklist
func (h *AuthHandlers) RemoveDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Param("deviceId")

	if err := h.blacklistPort.RemoveMe(c.Request().Context(), appdto.RemoveDeviceBlacklistCommand{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      deviceId,
	}); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchAdminDeviceBlacklistHandler consulta blacklist com filtros paginados (Admin)
func (h *AuthHandlers) SearchAdminDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
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

	blacklist, err := h.blacklistPort.SearchAdmin(c.Request().Context(), appdto.SearchAdminDeviceBlacklistQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		QueryParams:   queryParams,
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}

	return c.JSON(http.StatusOK, blacklist)
}

// AdminAddDeviceBlacklistHandler adiciona dispositivo à blacklist (Admin)
func (h *AuthHandlers) AdminAddDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")

	var req dto.AdminAddDeviceBlacklistRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}

	if err := h.blacklistPort.AdminAdd(c.Request().Context(), mapper.ToAdminAddDeviceBlacklistCommand(req, token, tenantId, correlationID)); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

// AdminRemoveDeviceBlacklistHandler remove dispositivo da blacklist (Admin)
func (h *AuthHandlers) AdminRemoveDeviceBlacklistHandler(c echo.Context) error {
	correlationID := GetCorrelationID(c)

	tenantId, _, err := GetTenantAndClientId(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_HEADER",
			Message:       err.Error(),
			CorrelationID: correlationID,
		})
	}

	token := strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer ")
	deviceId := c.Param("deviceId")
	userId := c.QueryParam("userId")

	if userId == "" {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "MISSING_PARAM",
			Message:       "Parâmetro 'userId' é obrigatório",
			CorrelationID: correlationID,
		})
	}

	if err := h.blacklistPort.AdminRemove(c.Request().Context(), appdto.AdminRemoveDeviceBlacklistCommand{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      deviceId,
		UserID:        userId,
	}); err != nil {
		return handleError(c, err, correlationID)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandlers) tenantAuthHeaders(c echo.Context) (correlationID, tenantId, token string, err error) {
	correlationID = GetCorrelationID(c)
	tenantId, _, err = GetTenantAndClientId(c)
	if err != nil {
		return "", "", "", pkg.NewAppError("MISSING_HEADER", err.Error(), http.StatusBadRequest)
	}
	token = strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
	if token == "" {
		return correlationID, tenantId, "", pkg.NewAppError("UNAUTHORIZED", "Token de autorização não fornecido ou inválido", http.StatusUnauthorized)
	}
	return correlationID, tenantId, token, nil
}

func (h *AuthHandlers) ListTenantUserSessionsHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	userId := c.Param("userId")
	sessions, err := h.sessionPort.ListTenantUser(c.Request().Context(), appdto.ListTenantUserSessionsQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        userId,
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, sessions)
}

func (h *AuthHandlers) RevokeTenantUserSessionHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	if err := h.sessionPort.RevokeTenantUser(c.Request().Context(), appdto.RevokeTenantUserSessionCommand{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        c.Param("userId"),
		DeviceID:      c.Param("deviceId"),
	}); err != nil {
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandlers) ListTenantUserBlacklistHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	blacklist, err := h.blacklistPort.ListTenantUser(c.Request().Context(), appdto.ListTenantUserBlacklistQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		UserID:        c.Param("userId"),
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, blacklist)
}

func (h *AuthHandlers) AddTenantUserBlacklistHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	var req dto.AddDeviceBlacklistRequestDTO
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, pkg.ErrorResponse{
			Error:         "INVALID_REQUEST",
			Message:       "Requisição inválida",
			CorrelationID: correlationID,
		})
	}
	if err := h.blacklistPort.AdminAdd(c.Request().Context(), mapper.ToTenantAddDeviceBlacklistCommand(req, c.Param("userId"), token, tenantId, correlationID)); err != nil {
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandlers) RemoveTenantUserBlacklistHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	if err := h.blacklistPort.AdminRemove(c.Request().Context(), appdto.AdminRemoveDeviceBlacklistCommand{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		DeviceID:      c.Param("deviceId"),
		UserID:        c.Param("userId"),
	}); err != nil {
		return handleError(c, err, correlationID)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandlers) SearchTenantSessionsHandler(c echo.Context) error {
	correlationID, tenantId, token, err := h.tenantAuthHeaders(c)
	if err != nil {
		return handleError(c, err, correlationID)
	}
	queryParams := map[string]string{
		"userId":   c.QueryParam("userId"),
		"deviceId": c.QueryParam("deviceId"),
		"page":     c.QueryParam("page"),
		"size":     c.QueryParam("size"),
		"sort":     c.QueryParam("sort"),
	}
	sessions, err := h.sessionPort.SearchTenant(c.Request().Context(), appdto.SearchTenantSessionsQuery{
		TenantID:      tenantId,
		CorrelationID: correlationID,
		Token:         token,
		QueryParams:   queryParams,
	})
	if err != nil {
		return handleError(c, err, correlationID)
	}
	return c.JSON(http.StatusOK, sessions)
}

// handleError trata erros de forma padronizada
func handleError(c echo.Context, err error, correlationID string) error {
	statusCode := http.StatusInternalServerError
	errorCode := "INTERNAL_ERROR"
	msg := "Erro interno do servidor"

	if err != nil && (strings.Contains(err.Error(), "circuit breaker is open") || strings.Contains(err.Error(), "circuit breaker is half-open")) {
		return c.JSON(http.StatusServiceUnavailable, pkg.ErrorResponse{
			Error:         "SERVICE_TEMPORARILY_UNAVAILABLE",
			Message:       "O serviço está temporariamente indisponível. Por favor, tente novamente em instantes.",
			CorrelationID: correlationID,
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
		Error:         errorCode,
		Message:       msg,
		CorrelationID: correlationID,
	})
}
