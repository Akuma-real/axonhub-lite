package requestlog_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/looplj/axonhub/internal/log"
	"github.com/looplj/axonhub/internal/requestlog"
)

func TestFieldsHook(t *testing.T) {
	hook := log.HookFunc(requestlog.FieldsHook)

	t.Run("with request ID", func(t *testing.T) {
		ctx := requestlog.WithRequestID(context.Background(), "ar-test-request-id")
		fields := hook.Apply(ctx, "test message")
		assert.Len(t, fields, 1)
		assert.Equal(t, "request_id", fields[0].Key)
		assert.Equal(t, "ar-test-request-id", fields[0].String)
	})

	t.Run("with operation name", func(t *testing.T) {
		ctx := requestlog.WithOperationName(context.Background(), "test-operation-name")
		fields := hook.Apply(ctx, "test message")
		assert.Len(t, fields, 1)
		assert.Equal(t, "operation_name", fields[0].Key)
		assert.Equal(t, "test-operation-name", fields[0].String)
	})

	t.Run("with context that doesn't have request ID", func(t *testing.T) {
		ctx := context.Background()
		fields := hook.Apply(ctx, "test message")
		assert.Len(t, fields, 0)
	})

	t.Run("with nil context", func(t *testing.T) {
		fields := hook.Apply(nil, "test message")
		assert.Len(t, fields, 0)
	})
}
