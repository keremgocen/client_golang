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

// An example demonstrating metric vectors (metrics with labels) and how to use them
// to partition metrics by different dimensions.
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

// Application metrics with labels
type Metrics struct {
	// CounterVec: Counters partitioned by labels
	requestsTotal    *prometheus.CounterVec
	errorsTotal      *prometheus.CounterVec
	bytesTransferred *prometheus.CounterVec

	// GaugeVec: Gauges partitioned by labels
	activeUsers   *prometheus.GaugeVec
	queueSize     *prometheus.GaugeVec
	cacheHitRatio *prometheus.GaugeVec

	// HistogramVec: Histograms partitioned by labels
	requestDuration *prometheus.HistogramVec
	dbQueryDuration *prometheus.HistogramVec

	// SummaryVec: Summaries partitioned by labels
	apiLatency *prometheus.SummaryVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		// CounterVec examples with different label combinations
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "status", "endpoint"}, // Labels to partition by
		),
		errorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "application_errors_total",
				Help: "Total number of application errors.",
			},
			[]string{"error_type", "service"},
		),
		bytesTransferred: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "bytes_transferred_total",
				Help: "Total bytes transferred.",
			},
			[]string{"direction"}, // "in" or "out"
		),

		// GaugeVec examples
		activeUsers: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_users",
				Help: "Number of active users.",
			},
			[]string{"region", "subscription_type"},
		),
		queueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_size",
				Help: "Current queue size.",
			},
			[]string{"queue_name", "priority"},
		),
		cacheHitRatio: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cache_hit_ratio",
				Help: "Cache hit ratio between 0 and 1.",
			},
			[]string{"cache_name"},
		),

		// HistogramVec examples
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request durations in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		dbQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query durations in seconds.",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~4s
			},
			[]string{"operation", "table"},
		),

		// SummaryVec examples
		apiLatency: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "api_latency_seconds",
				Help: "API latency in seconds.",
				Objectives: map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.99: 0.001,
				},
			},
			[]string{"api_version", "endpoint"},
		),
	}

	// Register all metrics
	reg.MustRegister(
		m.requestsTotal,
		m.errorsTotal,
		m.bytesTransferred,
		m.activeUsers,
		m.queueSize,
		m.cacheHitRatio,
		m.requestDuration,
		m.dbQueryDuration,
		m.apiLatency,
	)

	return m
}

