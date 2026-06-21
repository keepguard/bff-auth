package valueobjects

import (
	"fmt"
	"strings"
)

// SessionID representa o identificador único de uma sessão
type SessionID struct {
	value string
}

// NewSessionID cria um novo SessionID
func NewSessionID(value string) (SessionID, error) {
	if strings.TrimSpace(value) == "" {
		return SessionID{}, fmt.Errorf("session ID cannot be empty")
	}

	// Validações adicionais podem ser adicionadas aqui
	// Por exemplo: formato UUID, tamanho máximo, etc.

	return SessionID{value: strings.TrimSpace(value)}, nil
}

// String retorna a representação string do SessionID
func (id SessionID) String() string {
	return id.value
}

// Value retorna o valor interno
func (id SessionID) Value() string {
	return id.value
}

// Equals compara dois SessionIDs
func (id SessionID) Equals(other SessionID) bool {
	return id.value == other.value
}

// IsEmpty verifica se o SessionID está vazio
func (id SessionID) IsEmpty() bool {
	return id.value == ""
}
