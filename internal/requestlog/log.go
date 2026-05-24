package requestlog

import (
	"context"

	"github.com/looplj/axonhub/internal/log"
)

func SetupLogger(logger *log.Logger) {
	logger.AddHook(log.HookFunc(FieldsHook))
}

// FieldsHook adds request ID and operation name to log entries if they exist in context.
func FieldsHook(ctx context.Context, msg string, fields ...log.Field) []log.Field {
	if ctx == nil {
		return fields
	}

	if requestID, ok := GetRequestID(ctx); ok {
		fields = append(fields, log.String("request_id", requestID))
	}

	if operationName, ok := GetOperationName(ctx); ok {
		fields = append(fields, log.String("operation_name", operationName))
	}

	return fields
}
