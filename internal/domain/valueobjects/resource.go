package valueobjects

import (
	"fmt"
	"strings"
)

// Resource representa um recurso do sistema
type Resource struct {
	value string
}

// Resources pré-definidos do sistema
var (
	ResourceUser        = Resource{value: "user"}
	ResourceCompany     = Resource{value: "company"}
	ResourcePermission  = Resource{value: "permission"}
	ResourceRole        = Resource{value: "role"}
	ResourceApplication = Resource{value: "application"}
	ResourceSession     = Resource{value: "session"}
	ResourceToken       = Resource{value: "token"}
	ResourceAudit       = Resource{value: "audit"}
)

// NewResource cria um novo Resource
func NewResource(value string) (Resource, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Resource{}, fmt.Errorf("resource cannot be empty")
	}

	// Valida se é um resource válido
	validResources := map[string]bool{
		"user":        true,
		"company":     true,
		"permission":  true,
		"role":        true,
		"application": true,
		"session":     true,
		"token":       true,
		"audit":       true,
	}

	if !validResources[trimmed] {
		return Resource{}, fmt.Errorf("invalid resource: %s", trimmed)
	}

	return Resource{value: trimmed}, nil
}

// String retorna a representação string do Resource
func (r Resource) String() string {
	return r.value
}

// Value retorna o valor interno
func (r Resource) Value() string {
	return r.value
}

// Equals compara dois Resources
func (r Resource) Equals(other Resource) bool {
	return r.value == other.value
}

// IsEmpty verifica se o Resource está vazio
func (r Resource) IsEmpty() bool {
	return r.value == ""
}

// IsUser verifica se o resource é de usuário
func (r Resource) IsUser() bool {
	return r.value == "user"
}

// IsCompany verifica se o resource é de empresa
func (r Resource) IsCompany() bool {
	return r.value == "company"
}

// IsPermission verifica se o resource é de permissão
func (r Resource) IsPermission() bool {
	return r.value == "permission"
}

// IsRole verifica se o resource é de role
func (r Resource) IsRole() bool {
	return r.value == "role"
}

// IsApplication verifica se o resource é de aplicação
func (r Resource) IsApplication() bool {
	return r.value == "application"
}

// IsSession verifica se o resource é de sessão
func (r Resource) IsSession() bool {
	return r.value == "session"
}

// IsToken verifica se o resource é de token
func (r Resource) IsToken() bool {
	return r.value == "token"
}

// IsAudit verifica se o resource é de auditoria
func (r Resource) IsAudit() bool {
	return r.value == "audit"
}
