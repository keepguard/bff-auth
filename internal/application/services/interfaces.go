package services

import "context"

// TokenService define a interface para serviços de token
type TokenService interface {
	ValidateToken(ctx context.Context, token string) (string, error)
	ExtractUserID(ctx context.Context, token string) (string, error)
	ExtractPermissions(ctx context.Context, token string) ([]string, error)
}

// ValidationService define a interface para serviços de validação
type ValidationService interface {
	ValidateLoginRequest(req interface{}) error
	ValidateRefreshRequest(req interface{}) error
	ValidateLogoutRequest(req interface{}) error
}

// AuthService define a interface para serviços de autenticação
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) error
	GenerateTokens(ctx context.Context, userID string) (accessToken, refreshToken string, err error)
	RevokeToken(ctx context.Context, token string) error
}
