package auth

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStringCache struct {
	mock.Mock
}

func (m *mockStringCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func TestRedisUserCacheDecorator_HitSkipsHTTP(t *testing.T) {
	inner := new(MockAuthClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:auth:codeuser:user-1").
		Return(`{"idUserExternal":"ext-1","codeUser":"user-1","companyId":"co-1"}`, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	ctx := authclient.WithCompanyID(context.Background(), "co-1")
	result, err := decorator.GetUserByCodeUser(ctx, "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "ext-1", result.ExternalID())
	inner.AssertNotCalled(t, "GetUserByCodeUser")
}

func TestRedisUserCacheDecorator_CompanyMismatchFallsBackToHTTP(t *testing.T) {
	inner := new(MockAuthClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:auth:codeuser:user-1").
		Return(`{"idUserExternal":"ext-1","companyId":"other-co"}`, nil)
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-http"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	ctx := authclient.WithCompanyID(context.Background(), "co-1")
	result, err := decorator.GetUserByCodeUser(ctx, "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "ext-http", result.ExternalID())
	inner.AssertExpectations(t)
}

func TestRedisUserCacheDecorator_MissFallsBackToHTTP(t *testing.T) {
	inner := new(MockAuthClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, mock.Anything).Return("", nil)
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-http"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "ext-http", result.ExternalID())
	inner.AssertExpectations(t)
}

func TestRedisUserCacheDecorator_RedisErrorFallsBackToHTTP(t *testing.T) {
	inner := new(MockAuthClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:auth:codeuser:user-1").
		Return("", errors.New("redis down"))
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-http"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "ext-http", result.ExternalID())
}

func TestRedisUserCacheDecorator_NilCacheUsesInner(t *testing.T) {
	inner := new(MockAuthClient)
	inner.On("GetUserByCodeUser", mock.Anything, "user-1", "tok", "tenant-1", "corr-1").
		Return(outboundDto.UserByCodeResponseDTO{IDUserExternal: "ext-http"}, nil)

	decorator := NewRedisUserCacheDecorator(inner, nil, nil, nil)
	result, err := decorator.GetUserByCodeUser(context.Background(), "user-1", "tok", "tenant-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "ext-http", result.ExternalID())
}
