package services

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

// AuthDomainService define as regras de negócio para autenticação
type AuthDomainService interface {
	ValidateTokenFormat(token string) error
	ValidateLoginCredentials(username, password string) error
	ShouldRefreshToken(expiresIn int) error
	ValidateTokenExpiration(expiresIn int) error
	CalculateTokenTTL(tokenType string) (int, error)
	ValidateRefreshToken(refreshToken string) error
	GenerateSessionID() string
	ValidateSessionTimeout(sessionDuration time.Duration) error
	ShouldLogoutUser(lastActivity time.Duration) bool
	ValidateXApplication(xApplication string) error
}

type authDomainServiceImpl struct{}

// NewAuthDomainService cria um novo serviço de domínio para autenticação
func NewAuthDomainService() AuthDomainService {
	return &authDomainServiceImpl{}
}

// ValidateTokenFormat valida o formato de um token
func (s *authDomainServiceImpl) ValidateTokenFormat(token string) error {
	if token == "" {
		return errors.New("token cannot be empty")
	}
	
	// Verificar se o token não é muito longo (limite de segurança)
	if len(token) > 8192 { // 8KB
		return errors.New("token is too long")
	}
	
	// Verificar se o token tem o formato JWT (3 partes separadas por ponto)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		// É um JWT, verificar se cada parte não está vazia
		for i, part := range parts {
			if part == "" {
				return errors.New("invalid token format: part " + string(rune(i+1)) + " is empty")
			}
		}
	}
	// Se não for JWT, aceitar tokens simples
	
	return nil
}

// ValidateLoginCredentials valida as credenciais de login
func (s *authDomainServiceImpl) ValidateLoginCredentials(username, password string) error {
	if username == "" {
		return errors.New("username is required")
	}
	
	if password == "" {
		return errors.New("password is required")
	}
	
	// Verificar se o username não é muito longo
	if len(username) > 255 {
		return errors.New("username is too long")
	}
	
	// Verificar se a senha não é muito longa
	if len(password) > 128 {
		return errors.New("password is too long")
	}
	
	// Verificar se o username contém apenas caracteres válidos
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validUsername.MatchString(username) {
		return errors.New("username contains invalid characters")
	}
	
	return nil
}

// ShouldRefreshToken verifica se um token deve ser renovado
func (s *authDomainServiceImpl) ShouldRefreshToken(expiresIn int) error {
	if expiresIn <= 0 {
		return errors.New("invalid expiration time")
	}
	
	// Verificar se o token não está muito próximo do vencimento (menos de 5 minutos)
	if expiresIn < 300 { // 5 minutos em segundos
		return errors.New("token is about to expire")
	}
	
	return nil
}

// ValidateTokenExpiration valida o tempo de expiração de um token
func (s *authDomainServiceImpl) ValidateTokenExpiration(expiresIn int) error {
	if expiresIn <= 0 {
		return errors.New("invalid expiration time")
	}
	
	// Verificar se o token não expira muito cedo (menos de 1 minuto)
	if expiresIn < 60 {
		return errors.New("token expires too soon")
	}
	
	// Verificar se o token não expira muito tarde (mais de 24 horas)
	if expiresIn > 86400 { // 24 horas em segundos
		return errors.New("token expires too late")
	}
	
	return nil
}

// CalculateTokenTTL calcula o TTL apropriado para um tipo de token
func (s *authDomainServiceImpl) CalculateTokenTTL(tokenType string) (int, error) {
	switch strings.ToLower(tokenType) {
	case "access":
		return 3600, nil // 1 hora
	case "refresh":
		return 86400, nil // 24 horas
	case "password_reset":
		return 1800, nil // 30 minutos
	case "email_verification":
		return 3600, nil // 1 hora
	case "api":
		return 7200, nil // 2 horas
	default:
		return 0, errors.New("unknown token type")
	}
}

// ValidateRefreshToken valida um refresh token
func (s *authDomainServiceImpl) ValidateRefreshToken(refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token is required")
	}
	
	// Verificar se o refresh token tem o formato correto
	if err := s.ValidateTokenFormat(refreshToken); err != nil {
		return err
	}
	
	// Verificar se o refresh token não é muito curto
	if len(refreshToken) < 32 {
		return errors.New("refresh token is too short")
	}
	
	return nil
}

// GenerateSessionID gera um ID único para sessão
func (s *authDomainServiceImpl) GenerateSessionID() string {
	// Implementação simplificada - em produção usar UUID ou algoritmo mais robusto
	timestamp := time.Now().UnixNano()
	return "session_" + string(rune(timestamp)) + "_" + generateRandomString(8)
}

// ValidateSessionTimeout valida o timeout de uma sessão
func (s *authDomainServiceImpl) ValidateSessionTimeout(sessionDuration time.Duration) error {
	if sessionDuration <= 0 {
		return errors.New("session duration must be positive")
	}
	
	// Verificar se a sessão não é muito curta (menos de 5 minutos)
	if sessionDuration < 5*time.Minute {
		return errors.New("session duration is too short")
	}
	
	// Verificar se a sessão não é muito longa (mais de 24 horas)
	if sessionDuration > 24*time.Hour {
		return errors.New("session duration is too long")
	}
	
	return nil
}

// ShouldLogoutUser verifica se um usuário deve ser deslogado
func (s *authDomainServiceImpl) ShouldLogoutUser(lastActivity time.Duration) bool {
	// Deslogar se inativo por mais de 2 horas
	return lastActivity > 2*time.Hour
}

// ValidateXApplication valida o X-Application header
func (s *authDomainServiceImpl) ValidateXApplication(xApplication string) error {
	if xApplication == "" {
		return errors.New("x-application header is required")
	}
	
	if len(xApplication) < 3 {
		return errors.New("x-application must have at least 3 characters")
	}
	
	if len(xApplication) > 50 {
		return errors.New("x-application must have at most 50 characters")
	}
	
	// Verificar se contém apenas caracteres válidos
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validChars.MatchString(xApplication) {
		return errors.New("x-application contains invalid characters")
	}
	
	return nil
}

// generateRandomString gera uma string aleatória
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
