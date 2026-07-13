package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics descriptors
var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "proxi_pulse_requests_total",
			Help: "Total number of proxied requests",
		},
		[]string{"host", "country", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "proxi_pulse_request_duration_ms",
			Help:    "Request duration in milliseconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"host", "country"},
	)

	ActiveConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "proxi_pulse_active_connections",
			Help: "Number of active connections",
		},
	)

	GeoLookupDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "proxi_pulse_geo_lookup_duration_ms",
			Help:    "GeoIP lookup duration in milliseconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1},
		},
	)

	SitesTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "proxi_pulse_sites_total",
			Help: "Total number of configured sites",
		},
	)
)

func init() {
	prometheus.MustRegister(RequestsTotal, RequestDuration, ActiveConnections, GeoLookupDuration, SitesTotal)
}

func Handler() http.Handler {
	return promhttp.Handler()
}