package client

import "context"

// LoginTokenPresence indica se o JWT de login ainda existe no Redis (tokenlogin).
type LoginTokenPresence int

const (
	// LoginTokenPresent: chave tokenlogin:{codeUser}:{token} existe.
	LoginTokenPresent LoginTokenPresence = iota
	// LoginTokenAbsent: sessão revogada, expirada ou nunca emitida.
	LoginTokenAbsent
	// LoginTokenUnknown: Redis indisponível, JWT ilegível ou checker nulo — seguir para o ms-auth.
	LoginTokenUnknown
)

// LoginTokenChecker consulta se o access token ainda está ativo no Redis.
type LoginTokenChecker interface {
	Check(ctx context.Context, token string) (LoginTokenPresence, string, error)
}
