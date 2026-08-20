package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurityDomainService(t *testing.T) {
	service := NewSecurityDomainService()
	assert.NotNil(t, service)
	assert.IsType(t, &securityDomainServiceImpl{}, service)
}

func TestSecurityDomainService_ValidatePasswordStrength(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "strong password",
			password: "MyComplexP@ssw0rd2024!",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  true,
			errMsg:   "password is required",
		},
		{
			name:     "password too short",
			password: "My123!",
			wantErr:  true,
			errMsg:   "password must have at least 8 characters",
		},
		{
			name:     "password too long",
			password: generateLongString(129),
			wantErr:  true,
			errMsg:   "password must have at most 128 characters",
		},
		{
			name:     "password without lowercase",
			password: "MYSECURE123!",
			wantErr:  true,
			errMsg:   "password must contain at least one lowercase letter",
		},
		{
			name:     "password without uppercase",
			password: "mysecure123!",
			wantErr:  true,
			errMsg:   "password must contain at least one uppercase letter",
		},
		{
			name:     "password without number",
			password: "MySecurePassword!",
			wantErr:  true,
			errMsg:   "password must contain at least one number",
		},
		{
			name:     "password without special character",
			password: "MySecure123",
			wantErr:  true,
			errMsg:   "password must contain at least one special character",
		},
		{
			name:     "password with common sequences",
			password: "Password123!",
			wantErr:  true,
			errMsg:   "password contains common sequences",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePasswordStrength(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityDomainService_GenerateSecureToken(t *testing.T) {
	service := NewSecurityDomainService()

	token1 := service.GenerateSecureToken()
	token2 := service.GenerateSecureToken()

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2) // Deve ser único
	assert.Len(t, token1, 64)          // 32 bytes em hex = 64 caracteres
}

func TestSecurityDomainService_ValidatePasswordHistory(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name     string
		password string
		history  []string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "new password",
			password: "NewPassword123!",
			history:  []string{"OldPassword123!", "AnotherPassword123!"},
			wantErr:  false,
		},
		{
			name:     "password in history",
			password: "OldPassword123!",
			history:  []string{"OldPassword123!", "AnotherPassword123!"},
			wantErr:  true,
			errMsg:   "password was used recently",
		},
		{
			name:     "empty history",
			password: "AnyPassword123!",
			history:  []string{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePasswordHistory(tt.password, tt.history)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityDomainService_CalculatePasswordEntropy(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name       string
		password   string
		minEntropy float64
	}{
		{
			name:       "simple password",
			password:   "password",
			minEntropy: 20.0,
		},
		{
			name:       "complex password",
			password:   "MySecure123!@#",
			minEntropy: 50.0,
		},
		{
			name:       "empty password",
			password:   "",
			minEntropy: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entropy := service.CalculatePasswordEntropy(tt.password)
			assert.GreaterOrEqual(t, entropy, tt.minEntropy)
		})
	}
}

func TestSecurityDomainService_ShouldRequirePasswordChange(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name         string
		lastChange   time.Time
		shouldChange bool
	}{
		{
			name:         "recent password change",
			lastChange:   time.Now().Add(-30 * 24 * time.Hour), // 30 dias atrás
			shouldChange: false,
		},
		{
			name:         "old password",
			lastChange:   time.Now().Add(-100 * 24 * time.Hour), // 100 dias atrás
			shouldChange: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldChange := service.ShouldRequirePasswordChange(tt.lastChange)
			assert.Equal(t, tt.shouldChange, shouldChange)
		})
	}
}

func TestSecurityDomainService_ValidatePasswordComplexity(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "complex password",
			password: "MySecure123!",
			wantErr:  false,
		},
		{
			name:     "password without lowercase",
			password: "MYSECURE123!",
			wantErr:  true,
			errMsg:   "password must contain at least one lowercase letter",
		},
		{
			name:     "password without uppercase",
			password: "mysecure123!",
			wantErr:  true,
			errMsg:   "password must contain at least one uppercase letter",
		},
		{
			name:     "password without number",
			password: "MySecurePassword!",
			wantErr:  true,
			errMsg:   "password must contain at least one number",
		},
		{
			name:     "password without special character",
			password: "MySecure123",
			wantErr:  true,
			errMsg:   "password must contain at least one special character",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePasswordComplexity(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityDomainService_GenerateSecureRandomString(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "short string",
			length: 8,
		},
		{
			name:   "medium string",
			length: 16,
		},
		{
			name:   "long string",
			length: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str1 := service.GenerateSecureRandomString(tt.length)
			str2 := service.GenerateSecureRandomString(tt.length)

			assert.Len(t, str1, tt.length)
			assert.Len(t, str2, tt.length)
			assert.NotEqual(t, str1, str2) // Deve ser único
		})
	}
}

func TestSecurityDomainService_ValidateSecurityHeaders(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name    string
		headers map[string]string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid headers",
			headers: map[string]string{
				"X-Tenant-Id":    "my-app",
				"X-Correlation-ID": "correlation-123",
			},
			wantErr: false,
		},
		{
			name: "missing X-Tenant-Id",
			headers: map[string]string{
				"X-Correlation-ID": "correlation-123",
			},
			wantErr: true,
			errMsg:  "missing required security header: X-Tenant-Id",
		},
		{
			name: "missing X-Correlation-ID",
			headers: map[string]string{
				"X-Tenant-Id": "my-app",
			},
			wantErr: true,
			errMsg:  "missing required security header: X-Correlation-ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSecurityHeaders(tt.headers)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityDomainService_ShouldBlockUser(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name        string
		attempts    int
		lastAttempt time.Time
		shouldBlock bool
	}{
		{
			name:        "few attempts, recent",
			attempts:    2,
			lastAttempt: time.Now().Add(-5 * time.Minute),
			shouldBlock: false,
		},
		{
			name:        "many attempts, recent",
			attempts:    6,
			lastAttempt: time.Now().Add(-5 * time.Minute),
			shouldBlock: true,
		},
		{
			name:        "many attempts, old",
			attempts:    6,
			lastAttempt: time.Now().Add(-30 * time.Minute),
			shouldBlock: false,
		},
		{
			name:        "too many attempts, recent",
			attempts:    12,
			lastAttempt: time.Now().Add(-30 * time.Minute),
			shouldBlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldBlock := service.ShouldBlockUser(tt.attempts, tt.lastAttempt)
			assert.Equal(t, tt.shouldBlock, shouldBlock)
		})
	}
}

func TestSecurityDomainService_CalculatePasswordScore(t *testing.T) {
	service := NewSecurityDomainService()

	tests := []struct {
		name     string
		password string
		minScore int
	}{
		{
			name:     "weak password",
			password: "password",
			minScore: 0,
		},
		{
			name:     "medium password",
			password: "Password123",
			minScore: 50,
		},
		{
			name:     "strong password",
			password: "MySecure123!@#",
			minScore: 70,
		},
		{
			name:     "empty password",
			password: "",
			minScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.CalculatePasswordScore(tt.password)
			assert.GreaterOrEqual(t, score, tt.minScore)
			assert.LessOrEqual(t, score, 100) // Score máximo é 100
		})
	}
}
