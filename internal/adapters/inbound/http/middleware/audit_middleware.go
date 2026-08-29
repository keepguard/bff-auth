package http

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	auditport "github.com/keepguard/bff-auth/internal/domain/ports/audit"
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
			outcome := "SUCCESS"
			if status >= 400 {
				outcome = "FAILURE"
			}
			if status == http.StatusForbidden || status == http.StatusUnauthorized {
				outcome = "DENIED"
			}
			event := auditport.Event{
				EventID:       newUUID(),
				OccurredAt:    time.Now().UTC().Format(time.RFC3339),
				SchemaVersion: 1,
				SourceService: sourceService,
				CorrelationID: GetCorrelationID(c),
				RequestID:     c.Response().Header().Get(echo.HeaderXRequestID),
				TenantID:      c.Request().Header.Get("X-Tenant-Id"),
				CompanyID:     c.Request().Header.Get("X-Company-Id"),
				Actor: auditport.Actor{
					Type:     actorType(c.Request().Header.Get("X-User-ID")),
					CodeUser: c.Request().Header.Get("X-User-ID"),
					ClientIP: c.RealIP(),
					DeviceID: c.Request().Header.Get("X-Device-Id"),
				},
				Action:   mapAuditAction(c.Request().Method, path),
				Resource: auditport.Resource{Type: "HTTP", ID: path},
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
	if method == http.MethodGet || method == http.MethodOptions || method == http.MethodHead {
		return true
	}
	if path == "/api/v1/auth/validate" {
		return true
	}
	return false
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
