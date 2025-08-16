package server

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	SearchLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "nodimus_search_latency_ms",
			Help: "Latency of memory searches in milliseconds.",
		},
		[]string{"type"},
	)

	StorageBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nodimus_storage_bytes",
			Help: "Size of storage in bytes.",
		},
		[]string{"type"},
	)

	EmbedCacheHitRatio = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "nodimus_embed_cache_hit_ratio",
			Help: "Cache hit ratio for embeddings.",
		},
	)
)

func init() {
	prometheus.MustRegister(SearchLatency)
	prometheus.MustRegister(StorageBytes)
	prometheus.MustRegister(EmbedCacheHitRatio)
}

// MetricsServer is the server for Prometheus metrics.
type MetricsServer struct {
	*http.Server
}

// NewMetricsServer creates a new metrics server.
func NewMetricsServer(port int, bind string) *MetricsServer {
	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())

	return &MetricsServer{
		Server: &http.Server{
			Addr:    fmt.Sprintf("%s:%d", bind, port),
			Handler: router,
		},
	}
}