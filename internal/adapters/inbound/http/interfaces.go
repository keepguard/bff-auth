package http

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// Handler define a interface para handlers HTTP
type Handler interface {
	LoginHandler(c echo.Context) error
	RefreshHandler(c echo.Context) error
	LogoutHandler(c echo.Context) error
	ValidateTokenHandler(c echo.Context) error
	ChangePasswordHandler(c echo.Context) error
	ResetPasswordHandler(c echo.Context) error
	SendResetPasswordMessageHandler(c echo.Context) error
	SendDeviceChallengeHandler(c echo.Context) error
	VerifyDeviceChallengeHandler(c echo.Context) error
	ListUserSessionsHandler(c echo.Context) error
	RevokeSessionHandler(c echo.Context) error
	RevokeAllOtherSessionsHandler(c echo.Context) error
}

// Middleware define a interface para middlewares HTTP
type Middleware interface {
	RequestIDMiddleware() echo.MiddlewareFunc
	LoggingMiddleware() echo.MiddlewareFunc
	RecoveryMiddleware() echo.MiddlewareFunc
	CORSMiddleware() echo.MiddlewareFunc
	SecurityMiddleware() echo.MiddlewareFunc
	MetricsMiddleware() echo.MiddlewareFunc
	TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc
}

// Server define a interface para o servidor HTTP
type Server interface {
	Start() error
	Stop(ctx context.Context) error
	SetupRoutes(handlers Handler)
}

// Logger define a interface para logging
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}
