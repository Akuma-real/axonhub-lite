package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/looplj/axonhub/internal/requestlog"
)

func TestWithRequestLogging(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	engine := gin.New()
	engine.Use(WithRequestLogging(requestlog.Config{}))

	engine.GET("/", func(c *gin.Context) {
		requestID, ok := requestlog.GetRequestID(c.Request.Context())
		assert.True(t, ok)
		assert.NotEmpty(t, requestID)
		assert.Contains(t, requestID, "ar-")
		c.Status(http.StatusOK)
	})

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("Ah-Request-Id"))
}

func TestWithRequestLoggingCustomHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	engine := gin.New()
	engine.Use(WithRequestLogging(requestlog.Config{
		RequestHeader: "X-Custom-Request-Id",
	}))

	engine.GET("/", func(c *gin.Context) {
		requestID, ok := requestlog.GetRequestID(c.Request.Context())
		assert.True(t, ok)
		assert.NotEmpty(t, requestID)
		c.Status(http.StatusOK)
	})

	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Ah-Request-Id"))
	assert.NotEmpty(t, w.Header().Get("X-Custom-Request-Id"))
}
