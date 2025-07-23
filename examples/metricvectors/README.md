# Metric Vectors Example

This example demonstrates how to use metric vectors (metrics with labels) to partition your metrics by different dimensions.

## What are Metric Vectors?

Metric vectors allow you to create multiple time series from a single metric definition by using labels. This is essential for creating meaningful, queryable metrics in Prometheus.

## Examples in this Demo

### CounterVec

- `http_requests_total{method, status, endpoint}` - Track requests by HTTP method, status code, and endpoint
- `application_errors_total{error_type, service}` - Track errors by type and service
- `bytes_transferred_total{direction}` - Track data transfer by direction (in/out)

### GaugeVec

- `active_users{region, subscription_type}` - Track active users by region and subscription tier
- `queue_size{queue_name, priority}` - Track queue sizes by queue name and priority level
- `cache_hit_ratio{cache_name}` - Track cache performance by cache type

### HistogramVec

- `http_request_duration_seconds{method, endpoint}` - Track request latency by method and endpoint
- `database_query_duration_seconds{operation, table}` - Track database performance by operation and table

### SummaryVec

- `api_latency_seconds{api_version, endpoint}` - Track API latency by version and endpoint

## Running the Example

```bash
go run main.go
```

Then visit:

- http://localhost:8080/ - Demo page with explanations
- http://localhost:8080/metrics - Raw metrics output
- http://localhost:8080/api/test - Test endpoint that records metrics

## Key Concepts

### Using WithLabelValues()

```go
// Define a CounterVec
requestsTotal := prometheus.NewCounterVec(
    prometheus.CounterOpts{Name: "requests_total"},
    []string{"method", "status"},
)

// Use it with specific label values
requestsTotal.WithLabelValues("GET", "200").Inc()
requestsTotal.WithLabelValues("POST", "404").Inc()
```

### Using With()

```go
// Alternative syntax using a map
requestsTotal.With(prometheus.Labels{
    "method": "GET",
    "status": "200",
}).Inc()
```

## Useful PromQL Queries

Once you have labeled metrics, you can use powerful PromQL queries:

```promql
# Request rate by endpoint
rate(http_requests_total[5m])

# Total active users per region
sum(active_users) by (region)

# Error rate by service
rate(application_errors_total[5m]) / rate(http_requests_total[5m])

# 95th percentile latency by endpoint
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Top endpoints by request volume
topk(5, sum(rate(http_requests_total[5m])) by (endpoint))
```

## Best Practices

1. **Choose meaningful labels** - Labels should be dimensions you want to query by
2. **Avoid high cardinality** - Don't use user IDs, timestamps, or other unique values as labels
3. **Be consistent** - Use the same label names across related metrics
4. **Pre-register label combinations** - Call `WithLabelValues()` early to avoid missing time series
5. **Use `rate()` with counters** - Always use rate() or increase() when querying counter metrics

## Label Cardinality Warning

Be careful with label cardinality! Each unique combination of label values creates a new time series. For example:

- 5 methods × 10 status codes × 20 endpoints = 1,000 time series
- This is manageable, but avoid labels with unbounded values
