package http

import (
	"net/http"
	"strings"

	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
)

const (
	AuthoritySessionRead  = "session:read"
	AuthoritySessionWrite = "session:write"
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

func RequireAuthority(authority string) echo.MiddlewareFunc {
	message := "Acesso restrito a administradores ou à permissão " + authority
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token := strings.TrimSpace(strings.TrimPrefix(c.Request().Header.Get("Authorization"), "Bearer "))
			claims, err := pkg.ExtractAllClaims(token)
			if err == nil && claims != nil &&
				(pkg.HasAnyRole(claims.Roles, "ADMIN", "SYSTEM") || pkg.HasAuthority(claims.Authorities, authority)) {
				return next(c)
			}
			return c.JSON(http.StatusForbidden, pkg.ErrorResponse{
				Error:         "FORBIDDEN",
				Message:       message,
				CorrelationID: GetCorrelationID(c),
			})
		}
	}
}

func RequireSessionRead() echo.MiddlewareFunc {
	return RequireAuthority(AuthoritySessionRead)
}

func RequireSessionWrite() echo.MiddlewareFunc {
	return RequireAuthority(AuthoritySessionWrite)
}
