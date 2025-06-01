package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	NotificationCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_total",
			Help: "Total number of notifications sent",
		},
		[]string{"type", "status"},
	)

	NotificationLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_latency_seconds",
			Help:    "Time taken to send notifications",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)
) 