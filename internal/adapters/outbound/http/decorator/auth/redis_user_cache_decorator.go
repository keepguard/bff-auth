package auth

import (
	"context"
	"encoding/json"
	"strings"

	companydecorator "github.com/keepguard/bff-auth/internal/adapters/outbound/http/decorator/company"
	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/infrastructure/metrics"
	"go.uber.org/zap"
)

const authUserByCodeCachePrefix = "user_cache:auth:codeuser:"

type redisUserCacheDecorator struct {
	authclient.AuthClient
	cache   companydecorator.StringValueCache
	metrics *metrics.Metrics
	logger  *zap.Logger
}

// NewRedisUserCacheDecorator lê user_cache:auth:codeuser:{codeUser} (gravado pelo ms-auth) antes do HTTP.
func NewRedisUserCacheDecorator(
	inner authclient.AuthClient,
	cache companydecorator.StringValueCache,
	metricsInstance *metrics.Metrics,
	logger *zap.Logger,
) authclient.AuthClient {
	if cache == nil {
		return inner
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &redisUserCacheDecorator{
		AuthClient: inner,
		cache:      cache,
		metrics:    metricsInstance,
		logger:     logger,
	}
}

func (d *redisUserCacheDecorator) GetUserByCodeUser(ctx context.Context, codeUser, token, tenantId, correlationID string) (outboundDto.UserByCodeResponseDTO, error) {
	if codeUser != "" {
		if user, ok := d.lookup(ctx, codeUser); ok {
			d.recordHit()
			return user, nil
		}
	}

	d.recordMiss()
	return d.AuthClient.GetUserByCodeUser(ctx, codeUser, token, tenantId, correlationID)
}

func (d *redisUserCacheDecorator) lookup(ctx context.Context, codeUser string) (outboundDto.UserByCodeResponseDTO, bool) {
	key := authUserByCodeCachePrefix + strings.ToLower(strings.TrimSpace(codeUser))
	val, err := d.cache.Get(ctx, key)
	if err != nil {
		d.logger.Warn("Falha ao ler usuário no Redis; fallback HTTP",
			zap.String("key", key),
			zap.Error(err),
		)
		return outboundDto.UserByCodeResponseDTO{}, false
	}
	if val == "" {
		return outboundDto.UserByCodeResponseDTO{}, false
	}

	var user outboundDto.UserByCodeResponseDTO
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return outboundDto.UserByCodeResponseDTO{}, false
	}
	if user.ExternalID() == "" && user.CodeUser == "" && user.ID == "" {
		return outboundDto.UserByCodeResponseDTO{}, false
	}

	if companyID := authclient.CompanyIDFromContext(ctx); companyID != "" && user.CompanyID != "" && user.CompanyID != companyID {
		return outboundDto.UserByCodeResponseDTO{}, false
	}

	return user, true
}

func (d *redisUserCacheDecorator) recordHit() {
	if d.metrics != nil {
		d.metrics.RecordCacheHit("auth-user-redis", "codeUser")
	}
}

func (d *redisUserCacheDecorator) recordMiss() {
	if d.metrics != nil {
		d.metrics.RecordCacheMiss("auth-user-redis", "codeUser")
	}
}
