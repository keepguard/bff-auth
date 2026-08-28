package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

// RateLimitConfig configuração para rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int // Número de requisições permitidas por segundo
	Burst             int // Número de requisições em burst permitidas
}

// DefaultRateLimitConfig retorna uma configuração padrão de rate limit
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerSecond: 100, // 100 req/s
		Burst:             20,  // Permite burst de 20 requests
	}
}

// RateLimitError erro retornado quando rate limit é excedido
type RateLimitError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (retry after: %v)", e.Message, e.RetryAfter)
}

// tokenBucket implementa o algoritmo Token Bucket para rate limiting
type tokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64 // tokens por segundo
	lastRefillTime time.Time
	mu             sync.Mutex
}

// newTokenBucket cria um novo token bucket
func newTokenBucket(requestsPerSecond, burst int) *tokenBucket {
	maxTokens := float64(burst)
	if maxTokens < float64(requestsPerSecond) {
		maxTokens = float64(requestsPerSecond)
	}

	return &tokenBucket{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		refillRate:     float64(requestsPerSecond),
		lastRefillTime: time.Now(),
	}
}

// refill reabastece os tokens baseado no tempo decorrido
func (tb *tokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefillTime).Seconds()

	// Adiciona tokens baseado no tempo decorrido
	tb.tokens += elapsed * tb.refillRate

	// Não pode exceder o máximo
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}

	tb.lastRefillTime = now
}

// tryConsume tenta consumir um token
// Retorna true se conseguiu, false se rate limit foi excedido
func (tb *tokenBucket) tryConsume() (bool, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true, 0
	}

	// Calcula quanto tempo até ter 1 token disponível
	tokensNeeded := 1.0 - tb.tokens
	retryAfter := time.Duration(tokensNeeded/tb.refillRate) * time.Second

	return false, retryAfter
}

// rateLimitDecorator implementa o padrão decorator para rate limiting
type rateLimitDecorator struct {
	inner  authclient.AuthClient
	bucket *tokenBucket
	config RateLimitConfig
}

// NewRateLimitDecorator cria um decorator de rate limit para AuthClient
func NewRateLimitDecorator(
	inner authclient.AuthClient,
	config RateLimitConfig,
) authclient.AuthClient {
	return &rateLimitDecorator{
		inner:  inner,
		bucket: newTokenBucket(config.RequestsPerSecond, config.Burst),
		config: config,
	}
}

// checkRateLimit verifica se a requisição está dentro do rate limit
func (d *rateLimitDecorator) checkRateLimit() error {
	allowed, retryAfter := d.bucket.tryConsume()
	if !allowed {
		return &RateLimitError{
			Message:    fmt.Sprintf("Rate limit of %d requests/second exceeded", d.config.RequestsPerSecond),
			RetryAfter: retryAfter,
		}
	}
	return nil
}

// Login implementa o método Login com rate limiting
func (d *rateLimitDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return inboundDto.AuthResponseDTO{}, err
	}
	return d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// RefreshToken implementa o método RefreshToken com rate limiting
func (d *rateLimitDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return inboundDto.RefreshTokenResponseDTO{}, err
	}
	return d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)
}

// Logout implementa o método Logout com rate limiting
func (d *rateLimitDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.Logout(ctx, token, tenantId, correlationID)
}

// ValidateToken implementa o método ValidateToken com rate limiting
func (d *rateLimitDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.ValidateToken(ctx, token, tenantId, correlationID)
}

// ChangePassword implementa o método ChangePassword com rate limiting
func (d *rateLimitDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.ChangePassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// ResetPassword implementa o método ResetPassword com rate limiting
func (d *rateLimitDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.ResetPassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// GenerateResetToken implementa o método GenerateResetToken com rate limiting
func (d *rateLimitDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return outboundDto.GenerateResetTokenMSResponseDTO{}, err
	}
	return d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)
}

// SendDeviceChallenge implementa o método SendDeviceChallenge com rate limiting
func (d *rateLimitDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	if err := d.checkRateLimit(); err != nil {
		return nil, err
	}
	return d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge com rate limiting
func (d *rateLimitDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return inboundDto.AuthResponseDTO{}, err
	}
	return d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
}

// ListUserSessions implementa o método ListUserSessions com rate limiting
func (d *rateLimitDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return nil, err
	}
	return d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
}

// RevokeSession implementa o método RevokeSession com rate limiting
func (d *rateLimitDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions com rate limiting
func (d *rateLimitDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
}

// QuickRevoke implementa o método QuickRevoke com rate limiting
func (d *rateLimitDecorator) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	if err := d.checkRateLimit(); err != nil {
		return nil, err
	}
	return d.inner.QuickRevoke(ctx, token, blacklist, tenantId, correlationID)
}

// ListDeviceBlacklist implementa o método ListDeviceBlacklist com rate limiting
func (d *rateLimitDecorator) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return nil, err
	}
	return d.inner.ListDeviceBlacklist(ctx, token, tenantId, correlationID)
}

// AddDeviceToBlacklist implementa o método AddDeviceToBlacklist com rate limiting
func (d *rateLimitDecorator) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.AddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

// RemoveDeviceFromBlacklist implementa o método RemoveDeviceFromBlacklist com rate limiting
func (d *rateLimitDecorator) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.RemoveDeviceFromBlacklist(ctx, deviceId, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) SearchAdminDeviceBlacklist(ctx context.Context, queryParams map[string]string, token, tenantId, correlationID string) (inboundDto.PaginatedDeviceBlacklistResponseDTO, error) {
	return d.inner.SearchAdminDeviceBlacklist(ctx, queryParams, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) AdminAddDeviceToBlacklist(ctx context.Context, req inboundDto.AdminAddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.inner.AdminAddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) AdminRemoveDeviceFromBlacklist(ctx context.Context, deviceId, userId, token, tenantId, correlationID string) error {
	return d.inner.AdminRemoveDeviceFromBlacklist(ctx, deviceId, userId, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error) {
	if err := d.checkRateLimit(); err != nil {
		return outboundDto.UserByCodeResponseDTO{}, err
	}
	return d.inner.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) BlockUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.BlockUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
}

func (d *rateLimitDecorator) DeleteUser(ctx context.Context, idUserExternal, reason, token, tenantId, correlationID string) error {
	if err := d.checkRateLimit(); err != nil {
		return err
	}
	return d.inner.DeleteUser(ctx, idUserExternal, reason, token, tenantId, correlationID)
}
