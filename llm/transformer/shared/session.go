package shared

import (
	"context"
)

// sessionContextKey is the key used to store and retrieve the session ID from the context.
type sessionContextKey struct{}

// WithSessionID sets the session ID in the context.
// This is essential for features that require cross-request state, such as:
// 1. Prompt Caching: Providers like Anthropic use session IDs to optimize cache hits.
// 2. Request logging: It allows linking the transformation pipeline with request log correlation.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionContextKey{}, sessionID)
}

// GetSessionID retrieves the session ID from the context.
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionContextKey{}).(string)
	return sessionID, ok
}
