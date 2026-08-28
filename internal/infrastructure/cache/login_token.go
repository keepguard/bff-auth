package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/keepguard/bff-auth/internal/pkg"
	"github.com/redis/go-redis/v9"
)

const loginTokenPrefix = "tokenlogin"

// LoginTokenKey replica a chave gravada pelo ms-auth: tokenlogin:{codeUser}:{jwt}
func LoginTokenKey(codeUser, token string) string {
	return fmt.Sprintf("%s:%s:%s", loginTokenPrefix, strings.ToLower(strings.TrimSpace(codeUser)), token)
}

// RedisLoginTokenChecker consulta tokenlogin no Redis.
type RedisLoginTokenChecker struct {
	client *redis.Client
}

func NewRedisLoginTokenChecker(redisClient *redis.Client) *RedisLoginTokenChecker {
	return &RedisLoginTokenChecker{client: redisClient}
}

func (c *RedisLoginTokenChecker) Check(ctx context.Context, token string) (client.LoginTokenPresence, string, error) {
	return LookupLoginToken(ctx, c.client, token)
}

// LookupLoginToken verifica EXISTS da chave de login. JWT malformado ou Redis fora devolve Unknown.
func LookupLoginToken(ctx context.Context, redisClient *redis.Client, token string) (client.LoginTokenPresence, string, error) {
	if redisClient == nil {
		return client.LoginTokenUnknown, "", nil
	}

	codeUser, err := pkg.ExtractCodeUserFromToken(token)
	if err != nil || codeUser == "" {
		return client.LoginTokenUnknown, "", nil
	}

	lookupCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	exists, err := redisClient.Exists(lookupCtx, LoginTokenKey(codeUser, token)).Result()
	if err != nil {
		return client.LoginTokenUnknown, codeUser, err
	}
	if exists == 0 {
		return client.LoginTokenAbsent, codeUser, nil
	}
	return client.LoginTokenPresent, codeUser, nil
}
