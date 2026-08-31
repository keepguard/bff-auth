package http

import (
	"net/http"
	"strings"

	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
)

func RequireBearer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if !strings.HasPrefix(authHeader, "Bearer ") || token == "" {
				return c.JSON(http.StatusUnauthorized, pkg.ErrorResponse{
					Error:         "UNAUTHORIZED",
					Message:       "Token de autorização não fornecido ou inválido",
					CorrelationID: GetCorrelationID(c),
				})
			}
			return next(c)
		}
	}
}

func RequireAnyRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
			claims, err := pkg.ExtractAllClaims(token)
			if err != nil || claims == nil || !pkg.HasAnyRole(claims.Roles, roles...) {
				return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
					Error:         "FORBIDDEN",
					Message:       "Acesso restrito a gestores e administradores",
					CorrelationID: GetCorrelationID(c),
				})
			}
			return next(c)
		}
	}
}
