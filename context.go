package rona

import "context"

type contextKey int

const (
	flashContextKey = contextKey(iota + 1)
)

// NewContextWithFlash creates a context with the flash value.
func NewContextWithFlash(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, flashContextKey, v)
}

// FlashFromContext returns the flash value for the current request.
func FlashFromContext(ctx context.Context) string {
	v, _ := ctx.Value(flashContextKey).(string)
	return v
}
