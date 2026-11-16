package pkg

// UserStatusActive é o único status do usuário (ms-user) que pode seguir no fluxo de recuperação de senha.
const UserStatusActive = "ACTIVE"

// IsActiveUserStatus diz se o status do usuário permite a recuperação de senha. Só ACTIVE: PENDING,
// PENDING_ACTIVATION (cadastro não concluído), BLOCKED, DELETED, INACTIVE e qualquer outro valor são recusados.
func IsActiveUserStatus(status string) bool {
	return status == UserStatusActive
}
