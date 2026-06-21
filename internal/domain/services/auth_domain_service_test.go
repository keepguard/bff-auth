package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAuthDomainService(t *testing.T) {
	service := NewAuthDomainService()
	assert.NotNil(t, service)
	assert.IsType(t, &authDomainServiceImpl{}, service)
}

func TestAuthDomainService_ValidateTokenFormat(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid JWT token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: false,
		},
		{
			name:    "valid simple token",
			token:   "simple-token-123",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "token cannot be empty",
		},
		{
			name:    "invalid JWT format",
			token:   "invalid.jwt",
			wantErr: false, // Agora aceita tokens simples
		},
		{
			name:    "token too long",
			token:   generateLongString(8193),
			wantErr: true,
			errMsg:  "token is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTokenFormat(tt.token)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_ValidateLoginCredentials(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			password: "password123",
			wantErr:  true,
			errMsg:   "username is required",
		},
		{
			name:     "empty password",
			username: "testuser",
			password: "",
			wantErr:  true,
			errMsg:   "password is required",
		},
		{
			name:     "username too long",
			username: generateLongString(256),
			password: "password123",
			wantErr:  true,
			errMsg:   "username is too long",
		},
		{
			name:     "password too long",
			username: "testuser",
			password: generateLongString(129),
			wantErr:  true,
			errMsg:   "password is too long",
		},
		{
			name:     "username with invalid characters",
			username: "test@user",
			password: "password123",
			wantErr:  true,
			errMsg:   "username contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateLoginCredentials(tt.username, tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_ShouldRefreshToken(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name      string
		expiresIn int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "token with good expiration",
			expiresIn: 3600, // 1 hora
			wantErr:   false,
		},
		{
			name:      "token about to expire",
			expiresIn: 200, // menos de 5 minutos
			wantErr:   true,
			errMsg:    "token is about to expire",
		},
		{
			name:      "expired token",
			expiresIn: -100,
			wantErr:   true,
			errMsg:    "invalid expiration time",
		},
		{
			name:      "invalid expiration",
			expiresIn: 0,
			wantErr:   true,
			errMsg:    "invalid expiration time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ShouldRefreshToken(tt.expiresIn)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_ValidateTokenExpiration(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name      string
		expiresIn int
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid expiration",
			expiresIn: 3600, // 1 hora
			wantErr:   false,
		},
		{
			name:      "expires too soon",
			expiresIn: 30, // menos de 1 minuto
			wantErr:   true,
			errMsg:    "token expires too soon",
		},
		{
			name:      "expires too late",
			expiresIn: 100000, // mais de 24 horas
			wantErr:   true,
			errMsg:    "token expires too late",
		},
		{
			name:      "invalid expiration",
			expiresIn: 0,
			wantErr:   true,
			errMsg:    "invalid expiration time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateTokenExpiration(tt.expiresIn)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_CalculateTokenTTL(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name      string
		tokenType string
		wantTTL   int
		wantErr   bool
	}{
		{
			name:      "access token",
			tokenType: "access",
			wantTTL:   3600,
			wantErr:   false,
		},
		{
			name:      "refresh token",
			tokenType: "refresh",
			wantTTL:   86400,
			wantErr:   false,
		},
		{
			name:      "password reset token",
			tokenType: "password_reset",
			wantTTL:   1800,
			wantErr:   false,
		},
		{
			name:      "email verification token",
			tokenType: "email_verification",
			wantTTL:   3600,
			wantErr:   false,
		},
		{
			name:      "api token",
			tokenType: "api",
			wantTTL:   7200,
			wantErr:   false,
		},
		{
			name:      "unknown token type",
			tokenType: "unknown",
			wantTTL:   0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ttl, err := service.CalculateTokenTTL(tt.tokenType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTTL, ttl)
			}
		})
	}
}

func TestAuthDomainService_ValidateRefreshToken(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name         string
		refreshToken string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid refresh token",
			refreshToken: "valid-refresh-token-12345678901234567890123",
			wantErr:      false,
		},
		{
			name:         "empty refresh token",
			refreshToken: "",
			wantErr:      true,
			errMsg:       "refresh token is required",
		},
		{
			name:         "refresh token too short",
			refreshToken: "short",
			wantErr:      true,
			errMsg:       "refresh token is too short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateRefreshToken(tt.refreshToken)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_GenerateSessionID(t *testing.T) {
	service := NewAuthDomainService()

	sessionID1 := service.GenerateSessionID()
	sessionID2 := service.GenerateSessionID()

	assert.NotEmpty(t, sessionID1)
	assert.NotEmpty(t, sessionID2)
	assert.NotEqual(t, sessionID1, sessionID2) // Deve ser único
	assert.Contains(t, sessionID1, "session_")
}

func TestAuthDomainService_ValidateSessionTimeout(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name            string
		sessionDuration time.Duration
		wantErr         bool
		errMsg          string
	}{
		{
			name:            "valid session duration",
			sessionDuration: 30 * time.Minute,
			wantErr:         false,
		},
		{
			name:            "session too short",
			sessionDuration: 1 * time.Minute,
			wantErr:         true,
			errMsg:          "session duration is too short",
		},
		{
			name:            "session too long",
			sessionDuration: 48 * time.Hour,
			wantErr:         true,
			errMsg:          "session duration is too long",
		},
		{
			name:            "zero duration",
			sessionDuration: 0,
			wantErr:         true,
			errMsg:          "session duration must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSessionTimeout(tt.sessionDuration)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthDomainService_ShouldLogoutUser(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name         string
		lastActivity time.Duration
		wantLogout   bool
	}{
		{
			name:         "user recently active",
			lastActivity: 30 * time.Minute,
			wantLogout:   false,
		},
		{
			name:         "user inactive for 3 hours",
			lastActivity: 3 * time.Hour,
			wantLogout:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldLogout := service.ShouldLogoutUser(tt.lastActivity)
			assert.Equal(t, tt.wantLogout, shouldLogout)
		})
	}
}

func TestAuthDomainService_ValidateXApplication(t *testing.T) {
	service := NewAuthDomainService()

	tests := []struct {
		name         string
		xApplication string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid x-application",
			xApplication: "my-app",
			wantErr:      false,
		},
		{
			name:         "empty x-application",
			xApplication: "",
			wantErr:      true,
			errMsg:       "x-application header is required",
		},
		{
			name:         "x-application too short",
			xApplication: "ab",
			wantErr:      true,
			errMsg:       "x-application must have at least 3 characters",
		},
		{
			name:         "x-application too long",
			xApplication: generateLongString(51),
			wantErr:      true,
			errMsg:       "x-application must have at most 50 characters",
		},
		{
			name:         "x-application with invalid characters",
			xApplication: "my@app",
			wantErr:      true,
			errMsg:       "x-application contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateXApplication(tt.xApplication)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to generate long strings
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}
