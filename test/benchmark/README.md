# Benchmark Tests

This directory contains benchmark tests for Envoy Gateway to verify performance and scalability under various load conditions.

## Available Tests

### ScaleHTTPRoute

Tests gateway scalability with different scales of HTTPRoutes (10, 50, 100, 300, 500, 1000 routes) distributed across multiple hostnames.

**Location:** `tests/scale_httproutes.go`

**Purpose:** Verify that the gateway can handle moderate scale deployments and measure convergence time and performance metrics.

### ScaleHTTPRoute30K

Tests gateway scalability with over 30,000 HTTPRoutes to verify it can handle large-scale enterprise deployments.

**Location:** `tests/scale_httproutes_30k.go`

**Purpose:** Stress test the gateway with a large number of routes (30K+) to:
- Verify the gateway can handle enterprise-scale deployments
- Measure route creation and deletion performance at scale
- Identify potential bottlenecks or memory issues
- Validate that the gateway remains stable under heavy load

**Key Features:**
- Scales up to 30,000 HTTPRoutes
- Distributes routes across 10 different hostnames (3,000 routes per host)
- Logs progress every 1,000 routes for monitoring
- Measures time to create and delete all routes
- Verifies gateway accepts all routes

## Running Benchmark Tests

Benchmark tests are built with the `benchmark` build tag and require a Kubernetes cluster with Envoy Gateway installed.

```bash
# Run all benchmark tests
go test -tags benchmark ./test/benchmark -v

# Run specific test
go test -tags benchmark ./test/benchmark -v -run TestBenchmark
```

## Configuration

Benchmark tests can be configured with flags:
- `--rps`: Requests per second for load testing
- `--connections`: Number of concurrent connections
- `--duration`: Duration of the benchmark test
- `--concurrency`: Number of concurrent workers
- `--report-save-dir`: Directory to save benchmark reports

## Requirements

- Kubernetes cluster with Envoy Gateway installed
- Sufficient cluster resources for the scale being tested
- For the 30K routes test, ensure adequate cluster capacity

## Notes

- The 30K routes test may take considerable time to complete
- Ensure your cluster has sufficient resources (CPU, memory) for large-scale tests
- Monitor cluster resources during test execution
