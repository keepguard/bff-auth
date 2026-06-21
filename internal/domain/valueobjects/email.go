package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// Email representa um endereço de email
type Email struct {
	value string
}

// NewEmail cria um novo Email
func NewEmail(value string) (Email, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Email{}, fmt.Errorf("email cannot be empty")
	}

	// Regex básico para validar email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(trimmed) {
		return Email{}, fmt.Errorf("invalid email format")
	}

	return Email{value: trimmed}, nil
}

// String retorna a representação string do Email
func (e Email) String() string {
	return e.value
}

// Value retorna o valor interno
func (e Email) Value() string {
	return e.value
}

// Equals compara dois Emails
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// IsEmpty verifica se o Email está vazio
func (e Email) IsEmpty() bool {
	return e.value == ""
}

// Domain retorna o domínio do email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// LocalPart retorna a parte local do email (antes do @)
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}
