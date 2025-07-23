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

// A comprehensive example demonstrating the four basic Prometheus metric types:
// Counter, Gauge, Histogram, and Summary.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Application metrics
type Metrics struct {
	// Counter: A monotonically increasing value (only goes up or resets to zero)
	// Examples: total requests, total errors, total bytes processed
	totalRequests prometheus.Counter
	totalErrors   prometheus.Counter

	// Gauge: A value that can go up or down
	// Examples: current memory usage, number of active connections, temperature
	activeConnections prometheus.Gauge
	memoryUsageBytes  prometheus.Gauge

	// Histogram: Samples observations (typically durations or sizes) into configurable buckets
	// Examples: request durations, response sizes
	requestDuration prometheus.Histogram
	responseSize    prometheus.Histogram

	// Summary: Similar to histogram but calculates streaming quantiles on the client side
	// Examples: request latencies, processing times
	processingTime prometheus.Summary
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		// Counter examples
		totalRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		}),
		totalErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP errors encountered.",
		}),

		// Gauge examples
		activeConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections.",
		}),
		memoryUsageBytes: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Current memory usage in bytes.",
		}),

		// Histogram examples
		requestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations in seconds.",
			Buckets: prometheus.DefBuckets, // Default buckets: .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10
		}),
		responseSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Histogram of HTTP response sizes in bytes.",
			Buckets: prometheus.ExponentialBuckets(100, 2, 10), // 100, 200, 400, 800, 1600, 3200, 6400, 12800, 25600, 51200
		}),

		// Summary examples
		processingTime: prometheus.NewSummary(prometheus.SummaryOpts{
			Name: "processing_time_seconds",
			Help: "Summary of processing times in seconds.",
			Objectives: map[float64]float64{
				0.5:  0.05,  // 50th percentile with 5% error
				0.9:  0.01,  // 90th percentile with 1% error
				0.99: 0.001, // 99th percentile with 0.1% error
			},
		}),
	}

	// Register all metrics
	reg.MustRegister(
		m.totalRequests,
		m.totalErrors,
		m.activeConnections,
		m.memoryUsageBytes,
		m.requestDuration,
		m.responseSize,
		m.processingTime,
	)

	return m
}

func (m *Metrics) simulateTraffic() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Simulate HTTP requests
		m.totalRequests.Inc()

		// Simulate request duration (0.01 to 2 seconds)
		duration := 0.01 + rand.Float64()*1.99
		m.requestDuration.Observe(duration)

		// Simulate processing time
		processingTime := 0.001 + rand.Float64()*0.1
		m.processingTime.Observe(processingTime)

		// Simulate response size (100 to 10000 bytes)
		responseSize := 100 + rand.Float64()*9900
		m.responseSize.Observe(responseSize)

		// Simulate errors (10% chance)
		if rand.Float64() < 0.1 {
			m.totalErrors.Inc()
		}

		// Simulate active connections (fluctuating between 0-50)
		connections := rand.Intn(51)
		m.activeConnections.Set(float64(connections))

		// Simulate memory usage (fluctuating between 100MB-1GB)
		memoryMB := 100 + rand.Intn(924)
		m.memoryUsageBytes.Set(float64(memoryMB * 1024 * 1024))
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

	// Start simulating traffic
	go metrics.simulateTraffic()

	// Set up HTTP handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
	<title>Basic Metrics Example</title>
</head>
<body>
	<h1>Prometheus Basic Metrics Example</h1>
	<p>This example demonstrates the four basic Prometheus metric types:</p>
	<ul>
		<li><strong>Counter:</strong> total_requests, total_errors</li>
		<li><strong>Gauge:</strong> active_connections, memory_usage_bytes</li>
		<li><strong>Histogram:</strong> request_duration_seconds, response_size_bytes</li>
		<li><strong>Summary:</strong> processing_time_seconds</li>
	</ul>
	<p><a href="/metrics">View Metrics</a></p>
	<p>Metrics are automatically generated with simulated traffic data.</p>
</body>
</html>`)
	})

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	log.Printf("Starting server on %s", *addr)
	log.Printf("Metrics available at http://localhost%s/metrics", *addr)
	log.Printf("Demo page available at http://localhost%s/", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
