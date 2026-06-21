package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// XApplication representa o identificador de aplicação
type XApplication struct {
	value string
}

// NewXApplication cria um novo XApplication
func NewXApplication(value string) (XApplication, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return XApplication{}, fmt.Errorf("x-application cannot be empty")
	}

	if len(trimmed) < 3 {
		return XApplication{}, fmt.Errorf("x-application must be at least 3 characters long")
	}

	if len(trimmed) > 50 {
		return XApplication{}, fmt.Errorf("x-application must be at most 50 characters long")
	}

	// Regex para validar x-application: apenas letras, números, hífens e underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(trimmed) {
		return XApplication{}, fmt.Errorf("x-application can only contain letters, numbers, underscore and hyphen")
	}

	return XApplication{value: trimmed}, nil
}

// String retorna a representação string do XApplication
func (xa XApplication) String() string {
	return xa.value
}

// Value retorna o valor interno
func (xa XApplication) Value() string {
	return xa.value
}

// Equals compara dois XApplications
func (xa XApplication) Equals(other XApplication) bool {
	return xa.value == other.value
}

// IsEmpty verifica se o XApplication está vazio
func (xa XApplication) IsEmpty() bool {
	return xa.value == ""
}
