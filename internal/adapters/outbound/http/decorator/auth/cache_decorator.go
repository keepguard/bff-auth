package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// CacheConfig configuração para cache
type CacheConfig struct {
	TTL             time.Duration // Tempo de vida do cache
	MaxSize         int           // Número máximo de entradas
	CleanupInterval time.Duration // Intervalo de limpeza de entradas expiradas
}

// DefaultCacheConfig retorna uma configuração padrão de cache
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:             5 * time.Minute,
		MaxSize:         1000,
		CleanupInterval: 1 * time.Minute,
	}
}

// cacheEntry representa uma entrada no cache
type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// isExpired verifica se a entrada está expirada
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

// cache implementa um cache thread-safe em memória
type cache struct {
	entries map[string]*cacheEntry
	mu      sync.RWMutex
	config  CacheConfig
	metrics *metrics.Metrics
}

// newCache cria um novo cache
func newCache(config CacheConfig, metrics *metrics.Metrics) *cache {
	c := &cache{
		entries: make(map[string]*cacheEntry),
		config:  config,
		metrics: metrics,
	}

	// Inicia goroutine de limpeza
	if config.CleanupInterval > 0 {
		go c.cleanup()
	}

	return c
}

// cleanup remove entradas expiradas periodicamente
func (c *cache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.entries {
			if entry.isExpired() {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// get obtém um valor do cache
func (c *cache) get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("auth-client", "auth")
		}
		return nil, false
	}

	if entry.isExpired() {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("auth-client", "auth")
		}
		return nil, false
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("auth-client", "auth")
	}
	return entry.value, true
}

// set armazena um valor no cache
func (c *cache) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Se atingiu o tamanho máximo, remove entrada mais antiga
	if len(c.entries) >= c.config.MaxSize {
		// Remove primeira entrada (simples, mas não LRU)
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(c.config.TTL),
	}
}

// invalidate remove uma chave específica do cache
func (c *cache) invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// clear limpa todo o cache
func (c *cache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*cacheEntry)
}

// cacheDecorator implementa o padrão decorator para cache
type cacheDecorator struct {
	inner  authclient.AuthClient
	cache  *cache
	config CacheConfig
}

// NewCacheDecorator cria um decorator de cache para AuthClient
func NewCacheDecorator(
	inner authclient.AuthClient,
	config CacheConfig,
	metrics *metrics.Metrics,
) authclient.AuthClient {
	return &cacheDecorator{
		inner:  inner,
		cache:  newCache(config, metrics),
		config: config,
	}
}

