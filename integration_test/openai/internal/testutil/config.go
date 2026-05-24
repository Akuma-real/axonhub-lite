package testutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

// Config holds configuration for tests
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	Model      string // Default model for tests
}

// DefaultConfig returns a default configuration for tests
func DefaultConfig() *Config {
	return DefaultConfigWithPrefix("")
}

// DefaultConfigWithPrefix returns a default configuration for tests with custom prefix for IDs
func DefaultConfigWithPrefix(_ string) *Config {
	config := &Config{
		APIKey:     getEnvOrDefault("TEST_AXONHUB_API_KEY", ""),
		BaseURL:    getEnvOrDefault("TEST_OPENAI_BASE_URL", "http://localhost:8090/v1"),
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		Model:      getEnvOrDefault("TEST_MODEL", "deepseek-chat"),
	}

	return config
}

// NewClient creates a new OpenAI client with the given configuration
func (c *Config) NewClient() openai.Client {
	if c.APIKey == "" {
		panic("TEST_AXONHUB_API_KEY environment variable is required")
	}

	opts := []option.RequestOption{
		option.WithAPIKey(c.APIKey),
		option.WithBaseURL(c.BaseURL),
	}
	// Remove headers from client initialization - they will be passed at call point
	return openai.NewClient(opts...)
}

// WithHeaders creates a context with the configured headers
func (c *Config) WithHeaders(ctx context.Context) context.Context {
	return ctx
}

// GetHeaderOptions returns request options with the configured headers for call-time usage
func (c *Config) GetHeaderOptions() []option.RequestOption {
	var opts []option.RequestOption
	for k, v := range c.GetHeaders() {
		opts = append(opts, option.WithHeader(k, v))
	}
	return opts
}

// GetHeaders returns the standard headers used in axonhub
func (c *Config) GetHeaders() map[string]string {
	return map[string]string{}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ValidateConfig validates the test configuration
func (c *Config) ValidateConfig() error {
	if c.APIKey == "" {
		return fmt.Errorf("API key is required (set TEST_AXONHUB_API_KEY environment variable)")
	}

	if c.Model == "" {
		return fmt.Errorf("model is required (set TEST_MODEL environment variable)")
	}
	return nil
}

// GetModel returns the configured model as a ChatModel type
func (c *Config) GetModel() openai.ChatModel {
	return openai.ChatModel(c.Model)
}

// GetModelWithFallback returns the configured model, or fallback to GPT-4o if empty
func (c *Config) GetModelWithFallback(fallback openai.ChatModel) openai.ChatModel {
	if c.Model != "" {
		return openai.ChatModel(c.Model)
	}
	return fallback
}

// SetModel sets the model configuration
func (c *Config) SetModel(model string) {
	c.Model = model
}

// IsModelSet returns true if a model is configured
func (c *Config) IsModelSet() bool {
	return c.Model != ""
}
