package client

import (
	"context"

	authclient "github.com/keepguard/bff-auth/internal/application/port"
)

func companyHeader(ctx context.Context) string {
	return authclient.CompanyIDFromContext(ctx)
}