// generateCacheKey gera uma chave de cache baseada nos parâmetros
func (d *cacheDecorator) generateCacheKey(operation string, params ...interface{}) string {
	data, _ := json.Marshal(map[string]interface{}{
		"operation": operation,
		"params":    params,
	})
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Login - NÃO CACHEIA (autenticação sempre deve ser validada)
func (d *cacheDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent string) (inboundDto.AuthResponseDTO, error) {
	return d.inner.Login(ctx, req, tenantId, correlationID, clientId, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// RefreshToken - NÃO CACHEIA (tokens sempre devem ser validados)
func (d *cacheDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, tenantId, correlationID, clientId string) (inboundDto.RefreshTokenResponseDTO, error) {
	return d.inner.RefreshToken(ctx, req, tenantId, correlationID, clientId)
}

// Logout - NÃO CACHEIA e INVALIDA cache relacionado
func (d *cacheDecorator) Logout(ctx context.Context, token, tenantId, correlationID string) error {
	err := d.inner.Logout(ctx, token, tenantId, correlationID)

	// Invalida cache de validação de token
	cacheKey := d.generateCacheKey("ValidateToken", token, tenantId)
	d.cache.invalidate(cacheKey)

	return err
}

// ValidateToken - CACHEIA (validação de token pode ser cacheada por curto período)
func (d *cacheDecorator) ValidateToken(ctx context.Context, token, tenantId, correlationID string) error {
	cacheKey := d.generateCacheKey("ValidateToken", token, tenantId)

	// Tenta obter do cache
	if cached, found := d.cache.get(cacheKey); found {
		if err, ok := cached.(error); ok {
			return err
		}
		return nil
	}

	// Se não encontrou no cache, executa a operação
	err := d.inner.ValidateToken(ctx, token, tenantId, correlationID)

	// Armazena no cache (mesmo erros, para evitar validações repetidas)
	d.cache.set(cacheKey, err)

	return err
}

// ChangePassword - NÃO CACHEIA (operação de escrita)
func (d *cacheDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	return d.inner.ChangePassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// ResetPassword - NÃO CACHEIA (operação de escrita)
func (d *cacheDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent string) error {
	return d.inner.ResetPassword(ctx, req, tenantId, correlationID, deviceId, deviceName, deviceType, ipAddress, userAgent)
}

// GenerateResetToken implementa o método GenerateResetToken (sem cache)
func (d *cacheDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, tenantId, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	return d.inner.GenerateResetToken(ctx, req, tenantId, correlationID)
}

// SendDeviceChallenge implementa o método SendDeviceChallenge (sem cache)
func (d *cacheDecorator) SendDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeSendRequestDTO, tenantId, correlationID string) (map[string]interface{}, error) {
	return d.inner.SendDeviceChallenge(ctx, req, tenantId, correlationID)
}

// VerifyDeviceChallenge implementa o método VerifyDeviceChallenge (sem cache)
func (d *cacheDecorator) VerifyDeviceChallenge(ctx context.Context, req inboundDto.DeviceChallengeVerifyRequestDTO, tenantId, correlationID string) (inboundDto.AuthResponseDTO, error) {
	return d.inner.VerifyDeviceChallenge(ctx, req, tenantId, correlationID)
}

// ListUserSessions implementa o método ListUserSessions (sem cache)
func (d *cacheDecorator) ListUserSessions(ctx context.Context, token, deviceId, tenantId, correlationID string) ([]inboundDto.DeviceSessionDTO, error) {
	return d.inner.ListUserSessions(ctx, token, deviceId, tenantId, correlationID)
}

// RevokeSession implementa o método RevokeSession (sem cache)
func (d *cacheDecorator) RevokeSession(ctx context.Context, deviceIdToRevoke, token, tenantId, correlationID string) error {
	return d.inner.RevokeSession(ctx, deviceIdToRevoke, token, tenantId, correlationID)
}

// RevokeAllOtherSessions implementa o método RevokeAllOtherSessions (sem cache)
func (d *cacheDecorator) RevokeAllOtherSessions(ctx context.Context, token, currentDeviceId, tenantId, correlationID string) error {
	return d.inner.RevokeAllOtherSessions(ctx, token, currentDeviceId, tenantId, correlationID)
}

// QuickRevoke implementa o método QuickRevoke (sem cache)
func (d *cacheDecorator) QuickRevoke(ctx context.Context, token string, blacklist bool, tenantId, correlationID string) (map[string]interface{}, error) {
	return d.inner.QuickRevoke(ctx, token, blacklist, tenantId, correlationID)
}

// ListDeviceBlacklist implementa o método ListDeviceBlacklist (sem cache)
func (d *cacheDecorator) ListDeviceBlacklist(ctx context.Context, token, tenantId, correlationID string) ([]inboundDto.DeviceBlacklistDTO, error) {
	return d.inner.ListDeviceBlacklist(ctx, token, tenantId, correlationID)
}

// AddDeviceToBlacklist implementa o método AddDeviceToBlacklist (sem cache)
func (d *cacheDecorator) AddDeviceToBlacklist(ctx context.Context, req inboundDto.AddDeviceBlacklistRequestDTO, token, tenantId, correlationID string) error {
	return d.inner.AddDeviceToBlacklist(ctx, req, token, tenantId, correlationID)
}

// RemoveDeviceFromBlacklist implementa o método RemoveDeviceFromBlacklist (sem cache)
func (d *cacheDecorator) RemoveDeviceFromBlacklist(ctx context.Context, deviceId, token, tenantId, correlationID string) error {
	return d.inner.RemoveDeviceFromBlacklist(ctx, deviceId, token, tenantId, correlationID)
}
