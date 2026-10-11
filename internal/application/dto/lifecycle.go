package dto

type AccountLifecycleCommand struct {
	TenantID      string
	Token         string
	CodeUser      string
	Reason        string
	CorrelationID string
}

// AdminDeleteUserCommand: admin/gestor da company exclui (soft delete) a
// conta de outro usuário. A eliminação definitiva vem depois, pelo job do
// ms-auth (account_purge).
type AdminDeleteUserCommand struct {
	TenantID       string
	Token          string
	IDUserExternal string
	Reason         string
	CorrelationID  string
}

// AdminDeleteUserByCodeCommand: admin/gestor da company exclui (soft delete) a
// conta de outro usuário a partir do codeUser (sub do JWT).
type AdminDeleteUserByCodeCommand struct {
	TenantID      string
	Token         string
	CodeUser      string
	Reason        string
	CorrelationID string
}
