package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	auditport "github.com/keepguard/bff-auth/internal/domain/ports/audit"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type recPub struct {
	events []auditport.Event
}

func (r *recPub) Publish(_ context.Context, event auditport.Event) {
	r.events = append(r.events, event)
}

func (r *recPub) Close() error { return nil }

func TestMapAuditActionLogin(t *testing.T) {
	require.Equal(t, "LOGIN", mapAuditAction(http.MethodPost, "/api/v1/auth/login"))
	require.Equal(t, "LOGOUT", mapAuditAction(http.MethodPost, "/api/v1/auth/logout"))
}

func TestShouldSkipValidateAndGet(t *testing.T) {
	require.True(t, shouldSkipAudit(http.MethodPost, "/api/v1/auth/validate"))
	require.True(t, shouldSkipAudit(http.MethodGet, "/api/v1/users/me/sessions"))
	require.False(t, shouldSkipAudit(http.MethodPost, "/api/v1/auth/login"))
}

func TestAuditMiddlewareEmitsLoginWithCorrelationID(t *testing.T) {
	rec := &recPub{}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-Correlation-ID", "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	rr := httptest.NewRecorder()
	c := e.NewContext(req, rr)
	c.SetPath("/api/v1/auth/login")

	mw := AuditMiddleware(rec, "bff-auth")
	err := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})(c)
	require.NoError(t, err)
	require.Len(t, rec.events, 1)
	require.Equal(t, "LOGIN", rec.events[0].Action)
	require.Equal(t, "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", rec.events[0].CorrelationID)
	require.Equal(t, "SUCCESS", rec.events[0].Outcome)
	require.Equal(t, "bff-auth", rec.events[0].SourceService)
}
