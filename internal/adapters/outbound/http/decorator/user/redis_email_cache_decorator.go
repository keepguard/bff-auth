package user

import (
	"context"
	"encoding/json"
	"strings"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	companydecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/company"
	portsclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

const authUserByEmailCachePrefix = "user_cache:auth:email:"

type redisEmailCacheDecorator struct {
	portsclient.UserClient
	cache   companydecorator.StringValueCache
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewRedisEmailCacheDecorator lê user_cache:auth:email:{companyId}:{email} antes do HTTP ao ms-auth.
func NewRedisEmailCacheDecorator(
	inner portsclient.UserClient,
	cache companydecorator.StringValueCache,
	metricsInstance *metrics.Metrics,
	logger *zap.Logger,
) portsclient.UserClient {
	if cache == nil {
		return inner
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &redisEmailCacheDecorator{
		UserClient: inner,
		cache:      cache,
		metrics:    metricsInstance,
		logger:     logger,
	}
}

func (d *redisEmailCacheDecorator) GetByEmail(ctx context.Context, email, tenantId, companyId, correlationID string) (outboundDto.UserByEmailResponseDTO, error) {
	if email != "" && companyId != "" {
		if user, ok := d.lookup(ctx, companyId, email); ok {
			d.recordHit()
			return user, nil
		}
	}

	d.recordMiss()
	return d.UserClient.GetByEmail(ctx, email, tenantId, companyId, correlationID)
}

func (d *redisEmailCacheDecorator) lookup(ctx context.Context, companyId, email string) (outboundDto.UserByEmailResponseDTO, bool) {
	key := authUserByEmailCachePrefix + companyId + ":" + strings.ToLower(strings.TrimSpace(email))
	val, err := d.cache.Get(ctx, key)
	if err != nil {
		d.logger.Warn("Falha ao ler usuário por email no Redis; fallback HTTP",
			zap.String("key", key),
			zap.Error(err),
		)
		return outboundDto.UserByEmailResponseDTO{}, false
	}
	if val == "" {
		return outboundDto.UserByEmailResponseDTO{}, false
	}

	var user outboundDto.UserByEmailResponseDTO
	if err := json.Unmarshal([]byte(val), &user); err != nil || (user.Email == "" && user.CodeUser == "") {
		return outboundDto.UserByEmailResponseDTO{}, false
	}
	return user, true
}

func (d *redisEmailCacheDecorator) recordHit() {
	if d.metrics != nil {
		d.metrics.RecordCacheHit("auth-user-redis", "email")
	}
}

func (d *redisEmailCacheDecorator) recordMiss() {
	if d.metrics != nil {
		d.metrics.RecordCacheMiss("auth-user-redis", "email")
	}
}
