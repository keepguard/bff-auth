package valueobjects

import (
	"fmt"
	"strings"
)

// Role representa um papel/função do usuário
type Role struct {
	value string
}

// Roles pré-definidos do sistema
var (
	RoleAdmin    = Role{value: "admin"}
	RoleUser     = Role{value: "user"}
	RoleManager  = Role{value: "manager"}
	RoleViewer   = Role{value: "viewer"}
	RoleOperator = Role{value: "operator"}
)

// NewRole cria um novo Role
func NewRole(value string) (Role, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Role{}, fmt.Errorf("role cannot be empty")
	}

	// Valida se é um role válido
	validRoles := map[string]bool{
		"admin":    true,
		"user":     true,
		"manager":  true,
		"viewer":   true,
		"operator": true,
	}

	if !validRoles[trimmed] {
		return Role{}, fmt.Errorf("invalid role: %s", trimmed)
	}

	return Role{value: trimmed}, nil
}

// String retorna a representação string do Role
func (r Role) String() string {
	return r.value
}

// Value retorna o valor interno
func (r Role) Value() string {
	return r.value
}

// Equals compara dois Roles
func (r Role) Equals(other Role) bool {
	return r.value == other.value
}

// IsEmpty verifica se o Role está vazio
func (r Role) IsEmpty() bool {
	return r.value == ""
}

// IsAdmin verifica se o role é de administrador
func (r Role) IsAdmin() bool {
	return r.value == "admin"
}

// IsManager verifica se o role é de gerente
func (r Role) IsManager() bool {
	return r.value == "manager"
}

// HasPermission verifica se o role tem uma permissão específica
func (r Role) HasPermission(permission string) bool {
	// Mapeamento de roles para permissões
	rolePermissions := map[string][]string{
		"admin":    {"*"}, // Admin tem todas as permissões
		"manager":  {"read", "write", "update"},
		"user":     {"read", "write"},
		"operator": {"read", "execute"},
		"viewer":   {"read"},
	}

	permissions, exists := rolePermissions[r.value]
	if !exists {
		return false
	}

	// Admin tem todas as permissões
	if len(permissions) == 1 && permissions[0] == "*" {
		return true
	}

	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}

	return false
}
