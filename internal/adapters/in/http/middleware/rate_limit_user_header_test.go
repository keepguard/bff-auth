package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/keepguard/bff-auth/internal/infrastructure/config"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSecurityMiddleware_StripsClientUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-User-ID", "forjado-pelo-cliente")
	c := e.NewContext(req, httptest.NewRecorder())

	handler := NewMiddleware(zap.NewNop()).SecurityMiddleware()(func(c echo.Context) error {
		assert.Empty(t, GetUserID(c), "X-User-ID só vale vindo de JWT validado")
		return nil
	})
	require.NoError(t, handler(c))
}

// TestRateLimit_RotatingUserIDDoesNotBypassLogin: Security + rate limit de
// login com Redis real. Trocar o X-User-ID a cada tentativa não fura o
// limite — a chave é o IP. Roda só com BFF_AUTH_IT_REDIS_ADDR (Redis
// descartável), ex.: docker run --rm -d -p 56379:6379 redis:7.2-alpine.
func TestRateLimit_RotatingUserIDDoesNotBypassLogin(t *testing.T) {
	addr := os.Getenv("BFF_AUTH_IT_REDIS_ADDR")
	if addr == "" {
		t.Skip("BFF_AUTH_IT_REDIS_ADDR não definida (teste de integração com Redis)")
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	t.Cleanup(func() { _ = rdb.Close() })
	ip := fmt.Sprintf("203.0.113.%d", time.Now().UnixNano()%250+1)
	require.NoError(t, rdb.Del(context.Background(), "rl:bff-auth:login:"+ip).Err())

	const limit = 3
	rl := NewRateLimiterMiddleware(rdb, config.RateLimitConfig{Enabled: true}, zap.NewNop(), nil)
	e := echo.New()
	e.Use(NewMiddleware(zap.NewNop()).SecurityMiddleware())
	e.POST("/api/v1/auth/login", func(c echo.Context) error { return c.NoContent(http.StatusOK) },
		rl.Limit("login", config.RateLimitRule{Limit: limit, Window: time.Minute}))

	var codes []int
	for i := 0; i < limit+2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		req.Header.Set("X-Real-IP", ip)
		req.Header.Set("X-User-ID", fmt.Sprintf("atacante-%d", i))
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}

	assert.Equal(t, []int{200, 200, 200, 429, 429}, codes)
}
