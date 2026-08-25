package http

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/keepguard/bff-auth/internal/infrastructure/logger"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

// middlewareImpl representa os middlewares HTTP
type middlewareImpl struct {
	logger  *zap.Logger
	metrics *metrics.Metrics
}

// NewMiddleware cria um novo middleware
func NewMiddleware(logger *zap.Logger) Middleware {
	return &middlewareImpl{
		logger: logger,
	}
}

// NewMiddlewareWithMetrics cria um novo middleware com suporte a métricas
func NewMiddlewareWithMetrics(logger *zap.Logger, metrics *metrics.Metrics) Middleware {
	return &middlewareImpl{
		logger:  logger,
		metrics: metrics,
	}
}

// NewMiddlewareWithLogger cria um novo middleware usando a interface logger.Logger
func NewMiddlewareWithLogger(log logger.Logger) Middleware {
	zapLogger, _ := zap.NewDevelopment()
	return &middlewareImpl{
		logger: zapLogger,
	}
}

// RequestIDMiddleware adiciona ID de requisição
func (m *middlewareImpl) RequestIDMiddleware() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return generateRequestID()
		},
	})
}

// LoggingMiddleware registra logs das requisições
func (m *middlewareImpl) LoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Extrai informações da requisição
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			userID := c.Request().Header.Get("X-User-ID")
			method := c.Request().Method
			path := c.Path()
			uri := c.Request().RequestURI
			userAgent := c.Request().UserAgent()
			remoteIP := c.RealIP()

			// Executa o handler
			err := next(c)

			// Calcula métricas
			duration := time.Since(start)
			status := c.Response().Status
			responseSize := c.Response().Size

			// Determina o nível do log baseado no status
			var logFunc func(string, ...zap.Field)

			switch {
			case status >= 500:
				logFunc = m.logger.Error
			case status >= 400:
				logFunc = m.logger.Warn
			default:
				logFunc = m.logger.Info
			}

			// Log estruturado otimizado para Grafana/Loki
			standardFields := logger.NewStandardFields("bff-auth", "bff-auth", getEnvOrDefault("ENV", "local"), "1.0.0").
				WithTraceID(requestID).
				WithSpanID(requestID). // Usando requestID como spanId por simplicidade
				WithRequestID(requestID).
				WithUserID(userID)

			httpFields := logger.NewHTTPRequestFields(standardFields).
				WithMethod(method).
				WithPath(path).
				WithURI(uri).
				WithStatus(status).
				WithLatency(duration.Milliseconds(), duration.String()).
				WithUserAgent(userAgent).
				WithIP(remoteIP).
				WithResponseSize(responseSize)

			logFunc("HTTP request completed", httpFields.ToZapFields()...)

			// Log adicional para usuário autenticado
			if userID != "" {
				authFields := standardFields.
					WithTraceID(requestID).
					WithRequestID(requestID).
					WithUserID(userID)

				m.logger.Info("Authenticated request", authFields.ToZapFields()...)
			}

			return err
		}
	}
}

// RecoverMiddleware recupera de panics
func (m *middlewareImpl) RecoveryMiddleware() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			userID := c.Request().Header.Get("X-User-ID")

			m.logger.Error("Panic recovered",
				zap.String("traceId", requestID),
				zap.String("spanId", requestID),
				zap.String("requestId", requestID),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
				zap.String("uri", c.Request().RequestURI),
				zap.String("userAgent", c.Request().UserAgent()),
				zap.String("ip", c.RealIP()),
				zap.String("userId", userID),
				zap.Error(err),
				zap.String("stack", string(stack)),
				zap.String("component", "bff-auth"),
				zap.String("service", "bff-auth"),
				zap.String("environment", getEnvOrDefault("ENV", "local")),
				zap.String("version", "1.0.0"),
			)
			return nil
		},
	})
}

// CORSMiddleware configura CORS
func (m *middlewareImpl) CORSMiddleware() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestID,
			"X-User-ID",
			"X-Tenant-Id",
			"X-Correlation-ID",
			"X-Client-Id",
			"X-Client-ID",
			"X-Device-Id",
			"X-Device-Name",
			"X-Device-Type",
			"Idempotency-Key",
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
			"X-User-ID",
			"X-Correlation-ID",
			"X-Tenant-Id",
			"X-Device-Id",
		},
		MaxAge: 86400, // 24 horas
	})
}

// SecurityMiddleware adiciona headers de segurança
func (m *middlewareImpl) SecurityMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Remove header do servidor
			c.Response().Header().Set(echo.HeaderServer, "")

			// Adiciona headers de segurança
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			return next(c)
		}
	}
}

// MetricsMiddleware registra métricas das requisições
func (m *middlewareImpl) MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			status := c.Response().Status

			if m.metrics != nil {
				path := c.Path()
				if path == "" {
					path = c.Request().URL.Path
				}
				m.metrics.RecordHTTPRequest(c.Request().Method, path, status, duration)
			}

			m.logger.Info("HTTP request completed",
				zap.String("method", c.Request().Method),
				zap.String("path", c.Path()),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.String("traceId", c.Response().Header().Get(echo.HeaderXRequestID)),
			)

			return err
		}
	}
}

// TimeoutMiddleware adiciona timeout às requisições
func (m *middlewareImpl) TimeoutMiddleware(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
	})
}

// generateRequestID gera um ID único para a requisição
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + strconv.FormatInt(time.Now().Unix(), 36)
}

// GetTraceID extrai o trace ID do contexto
func GetTraceID(c echo.Context) string {
	return c.Response().Header().Get(echo.HeaderXRequestID)
}

// GetUserID extrai o user ID do contexto (se disponível)
func GetUserID(c echo.Context) string {
	return c.Request().Header.Get("X-User-ID")
}

// SetUserID define o user ID no contexto
func SetUserID(c echo.Context, userID string) {
	c.Request().Header.Set("X-User-ID", userID)
	c.Response().Header().Set("X-User-ID", userID)
}

// getEnvOrDefault retorna valor da variável de ambiente ou valor padrão
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
