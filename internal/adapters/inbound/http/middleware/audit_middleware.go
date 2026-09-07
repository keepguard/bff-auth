package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	authclient "github.com/keepguard/bff-auth/internal/application/port"
	auditport "github.com/keepguard/bff-auth/internal/domain/ports/audit"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/labstack/echo/v4"
)

func AuditMiddleware(publisher auditport.EventPublisher, sourceService string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if publisher == nil {
				return err
			}
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			if shouldSkipAudit(c.Request().Method, path) {
				return err
			}
			status := c.Response().Status
			denied := status == http.StatusForbidden || status == http.StatusUnauthorized
			if !privilegedAuditRead(path) && domainCoveredByMS(path) {
				return err
			}
			outcome := "SUCCESS"
			if status >= 400 {
				outcome = "FAILURE"
			}
			if denied {
				outcome = "DENIED"
			}
			codeUser, tenantID, companyID, deviceID := auditIdentity(c)
			action := mapAuditAction(c.Request().Method, path)
			resource := auditport.Resource{Type: "HTTP", ID: path}
			if privilegedAuditRead(path) && c.Request().Method == http.MethodGet {
				action, resource = privilegedReadAction(path, c)
			}
			event := auditport.Event{
				EventID:       newUUID(),
				OccurredAt:    time.Now().UTC().Format(time.RFC3339),
				SchemaVersion: 1,
				SourceService: sourceService,
				CorrelationID: GetCorrelationID(c),
				RequestID:     c.Response().Header().Get(echo.HeaderXRequestID),
				TenantID:      tenantID,
				CompanyID:     companyID,
				Actor: auditport.Actor{
					Type:     actorType(codeUser),
					CodeUser: codeUser,
					ClientIP: c.RealIP(),
					DeviceID: deviceID,
				},
				Action:   action,
				Resource: resource,
				Outcome:  outcome,
				Metadata: map[string]any{
					"method": c.Request().Method,
					"status": status,
				},
			}
			publisher.Publish(c.Request().Context(), event)
			return err
		}
	}
}

func shouldSkipAudit(method, path string) bool {
	if path == "/health" || strings.HasPrefix(path, "/swagger") || path == "/metrics" {
		return true
	}
	if method == http.MethodOptions || method == http.MethodHead {
		return true
	}
	if path == "/api/v1/auth/validate" {
		return true
	}
	if strings.Contains(path, "/auth/refresh") {
		return true
	}
	if method == http.MethodGet {
		return !privilegedAuditRead(path)
	}
	return false
}

func privilegedAuditRead(path string) bool {
	if strings.Contains(path, "/users/me") {
		return false
	}
	if strings.Contains(path, "/users/") && strings.Contains(path, "/sessions") {
		return true
	}
	if strings.Contains(path, "/users/") && strings.Contains(path, "/devices/blacklist") {
		return true
	}
	if path == "/api/v1/sessions" {
		return true
	}
	if path == "/api/v1/devices/blacklist" || strings.HasPrefix(path, "/api/v1/admin/devices/blacklist") {
		return true
	}
	return false
}

func privilegedReadAction(path string, c echo.Context) (string, auditport.Resource) {
	if strings.Contains(path, "/sessions") {
		return "SESSION_LIST_TENANT", auditport.Resource{Type: "SESSION", ID: strings.TrimSpace(c.Param("userId"))}
	}
	if strings.Contains(path, "/devices/blacklist") {
		return "DEVICE_BLACKLIST_LIST_TENANT", auditport.Resource{Type: "DEVICE", ID: strings.TrimSpace(c.Param("userId"))}
	}
	return "SESSION_LIST_TENANT", auditport.Resource{Type: "HTTP", ID: path}
}

func domainCoveredByMS(path string) bool {
	switch {
	case strings.Contains(path, "/auth/login"):
		return true
	case strings.Contains(path, "/auth/logout"):
		return true
	case strings.Contains(path, "/change-password"):
		return true
	case strings.Contains(path, "/reset-password"):
		return true
	case strings.Contains(path, "/forgot-password"):
		return true
	case strings.Contains(path, "/device/challenge/send"):
		return true
	case strings.Contains(path, "/device/challenge/verify"):
		return true
	case strings.Contains(path, "/quick-revoke"):
		return true
	case strings.Contains(path, "/sessions"):
		return true
	case strings.Contains(path, "/devices/blacklist"):
		return true
	case strings.HasSuffix(path, "/block"):
		return true
	case strings.Contains(path, "/users/me"):
		return true
	default:
		return false
	}
}

func mapAuditAction(method, path string) string {
	switch {
	case strings.Contains(path, "/auth/login"):
		return "LOGIN"
	case strings.Contains(path, "/auth/logout"):
		return "LOGOUT"
	case strings.Contains(path, "/auth/refresh"):
		return "REFRESH_TOKEN"
	case strings.Contains(path, "/change-password"):
		return "CHANGE_PASSWORD"
	case strings.Contains(path, "/reset-password"):
		return "RESET_PASSWORD"
	case strings.Contains(path, "/forgot-password"):
		return "FORGOT_PASSWORD"
	case strings.Contains(path, "/device/challenge/send"):
		return "SEND_DEVICE_CHALLENGE"
	case strings.Contains(path, "/device/challenge/verify"):
		return "VERIFY_DEVICE_CHALLENGE"
	case strings.Contains(path, "/quick-revoke"):
		return "QUICK_REVOKE"
	case strings.Contains(path, "/sessions") && method == http.MethodDelete:
		return "REVOKE_SESSION"
	case strings.Contains(path, "/devices/blacklist") && method == http.MethodPost:
		return "ADD_DEVICE_BLACKLIST"
	case strings.Contains(path, "/devices/blacklist") && method == http.MethodDelete:
		return "REMOVE_DEVICE_BLACKLIST"
	case strings.HasSuffix(path, "/block"):
		return "BLOCK_ME"
	case strings.Contains(path, "/users/me") && method == http.MethodDelete:
		return "DELETE_ME"
	default:
		return method + "_" + strings.Trim(path, "/")
	}
}

func auditIdentity(c echo.Context) (codeUser, tenantID, companyID, deviceID string) {
	authHeader := c.Request().Header.Get("Authorization")
	var claims *pkg.JWTClaims
	if strings.HasPrefix(authHeader, "Bearer ") {
		claims, _ = pkg.ExtractAllClaims(authHeader)
	}
	if claims != nil {
		codeUser = firstNonBlank(claims.CodeUser, claims.Sub, claims.Username)
		tenantID = strings.TrimSpace(claims.TenantId)
		deviceID = strings.TrimSpace(claims.DeviceID)
	}
	if codeUser == "" {
		codeUser = strings.TrimSpace(GetUserID(c))
	}
	if tenantID == "" {
		tenantID = strings.TrimSpace(c.Request().Header.Get("X-Tenant-Id"))
	}
	companyID = authclient.CompanyIDFromContext(c.Request().Context())
	if companyID == "" {
		companyID = strings.TrimSpace(c.Request().Header.Get("X-Company-Id"))
	}
	if deviceID == "" {
		deviceID = strings.TrimSpace(c.Request().Header.Get("X-Device-Id"))
	}
	return codeUser, tenantID, companyID, deviceID
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func actorType(codeUser string) string {
	if strings.TrimSpace(codeUser) == "" {
		return "ANONYMOUS"
	}
	return "USER"
}

func newUUID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b)
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}
