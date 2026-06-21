package valueobjects

import (
	"fmt"
	"strings"
)

// Scope representa o escopo de uma permissão
type Scope struct {
	value string
}

// Scopes pré-definidos do sistema
var (
	ScopeSystem      = Scope{value: "system"}
	ScopeApplication = Scope{value: "application"}
	ScopeUser        = Scope{value: "user"}
	ScopeCompany     = Scope{value: "company"}
	ScopeGlobal      = Scope{value: "global"}
)

// NewScope cria um novo Scope
func NewScope(value string) (Scope, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Scope{}, fmt.Errorf("scope cannot be empty")
	}

	// Valida se é um scope válido
	validScopes := map[string]bool{
		"system":      true,
		"application": true,
		"user":        true,
		"company":     true,
		"global":      true,
	}

	if !validScopes[trimmed] {
		return Scope{}, fmt.Errorf("invalid scope: %s", trimmed)
	}

	return Scope{value: trimmed}, nil
}

// String retorna a representação string do Scope
func (s Scope) String() string {
	return s.value
}

// Value retorna o valor interno
func (s Scope) Value() string {
	return s.value
}

// Equals compara dois Scopes
func (s Scope) Equals(other Scope) bool {
	return s.value == other.value
}

// IsEmpty verifica se o Scope está vazio
func (s Scope) IsEmpty() bool {
	return s.value == ""
}

// IsSystem verifica se o scope é de sistema
func (s Scope) IsSystem() bool {
	return s.value == "system"
}

// IsApplication verifica se o scope é de aplicação
func (s Scope) IsApplication() bool {
	return s.value == "application"
}

// IsUser verifica se o scope é de usuário
func (s Scope) IsUser() bool {
	return s.value == "user"
}

// IsCompany verifica se o scope é de empresa
func (s Scope) IsCompany() bool {
	return s.value == "company"
}

// IsGlobal verifica se o scope é global
func (s Scope) IsGlobal() bool {
	return s.value == "global"
}
