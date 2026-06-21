package company

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
)

// CacheConfig configuração para cache do CompanyClient
type CacheConfig struct {
	TTL             time.Duration // Tempo de vida do cache
	MaxSize         int           // Número máximo de entradas
	CleanupInterval time.Duration // Intervalo de limpeza
}

// DefaultCacheConfig retorna configuração padrão
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:             5 * time.Minute,
		MaxSize:         1000,
		CleanupInterval: 1 * time.Minute,
	}
}

// cacheEntry representa uma entrada no cache
type cacheEntry struct {
	value     portsclient.CompanySimpleResponseDTO
	expiresAt time.Time
}

// isExpired verifica se a entrada está expirada
func (e *cacheEntry) isExpired() bool {
	return time.Now().After(e.expiresAt)
}

// companyCache implementa cache thread-safe para empresas
type companyCache struct {
	entries map[string]*cacheEntry
	mu      sync.RWMutex
	config  CacheConfig
	metrics *metrics.Metrics
}

// newCompanyCache cria um novo cache
func newCompanyCache(config CacheConfig, metrics *metrics.Metrics) *companyCache {
	c := &companyCache{
		entries: make(map[string]*cacheEntry),
		config:  config,
		metrics: metrics,
	}

	if config.CleanupInterval > 0 {
		go c.cleanup()
	}

	return c
}

// cleanup remove entradas expiradas periodicamente
func (c *companyCache) cleanup() {
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
func (c *companyCache) get(key string) (portsclient.CompanySimpleResponseDTO, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("company-client", "company")
		}
		return portsclient.CompanySimpleResponseDTO{}, false
	}

	if entry.isExpired() {
		if c.metrics != nil {
			c.metrics.RecordCacheMiss("company-client", "company")
		}
		return portsclient.CompanySimpleResponseDTO{}, false
	}

	if c.metrics != nil {
		c.metrics.RecordCacheHit("company-client", "company")
	}
	return entry.value, true
}

// set armazena um valor no cache
func (c *companyCache) set(key string, value portsclient.CompanySimpleResponseDTO) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Se atingiu o tamanho máximo, remove entrada mais antiga
	if len(c.entries) >= c.config.MaxSize {
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

// generateCacheKey gera chave de cache para xApplication
func generateCacheKey(xApplication string) string {
	hash := sha256.Sum256([]byte(xApplication))
	return hex.EncodeToString(hash[:])
}

// cacheDecorator implementa cache para CompanyClient
type cacheDecorator struct {
	inner  portsclient.CompanyClient
	cache  *companyCache
	config CacheConfig
}

// NewCacheDecorator cria um decorator de cache para CompanyClient
func NewCacheDecorator(
	inner portsclient.CompanyClient,
	config CacheConfig,
	metrics *metrics.Metrics,
) portsclient.CompanyClient {
	return &cacheDecorator{
		inner:  inner,
		cache:  newCompanyCache(config, metrics),
		config: config,
	}
}

// GetByXApplication implementa GetByXApplication com cache
func (d *cacheDecorator) GetByXApplication(ctx context.Context, xApplication, correlationID string) (portsclient.CompanySimpleResponseDTO, error) {
	cacheKey := generateCacheKey(xApplication)

	// Tenta obter do cache
	if cached, found := d.cache.get(cacheKey); found {
		return cached, nil
	}

	// Se não encontrou no cache, executa a operação
	response, err := d.inner.GetByXApplication(ctx, xApplication, correlationID)

	if err == nil {
		// Armazena no cache apenas se sucesso
		d.cache.set(cacheKey, response)
	}

	return response, err
}
