package valueobjects

import (
	"fmt"
	"strings"
)

// UserID representa o identificador único de um usuário
type UserID struct {
	value string
}

// NewUserID cria um novo UserID
func NewUserID(value string) (UserID, error) {
	if strings.TrimSpace(value) == "" {
		return UserID{}, fmt.Errorf("user ID cannot be empty")
	}

	// Validações adicionais podem ser adicionadas aqui
	// Por exemplo: formato UUID, tamanho máximo, etc.

	return UserID{value: strings.TrimSpace(value)}, nil
}

// String retorna a representação string do UserID
func (id UserID) String() string {
	return id.value
}

// Value retorna o valor interno
func (id UserID) Value() string {
	return id.value
}

// Equals compara dois UserIDs
func (id UserID) Equals(other UserID) bool {
	return id.value == other.value
}

// IsEmpty verifica se o UserID está vazio
func (id UserID) IsEmpty() bool {
	return id.value == ""
}
