package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// TenantId representa o identificador de aplicação
type TenantId struct {
	value string
}

// NewTenantId cria um novo TenantId
func NewTenantId(value string) (TenantId, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return TenantId{}, fmt.Errorf("x-tenant-id cannot be empty")
	}

	if len(trimmed) < 3 {
		return TenantId{}, fmt.Errorf("x-tenant-id must be at least 3 characters long")
	}

	if len(trimmed) > 50 {
		return TenantId{}, fmt.Errorf("x-tenant-id must be at most 50 characters long")
	}

	// Regex para validar x-tenant-id: apenas letras, números, hífens e underscores
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validPattern.MatchString(trimmed) {
		return TenantId{}, fmt.Errorf("x-tenant-id can only contain letters, numbers, underscore and hyphen")
	}

	return TenantId{value: trimmed}, nil
}

// String retorna a representação string do TenantId
func (xa TenantId) String() string {
	return xa.value
}

// Value retorna o valor interno
func (xa TenantId) Value() string {
	return xa.value
}

// Equals compara dois TenantIds
func (xa TenantId) Equals(other TenantId) bool {
	return xa.value == other.value
}

// IsEmpty verifica se o TenantId está vazio
func (xa TenantId) IsEmpty() bool {
	return xa.value == ""
}
