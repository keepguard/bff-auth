package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func encodeUnsignedJWT(t *testing.T, payload map[string]any) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{"alg": "none", "typ": "JWT"})
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(body) + ".sig"
}

func runAuthorityMiddleware(t *testing.T, mw echo.MiddlewareFunc, token string) int {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set("X-Correlation-ID", "corr-1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	handler := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return rec.Code
}

func TestRequireSessionRead_AllowsAuthority(t *testing.T) {
	token := encodeUnsignedJWT(t, map[string]any{
		"roles":       []string{"ROLE_MANAGER"},
		"authorities": []string{"session:read"},
	})
	if code := runAuthorityMiddleware(t, RequireSessionRead(), token); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestRequireSessionRead_RejectsManagerWithoutAuthority(t *testing.T) {
	token := encodeUnsignedJWT(t, map[string]any{
		"roles": []string{"ROLE_MANAGER"},
	})
	if code := runAuthorityMiddleware(t, RequireSessionRead(), token); code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestRequireSessionWrite_AllowsAdminWithoutAuthorities(t *testing.T) {
	token := encodeUnsignedJWT(t, map[string]any{
		"roles": []string{"ADMIN"},
	})
	if code := runAuthorityMiddleware(t, RequireSessionWrite(), token); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestRequireSessionWrite_RejectsReadOnly(t *testing.T) {
	token := encodeUnsignedJWT(t, map[string]any{
		"roles":       []string{"ROLE_MANAGER"},
		"authorities": []string{"session:read"},
	})
	if code := runAuthorityMiddleware(t, RequireSessionWrite(), token); code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}
