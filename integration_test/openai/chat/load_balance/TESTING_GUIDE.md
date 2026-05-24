# Load Balance Testing Guide

Quick reference for running and interpreting load balance tests.

## Quick Start

```bash
export TEST_AXONHUB_API_KEY="your-api-key"

make verify
make test
make test-weight
make test-advanced
```

## Test Categories

### Weight-Based Load Balancing

Tests: `TestWeight*`

What it verifies:

- Channels receive requests proportional to their weights.
- Higher weight channels handle more requests before score drops.
- `WeightRoundRobinStrategy` normalizes load by configured weight.

### Connection-Aware Load Balancing

Tests: `TestConnectionAware*`

What it verifies:

- Active upstream requests are tracked.
- Saturated channels are deprioritized when concurrency signals are available.
- Concurrent requests spread across available channels.

### Error-Aware Load Balancing

Tests: `TestErrorAware*`

What it verifies:

- Channels with failures are penalized.
- Healthy channels are preferred.
- The system routes around failing channels.

### Advanced Scenarios

Tests: `TestLoadBalancing*`

What it verifies:

- Top-K selection.
- Priority group handling.
- Inactivity decay.
- Scaling factor effects.
- Composite strategies.

## Interpreting Results

Success indicators:

- Tests pass without API errors.
- Response times stay within `TEST_TIMEOUT`.
- Request distribution roughly follows weights.
- Failover tests recover through alternative channels.

Common failures:

- Uneven distribution: check model support, channel enabled status, and configured weights.
- High failure rate: check provider credentials, channel health, and server logs.
- Connection timeout: increase `TEST_TIMEOUT` and inspect server/provider latency.

## Debugging

Enable server-side debug logs:

```bash
export AXONHUB_DEBUG_LOAD_BALANCER_ENABLED=true
```

Relevant log fields include candidate channel count, selected channel, strategy scores, total score, retry state, and execution time.
