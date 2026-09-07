package port

import "context"

type LoginTokenPresence int

const (
	LoginTokenPresent LoginTokenPresence = iota
	LoginTokenAbsent
	LoginTokenUnknown
)

type LoginTokenChecker interface {
	Check(ctx context.Context, token string) (LoginTokenPresence, string, error)
}
