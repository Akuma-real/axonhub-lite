# Load Balance Integration Tests

This directory contains integration tests for AxonHub's load balancing system. The tests verify that the active strategies route requests across multiple channels correctly.

## Strategies Covered

AxonHub combines these strategies when ranking candidate channels:

1. **ErrorAwareStrategy**
   - Monitors channel health and recent failures.
   - Penalizes channels with consecutive or recent errors.
   - Rewards healthy channels.

2. **WeightRoundRobinStrategy**
   - Distributes load proportionally based on channel weight.
   - Uses `normalizedCount = effectiveCount / (weight / 100.0)`.
   - Applies inactivity decay so old traffic does not permanently penalize a channel.

3. **LatencyAwareStrategy**
   - Prefers lower EWMA first-token latency and higher EWMA output throughput for streaming.
   - Prefers lower EWMA end-to-end latency for non-streaming.

4. **RateLimitAwareStrategy**
   - Respects RPM, TPM, concurrency limits, and cooldowns.
   - Falls back to connection tracker capacity when `MaxConcurrent` is not configured.

## Test Files

`load_balance_test.go` covers the core behavior:

- `TestWeightBasedLoadBalancing`
- `TestWeightRoundRobinLoadBalancing`
- `TestLoadBalancingWithRetry`
- `TestLoadBalancingDebugMode`
- `TestLoadBalancingStrategyComposition`

`advanced_test.go` covers edge cases and optimizations:

- `TestConnectionAwareLoadBalancing`
- `TestErrorAwareLoadBalancing`
- `TestLoadBalancingWithChannelFailover`
- `TestLoadBalancingTopKSelection`
- `TestLoadBalancingPriorityGroups`
- `TestLoadBalancingInactivityDecay`
- `TestLoadBalancingScalingFactor`
- `TestLoadBalancingWeightNormalization`
- `TestLoadBalancingCompositeStrategy`

## Running Tests

Prerequisites:

1. AxonHub server running on `http://localhost:8090` or `TEST_OPENAI_BASE_URL`.
2. `TEST_AXONHUB_API_KEY` set.
3. Multiple enabled channels with different weights.
4. A model available on those channels.

```bash
export TEST_AXONHUB_API_KEY="your-api-key"
export TEST_OPENAI_BASE_URL="http://localhost:8090/v1"
export TEST_MODEL="deepseek-chat"

make test
go test -v ./...
go test -v -run TestWeight
go test -v -run TestConnection
```

## Debug Mode

Enable debug mode on the server to inspect decisions:

```bash
export AXONHUB_DEBUG_LOAD_BALANCER_ENABLED=true
```

Debug logs include individual strategy scores, final rankings, execution times, and the selected channel.

## Expected Results

Success criteria:

- Weight distribution roughly follows configured channel weights.
- Concurrent requests are spread across available channels.
- Unhealthy channels are penalized.
- Retry/failover succeeds when an initial channel fails.
- Priority groups and rate-limit signals affect ranking as configured.

Common failures:

- All requests go to one channel: check channel weights, model support, and enabled status.
- Many requests fail: check channel health, credentials, retry settings, and model availability.
- Requests time out: increase `TEST_TIMEOUT` and check server/channel latency.

## References

- [Load balancing docs](../../../../internal/server/orchestrator/load-balancing.md)
- [Orchestrator](../../../../internal/server/orchestrator/orchestrator.go)
- [Strategy implementations](../../../../internal/server/orchestrator/lb_strategy_*.go)
