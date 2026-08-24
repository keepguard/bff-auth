package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// Username representa o nome de usuário
type Username struct {
	value string
}

// NewUsername cria um novo Username
func NewUsername(value string) (Username, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return Username{}, fmt.Errorf("username cannot be empty")
	}

	if len(trimmed) < 3 {
		return Username{}, fmt.Errorf("username must be at least 3 characters long")
	}

	if len(trimmed) > 50 {
		return Username{}, fmt.Errorf("username must be at most 50 characters long")
	}

	// Regex para validar username: apenas letras, números, underscore e hífen
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(trimmed) {
		return Username{}, fmt.Errorf("username can only contain letters, numbers, underscore and hyphen")
	}

	return Username{value: trimmed}, nil
}

// String retorna a representação string do Username
func (u Username) String() string {
	return u.value
}

// Value retorna o valor interno
func (u Username) Value() string {
	return u.value
}

// Equals compara dois Usernames
func (u Username) Equals(other Username) bool {
	return u.value == other.value
}

// IsEmpty verifica se o Username está vazio
func (u Username) IsEmpty() bool {
	return u.value == ""
}

// NewUsernameUnsafe cria um Username sem validação (usado para testes ou bypass)
func NewUsernameUnsafe(value string) Username {
	return Username{value: value}
}
