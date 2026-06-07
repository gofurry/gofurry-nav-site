package report

import "context"

type runIDContextKey struct{}

// ContextWithRunID returns a child context carrying the current v2 run id.
func ContextWithRunID(ctx context.Context, runID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, runIDContextKey{}, runID)
}

// RunIDFromContext returns the v2 run id carried by ctx, if present.
func RunIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(runIDContextKey{}).(string)
	return value
}
