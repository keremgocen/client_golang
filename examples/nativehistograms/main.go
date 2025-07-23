// Copyright 2024 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// An example demonstrating Prometheus native histograms, which provide
// better resolution and performance compared to traditional histograms.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	// Traditional histogram for comparison
	traditionalLatency prometheus.Histogram

	// Native histogram - better resolution and performance
	nativeLatency prometheus.Histogram

	// Native histogram with exemplars for tracing integration
	nativeLatencyWithExemplars prometheus.Histogram

	// Native histogram vector for multiple dimensions
	serviceLatency *prometheus.HistogramVec

	// Native histogram for different use cases
	requestSize    prometheus.Histogram
	processingTime prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		// Traditional histogram with fixed buckets
		traditionalLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_traditional_seconds",
			Help:    "Traditional histogram of HTTP request durations.",
			Buckets: prometheus.DefBuckets, // Fixed buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		}),

		// Native histogram with automatic bucket management
		nativeLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "http_request_duration_native_seconds",
			Help:                        "Native histogram of HTTP request durations.",
			NativeHistogramBucketFactor: 1.1, // Enables native histograms with exponential bucket growth
		}),

		// Native histogram with exemplars for distributed tracing
		nativeLatencyWithExemplars: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "http_request_duration_exemplars_seconds",
			Help:                        "Native histogram with exemplars for tracing.",
			NativeHistogramBucketFactor: 1.1,
			NativeHistogramMaxExemplars: 10,              // Maximum number of exemplars to store
			NativeHistogramExemplarTTL:  5 * time.Minute, // How long to keep exemplars
		}),

		// Native histogram vector for multiple services
		serviceLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:                         "service_request_duration_seconds",
				Help:                         "Native histogram of service request durations.",
				NativeHistogramBucketFactor:  1.2,   // Different growth factor
				NativeHistogramZeroThreshold: 0.001, // Observations <= 1ms go in zero bucket
			},
			[]string{"service", "operation"},
		),

		// Native histogram for request sizes with different configuration
		requestSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                           "http_request_size_bytes",
			Help:                           "Native histogram of HTTP request sizes.",
			NativeHistogramBucketFactor:    1.5, // Larger growth factor for sizes
			NativeHistogramMaxBucketNumber: 100, // Limit total number of buckets
		}),

		// Native histogram for processing time with fine resolution
		processingTime: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "processing_time_seconds",
			Help:                        "Native histogram of processing times.",
			NativeHistogramBucketFactor: 1.05, // Very fine resolution
		}),
	}

	// Register all metrics
	reg.MustRegister(
		m.traditionalLatency,
		m.nativeLatency,
		m.nativeLatencyWithExemplars,
		m.serviceLatency,
		m.requestSize,
		m.processingTime,
	)

	return m
}

func (m *Metrics) simulateRequests() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	services := []string{"auth", "user", "order", "payment", "notification"}
	operations := []string{"create", "read", "update", "delete", "list"}

	for range ticker.C {
		// Generate request latency with different distributions

		// Most requests are fast (normal distribution around 50ms)
		var latency float64
		if rand.Float64() < 0.8 {
			// Fast requests: normal distribution around 50ms
			latency = math.Max(0.001, rand.NormFloat64()*0.02+0.05)
		} else if rand.Float64() < 0.95 {
			// Slower requests: normal distribution around 200ms
			latency = math.Max(0.001, rand.NormFloat64()*0.05+0.2)
		} else {
			// Very slow requests: exponential distribution
			latency = 0.5 + rand.ExpFloat64()*2
		}

		// Record in both traditional and native histograms
		m.traditionalLatency.Observe(latency)
		m.nativeLatency.Observe(latency)

		// Record with exemplar (simulating trace ID)
		if rand.Float64() < 0.1 { // 10% of requests have exemplars
			traceID := fmt.Sprintf("trace_%d", rand.Intn(100000))
			// Type assert to access exemplar functionality
			if exemplarObserver, ok := m.nativeLatencyWithExemplars.(prometheus.ExemplarObserver); ok {
				exemplarObserver.ObserveWithExemplar(
					latency,
					prometheus.Labels{"trace_id": traceID},
				)
			}
		} else {
			m.nativeLatencyWithExemplars.Observe(latency)
		}

		// Record service-specific latencies
		service := services[rand.Intn(len(services))]
		operation := operations[rand.Intn(len(operations))]
		serviceLatency := latency * (0.8 + rand.Float64()*0.4) // Add some variation
		m.serviceLatency.WithLabelValues(service, operation).Observe(serviceLatency)

		// Record request sizes (exponential distribution)
		requestSize := 100 + rand.ExpFloat64()*5000 // 100 bytes to ~5KB
		m.requestSize.Observe(requestSize)

		// Record processing times (mostly very fast with occasional spikes)
		var processingTime float64
		if rand.Float64() < 0.95 {
			// Fast processing: 1-10ms
			processingTime = 0.001 + rand.Float64()*0.009
		} else {
			// Slow processing: 50-200ms
			processingTime = 0.05 + rand.Float64()*0.15
		}
		m.processingTime.Observe(processingTime)
	}
}

