package valueobjects

import (
	"fmt"
	"strings"
)

// PermissionID representa o identificador único de uma permissão
type PermissionID struct {
	value string
}

// NewPermissionID cria um novo PermissionID
func NewPermissionID(value string) (PermissionID, error) {
	if strings.TrimSpace(value) == "" {
		return PermissionID{}, fmt.Errorf("permission ID cannot be empty")
	}

	// Validações adicionais podem ser adicionadas aqui
	// Por exemplo: formato UUID, tamanho máximo, etc.

	return PermissionID{value: strings.TrimSpace(value)}, nil
}

// String retorna a representação string do PermissionID
func (id PermissionID) String() string {
	return id.value
}

// Value retorna o valor interno
func (id PermissionID) Value() string {
	return id.value
}

// Equals compara dois PermissionIDs
func (id PermissionID) Equals(other PermissionID) bool {
	return id.value == other.value
}

// IsEmpty verifica se o PermissionID está vazio
func (id PermissionID) IsEmpty() bool {
	return id.value == ""
}
