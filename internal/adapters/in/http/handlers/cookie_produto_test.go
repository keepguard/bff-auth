package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type cabecalhos struct{ host, forwardedHost, origin, referer string }

func contextoCookie(cb cabecalhos) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodPost, "http://"+cb.host+"/api/v1/auth/login", nil)
	for k, v := range map[string]string{"X-Forwarded-Host": cb.forwardedHost, "Origin": cb.origin, "Referer": cb.referer} {
		if v != "" {
			req.Header.Set(k, v)
		}
	}
	rec := httptest.NewRecorder()
	return echo.New().NewContext(req, rec), rec
}

// Envs de produção (08/11/2026): Domain=.keepguard.com.br, SameSite=None e Secure para todos os pedidos.
func configurarCookieDoKeepguard(t *testing.T) {
	t.Helper()
	t.Setenv("BFF_AUTH_COOKIE_DOMAIN", ".keepguard.com.br")
	t.Setenv("BFF_AUTH_COOKIE_SAMESITE", "None")
	t.Setenv("BFF_AUTH_COOKIE_SECURE", "true")
	t.Setenv("BFF_AUTH_COOKIE_PATH", "")
}

// Produto em domínio próprio (Echouno pelo proxy do nginx do front): cookie do host, sem Domain, Secure e Lax, mesmo com o env global.
func TestCookieConfig_ForaDoKeepguardViraHostOnly(t *testing.T) {
	configurarCookieDoKeepguard(t)
	h := &AuthHandlers{}

	casos := map[string]cabecalhos{
		"nginx do app (host e forwarded)":  {host: "app.echouno.com", forwardedHost: "app.echouno.com", origin: "https://app.echouno.com", referer: "https://app.echouno.com/entrar"},
		"nginx do admin":                   {host: "admin.echouno.com", forwardedHost: "admin.echouno.com", origin: "https://admin.echouno.com"},
		"só x-forwarded-host":              {host: "bff-auth:8381", forwardedHost: "app.echouno.com"},
		"host com porta":                   {host: "app.echouno.com:443"},
		"forwarded em lista":               {host: "bff-auth:8381", forwardedHost: "app.echouno.com, 10.0.0.1"},
		"servidor para servidor":           {host: "bff-auth:8381"},
		"domínio parecido não é keepguard": {host: "keepguard.com.evil.com"},
	}
	for nome, cb := range casos {
		c, _ := contextoCookie(cb)
		domain, sameSite, secure, path := h.getCookieConfig(c)
		if domain != "" || sameSite != http.SameSiteLaxMode || !secure || path != "/" {
			t.Errorf("%s: domain=%q sameSite=%v secure=%v path=%q; esperava host-only, Lax, Secure, /", nome, domain, sameSite, secure, path)
		}
	}
}

// *.keepguard.com.br: nada muda, valem os envs de sempre.
func TestCookieConfig_KeepguardNaoMuda(t *testing.T) {
	configurarCookieDoKeepguard(t)
	h := &AuthHandlers{}

	for nome, cb := range map[string]cabecalhos{
		"host":      {host: "app-core.keepguard.com.br"},
		"forwarded": {host: "bff-auth:8381", forwardedHost: "app-core.keepguard.com.br"},
		"origin":    {host: "bff-auth:8381", origin: "https://app-core.keepguard.com.br"},
		"referer":   {host: "bff-auth:8381", referer: "https://app-core.keepguard.com.br/login"},
	} {
		c, _ := contextoCookie(cb)
		domain, sameSite, secure, path := h.getCookieConfig(c)
		if domain != ".keepguard.com.br" || sameSite != http.SameSiteNoneMode || !secure || path != "/" {
			t.Errorf("%s: domain=%q sameSite=%v secure=%v path=%q; o KeepGuard devia seguir igual", nome, domain, sameSite, secure, path)
		}
	}
}

// Desenvolvimento local sem HTTPS: host-only e sem Secure (senão o navegador recusa o cookie em http://localhost).
func TestCookieConfig_LocalhostSemSecure(t *testing.T) {
	for _, k := range []string{"BFF_AUTH_COOKIE_DOMAIN", "BFF_AUTH_COOKIE_SAMESITE", "BFF_AUTH_COOKIE_SECURE", "BFF_AUTH_COOKIE_PATH"} {
		t.Setenv(k, "")
	}
	h := &AuthHandlers{}
	c, _ := contextoCookie(cabecalhos{host: "localhost:8381"})
	domain, sameSite, secure, _ := h.getCookieConfig(c)
	if domain != "" || sameSite != http.SameSiteLaxMode || secure {
		t.Errorf("domain=%q sameSite=%v secure=%v; esperava host-only, Lax, sem Secure", domain, sameSite, secure)
	}
}

// O Set-Cookie que sai de verdade no login e no logout, nos dois casos.
func TestSetCookie_CabecalhoNosDoisCasos(t *testing.T) {
	configurarCookieDoKeepguard(t)
	h := &AuthHandlers{logger: zap.NewNop()}

	c, rec := contextoCookie(cabecalhos{host: "app.echouno.com", forwardedHost: "app.echouno.com"})
	h.setRefreshTokenCookie(c, "rt_teste")
	sc := strings.Join(rec.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(sc, "keepguard_refresh_token=rt_teste") || strings.Contains(sc, "Domain=") ||
		!strings.Contains(sc, "HttpOnly") || !strings.Contains(sc, "Secure") || !strings.Contains(sc, "SameSite=Lax") {
		t.Errorf("Echouno: Set-Cookie inesperado:\n%s", sc)
	}

	c, rec = contextoCookie(cabecalhos{host: "app.echouno.com", forwardedHost: "app.echouno.com"})
	h.clearRefreshTokenCookie(c)
	if sc := strings.Join(rec.Header().Values("Set-Cookie"), "\n"); strings.Contains(sc, "Domain=") || !strings.Contains(sc, "Max-Age=0") {
		t.Errorf("Echouno logout: Set-Cookie inesperado:\n%s", sc)
	}

	c, rec = contextoCookie(cabecalhos{host: "app-core.keepguard.com.br"})
	h.setRefreshTokenCookie(c, "rt_teste")
	sc = strings.Join(rec.Header().Values("Set-Cookie"), "\n")
	if !strings.Contains(sc, "Domain=keepguard.com.br") || !strings.Contains(sc, "SameSite=None") || !strings.Contains(sc, "Secure") {
		t.Errorf("KeepGuard: Set-Cookie inesperado:\n%s", sc)
	}
}
