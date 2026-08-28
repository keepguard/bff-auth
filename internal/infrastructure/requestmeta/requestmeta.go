package requestmeta

import "context"

type contextKey string

const clientIPKey contextKey = "keepguard.clientIP"

func WithClientIP(ctx context.Context, ip string) context.Context {
	if ip == "" {
		return ctx
	}
	return context.WithValue(ctx, clientIPKey, ip)
}

func ClientIP(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey).(string)
	return ip
}
