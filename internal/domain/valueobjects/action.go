package valueobjects

import (
	"fmt"
	"strings"
)

// Action representa uma ação que pode ser realizada em um recurso
type Action struct {
	value string
}

// Actions pré-definidas do sistema
var (
	ActionCreate   = Action{value: "create"}
	ActionRead     = Action{value: "read"}
	ActionUpdate   = Action{value: "update"}
	ActionDelete   = Action{value: "delete"}
	ActionList     = Action{value: "list"}
	ActionExecute  = Action{value: "execute"}
	ActionManage   = Action{value: "manage"}
	ActionAudit    = Action{value: "audit"}
	ActionLogin    = Action{value: "login"}
	ActionLogout   = Action{value: "logout"}
	ActionRefresh  = Action{value: "refresh"}
	ActionValidate = Action{value: "validate"}
)

// NewAction cria um novo Action
func NewAction(value string) (Action, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))

	if trimmed == "" {
		return Action{}, fmt.Errorf("action cannot be empty")
	}

	// Valida se é uma action válida
	validActions := map[string]bool{
		"create":   true,
		"read":     true,
		"update":   true,
		"delete":   true,
		"list":     true,
		"execute":  true,
		"manage":   true,
		"audit":    true,
		"login":    true,
		"logout":   true,
		"refresh":  true,
		"validate": true,
	}

	if !validActions[trimmed] {
		return Action{}, fmt.Errorf("invalid action: %s", trimmed)
	}

	return Action{value: trimmed}, nil
}

// String retorna a representação string do Action
func (a Action) String() string {
	return a.value
}

// Value retorna o valor interno
func (a Action) Value() string {
	return a.value
}

// Equals compara dois Actions
func (a Action) Equals(other Action) bool {
	return a.value == other.value
}

// IsEmpty verifica se o Action está vazio
func (a Action) IsEmpty() bool {
	return a.value == ""
}

// IsCreate verifica se a ação é de criação
func (a Action) IsCreate() bool {
	return a.value == "create"
}

// IsRead verifica se a ação é de leitura
func (a Action) IsRead() bool {
	return a.value == "read"
}

// IsUpdate verifica se a ação é de atualização
func (a Action) IsUpdate() bool {
	return a.value == "update"
}

// IsDelete verifica se a ação é de exclusão
func (a Action) IsDelete() bool {
	return a.value == "delete"
}

// IsList verifica se a ação é de listagem
func (a Action) IsList() bool {
	return a.value == "list"
}

// IsExecute verifica se a ação é de execução
func (a Action) IsExecute() bool {
	return a.value == "execute"
}

// IsManage verifica se a ação é de gerenciamento
func (a Action) IsManage() bool {
	return a.value == "manage"
}

// IsAudit verifica se a ação é de auditoria
func (a Action) IsAudit() bool {
	return a.value == "audit"
}

// IsLogin verifica se a ação é de login
func (a Action) IsLogin() bool {
	return a.value == "login"
}

// IsLogout verifica se a ação é de logout
func (a Action) IsLogout() bool {
	return a.value == "logout"
}

// IsRefresh verifica se a ação é de refresh
func (a Action) IsRefresh() bool {
	return a.value == "refresh"
}

// IsValidate verifica se a ação é de validação
func (a Action) IsValidate() bool {
	return a.value == "validate"
}
