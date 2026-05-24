package testutil

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/genai"
)

// Config holds configuration for Gemini tests
type Config struct {
	APIKey     string
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	Model      string // Default model for tests
}

// DefaultConfig returns a default configuration for Gemini tests
func DefaultConfig() *Config {
	return DefaultConfigWithPrefix("")
}

// DefaultConfigWithPrefix returns a default configuration for Gemini tests with custom prefix for IDs
func DefaultConfigWithPrefix(_ string) *Config {
	config := &Config{
		APIKey:     getEnvOrDefault("TEST_AXONHUB_API_KEY", ""),
		BaseURL:    getEnvOrDefault("TEST_GEMINI_BASE_URL", "http://localhost:8090/gemini"),
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		Model:      getEnvOrDefault("TEST_MODEL", "gemini-2.5-flash"),
	}

	return config
}

// NewClient creates a new Gemini client with the given configuration
func (c *Config) NewClient() (*genai.Client, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("TEST_AXONHUB_API_KEY environment variable is required")
	}

	ctx := context.Background()

	// For AxonHub integration, we'll use Gemini API backend
	clientConfig := &genai.ClientConfig{
		APIKey:  c.APIKey,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			BaseURL: c.BaseURL,
		},
	}

	// If custom base URL is provided, we need to handle it differently
	// For now, we'll use the standard Gemini API endpoint
	client, err := genai.NewClient(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return client, nil
}

// WithHeaders creates a context with the configured headers
func (c *Config) WithHeaders(ctx context.Context) context.Context {
	return ctx
}

// GetHeaders returns the standard headers used in AxonHub
func (c *Config) GetHeaders() map[string]string {
	return map[string]string{}
}

// GetHTTPOptions returns HTTPOptions with the configured headers for call-time usage
func (c *Config) GetHTTPOptions() *genai.HTTPOptions {
	headers := c.GetHeaders()
	httpHeaders := make(map[string][]string)
	for k, v := range headers {
		httpHeaders[k] = []string{v}
	}
	return &genai.HTTPOptions{
		Headers: httpHeaders,
	}
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

// GetModel returns the configured model
func (c *Config) GetModel() string {
	return c.Model
}

// GetModelWithFallback returns the configured model, or fallback if empty
func (c *Config) GetModelWithFallback(fallback string) string {
	if c.Model != "" {
		return c.Model
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
