package requestlog

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/looplj/axonhub/internal/contexts"
)

type Config struct {
	// RequestHeader is the response header name for the generated request ID.
	// Default to "AH-Request-Id".
	RequestHeader string `conf:"request_header" yaml:"request_header" json:"request_header"`
}

// GenerateRequestID generates a request ID, formatted as ar-{{uuid}}.
func GenerateRequestID() string {
	id := uuid.New()
	return fmt.Sprintf("ar-%s", id.String())
}

// WithRequestID stores request ID in context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return contexts.WithRequestID(ctx, requestID)
}

// GetRequestID gets request ID from context.
func GetRequestID(ctx context.Context) (string, bool) {
	return contexts.GetRequestID(ctx)
}

// WithOperationName stores operation name in context.
func WithOperationName(ctx context.Context, name string) context.Context {
	return contexts.WithOperationName(ctx, name)
}

// GetOperationName gets operation name from context.
func GetOperationName(ctx context.Context) (string, bool) {
	return contexts.GetOperationName(ctx)
}
