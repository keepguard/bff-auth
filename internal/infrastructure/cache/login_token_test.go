package cache

import (
	"context"
	"testing"

	"github.com/keepguard/bff-auth/internal/domain/ports/client"
	"github.com/stretchr/testify/assert"
)

func TestLoginTokenKey(t *testing.T) {
	assert.Equal(t, "tokenlogin:user-1:jwt-abc", LoginTokenKey("user-1", "jwt-abc"))
}

func TestLookupLoginToken_NilClient(t *testing.T) {
	presence, codeUser, err := LookupLoginToken(context.Background(), nil, "any")
	assert.Equal(t, client.LoginTokenUnknown, presence)
	assert.Empty(t, codeUser)
	assert.NoError(t, err)
}

func TestLookupLoginToken_MalformedJWT(t *testing.T) {
	checker := NewRedisLoginTokenChecker(nil)
	presence, _, err := checker.Check(context.Background(), "not-a-jwt")
	assert.Equal(t, client.LoginTokenUnknown, presence)
	assert.NoError(t, err)
}
