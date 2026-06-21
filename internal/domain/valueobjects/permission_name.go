package valueobjects

import (
	"fmt"
	"regexp"
	"strings"
)

// PermissionName representa o nome de uma permissão
type PermissionName struct {
	value string
}

// NewPermissionName cria um novo PermissionName
func NewPermissionName(value string) (PermissionName, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return PermissionName{}, fmt.Errorf("permission name cannot be empty")
	}

	if len(trimmed) < 3 {
		return PermissionName{}, fmt.Errorf("permission name must be at least 3 characters long")
	}

	if len(trimmed) > 100 {
		return PermissionName{}, fmt.Errorf("permission name must be at most 100 characters long")
	}

	// Regex para validar permission name: letras, números, pontos, underscores e hífens
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validPattern.MatchString(trimmed) {
		return PermissionName{}, fmt.Errorf("permission name can only contain letters, numbers, dots, underscores and hyphens")
	}

	return PermissionName{value: trimmed}, nil
}

// String retorna a representação string do PermissionName
func (pn PermissionName) String() string {
	return pn.value
}

// Value retorna o valor interno
func (pn PermissionName) Value() string {
	return pn.value
}

// Equals compara dois PermissionNames
func (pn PermissionName) Equals(other PermissionName) bool {
	return pn.value == other.value
}

// IsEmpty verifica se o PermissionName está vazio
func (pn PermissionName) IsEmpty() bool {
	return pn.value == ""
}
