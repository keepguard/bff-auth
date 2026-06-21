package valueobjects

import (
	"fmt"
	"strings"
)

// Application representa uma aplicação do sistema
type Application struct {
	value string
}

// Applications pré-definidas do sistema
var (
	ApplicationAdmin  = Application{value: "admin"}
	ApplicationUser   = Application{value: "user"}
	ApplicationMobile = Application{value: "mobile"}
	ApplicationWeb    = Application{value: "web"}
	ApplicationAPI    = Application{value: "api"}
)

// NewApplication cria uma nova Application
func NewApplication(value string) (Application, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Application{}, fmt.Errorf("application cannot be empty")
	}

	// Valida se é uma aplicação válida
	validApplications := map[string]bool{
		"admin":  true,
		"user":   true,
		"mobile": true,
		"web":    true,
		"api":    true,
	}

	if !validApplications[trimmed] {
		return Application{}, fmt.Errorf("invalid application: %s", trimmed)
	}

	return Application{value: trimmed}, nil
}

// String retorna a representação string da Application
func (a Application) String() string {
	return a.value
}

// Value retorna o valor interno
func (a Application) Value() string {
	return a.value
}

// Equals compara duas Applications
func (a Application) Equals(other Application) bool {
	return a.value == other.value
}

// IsEmpty verifica se a Application está vazia
func (a Application) IsEmpty() bool {
	return a.value == ""
}

// IsAdmin verifica se a aplicação é de administração
func (a Application) IsAdmin() bool {
	return a.value == "admin"
}

// IsUser verifica se a aplicação é de usuário
func (a Application) IsUser() bool {
	return a.value == "user"
}

// IsMobile verifica se a aplicação é mobile
func (a Application) IsMobile() bool {
	return a.value == "mobile"
}

// IsWeb verifica se a aplicação é web
func (a Application) IsWeb() bool {
	return a.value == "web"
}

// IsAPI verifica se a aplicação é API
func (a Application) IsAPI() bool {
	return a.value == "api"
}
