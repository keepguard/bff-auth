package http

import (
	"context"
	"net/http"
	"time"

	middlewarePkg "github.com/keepguard/bff-auth/internal/adapters/inbound/http/middleware"
	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/keepguard/bff-auth/internal/infrastructure/validation"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
)

// serverImpl representa o servidor HTTP
type serverImpl struct {
	echo        *echo.Echo
	config      *config.Config
	logger      logger.Logger
	metrics     *metrics.Metrics
	rateLimiter *middlewarePkg.RateLimiterMiddleware
}

// NewServer cria um novo servidor HTTP
func NewServer(
	config *config.Config,
	logger logger.Logger,
	metrics *metrics.Metrics,
	rateLimiter *middlewarePkg.RateLimiterMiddleware,
) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares
	zapLogger, _ := zap.NewDevelopment()
	middlewareInstance := middlewarePkg.NewMiddlewareWithMetrics(zapLogger, metrics)
	validator := validation.NewValidator()

	e.Use(middlewareInstance.RequestIDMiddleware())
	e.Use(middlewareInstance.RecoveryMiddleware())
	e.Use(middlewareInstance.LoggingMiddleware())
	e.Use(middlewarePkg.ValidationMiddleware(validator))
	e.Use(middlewareInstance.CORSMiddleware())
	e.Use(middlewareInstance.SecurityMiddleware())
	e.Use(middlewareInstance.MetricsMiddleware())
	e.Use(middlewareInstance.TimeoutMiddleware(30 * time.Second))

	return &serverImpl{
		echo:        e,
		config:      config,
		logger:      logger,
		metrics:     metrics,
		rateLimiter: rateLimiter,
	}
}

// Start inicia o servidor
func (s *serverImpl) Start() error {
	return s.echo.Start(":" + s.config.Server.Port)
}

// Stop para o servidor
func (s *serverImpl) Stop(ctx context.Context) error {
	return s.echo.Shutdown(ctx)
}

// SetupRoutes configura as rotas da API com proteção de Rate Limit
func (s *serverImpl) SetupRoutes(handlers Handler) {
	// Health check
	s.echo.GET("/health", s.HealthHandler)

	// Swagger
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	rl := s.rateLimiter
	rules := s.config.RateLimit.Rules

	// Auth routes
	authGroup := s.echo.Group("/api/v1/auth")
	authGroup.POST("/login", handlers.LoginHandler, rl.Limit("login", rules.Login))
	authGroup.POST("/refresh", handlers.RefreshHandler, rl.Limit("refresh", rules.Refresh))
	authGroup.POST("/logout", handlers.LogoutHandler, rl.Limit("logout", rules.Logout))
	authGroup.POST("/validate", handlers.ValidateTokenHandler, rl.Limit("validate", rules.Validate))
	authGroup.POST("/change-password", handlers.ChangePasswordHandler, rl.Limit("change_password", rules.ChangePassword))
	authGroup.POST("/reset-password", handlers.ResetPasswordHandler, rl.Limit("reset_password", rules.ResetPassword))
	authGroup.POST("/forgot-password", handlers.SendResetPasswordMessageHandler, rl.Limit("forgot_password", rules.ForgotPassword))
	authGroup.POST("/device/challenge/send", handlers.SendDeviceChallengeHandler, rl.Limit("device_challenge_send", rules.DeviceChallengeSend))
	authGroup.POST("/device/challenge/verify", handlers.VerifyDeviceChallengeHandler, rl.Limit("device_challenge_verify", rules.DeviceChallengeVerify))
	authGroup.POST("/device/quick-revoke", handlers.QuickRevokeHandler, rl.Limit("device_quick_revoke", rules.DeviceQuickRevoke))
	authGroup.GET("/device/quick-revoke", handlers.QuickRevokeHandler, rl.Limit("device_quick_revoke", rules.DeviceQuickRevoke))

	// User Sessions & Device Blacklist routes
	userGroup := s.echo.Group("/api/v1/users/me")
	userGroup.GET("/sessions", handlers.ListUserSessionsHandler)
	userGroup.DELETE("/sessions/:deviceId", handlers.RevokeSessionHandler)
	userGroup.DELETE("/sessions", handlers.RevokeAllOtherSessionsHandler)
	userGroup.GET("/devices/blacklist", handlers.ListDeviceBlacklistHandler)
	userGroup.POST("/devices/blacklist", handlers.AddDeviceBlacklistHandler, rl.Limit("device_blacklist", rules.DeviceBlacklist))
	userGroup.DELETE("/devices/blacklist/:deviceId", handlers.RemoveDeviceBlacklistHandler)

	// Admin Device Blacklist routes (ROLE_ADMIN, ROLE_MANAGER)
	adminGroup := s.echo.Group("/api/v1/admin/devices/blacklist")
	adminGroup.GET("", handlers.SearchAdminDeviceBlacklistHandler)
	adminGroup.POST("", handlers.AdminAddDeviceBlacklistHandler, rl.Limit("admin_device_blacklist", rules.DeviceBlacklist))
	adminGroup.DELETE("/:deviceId", handlers.AdminRemoveDeviceBlacklistHandler)

	s.logger.Info("Rotas configuradas com sucesso com proteção de Rate Limit",
		zap.String("port", s.config.Server.Port),
	)
}

// HealthHandler verifica o status do serviço
// @Summary Health check
// @Description Verifica se o serviço está funcionando corretamente
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string "Serviço funcionando"
// @Router /health [get]
func (s *serverImpl) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "bff-auth",
		"version": "1.0.0",
	})
}
