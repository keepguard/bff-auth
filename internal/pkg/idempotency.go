package pkg

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"
)

// IdempotencyManager gerencia chaves de idempotência
type IdempotencyManager interface {
	IsProcessed(ctx context.Context, key string) (bool, error)
	MarkAsProcessed(ctx context.Context, key string, ttl time.Duration) error
	GenerateKey(method, path, body string) string
}

// DefaultTTL é o TTL padrão para chaves de idempotência (24 horas)
const DefaultTTL = 24 * time.Hour

// NoopIdempotencyManager implementa IdempotencyManager sem cache (para BFF sem cache)
type NoopIdempotencyManager struct{}

// NewNoopIdempotencyManager cria um novo gerenciador de idempotência sem cache
func NewNoopIdempotencyManager() *NoopIdempotencyManager {
	return &NoopIdempotencyManager{}
}

// IsProcessed sempre retorna false (nunca foi processado)
func (m *NoopIdempotencyManager) IsProcessed(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// MarkAsProcessed não faz nada (sem cache)
func (m *NoopIdempotencyManager) MarkAsProcessed(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

// GenerateKey gera uma chave de idempotência baseada no método, path e body
func (m *NoopIdempotencyManager) GenerateKey(method, path, body string) string {
	data := fmt.Sprintf("%s:%s:%s", method, path, body)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("idempotency:%x", hash)
}
