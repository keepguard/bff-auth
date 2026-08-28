package http

import (
	"net/http"
	"strings"

	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/cache"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenRevocationMiddleware verifica no Redis se o Bearer token ainda existe e não foi revogado.
func TokenRevocationMiddleware(redisClient *redis.Client, logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				return next(c)
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return next(c)
			}

			presence, codeUser, err := cache.LookupLoginToken(c.Request().Context(), redisClient, token)
			if err != nil {
				logger.Warn("Falha ao consultar existência de token no Redis (Bypass ativado)",
					zap.String("codeUser", codeUser),
					zap.Error(err),
				)
				return next(c)
			}
			if presence != client.LoginTokenAbsent {
				return next(c)
			}

			logger.Warn("Requisição bloqueada na borda pelo BFF: Token revogado ou inexistente no Redis",
				zap.String("codeUser", codeUser),
				zap.String("path", c.Path()),
			)
			correlationID := c.Response().Header().Get(echo.HeaderXRequestID)
			return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
				Error:   "TOKEN_REVOKED",
				Message: "Sessão revogada ou expirada. Por favor, realize login novamente.",
				TraceID: correlationID,
			})
		}
	}
}
