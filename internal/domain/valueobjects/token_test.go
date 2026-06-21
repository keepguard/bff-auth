package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewToken(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple token",
			value:   "simple-token-123",
			wantErr: false,
		},
		{
			name:    "valid JWT token",
			value:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: false,
		},
		{
			name:    "empty token",
			value:   "",
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "whitespace only",
			value:   "   ",
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "too short",
			value:   "short",
			wantErr: true,
			errMsg:  "token must be at least 10 characters long",
		},
		{
			name:    "too long",
			value:   generateLongString(8193),
			wantErr: true,
			errMsg:  "token is too long",
		},
		{
			name:    "invalid JWT format - missing parts",
			value:   "invalid.jwt",
			wantErr: false, // Agora aceita tokens simples
		},
		{
			name:    "invalid JWT format - empty part",
			value:   "header..signature",
			wantErr: true,
			errMsg:  "invalid JWT format: part 2 is empty",
		},
		{
			name:    "simple token with invalid characters",
			value:   "token with spaces",
			wantErr: true,
			errMsg:  "token contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := NewToken(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.True(t, token.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, token.IsEmpty())
				assert.Equal(t, tt.value, token.Value())
			}
		})
	}
}

func TestToken_String(t *testing.T) {
	token, _ := NewToken("simple-token-123")
	assert.Equal(t, "simple-token-123", token.String())
}

func TestToken_Value(t *testing.T) {
	token, _ := NewToken("simple-token-123")
	assert.Equal(t, "simple-token-123", token.Value())
}

func TestToken_Equals(t *testing.T) {
	token1, _ := NewToken("valid-token-123456")
	token2, _ := NewToken("valid-token-123456")
	token3, _ := NewToken("different-token-456789")

	assert.True(t, token1.Equals(token2))
	assert.False(t, token1.Equals(token3))
}

func TestToken_IsEmpty(t *testing.T) {
	token, _ := NewToken("valid-token-123456")
	assert.False(t, token.IsEmpty())

	emptyToken, _ := NewToken("")
	assert.True(t, emptyToken.IsEmpty())
}

func TestToken_IsJWT(t *testing.T) {
	// JWT token
	jwtToken, _ := NewToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	assert.True(t, jwtToken.IsJWT())

	// Simple token
	simpleToken, _ := NewToken("simple-token-123")
	assert.False(t, simpleToken.IsJWT())
}

func TestToken_GetJWTHeader(t *testing.T) {
	// JWT token
	jwtToken, _ := NewToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	header := jwtToken.GetJWTHeader()
	assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", header)

	// Simple token
	simpleToken, _ := NewToken("simple-token-123")
	header = simpleToken.GetJWTHeader()
	assert.Equal(t, "", header)
}

func TestToken_GetJWTPayload(t *testing.T) {
	// JWT token
	jwtToken, _ := NewToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	payload := jwtToken.GetJWTPayload()
	assert.Equal(t, "eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ", payload)

	// Simple token
	simpleToken, _ := NewToken("simple-token-123")
	payload = simpleToken.GetJWTPayload()
	assert.Equal(t, "", payload)
}

func TestToken_GetJWTSignature(t *testing.T) {
	// JWT token
	jwtToken, _ := NewToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")
	signature := jwtToken.GetJWTSignature()
	assert.Equal(t, "SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c", signature)

	// Simple token
	simpleToken, _ := NewToken("simple-token-123")
	signature = simpleToken.GetJWTSignature()
	assert.Equal(t, "", signature)
}
