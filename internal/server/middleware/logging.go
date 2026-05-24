package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/looplj/axonhub/internal/requestlog"
)

// WithRequestLogging saves the request ID and operation name to the request context.
// The logger can include those fields in subsequent logs.
func WithRequestLogging(config requestlog.Config) gin.HandlerFunc {
	// Use the configured request header name, or default to "AH-Request-Id"
	requestHeader := config.RequestHeader
	if requestHeader == "" {
		requestHeader = "AH-Request-Id"
	}

	return func(c *gin.Context) {
		// Generate request ID for each request
		requestID := requestlog.GenerateRequestID()

		// Set request ID header in response
		c.Header(requestHeader, requestID)

		ctx := requestlog.WithRequestID(c.Request.Context(), requestID)

		if !strings.HasSuffix(c.FullPath(), "/graphql") {
			operationName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
			ctx = requestlog.WithOperationName(ctx, operationName)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
