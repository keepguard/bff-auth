package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	inboundDto "github.com/keepguard/bff-auth/internal/adapters/inbound/http/dto"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// SmartCacheConfig configuração para cache inteligente
type SmartCacheConfig struct {
	MaxTTL          time.Duration // TTL máximo permitido (segurança)
	MinTTL          time.Duration // TTL mínimo (para tokens com ExpiresIn muito baixo)
	MaxSize         int           // Número máximo de entradas
	CleanupInterval time.Duration // Intervalo de limpeza
}

// DefaultSmartCacheConfig retorna configuração padrão
func DefaultSmartCacheConfig() SmartCacheConfig {
	return SmartCacheConfig{
		MaxTTL:          10 * time.Minute, // Máximo 10 minutos por segurança
		MinTTL:          30 * time.Second, // Mínimo 30 segundos
		MaxSize:         1000,
		CleanupInterval: 1 * time.Minute,
	}
}

// tokenMetadata armazena metadados do token
type tokenMetadata struct {
	token        string
	expiresAt    time.Time
	createdAt    time.Time
	xApplication string
}

// isExpired verifica se o token expirou
func (tm *tokenMetadata) isExpired() bool {
	return time.Now().After(tm.expiresAt)
}

// getRemainingTTL retorna o tempo restante até expiração
func (tm *tokenMetadata) getRemainingTTL() time.Duration {
	remaining := time.Until(tm.expiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// smartCache implementa cache inteligente com TTL baseado no token
type smartCache struct {
	tokens  map[string]*tokenMetadata
	results map[string]*cacheEntry // Cache de resultados de ValidateToken
	mu      sync.RWMutex
	config  SmartCacheConfig
	metrics *metrics.Metrics
}

// newSmartCache cria um novo cache inteligente
func newSmartCache(config SmartCacheConfig, metrics *metrics.Metrics) *smartCache {
	c := &smartCache{
		tokens:  make(map[string]*tokenMetadata),
		results: make(map[string]*cacheEntry),
		config:  config,
		metrics: metrics,
	}

	if config.CleanupInterval > 0 {
		go c.cleanup()
	}

	return c
}

// cleanup remove entradas expiradas periodicamente
func (c *smartCache) cleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()

		// Limpa tokens expirados
		for key, metadata := range c.tokens {
			if metadata.isExpired() {
				delete(c.tokens, key)
			}
		}

		// Limpa resultados expirados
		for key, entry := range c.results {
			if entry.isExpired() {
				delete(c.results, key)
			}
		}

		c.mu.Unlock()
	}
}

// storeTokenMetadata armazena metadados de um token
func (c *smartCache) storeTokenMetadata(token string, expiresIn int64, xApplication string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calcula TTL baseado no ExpiresIn do token
	ttl := time.Duration(expiresIn) * time.Second

	// Aplica limites de segurança
	if ttl > c.config.MaxTTL {
		ttl = c.config.MaxTTL
	}
	if ttl < c.config.MinTTL {
		ttl = c.config.MinTTL
	}

	key := c.generateTokenKey(token, xApplication)
	c.tokens[key] = &tokenMetadata{
		token:        token,
		expiresAt:    time.Now().Add(ttl),
		createdAt:    time.Now(),
		xApplication: xApplication,
	}
}

// getTokenMetadata obtém metadados de um token
func (c *smartCache) getTokenMetadata(token, xApplication string) (*tokenMetadata, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateTokenKey(token, xApplication)
	metadata, exists := c.tokens[key]

	if !exists || metadata.isExpired() {
		return nil, false
	}

	return metadata, true
}

// invalidateToken invalida um token específico
func (c *smartCache) invalidateToken(token, xApplication string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateTokenKey(token, xApplication)
	delete(c.tokens, key)
	delete(c.results, key)
}

// cacheValidationResult cacheia resultado de ValidateToken
func (c *smartCache) cacheValidationResult(token, xApplication string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.generateTokenKey(token, xApplication)

	// Busca metadados do token para usar TTL correto
	metadata, exists := c.tokens[key]
	var ttl time.Duration

	if exists && !metadata.isExpired() {
		// Usa o tempo restante do token
		ttl = metadata.getRemainingTTL()
	} else {
		// Se não tem metadados, usa TTL mínimo por segurança
		ttl = c.config.MinTTL
	}

	c.results[key] = &cacheEntry{
		value:     err,
		expiresAt: time.Now().Add(ttl),
	}
}

