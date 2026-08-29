package http

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// JWTConfig configuração do middleware JWT
type JWTConfig struct {
	SecretKey               string
	SigningMethod           string
	Claims                  jwt.Claims
	SuccessHandler          echo.HandlerFunc
	ErrorHandler            echo.HandlerFunc
	ErrorHandlerWithContext func(c echo.Context, err error) error
	SigningKey              interface{}
	SigningKeys             map[string]interface{}
	KeyFunc                 jwt.Keyfunc
	TokenLookup             string
	AuthScheme              string
	ContextKey              string
}

// JWTMiddleware implementa middleware JWT
type JWTMiddleware struct {
	config JWTConfig
	logger *zap.Logger
}

// NewJWTMiddleware cria um novo middleware JWT
func NewJWTMiddleware(config JWTConfig, logger *zap.Logger) *JWTMiddleware {
	if config.AuthScheme == "" {
		config.AuthScheme = "Bearer"
	}
	if config.ContextKey == "" {
		config.ContextKey = "user"
	}
	if config.TokenLookup == "" {
		config.TokenLookup = "header:Authorization"
	}

	return &JWTMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware retorna o middleware JWT
func (j *JWTMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extrair token do header
			token, err := j.extractToken(c)
			if err != nil {
				j.logger.Error("Erro ao extrair token JWT",
					zap.String("correlationId", GetCorrelationID(c)),
					zap.Error(err),
				)
				return j.handleError(c, err)
			}

			// Validar token
			claims, err := j.validateToken(token)
			if err != nil {
				j.logger.Error("Token JWT inválido",
					zap.String("correlationId", GetCorrelationID(c)),
					zap.Error(err),
				)
				return j.handleError(c, err)
			}

			// Adicionar claims ao contexto
			c.Set(j.config.ContextKey, claims)

			// Extrair informações do usuário dos claims
			if userID := j.extractUserID(claims); userID != "" {
				SetUserID(c, userID)
			}

			return next(c)
		}
	}
}

// extractToken extrai o token JWT da requisição
func (j *JWTMiddleware) extractToken(c echo.Context) (string, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Token de autorização não fornecido")
	}

	// Verificar se o header começa com o scheme correto
	if !strings.HasPrefix(authHeader, j.config.AuthScheme+" ") {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Formato de token inválido")
	}

	// Extrair o token
	token := strings.TrimPrefix(authHeader, j.config.AuthScheme+" ")
	if token == "" {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Token vazio")
	}

	return token, nil
}

// validateToken valida o token JWT
func (j *JWTMiddleware) validateToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verificar método de assinatura
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "Método de assinatura inválido")
		}

		// Retornar chave de assinatura
		if j.config.KeyFunc != nil {
			return j.config.KeyFunc(token)
		}
		return []byte(j.config.SecretKey), nil
	})

	if err != nil {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Token inválido: "+err.Error())
	}

	if !token.Valid {
		return nil, echo.NewHTTPError(http.StatusUnauthorized, "Token inválido")
	}

	return token.Claims, nil
}

// extractUserID extrai o ID do usuário dos claims
func (j *JWTMiddleware) extractUserID(claims jwt.Claims) string {
	if claimsMap, ok := claims.(jwt.MapClaims); ok {
		if userID, ok := claimsMap["sub"].(string); ok {
			return userID
		}
		if userID, ok := claimsMap["user_id"].(string); ok {
			return userID
		}
	}
	return ""
}

// handleError trata erros do middleware JWT
func (j *JWTMiddleware) handleError(c echo.Context, err error) error {
	if j.config.ErrorHandlerWithContext != nil {
		return j.config.ErrorHandlerWithContext(c, err)
	}
	if j.config.ErrorHandler != nil {
		return j.config.ErrorHandler(c)
	}
	return err
}

// GetUserFromContext extrai o usuário do contexto
func GetUserFromContext(c echo.Context) jwt.Claims {
	if user := c.Get("user"); user != nil {
		if claims, ok := user.(jwt.Claims); ok {
			return claims
		}
	}
	return nil
}

// GetUserIDFromContext extrai o ID do usuário do contexto
func GetUserIDFromContext(c echo.Context) string {
	if claims := GetUserFromContext(c); claims != nil {
		if claimsMap, ok := claims.(jwt.MapClaims); ok {
			if userID, ok := claimsMap["sub"].(string); ok {
				return userID
			}
			if userID, ok := claimsMap["user_id"].(string); ok {
				return userID
			}
		}
	}
	return ""
}
