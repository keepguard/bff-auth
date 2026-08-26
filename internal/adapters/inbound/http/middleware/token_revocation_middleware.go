package http

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenRevocationMiddleware verifica no Redis se o Bearer token ainda existe e não foi revogado
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

			// Se o Redis não estiver inicializado ou com erro temporário, permite bypass com warning
			if redisClient == nil {
				return next(c)
			}

			// Decodifica payload do JWT para extrair 'sub' (codeUser)
			codeUser, err := extractCodeUserFromJWT(token)
			if err != nil || codeUser == "" {
				// Se o formato do token for inválido, deixa o ms-auth retornar o erro específico
				return next(c)
			}

			// Chave do Redis usada pelo ms-auth: tokenlogin:{codeUser}:{token}
			redisKey := fmt.Sprintf("tokenlogin:%s:%s", codeUser, token)

			ctx, cancel := context.WithTimeout(c.Request().Context(), 500*time.Millisecond)
			defer cancel()

			exists, err := redisClient.Exists(ctx, redisKey).Result()
			if err != nil {
				// Loga aviso e não bloqueia em caso de falha de conexão temporária com Redis
				logger.Warn("Falha ao consultar existência de token no Redis (Bypass ativado)",
					zap.String("codeUser", codeUser),
					zap.Error(err),
				)
				return next(c)
			}

			if exists == 0 {
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

			return next(c)
		}
	}
}

// extractCodeUserFromJWT decodifica a parte payload do JWT sem validar assinatura (rápido)
func extractCodeUserFromJWT(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("formato de JWT inválido")
	}

	payloadSegment := parts[1]
	// Normaliza base64url para base64 padrão
	if l := len(payloadSegment) % 4; l > 0 {
		payloadSegment += strings.Repeat("=", 4-l)
	}
	payloadBytes, err := base64.URLEncoding.DecodeString(payloadSegment)
	if err != nil {
		payloadBytes, err = base64.StdEncoding.DecodeString(payloadSegment)
		if err != nil {
			return "", err
		}
	}

	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", err
	}

	return claims.Sub, nil
}
