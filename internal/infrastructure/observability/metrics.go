package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// NotificationsProcessed records the number of notifications processed successfully or failed.
	NotificationsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_processed_total",
		Help: "The total number of processed notifications",
	}, []string{"channel", "status"})

	// NotificationLatency records the time taken to process a notification.
	NotificationLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "notification_processing_duration_seconds",
		Help:    "Histogram of response latency for processing notifications.",
		Buckets: prometheus.DefBuckets,
	}, []string{"channel"})

	// RateLimitHits records the number of times a rate limit was hit.
	RateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notification_rate_limit_hits_total",
		Help: "The total number of rate limit hits per channel",
	}, []string{"channel"})
)
