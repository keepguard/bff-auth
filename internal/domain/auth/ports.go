package auth

import "context"

// Permissions representa as permissões de um usuário
type Permissions struct {
	Roles        []string `json:"roles"`
	Applications []string `json:"applications"`
	Scopes       []string `json:"scopes"`
}

// LoginRequest representa a requisição de login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse representa a resposta de login
type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// RefreshRequest representa a requisição de refresh
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshResponse representa a resposta de refresh
type RefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// LogoutRequest representa a requisição de logout
type LogoutRequest struct {
	RefreshToken string `json:"refreshToken,omitempty"`
}

// User representa um usuário
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Apps     []string `json:"apps"`
	Active   bool     `json:"active"`
}

// MeResponse representa a resposta do endpoint /me
type MeResponse struct {
	User        User        `json:"user"`
	Permissions Permissions `json:"permissions"`
}

// CreateUserRequest representa a requisição de criação de usuário
type CreateUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
	Apps     []string `json:"apps"`
}

// CreateUserResponse representa a resposta de criação de usuário
type CreateUserResponse struct {
	IDExternal string `json:"idExternal"`
	ID         string `json:"id"`
}

// AuthReader interface para leitura de dados de autenticação
type AuthReader interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponse, error)
	GetUser(ctx context.Context, userID string) (*User, error)
	GetPermissions(ctx context.Context, userID string) (*Permissions, error)
}

// AuthWriter interface para escrita de dados de autenticação
type AuthWriter interface {
	Logout(ctx context.Context, req LogoutRequest) error
	CreateUser(ctx context.Context, req CreateUserRequest) (*CreateUserResponse, error)
	DeleteUser(ctx context.Context, userID string) error
	BlockUser(ctx context.Context, userID string) error
	UnblockUser(ctx context.Context, userID string) error
	AddRole(ctx context.Context, userID, role string) error
	RemoveRole(ctx context.Context, userID, role string) error
	AddApplication(ctx context.Context, userID, app string) error
	RemoveApplication(ctx context.Context, userID, app string) error
}