func (m *Metrics) simulateTraffic() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	methods := []string{"GET", "POST", "PUT", "DELETE"}
	statuses := []string{"200", "400", "404", "500"}
	endpoints := []string{"/api/users", "/api/orders", "/api/products", "/health"}
	regions := []string{"us-east-1", "us-west-2", "eu-west-1", "ap-southeast-1"}
	subscriptions := []string{"free", "premium", "enterprise"}
	errorTypes := []string{"validation", "timeout", "internal", "auth"}
	services := []string{"user-service", "order-service", "payment-service"}
	queues := []string{"email", "processing", "notifications"}
	priorities := []string{"high", "medium", "low"}
	caches := []string{"redis", "memcached", "application"}
	dbOps := []string{"SELECT", "INSERT", "UPDATE", "DELETE"}
	tables := []string{"users", "orders", "products", "sessions"}
	apiVersions := []string{"v1", "v2"}

	for range ticker.C {
		// Simulate HTTP requests with different labels
		method := methods[rand.Intn(len(methods))]
		status := statuses[rand.Intn(len(statuses))]
		endpoint := endpoints[rand.Intn(len(endpoints))]

		// Increment request counter with labels
		m.requestsTotal.WithLabelValues(method, status, endpoint).Inc()

		// Record request duration
		duration := 0.01 + rand.Float64()*0.5
		m.requestDuration.WithLabelValues(method, endpoint).Observe(duration)

		// Record API latency
		apiVersion := apiVersions[rand.Intn(len(apiVersions))]
		latency := 0.005 + rand.Float64()*0.1
		m.apiLatency.WithLabelValues(apiVersion, endpoint).Observe(latency)

		// Simulate bytes transferred
		bytes := 100 + rand.Float64()*9900
		direction := "out"
		if rand.Float64() < 0.3 {
			direction = "in"
		}
		m.bytesTransferred.WithLabelValues(direction).Add(bytes)

		// Simulate errors (less frequent)
		if rand.Float64() < 0.1 {
			errorType := errorTypes[rand.Intn(len(errorTypes))]
			service := services[rand.Intn(len(services))]
			m.errorsTotal.WithLabelValues(errorType, service).Inc()
		}

		// Update gauges with different label combinations

		// Active users per region and subscription type
		for _, region := range regions {
			for _, sub := range subscriptions {
				users := rand.Intn(1000)
				m.activeUsers.WithLabelValues(region, sub).Set(float64(users))
			}
		}

		// Queue sizes
		for _, queue := range queues {
			for _, priority := range priorities {
				size := rand.Intn(100)
				m.queueSize.WithLabelValues(queue, priority).Set(float64(size))
			}
		}

		// Cache hit ratios
		for _, cache := range caches {
			ratio := 0.7 + rand.Float64()*0.3 // Between 70% and 100%
			m.cacheHitRatio.WithLabelValues(cache).Set(ratio)
		}

		// Database operations
		if rand.Float64() < 0.3 {
			op := dbOps[rand.Intn(len(dbOps))]
			table := tables[rand.Intn(len(tables))]
			dbDuration := 0.001 + rand.Float64()*0.05
			m.dbQueryDuration.WithLabelValues(op, table).Observe(dbDuration)
		}
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
	<title>Metric Vectors Example</title>
</head>
<body>
	<h1>Prometheus Metric Vectors Example</h1>
	<p>This example demonstrates metric vectors (metrics with labels):</p>
	
	<h2>CounterVec Examples:</h2>
	<ul>
		<li><strong>http_requests_total{method, status, endpoint}</strong> - HTTP requests by method, status, and endpoint</li>
		<li><strong>application_errors_total{error_type, service}</strong> - Errors by type and service</li>
		<li><strong>bytes_transferred_total{direction}</strong> - Bytes transferred in/out</li>
	</ul>
	
	<h2>GaugeVec Examples:</h2>
	<ul>
		<li><strong>active_users{region, subscription_type}</strong> - Active users by region and subscription</li>
		<li><strong>queue_size{queue_name, priority}</strong> - Queue size by name and priority</li>
		<li><strong>cache_hit_ratio{cache_name}</strong> - Cache hit ratio by cache type</li>
	</ul>
	
	<h2>HistogramVec Examples:</h2>
	<ul>
		<li><strong>http_request_duration_seconds{method, endpoint}</strong> - Request durations by method and endpoint</li>
		<li><strong>database_query_duration_seconds{operation, table}</strong> - Database query durations</li>
	</ul>
	
	<h2>SummaryVec Examples:</h2>
	<ul>
		<li><strong>api_latency_seconds{api_version, endpoint}</strong> - API latency by version and endpoint</li>
	</ul>
	
	<p><a href="/metrics">View Metrics</a></p>
	<p>Example PromQL queries:</p>
	<ul>
		<li><code>rate(http_requests_total[5m])</code> - Request rate by labels</li>
		<li><code>sum(active_users) by (region)</code> - Total users per region</li>
		<li><code>histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))</code> - 95th percentile latency</li>
	</ul>
</body>
</html>`)
	})

	// Add an endpoint that demonstrates manual metric recording
	http.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Simulate some processing
		time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)

		// Record metrics for this specific request
		duration := time.Since(start).Seconds()
		metrics.requestDuration.WithLabelValues(r.Method, "/api/test").Observe(duration)
		metrics.requestsTotal.WithLabelValues(r.Method, "200", "/api/test").Inc()

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "ok", "duration": "%v"}`, time.Since(start))
	})

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	}))

	log.Printf("Starting server on %s", *addr)
	log.Printf("Metrics available at http://localhost%s/metrics", *addr)
	log.Printf("Demo page available at http://localhost%s/", *addr)
	log.Printf("Test endpoint available at http://localhost%s/api/test", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
