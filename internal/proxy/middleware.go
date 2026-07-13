package proxy

import (
	"net"
	"net/http"
	"time"

	"github.com/josesojo2828/proxi-pulse/internal/geo"
	"github.com/josesojo2828/proxi-pulse/internal/store"
)

// Middleware wraps incoming HTTP requests with geo-capture and visit logging.
// For configured domains, it proxies the request. Unknown domains fall through
// to the next handler (management API/UI).
type Middleware struct {
	manager *Manager
	geo     *geo.Cache
	visits  *store.VisitBuffer
}

// NewMiddleware creates a new proxy middleware.
func NewMiddleware(manager *Manager, geoCache *geo.Cache, visitBuffer *store.VisitBuffer) *Middleware {
	return &Middleware{
		manager: manager,
		geo:     geoCache,
		visits:  visitBuffer,
	}
}

// Handler wraps an http.Handler with geo-capture and visit logging.
// If the request matches a configured domain, it proxies; otherwise
// passes through to the next handler.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		domain := r.Host
		// Strip port if present
		if h, _, err := net.SplitHostPort(domain); err == nil {
			domain = h
		}

		proxy := m.manager.Get(domain)
		if proxy == nil {
			// Not a configured domain — pass through to API/UI
			next.ServeHTTP(w, r)
			return
		}

		// Geo lookup
		var country, city string
		if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			if parsed := net.ParseIP(ip); parsed != nil && !geo.IsPrivate(parsed) {
				if rec, err := m.geo.Get(parsed); err == nil {
					country = rec.Country
					city = rec.City
				}
			}
		}

		// Wrap response writer to capture status code
		wr := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Serve the proxy
		proxy.ServeHTTP(wr, r)

		latency := time.Since(start).Seconds() * 1000

		// Log visit asynchronously (non-blocking)
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

// responseWriter captures the HTTP status code for visit logging.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before delegating.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
