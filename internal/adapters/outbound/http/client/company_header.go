package client

import (
	"context"

	authclient "github.com/keepguard/bff-auth/internal/domain/ports/client"
)

func companyHeader(ctx context.Context) string {
	return authclient.CompanyIDFromContext(ctx)
}
