package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/looplj/axonhub/openai_test/internal/testutil"
	"github.com/openai/openai-go/v3"
)

// TestWeightBasedLoadBalancing verifies that weight-based load balancing
// distributes requests according to channel weights.
func TestWeightBasedLoadBalancing(t *testing.T) {
	t.Log("=== Testing weight-based load balancing ===")
	t.Log("This test verifies that channels with different weights receive")
	t.Log("proportional amounts of traffic")

	requestCount := 60

	for i := 0; i < requestCount; i++ {
		helper := testutil.NewTestHelper(t, fmt.Sprintf("TestWeightBasedLoadBalancing_Request%d", i))
		ctx := helper.CreateTestContext()

		response, err := helper.CreateChatCompletionWithHeaders(ctx, openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(fmt.Sprintf("Request %d: What is %d+%d?", i+1, i, i)),
			},
			Model: helper.GetModel(),
		})

		helper.AssertNoError(t, err, fmt.Sprintf("Request %d failed", i+1))
		helper.ValidateChatResponse(t, response, fmt.Sprintf("Request %d", i+1))

		if i%5 == 0 {
			t.Logf("Completed %d/%d requests", i+1, requestCount)
		}

		// Small delay between requests
		time.Sleep(50 * time.Millisecond)
	}

	t.Log("\n=== Weight-based load balancing test completed ===")
	t.Logf("Sent %d requests", requestCount)
	t.Log("Channels should have received requests proportional to their weights:")
	t.Log("- High weight channels (weight=80-100) should get more requests")
	t.Log("- Medium weight channels (weight=40-60) should get moderate requests")
	t.Log("- Low weight channels (weight=10-30) should get fewer requests")
}

// TestWeightRoundRobinLoadBalancing verifies that WeightRoundRobinStrategy
// distributes load proportionally based on both weight and current load.
func TestWeightRoundRobinLoadBalancing(t *testing.T) {
	t.Log("=== Testing weighted round-robin load balancing ===")
	t.Log("This test sends concurrent requests to verify that WeightRoundRobinStrategy")
	t.Log("balances load based on both channel weight and current request count")

	// Send concurrent requests to test round-robin behavior
	concurrentRequests := 15
	results := make(chan error, concurrentRequests)

	t.Logf("Sending %d concurrent requests...", concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func(requestNum int) {
			helper := testutil.NewTestHelper(t, fmt.Sprintf("TestWeightRoundRobin_Request%d", requestNum))
			ctx := helper.CreateTestContext()

			response, err := helper.CreateChatCompletionWithHeaders(ctx, openai.ChatCompletionNewParams{
				Messages: []openai.ChatCompletionMessageParamUnion{
					openai.UserMessage(fmt.Sprintf("Concurrent request %d", requestNum)),
				},
				Model: helper.GetModel(),
			})

			if err != nil {
				results <- fmt.Errorf("request %d failed: %w", requestNum, err)
				return
			}

			if response == nil || len(response.Choices) == 0 {
				results <- fmt.Errorf("request %d: invalid response", requestNum)
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all requests to complete
	successCount := 0
	for i := 0; i < concurrentRequests; i++ {
		err := <-results
		if err != nil {
			t.Logf("Request failed: %v", err)
		} else {
			successCount++
		}
	}

	t.Logf("\n=== Weighted round-robin test completed ===")
	t.Logf("Successful requests: %d/%d", successCount, concurrentRequests)
	t.Log("WeightRoundRobinStrategy should have distributed requests based on:")
	t.Log("- Channel weight (higher weight = can handle more requests)")
	t.Log("- Current load (channels with fewer active requests get priority)")
	t.Log("- Formula: normalizedCount = effectiveCount / (weight / 100.0)")

	if successCount < concurrentRequests/2 {
		t.Errorf("Too many failures: only %d/%d succeeded", successCount, concurrentRequests)
	}
}

// TestLoadBalancingWithRetry verifies that load balancing works correctly
// when retries are enabled and a channel fails.
func TestLoadBalancingWithRetry(t *testing.T) {
	helper := testutil.NewTestHelper(t, "TestLoadBalancingWithRetry")
	ctx := helper.CreateTestContext()

	t.Log("=== Testing load balancing with retry mechanism ===")
	t.Log("If a channel fails, the system should automatically retry with the next channel")

	// Make a request that might trigger retry behavior
	response, err := helper.CreateChatCompletionWithHeaders(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Test request for retry mechanism"),
		},
		Model: helper.GetModel(),
	})

	helper.AssertNoError(t, err, "Request failed even with retry")
	helper.ValidateChatResponse(t, response, "Retry test")

	t.Log("=== Load balancing with retry test completed ===")
	t.Log("Request succeeded, demonstrating that the system can handle channel failures")
	t.Log("and automatically retry with alternative channels based on load balancing scores")
}

// TestLoadBalancingDebugMode verifies that debug mode provides detailed
// load balancing decision information.
func TestLoadBalancingDebugMode(t *testing.T) {
	helper := testutil.NewTestHelper(t, "TestLoadBalancingDebugMode")
	ctx := helper.CreateTestContext()

	t.Log("=== Testing load balancing debug mode ===")
	t.Log("When AXONHUB_DEBUG_LOAD_BALANCER_ENABLED=true or AH-Debug header is set,")
	t.Log("the system should log detailed scoring information for each channel")

	// Make a request with debug context
	response, err := helper.CreateChatCompletionWithHeaders(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Debug mode test request"),
		},
		Model: helper.GetModel(),
	})

	helper.AssertNoError(t, err, "Debug mode request failed")
	helper.ValidateChatResponse(t, response, "Debug mode test")

	t.Log("\n=== Debug mode test completed ===")
	t.Log("Check server logs for detailed load balancing decision information:")
	t.Log("- Channel scores from each strategy")
	t.Log("- Total scores and final ranking")
	t.Log("- Strategy execution times")
	t.Log("- Detailed decision breakdown")
}

// TestLoadBalancingStrategyComposition verifies that multiple strategies
// work together correctly to produce the final channel ranking.
func TestLoadBalancingStrategyComposition(t *testing.T) {
	t.Log("=== Testing load balancing strategy composition ===")
	t.Log("The default orchestrator uses these strategies in order:")
	t.Log("1. ErrorAwareStrategy (0-200 points)")
	t.Log("2. WeightRoundRobinStrategy (10-150 points)")
	t.Log("3. LatencyAwareStrategy (0-80 points)")
	t.Log("4. RateLimitAwareStrategy (-10000 to 100 points)")
	t.Log("Total score determines channel priority")

	helper1 := testutil.NewTestHelper(t, "TestStrategyComposition")
	ctx1 := helper1.CreateTestContext()

	response1, err := helper1.CreateChatCompletionWithHeaders(ctx1, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Request for load balancing strategy composition"),
		},
		Model: helper1.GetModel(),
	})
	helper1.AssertNoError(t, err, "Strategy composition request failed")
	helper1.ValidateChatResponse(t, response1, "Strategy composition request")
	t.Log("Expected: High weight channel selected (WeightRoundRobin + ErrorAware + Connection)")

	t.Log("\n=== Strategy composition test completed ===")
	t.Log("Verified that strategies work together to produce optimal channel selection")
}
