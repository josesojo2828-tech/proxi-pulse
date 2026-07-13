package proxy

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/josesojo2828/proxi-pulse/internal/geo"
	"github.com/josesojo2828/proxi-pulse/internal/metrics"
	"github.com/josesojo2828/proxi-pulse/internal/store"
)

type Middleware struct {
	manager *Manager
	geo     *geo.Cache
	visits  *store.VisitBuffer
}

func NewMiddleware(manager *Manager, geoCache *geo.Cache, visitBuffer *store.VisitBuffer) *Middleware {
	return &Middleware{
		manager: manager,
		geo:     geoCache,
		visits:  visitBuffer,
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		domain := r.Host
		if h, _, err := net.SplitHostPort(domain); err == nil {
			domain = h
		}

		proxy := m.manager.Get(domain)
		if proxy == nil {
			next.ServeHTTP(w, r)
			return
		}

		var country, city string
		if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if parsed := net.ParseIP(ip); parsed != nil && !geo.IsPrivate(parsed) {
				if rec, err := m.geo.Get(parsed); err == nil {
					country = rec.Country
					city = rec.City
				}
			}
		}

		wr := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		proxy.ServeHTTP(wr, r)

		latency := time.Since(start).Seconds() * 1000

		metrics.RequestsTotal.WithLabelValues(domain, country, strconv.Itoa(wr.statusCode)).Inc()
		metrics.RequestDuration.WithLabelValues(domain, country).Observe(latency)

		m.visits.Log(store.Visit{
			IP:         r.RemoteAddr,
			Country:    country,
			City:       city,
			Path:       r.URL.Path,
			Method:     r.Method,
			StatusCode: wr.statusCode,
			LatencyMs:  latency,
			UserAgent:  r.UserAgent(),
			Referer:    r.Referer(),
		})
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
