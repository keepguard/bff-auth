package entities

import (
	"testing"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	tests := []struct {
		name          string
		id            valueobjects.SessionID
		userID        valueobjects.UserID
		accessToken   valueobjects.Token
		refreshToken  valueobjects.Token
		expiresIn     int
		ipAddress     string
		userAgent     string
		wantErr       bool
		errMsg        string
	}{
		{
			name:         "valid session",
			id:           func() valueobjects.SessionID { id, _ := valueobjects.NewSessionID("session-123"); return id }(),
			userID:       func() valueobjects.UserID { id, _ := valueobjects.NewUserID("user-123"); return id }(),
			accessToken:  func() valueobjects.Token { token, _ := valueobjects.NewToken("access-token-123"); return token }(),
			refreshToken: func() valueobjects.Token { token, _ := valueobjects.NewToken("refresh-token-123"); return token }(),
			expiresIn:    3600,
			ipAddress:    "192.168.1.1",
			userAgent:    "Mozilla/5.0",
			wantErr:      false,
		},
		{
			name:         "empty session ID",
			id:           func() valueobjects.SessionID { id, _ := valueobjects.NewSessionID(""); return id }(),
			userID:       func() valueobjects.UserID { id, _ := valueobjects.NewUserID("user-123"); return id }(),
			accessToken:  func() valueobjects.Token { token, _ := valueobjects.NewToken("access-token-123"); return token }(),
			refreshToken: func() valueobjects.Token { token, _ := valueobjects.NewToken("refresh-token-123"); return token }(),
			expiresIn:    3600,
			ipAddress:    "192.168.1.1",
			userAgent:    "Mozilla/5.0",
			wantErr:      true,
			errMsg:       "session ID cannot be empty",
		},
		{
			name:         "empty user ID",
			id:           func() valueobjects.SessionID { id, _ := valueobjects.NewSessionID("session-123"); return id }(),
			userID:       func() valueobjects.UserID { id, _ := valueobjects.NewUserID(""); return id }(),
			accessToken:  func() valueobjects.Token { token, _ := valueobjects.NewToken("access-token-123"); return token }(),
			refreshToken: func() valueobjects.Token { token, _ := valueobjects.NewToken("refresh-token-123"); return token }(),
			expiresIn:    3600,
			ipAddress:    "192.168.1.1",
			userAgent:    "Mozilla/5.0",
			wantErr:      true,
			errMsg:       "user ID cannot be empty",
		},
		{
			name:         "empty access token",
			id:           func() valueobjects.SessionID { id, _ := valueobjects.NewSessionID("session-123"); return id }(),
			userID:       func() valueobjects.UserID { id, _ := valueobjects.NewUserID("user-123"); return id }(),
			accessToken:  func() valueobjects.Token { token, _ := valueobjects.NewToken(""); return token }(),
			refreshToken: func() valueobjects.Token { token, _ := valueobjects.NewToken("refresh-token-123"); return token }(),
			expiresIn:    3600,
			ipAddress:    "192.168.1.1",
			userAgent:    "Mozilla/5.0",
			wantErr:      true,
			errMsg:       "access token cannot be empty",
		},
		{
			name:         "invalid expiration time",
			id:           func() valueobjects.SessionID { id, _ := valueobjects.NewSessionID("session-123"); return id }(),
			userID:       func() valueobjects.UserID { id, _ := valueobjects.NewUserID("user-123"); return id }(),
			accessToken:  func() valueobjects.Token { token, _ := valueobjects.NewToken("access-token-123"); return token }(),
			refreshToken: func() valueobjects.Token { token, _ := valueobjects.NewToken("refresh-token-123"); return token }(),
			expiresIn:    0,
			ipAddress:    "192.168.1.1",
			userAgent:    "Mozilla/5.0",
			wantErr:      true,
			errMsg:       "expiration time must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := NewSession(tt.id, tt.userID, tt.accessToken, tt.refreshToken, tt.expiresIn, tt.ipAddress, tt.userAgent)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, session)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, session)
				assert.Equal(t, tt.id, session.ID())
				assert.Equal(t, tt.userID, session.UserID())
				assert.Equal(t, tt.accessToken, session.AccessToken())
				assert.Equal(t, tt.refreshToken, session.RefreshToken())
				assert.Equal(t, tt.ipAddress, session.IPAddress())
				assert.Equal(t, tt.userAgent, session.UserAgent())
				assert.True(t, session.IsActive())
				assert.False(t, session.IsExpired())
			}
		})
	}
}

