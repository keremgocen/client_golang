# Native Histograms Example

This example demonstrates Prometheus native histograms, a new histogram implementation that provides better resolution and performance compared to traditional histograms.

## What are Native Histograms?

Native histograms are an improved histogram implementation introduced in Prometheus that addresses several limitations of traditional histograms:

### Traditional Histograms vs Native Histograms

| Feature | Traditional Histograms | Native Histograms |
|---------|----------------------|-------------------|
| **Bucket Definition** | Fixed, predefined buckets | Automatic, exponential buckets |
| **Resolution** | Limited by bucket count | High resolution across range |
| **Memory Usage** | Fixed per bucket | Grows with data distribution |
| **Quantile Accuracy** | Depends on bucket boundaries | Much more accurate |
| **Configuration** | Need to choose buckets carefully | Just set bucket factor |

### Key Benefits

1. **Automatic Bucket Management**: No need to predefine bucket boundaries
2. **High Resolution**: Exponential buckets provide fine granularity across the entire range
3. **Better Quantile Accuracy**: More precise percentile calculations
4. **Simplified Configuration**: Only need to set a bucket factor instead of choosing buckets
5. **Exemplar Support**: Built-in support for tracing integration

## Key Configuration Options

- `NativeHistogramBucketFactor`: Controls bucket resolution (1.05 = very fine, 2.0 = coarse)
- `NativeHistogramZeroThreshold`: Values below this go in the zero bucket
- `NativeHistogramMaxBucketNumber`: Limits total number of buckets to control memory usage
- `NativeHistogramMaxExemplars`: Maximum number of exemplars to store
- `NativeHistogramExemplarTTL`: How long to keep exemplars

## Examples in this Demo

This example includes several types of native histograms:

1. **Basic Native Histogram**: Simple configuration with bucket factor
2. **Histogram with Exemplars**: For distributed tracing integration
3. **Histogram Vector**: Multiple histograms with labels
4. **Comparison with Traditional**: Side-by-side comparison

## Running the Example

```bash
cd examples/nativehistograms
go run main.go
```

Then visit:
- http://localhost:8080/ - Demo page with explanations
- http://localhost:8080/metrics - Metrics endpoint
- http://localhost:8080/api/slow - Endpoint that records exemplars

## PromQL Queries

Native histograms work with the same PromQL functions as traditional histograms:

```promql
# 95th percentile latency
histogram_quantile(0.95, rate(http_request_duration_native_seconds[5m]))

# Request rate
histogram_count(rate(http_request_duration_native_seconds[5m]))

# Total request time
histogram_sum(rate(http_request_duration_native_seconds[5m]))

# Quantiles by service
histogram_quantile(0.99, service_request_duration_seconds)
```

## Requirements

Native histograms require:
- Prometheus v2.40+ with native histograms feature enabled
- For exemplars: OpenMetrics format (enabled in this example)

## Best Practices

1. **Choose appropriate bucket factors**:
   - 1.05-1.1: Very fine resolution for precise measurements
   - 1.2-1.5: Good balance of resolution and memory usage
   - 2.0+: Coarse resolution for rough measurements

2. **Set zero threshold** for metrics with many small values
3. **Limit bucket count** with `MaxBucketNumber` for memory control
4. **Use exemplars** for tracing integration when needed

## When to Use Native Histograms

Native histograms are ideal when:
- You need accurate quantiles across a wide range of values
- You don't know the expected value distribution in advance
- You want simplified configuration without choosing buckets
- You need high-resolution histograms for SLA monitoring
- You want to integrate with distributed tracing systems
