package entities

import (
	"errors"
	"time"

	"github.com/keepguard/bff-auth/internal/domain/valueobjects"
)

// Session representa uma sessão de autenticação
type Session struct {
	id           valueobjects.SessionID
	userID       valueobjects.UserID
	accessToken  valueobjects.Token
	refreshToken valueobjects.Token
	expiresAt    time.Time
	createdAt    time.Time
	updatedAt    time.Time
	lastActivity time.Time
	ipAddress    string
	userAgent    string
	active       bool
}

// NewSession cria uma nova sessão
func NewSession(
	id valueobjects.SessionID,
	userID valueobjects.UserID,
	accessToken valueobjects.Token,
	refreshToken valueobjects.Token,
	expiresIn int,
	ipAddress string,
	userAgent string,
) (*Session, error) {
	if id.IsEmpty() {
		return nil, errors.New("session ID cannot be empty")
	}
	
	if userID.IsEmpty() {
		return nil, errors.New("user ID cannot be empty")
	}
	
	if accessToken.IsEmpty() {
		return nil, errors.New("access token cannot be empty")
	}
	
	if refreshToken.IsEmpty() {
		return nil, errors.New("refresh token cannot be empty")
	}
	
	if expiresIn <= 0 {
		return nil, errors.New("expiration time must be positive")
	}
	
	now := time.Now()
	expiresAt := now.Add(time.Duration(expiresIn) * time.Second)
	
	return &Session{
		id:           id,
		userID:       userID,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		expiresAt:    expiresAt,
		createdAt:    now,
		updatedAt:    now,
		lastActivity: now,
		ipAddress:    ipAddress,
		userAgent:    userAgent,
		active:       true,
	}, nil
}

// Getters
func (s *Session) ID() valueobjects.SessionID {
	return s.id
}

func (s *Session) UserID() valueobjects.UserID {
	return s.userID
}

func (s *Session) AccessToken() valueobjects.Token {
	return s.accessToken
}

func (s *Session) RefreshToken() valueobjects.Token {
	return s.refreshToken
}

func (s *Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Session) UpdatedAt() time.Time {
	return s.updatedAt
}

func (s *Session) LastActivity() time.Time {
	return s.lastActivity
}

func (s *Session) IPAddress() string {
	return s.ipAddress
}

func (s *Session) UserAgent() string {
	return s.userAgent
}

func (s *Session) IsActive() bool {
	return s.active
}

// Business Methods
func (s *Session) Refresh(accessToken valueobjects.Token, expiresIn int) error {
	if s == nil {
		return errors.New("session is nil")
	}
	
	if !s.active {
		return errors.New("cannot refresh inactive session")
	}
	
	if s.IsExpired() {
		return errors.New("cannot refresh expired session")
	}
	
	if accessToken.IsEmpty() {
		return errors.New("access token cannot be empty")
	}
	
	if expiresIn <= 0 {
		return errors.New("expiration time must be positive")
	}
	
	s.accessToken = accessToken
	s.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	s.lastActivity = time.Now()
	s.updatedAt = time.Now()
	
	return nil
}

func (s *Session) Terminate() {
	s.active = false
	s.updatedAt = time.Now()
}

func (s *Session) UpdateActivity() {
	s.lastActivity = time.Now()
	s.updatedAt = time.Now()
}

func (s *Session) UpdateLocation(ipAddress, userAgent string) {
	s.ipAddress = ipAddress
	s.userAgent = userAgent
	s.lastActivity = time.Now()
	s.updatedAt = time.Now()
}

func (s *Session) IsExpired() bool {
	if s == nil {
		return true
	}
	return time.Now().After(s.expiresAt)
}

func (s *Session) IsIdle(idleTimeout time.Duration) bool {
	return time.Since(s.lastActivity) > idleTimeout
}

func (s *Session) ShouldBeTerminated(maxDuration time.Duration, idleTimeout time.Duration) bool {
	// Terminar se expirou
	if s.IsExpired() {
		return true
	}
	
	// Terminar se inativo por muito tempo
	if s.IsIdle(idleTimeout) {
		return true
	}
	
	// Terminar se a sessão é muito longa
	if time.Since(s.createdAt) > maxDuration {
		return true
	}
	
	return false
}

func (s *Session) GetRemainingTime() time.Duration {
	if s.IsExpired() {
		return 0
	}
	return time.Until(s.expiresAt)
}

func (s *Session) GetIdleTime() time.Duration {
	return time.Since(s.lastActivity)
}

func (s *Session) CanBeRefreshed() bool {
	return s.active && !s.IsExpired()
}

func (s *Session) IsValid() bool {
	return s.active && !s.IsExpired()
}