func TestSession_Refresh(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	newAccessToken, _ := valueobjects.NewToken("new-access-token-123")
	err := session.Refresh(newAccessToken, 7200)
	
	assert.NoError(t, err)
	assert.Equal(t, newAccessToken, session.AccessToken())
	assert.True(t, session.CanBeRefreshed())
}

func TestSession_Refresh_Invalid(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	// Teste com token vazio
	emptyToken, _ := valueobjects.NewToken("")
	err := session.Refresh(emptyToken, 7200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token cannot be empty")
	
	// Teste com tempo de expiração inválido
	validToken, _ := valueobjects.NewToken("valid-token")
	err = session.Refresh(validToken, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expiration time must be positive")
	
	// Teste com sessão expirada
	expiredSession := &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	err = expiredSession.Refresh(validToken, 7200)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot refresh expired session")
}

func TestSession_Terminate(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	session.Terminate()
	
	assert.False(t, session.IsActive())
	assert.False(t, session.CanBeRefreshed())
}

func TestSession_UpdateActivity(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	initialActivity := session.LastActivity()
	time.Sleep(10 * time.Millisecond) // Pequeno delay para garantir diferença
	
	session.UpdateActivity()
	
	assert.True(t, session.LastActivity().After(initialActivity))
}

func TestSession_UpdateLocation(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	newIP := "10.0.0.1"
	newUserAgent := "Chrome/91.0"
	
	session.UpdateLocation(newIP, newUserAgent)
	
	assert.Equal(t, newIP, session.IPAddress())
	assert.Equal(t, newUserAgent, session.UserAgent())
}

func TestSession_IsExpired(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	// Sessão válida
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	assert.False(t, session.IsExpired())
	
	// Sessão expirada
	expiredSession := &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	assert.True(t, expiredSession.IsExpired())
}

func TestSession_IsIdle(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	idleTimeout := 1 * time.Second
	assert.False(t, session.IsIdle(idleTimeout))
	
	time.Sleep(2 * time.Second)
	assert.True(t, session.IsIdle(idleTimeout))
}

func TestSession_ShouldBeTerminated(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	maxDuration := 1 * time.Hour
	idleTimeout := 30 * time.Minute
	
	// Sessão válida
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	assert.False(t, session.ShouldBeTerminated(maxDuration, idleTimeout))
	
	// Sessão expirada
	session = &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	assert.True(t, session.ShouldBeTerminated(maxDuration, idleTimeout))
}

func TestSession_GetRemainingTime(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	// Sessão válida
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	remaining := session.GetRemainingTime()
	assert.True(t, remaining > 0)
	assert.True(t, remaining <= 3600*time.Second)
	
	// Sessão expirada
	session = &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	remaining = session.GetRemainingTime()
	assert.Equal(t, time.Duration(0), remaining)
}

func TestSession_GetIdleTime(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	
	time.Sleep(10 * time.Millisecond)
	idleTime := session.GetIdleTime()
	assert.True(t, idleTime >= 0)
	
	session.UpdateActivity()
	
	newIdleTime := session.GetIdleTime()
	assert.True(t, newIdleTime < idleTime)
}

func TestSession_CanBeRefreshed(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	// Sessão válida
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	assert.True(t, session.CanBeRefreshed())
	
	// Sessão expirada
	session = &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	assert.False(t, session.CanBeRefreshed())
	
	// Sessão terminada
	session, _ = NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	session.Terminate()
	assert.False(t, session.CanBeRefreshed())
}

func TestSession_IsValid(t *testing.T) {
	id, _ := valueobjects.NewSessionID("session-123")
	userID, _ := valueobjects.NewUserID("user-123")
	accessToken, _ := valueobjects.NewToken("access-token-123")
	refreshToken, _ := valueobjects.NewToken("refresh-token-123")
	
	// Sessão válida
	session, _ := NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	assert.True(t, session.IsValid())
	
	// Sessão expirada
	session = &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    time.Now().Add(-1 * time.Hour),
		active:       true,
	}
	assert.False(t, session.IsValid())
	
	// Sessão terminada
	session, _ = NewSession(id, userID, accessToken, refreshToken, 3600, "192.168.1.1", "Mozilla/5.0")
	session.Terminate()
	assert.False(t, session.IsValid())
}
