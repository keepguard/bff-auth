package valueobjects

import (
	"fmt"
	"strings"
)

// CompanyID representa o identificador único de uma empresa
type CompanyID struct {
	value string
}

// NewCompanyID cria um novo CompanyID
func NewCompanyID(value string) (CompanyID, error) {
	if strings.TrimSpace(value) == "" {
		return CompanyID{}, fmt.Errorf("company ID cannot be empty")
	}

	// Validações adicionais podem ser adicionadas aqui
	// Por exemplo: formato UUID, tamanho máximo, etc.

	return CompanyID{value: strings.TrimSpace(value)}, nil
}

// String retorna a representação string do CompanyID
func (id CompanyID) String() string {
	return id.value
}

// Value retorna o valor interno
func (id CompanyID) Value() string {
	return id.value
}

// Equals compara dois CompanyIDs
func (id CompanyID) Equals(other CompanyID) bool {
	return id.value == other.value
}

// IsEmpty verifica se o CompanyID está vazio
func (id CompanyID) IsEmpty() bool {
	return id.value == ""
}
