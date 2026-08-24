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
	echo    *echo.Echo
	config  *config.Config
	logger  logger.Logger
	metrics *metrics.Metrics
}

// NewServer cria um novo servidor HTTP
func NewServer(
	config *config.Config,
	logger logger.Logger,
	metrics *metrics.Metrics,
) Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middlewares
	middlewareInstance := middlewarePkg.NewMiddlewareWithLogger(logger)
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
		echo:    e,
		config:  config,
		logger:  logger,
		metrics: metrics,
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

// SetupRoutes configura as rotas da API
func (s *serverImpl) SetupRoutes(handlers Handler) {
	// Health check
	s.echo.GET("/health", s.HealthHandler)

	// Swagger
	s.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Auth routes
	authGroup := s.echo.Group("/api/v1/auth")
	authGroup.POST("/login", handlers.LoginHandler)
	authGroup.POST("/refresh", handlers.RefreshHandler)
	authGroup.POST("/logout", handlers.LogoutHandler)
	authGroup.POST("/validate", handlers.ValidateTokenHandler)
	authGroup.POST("/change-password", handlers.ChangePasswordHandler)
	authGroup.POST("/reset-password", handlers.ResetPasswordHandler)
	authGroup.POST("/forgot-password", handlers.SendResetPasswordMessageHandler)
	authGroup.POST("/device/challenge/send", handlers.SendDeviceChallengeHandler)
	authGroup.POST("/device/challenge/verify", handlers.VerifyDeviceChallengeHandler)

	// User Sessions routes
	userGroup := s.echo.Group("/api/v1/users/me")
	userGroup.GET("/sessions", handlers.ListUserSessionsHandler)
	userGroup.DELETE("/sessions/:deviceId", handlers.RevokeSessionHandler)
	userGroup.DELETE("/sessions", handlers.RevokeAllOtherSessionsHandler)

	s.logger.Info("Rotas configuradas com sucesso",
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