func main() {
	var (
		addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	)
	flag.Parse()

	// Create a new registry
	reg := prometheus.NewRegistry()

	// Add Go runtime metrics
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// Create and register our application metrics
	metrics := NewMetrics(reg)

	// Start simulating requests
	go metrics.simulateRequests()

	// Set up HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Native Histograms Example</title>
	<style>
		body { font-family: Arial, sans-serif; margin: 40px; }
		code { background-color: #f4f4f4; padding: 2px 4px; border-radius: 3px; }
		.comparison { background-color: #f9f9f9; padding: 15px; border-radius: 5px; margin: 15px 0; }
	</style>
</head>
<body>
	<h1>Prometheus Native Histograms Example</h1>
	
	<div class="comparison">
		<h2>Native Histograms vs Traditional Histograms</h2>
		<table border="1" cellpadding="10" cellspacing="0">
			<tr>
				<th>Feature</th>
				<th>Traditional Histograms</th>
				<th>Native Histograms</th>
			</tr>
			<tr>
				<td>Bucket Definition</td>
				<td>Fixed, predefined buckets</td>
				<td>Automatic, exponential buckets</td>
			</tr>
			<tr>
				<td>Resolution</td>
				<td>Limited by bucket count</td>
				<td>High resolution across range</td>
			</tr>
			<tr>
				<td>Memory Usage</td>
				<td>Fixed per bucket</td>
				<td>Grows with data distribution</td>
			</tr>
			<tr>
				<td>Quantile Accuracy</td>
				<td>Depends on bucket boundaries</td>
				<td>Much more accurate</td>
			</tr>
			<tr>
				<td>Configuration</td>
				<td>Need to choose buckets carefully</td>
				<td>Just set bucket factor</td>
			</tr>
		</table>
	</div>

	<h2>Examples in this Demo</h2>
	<ul>
		<li><code>http_request_duration_traditional_seconds</code> - Traditional histogram with fixed buckets</li>
		<li><code>http_request_duration_native_seconds</code> - Native histogram with factor 1.1</li>
		<li><code>http_request_duration_exemplars_seconds</code> - Native histogram with tracing exemplars</li>
		<li><code>service_request_duration_seconds</code> - Native histogram vector by service and operation</li>
		<li><code>http_request_size_bytes</code> - Native histogram for request sizes</li>
		<li><code>processing_time_seconds</code> - Native histogram with fine resolution (factor 1.05)</li>
	</ul>

	<h2>Key Configuration Options</h2>
	<ul>
		<li><strong>NativeHistogramBucketFactor:</strong> Controls bucket resolution (1.05 = fine, 2.0 = coarse)</li>
		<li><strong>NativeHistogramZeroThreshold:</strong> Values below this go in the zero bucket</li>
		<li><strong>NativeHistogramMaxBucketNumber:</strong> Limits total number of buckets</li>
		<li><strong>NativeHistogramMaxExemplars:</strong> Maximum exemplars to store</li>
		<li><strong>NativeHistogramExemplarTTL:</strong> How long to keep exemplars</li>
	</ul>

	<h2>PromQL Queries</h2>
	<p>Native histograms work with the same PromQL functions:</p>
	<ul>
		<li><code>histogram_quantile(0.95, rate(http_request_duration_native_seconds[5m]))</code></li>
		<li><code>histogram_quantile(0.99, service_request_duration_seconds)</code></li>
		<li><code>histogram_sum(rate(processing_time_seconds[5m]))</code></li>
		<li><code>histogram_count(rate(processing_time_seconds[5m]))</code></li>
	</ul>

	<p><a href="/metrics">View Metrics</a></p>
	<p><strong>Note:</strong> Native histograms require Prometheus v2.40+ with the feature flag enabled.</p>
</body>
</html>`)
	})

	// Endpoint to demonstrate manual latency recording with exemplars
	http.HandleFunc("/api/slow", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Simulate some work
		workTime := 100 + rand.Intn(400) // 100-500ms
		time.Sleep(time.Duration(workTime) * time.Millisecond)

		duration := time.Since(start).Seconds()

		// Record with exemplar including request ID
		requestID := fmt.Sprintf("req_%d", rand.Intn(10000))
		if exemplarObserver, ok := metrics.nativeLatencyWithExemplars.(prometheus.ExemplarObserver); ok {
			exemplarObserver.ObserveWithExemplar(
				duration,
				prometheus.Labels{
					"request_id": requestID,
					"endpoint":   "/api/slow",
				},
			)
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "duration": "%.3fs", "request_id": "%s"}`, duration, requestID)
	})

	// Expose metrics endpoint with OpenMetrics format to support exemplars
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry:          reg,
		EnableOpenMetrics: true, // Required for exemplars
	}))

	log.Printf("Starting server on %s", *addr)
	log.Printf("Metrics available at http://localhost%s/metrics", *addr)
	log.Printf("Demo page available at http://localhost%s/", *addr)
	log.Printf("Slow endpoint available at http://localhost%s/api/slow", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
