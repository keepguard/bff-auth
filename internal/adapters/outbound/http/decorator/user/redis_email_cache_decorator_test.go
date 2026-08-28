package user

import (
	"context"
	"errors"
	"testing"

	outboundDto "github.com/keepguard/bff-auth/internal/adapters/outbound/http/dto"
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

func TestRedisEmailCacheDecorator_HitSkipsHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:auth:email:co-1:user@example.com").
		Return(`{"id":"id-1","codeUser":"user-1","email":"user@example.com","status":"ACTIVE"}`, nil)

	decorator := NewRedisEmailCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "user@example.com", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", result.Email)
	inner.AssertNotCalled(t, "GetByEmail")
}

func TestRedisEmailCacheDecorator_MissFallsBackToHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, mock.Anything).Return("", nil)
	inner.On("GetByEmail", mock.Anything, "user@example.com", "tenant-1", "co-1", "corr-1").
		Return(outboundDto.UserByEmailResponseDTO{Email: "http@example.com"}, nil)

	decorator := NewRedisEmailCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "user@example.com", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@example.com", result.Email)
	inner.AssertExpectations(t)
}

func TestRedisEmailCacheDecorator_RedisErrorFallsBackToHTTP(t *testing.T) {
	inner := new(MockUserClient)
	cache := new(mockStringCache)
	cache.On("Get", mock.Anything, "user_cache:auth:email:co-1:user@example.com").
		Return("", errors.New("redis down"))
	inner.On("GetByEmail", mock.Anything, "user@example.com", "tenant-1", "co-1", "corr-1").
		Return(outboundDto.UserByEmailResponseDTO{Email: "http@example.com"}, nil)

	decorator := NewRedisEmailCacheDecorator(inner, cache, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "user@example.com", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@example.com", result.Email)
}

func TestRedisEmailCacheDecorator_NilCacheUsesInner(t *testing.T) {
	inner := new(MockUserClient)
	inner.On("GetByEmail", mock.Anything, "user@example.com", "tenant-1", "co-1", "corr-1").
		Return(outboundDto.UserByEmailResponseDTO{Email: "http@example.com"}, nil)

	decorator := NewRedisEmailCacheDecorator(inner, nil, nil, nil)
	result, err := decorator.GetByEmail(context.Background(), "user@example.com", "tenant-1", "co-1", "corr-1")

	assert.NoError(t, err)
	assert.Equal(t, "http@example.com", result.Email)
}
