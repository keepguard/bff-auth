package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// Token representa um token de autenticação
type Token struct {
	value string
}

// NewToken cria um novo Token
func NewToken(value string) (Token, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return Token{}, fmt.Errorf("token cannot be empty")
	}

	if len(trimmed) < 10 {
		return Token{}, fmt.Errorf("token must be at least 10 characters long")
	}

	if len(trimmed) > 8192 {
		return Token{}, fmt.Errorf("token is too long")
	}

	// Verificar se é um JWT (3 partes separadas por ponto)
	parts := strings.Split(trimmed, ".")
	if len(parts) == 3 {
		// Validar formato JWT
		for i, part := range parts {
			if part == "" {
				return Token{}, fmt.Errorf("invalid JWT format: part %d is empty", i+1)
			}
		}
	} else {
		// Validar se é um token simples (apenas caracteres válidos)
		validToken := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
		if !validToken.MatchString(trimmed) {
			return Token{}, fmt.Errorf("token contains invalid characters")
		}
	}

	return Token{value: trimmed}, nil
}

// String retorna a representação string do Token
func (t Token) String() string {
	return t.value
}

// Value retorna o valor interno
func (t Token) Value() string {
	return t.value
}

// Equals compara dois Tokens
func (t Token) Equals(other Token) bool {
	return t.value == other.value
}

// IsEmpty verifica se o Token está vazio
func (t Token) IsEmpty() bool {
	return t.value == ""
}

// IsJWT verifica se o token é um JWT
func (t Token) IsJWT() bool {
	parts := strings.Split(t.value, ".")
	return len(parts) == 3
}

// GetJWTHeader retorna o header do JWT (se for JWT)
func (t Token) GetJWTHeader() string {
	if !t.IsJWT() {
		return ""
	}
	parts := strings.Split(t.value, ".")
	return parts[0]
}

// GetJWTPayload retorna o payload do JWT (se for JWT)
func (t Token) GetJWTPayload() string {
	if !t.IsJWT() {
		return ""
	}
	parts := strings.Split(t.value, ".")
	return parts[1]
}

// GetJWTSignature retorna a assinatura do JWT (se for JWT)
func (t Token) GetJWTSignature() string {
	if !t.IsJWT() {
		return ""
	}
	parts := strings.Split(t.value, ".")
	return parts[2]
}
