package requestmeta

import "context"

type contextKey string

const clientIPKey contextKey = "keepguard.clientIP"
const clientLocationKey contextKey = "keepguard.clientLocation"

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

func WithClientLocation(ctx context.Context, location string) context.Context {
	if location == "" {
		return ctx
	}
	return context.WithValue(ctx, clientLocationKey, location)
}

func ClientLocation(ctx context.Context) string {
	location, _ := ctx.Value(clientLocationKey).(string)
	return location
}
