package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/cache"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenRevocationMiddleware verifica no Redis se o Bearer token ainda existe e se o dispositivo não está na blacklist.
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

			correlationID := GetCorrelationID(c)

			presence, codeUser, err := cache.LookupLoginToken(c.Request().Context(), redisClient, token)
			if err != nil {
				logger.Warn("Falha ao consultar existência de token no Redis (Bypass ativado)",
					zap.String("codeUser", codeUser),
					zap.Error(err),
				)
				return next(c)
			}
			if presence == client.LoginTokenAbsent {
				logger.Warn("Requisição bloqueada na borda pelo BFF: Token revogado ou inexistente no Redis",
					zap.String("codeUser", codeUser),
					zap.String("path", c.Path()),
				)
				return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
					Error:         "TOKEN_REVOKED",
					Message:       "Sessão revogada ou expirada. Por favor, realize login novamente.",
					CorrelationID: correlationID,
				})
			}

			// Checagem de Blacklist de Dispositivo
			if redisClient != nil && codeUser != "" {
				claims, _ := pkg.ExtractAllClaims(token)
				deviceId := ""
				if claims != nil {
					deviceId = claims.DeviceID
				}
				if deviceId == "" {
					deviceId = c.Request().Header.Get("X-Device-Id")
				}
				if deviceId != "" {
					blacklistKey := fmt.Sprintf("device:blacklist:%s:%s", codeUser, deviceId)
					ctxB, cancelB := context.WithTimeout(c.Request().Context(), 500*time.Millisecond)
					blacklisted, errB := redisClient.Exists(ctxB, blacklistKey).Result()
					cancelB()
					if errB == nil && blacklisted > 0 {
						logger.Warn("Requisição bloqueada na borda pelo BFF: Dispositivo bloqueado na Blacklist",
							zap.String("codeUser", codeUser),
							zap.String("deviceId", deviceId),
							zap.String("path", c.Path()),
						)
						return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
							Error:         "DEVICE_BLACKLISTED",
							Message:       "Este dispositivo foi bloqueado para acesso a esta conta.",
							CorrelationID: correlationID,
						})
					}
				}
			}

			return next(c)
		}
	}
}

