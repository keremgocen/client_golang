# Basic Metrics Example

This example demonstrates the four basic Prometheus metric types:

## Metric Types

### Counter

A monotonically increasing metric that only goes up (or resets to zero on restart).

- `http_requests_total` - Total number of HTTP requests
- `http_errors_total` - Total number of HTTP errors

### Gauge

A metric that can increase or decrease over time.

- `active_connections` - Current number of active connections
- `memory_usage_bytes` - Current memory usage in bytes

### Histogram

A metric that samples observations and counts them in configurable buckets.

- `http_request_duration_seconds` - HTTP request durations
- `http_response_size_bytes` - HTTP response sizes

### Summary

A metric that samples observations and calculates configurable quantiles.

- `processing_time_seconds` - Processing time with 50th, 90th, and 99th percentiles

## Running the Example

```bash
go run main.go
```

Then visit:

- http://localhost:8080/ - Demo page
- http://localhost:8080/metrics - Raw metrics

The example automatically generates simulated traffic data to demonstrate how each metric type behaves.

## Key Learning Points

1. **Counters** only increase and are useful for counting events
2. **Gauges** can go up and down and represent current state
3. **Histograms** show distributions of values across predefined buckets
4. **Summaries** calculate quantiles on the client side

## Best Practices

- Use counters for total counts (requests, errors, bytes transferred)
- Use gauges for current state values (memory usage, active connections, queue size)
- Use histograms for measuring durations and sizes (prefer over summaries)
- Use summaries only when you need exact quantiles and can't use histograms
