package http

import (
	"net/http"
	"strings"

	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
)

// CompanyResolveMiddleware resolve tenant (JWT ou X-Tenant-Id) → company via cache/ms-company
// e anexa o companyId ao contexto das chamadas internas.
func CompanyResolveMiddleware(companyClient authclient.CompanyClient) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/swagger") {
				return next(c)
			}

			tenantId := tenantIDFromRequest(c)
			if tenantId == "" {
				return next(c)
			}

			correlationID := c.Request().Header.Get("X-Correlation-ID")
			company, err := companyClient.GetByTenantId(c.Request().Context(), tenantId, correlationID)
			if err != nil {
				return err
			}
			if company.ID == "" {
				return echo.NewHTTPError(http.StatusNotFound, "Empresa não encontrada para o tenant informado")
			}

			req := c.Request().WithContext(authclient.WithCompanyID(c.Request().Context(), company.ID))
			c.SetRequest(req)
			return next(c)
		}
	}
}

func tenantIDFromRequest(c echo.Context) string {
	headerTenant := strings.TrimSpace(c.Request().Header.Get("X-Tenant-Id"))
	authHeader := c.Request().Header.Get("Authorization")
	jwtTenant := ""
	if strings.HasPrefix(authHeader, "Bearer ") {
		if t, err := pkg.ExtractTenantIdFromToken(authHeader); err == nil {
			jwtTenant = strings.TrimSpace(t)
		}
	}
	if jwtTenant != "" {
		return jwtTenant
	}
	return headerTenant
}
