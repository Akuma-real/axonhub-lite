# OpenAI Go SDK Integration Tests

This directory contains comprehensive integration tests for the OpenAI Go SDK, demonstrating various API usage patterns with proper header handling for AxonHub integration.

## Test Structure

Each test case is organized in its own directory with dedicated tests and documentation:

### Test Cases

1. **[qa_simple](./qa_simple)** - Basic chat completion tests
   - Simple Q&A interactions using configured model
   - Multiple question testing with context preservation
   - Configurable model support

2. **[tool_single](./tool_single)** - Single tool calling tests
   - Weather information tool
   - Calculator tool
   - Tool call validation and continuation

3. **[tool_multiple](./tool_multiple)** - Multiple tool calling tests
   - Sequential tool execution
   - Parallel tool calls
   - Forced tool choice scenarios

4. **[streaming](./streaming)** - Streaming response tests
   - Real-time response streaming
   - Streaming with tools
   - Long response handling
   - Error handling in streams

5. **[conversation](./conversation)** - Multi-turn conversation tests
   - Context preservation
   - System prompt influence
   - Tool integration in conversations
   - Message history management

## Common Integration Features

### API Patterns Demonstrated
- **Chat Completions**: Basic and advanced chat interactions
- **Function Calling**: Tool definition, execution, and continuation
- **Streaming**: Real-time response processing
- **Context Management**: Multi-turn conversation state
- **Error Handling**: Proper error detection and recovery

### Testing Utilities
- **Configuration Management**: Environment-based configuration
- **Client Setup**: Proper OpenAI client initialization
- **Response Validation**: Consistent response verification
- **Tool Simulation**: Mock tool implementations for testing

## Prerequisites

### Environment Setup
1. **Go 1.25+**: Ensure Go is installed and configured
2. **AxonHub API Key**: Set the `TEST_AXONHUB_API_KEY` environment variable
3. **Dependencies**: Run `go mod tidy` to install required packages

### Required Environment Variables
```bash
export TEST_AXONHUB_API_KEY="your-api-key-here"
export TEST_OPENAI_BASE_URL="https://api.openai.com/v1"  # Optional, defaults to OpenAI
export TEST_PROJECT_ID="test-project"              # Optional, defaults provided
export TEST_MODEL="gpt-4o"                         # Optional, defaults to gpt-4o
```

## Running Tests

### Run All Tests
```bash
# From the integration test directory
cd /path/to/axonhub/integration_test/openai

# Run all tests
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Run Specific Test Cases
```bash
# Simple Q&A tests
go test -v ./qa_simple

# Single tool tests
go test -v ./tool_single

# Multiple tool tests
go test -v ./tool_multiple

# Streaming tests
go test -v ./streaming

# Conversation tests
go test -v ./conversation
```

### Run Individual Tests
```bash
# Run specific test function
go test -v -run TestSimpleQA ./qa_simple

# Run streaming with tools
go test -v -run TestStreamingWithTools ./streaming

# Run conversation context preservation
go test -v -run TestConversationContextPreservation ./conversation
```

## Test Configuration

### Configuration Structure
Tests use a centralized configuration system defined in `internal/testutil/`:

- **Config**: Environment variable management
- **Client**: OpenAI client setup with authentication
- **Headers**: AxonHub header generation
- **Helper**: Common test utilities and validation

### Custom Configuration
You can customize test behavior with environment variables:

```bash
# Custom API endpoint
export TEST_OPENAI_BASE_URL="https://your-proxy.com/v1"

# Custom model for tests
export TEST_MODEL="gpt-4o-mini"
```

## Model Configuration

All tests now use a configurable model system that allows you to specify which model to use for testing:

### Environment Variable
Set the `TEST_MODEL` environment variable to specify which model to use:

```bash
# Use GPT-4o (default)
export TEST_MODEL="gpt-4o"

# Use GPT-3.5 Turbo
export TEST_MODEL="gpt-3.5-turbo"

# Use GPT-4o Mini
export TEST_MODEL="gpt-4o-mini"

# Use O1 Mini
export TEST_MODEL="o1-mini"
```

### Code Usage
Tests automatically use the configured model through the `helper.GetModel()` method:

```go
// Get the configured model
model := helper.GetModel()

// Use in API calls
params := openai.ChatCompletionNewParams{
    Messages: messages,
    Model:    model,
}
```

If no `TEST_MODEL` is specified, the system defaults to `gpt-4o`.

## Integration with AxonHub

These tests demonstrate proper integration patterns for AxonHub:

### Context Management
Tests show how to maintain conversation context and state across multiple API calls using provider-native message history.

### Error Handling
Proper error handling and validation patterns that integrate well with AxonHub's error reporting and logging systems.

## Development

### Adding New Tests
1. Create a new directory under the test root
2. Add your test files following the existing patterns
3. Include a README.md with documentation
4. Update this main README if needed

### Test Patterns
Follow these patterns for consistency:

- Use `TestMain` for setup/teardown
- Leverage `internal/testutil` for common functionality
- Include comprehensive error checking
- Document test expectations and scenarios
- Use descriptive test names and logging

### Code Quality
- Follow Go testing best practices
- Include proper error handling
- Use descriptive variable names
- Add comments for complex logic
- Maintain consistent formatting

## Troubleshooting

### Common Issues

**Missing API Key**
```
Error: TEST_AXONHUB_API_KEY environment variable is required
Solution: Set TEST_AXONHUB_API_KEY environment variable
```

**Import Errors**
```
Solution: Run `go mod tidy` to resolve dependencies
```

**Network Issues**
```
Solution: Check OPENAI_BASE_URL and network connectivity
```

**Rate Limiting**
```
Solution: Tests include delays and respect rate limits
```

### Debug Mode
Enable verbose logging by setting test flags:

```bash
go test -v -args -test.debug=true
```
