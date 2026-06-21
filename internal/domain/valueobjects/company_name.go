package valueobjects

import (
	"fmt"
	"strings"
)

// CompanyName representa o nome de uma empresa
type CompanyName struct {
	value string
}

// NewCompanyName cria um novo CompanyName
func NewCompanyName(value string) (CompanyName, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return CompanyName{}, fmt.Errorf("company name cannot be empty")
	}

	if len(trimmed) < 2 {
		return CompanyName{}, fmt.Errorf("company name must be at least 2 characters long")
	}

	if len(trimmed) > 100 {
		return CompanyName{}, fmt.Errorf("company name must be at most 100 characters long")
	}

	return CompanyName{value: trimmed}, nil
}

// String retorna a representação string do CompanyName
func (cn CompanyName) String() string {
	return cn.value
}

// Value retorna o valor interno
func (cn CompanyName) Value() string {
	return cn.value
}

// Equals compara dois CompanyNames
func (cn CompanyName) Equals(other CompanyName) bool {
	return cn.value == other.value
}

// IsEmpty verifica se o CompanyName está vazio
func (cn CompanyName) IsEmpty() bool {
	return cn.value == ""
}