// getValidationResult obtém resultado cacheado de ValidateToken
func (c *smartCache) getValidationResult(token, xApplication string) (error, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.generateTokenKey(token, xApplication)
	entry, exists := c.results[key]

	if !exists || entry.isExpired() {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("auth-client", "validate-token")
		}
		return nil, false
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("auth-client", "validate-token")
	}

	if err, ok := entry.value.(error); ok {
		return err, true
	}
	return nil, true
}

// generateTokenKey gera chave de cache para o token
func (c *smartCache) generateTokenKey(token, xApplication string) string {
	data := token + ":" + xApplication
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// smartCacheDecorator implementa cache inteligente para AuthClient
type smartCacheDecorator struct {
	inner  authclient.AuthClient
	cache  *smartCache
	config SmartCacheConfig
}

// NewSmartCacheDecorator cria um decorator de cache inteligente
func NewSmartCacheDecorator(
	inner authclient.AuthClient,
	config SmartCacheConfig,
	metrics *metrics.Metrics,
) authclient.AuthClient {
	return &smartCacheDecorator{
		inner:  inner,
		cache:  newSmartCache(config, metrics),
		config: config,
	}
}

// Login - NÃO CACHEIA mas ARMAZENA metadados do token
func (d *smartCacheDecorator) Login(ctx context.Context, req inboundDto.AuthRequestDTO, xApplication, correlationID string) (inboundDto.AuthResponseDTO, error) {
	response, err := d.inner.Login(ctx, req, xApplication, correlationID)

	if err == nil {
		// Armazena metadados do token para uso futuro no cache
		d.cache.storeTokenMetadata(response.Token, response.ExpiresIn, xApplication)
	}

	return response, err
}

// RefreshToken - NÃO CACHEIA mas ARMAZENA metadados do token
func (d *smartCacheDecorator) RefreshToken(ctx context.Context, req inboundDto.RefreshTokenRequestDTO, xApplication, correlationID string) (inboundDto.RefreshTokenResponseDTO, error) {
	response, err := d.inner.RefreshToken(ctx, req, xApplication, correlationID)

	if err == nil {
		// Armazena metadados do novo token
		d.cache.storeTokenMetadata(response.Token, response.ExpiresIn, xApplication)
	}

	return response, err
}

// Logout - INVALIDA token do cache
func (d *smartCacheDecorator) Logout(ctx context.Context, token, xApplication, correlationID string) error {
	err := d.inner.Logout(ctx, token, xApplication, correlationID)

	// Invalida token (mesmo se logout falhar, por segurança)
	d.cache.invalidateToken(token, xApplication)

	return err
}

// ValidateToken - CACHEIA com TTL baseado no ExpiresIn do token
func (d *smartCacheDecorator) ValidateToken(ctx context.Context, token, xApplication, correlationID string) error {
	// Tenta obter do cache
	if cachedErr, found := d.cache.getValidationResult(token, xApplication); found {
		return cachedErr
	}

	// Se não encontrou no cache, executa validação
	err := d.inner.ValidateToken(ctx, token, xApplication, correlationID)

	// Armazena resultado no cache (com TTL baseado no token)
	d.cache.cacheValidationResult(token, xApplication, err)

	return err
}

// ChangePassword - NÃO CACHEIA
func (d *smartCacheDecorator) ChangePassword(ctx context.Context, req outboundDto.ChangePasswordMSRequestDTO, xApplication, correlationID string) error {
	return d.inner.ChangePassword(ctx, req, xApplication, correlationID)
}

// ResetPassword - NÃO CACHEIA
func (d *smartCacheDecorator) ResetPassword(ctx context.Context, req outboundDto.ResetPasswordMSRequestDTO, xApplication, correlationID string) error {
	return d.inner.ResetPassword(ctx, req, xApplication, correlationID)
}

// GenerateResetToken implementa o método GenerateResetToken (sem cache)
func (d *smartCacheDecorator) GenerateResetToken(ctx context.Context, req map[string]interface{}, xApplication, correlationID string) (outboundDto.GenerateResetTokenMSResponseDTO, error) {
	return d.inner.GenerateResetToken(ctx, req, xApplication, correlationID)
}
