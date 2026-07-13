package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/josesojo2828/proxi-pulse/internal/api"
	"github.com/josesojo2828/proxi-pulse/internal/config"
	"github.com/josesojo2828/proxi-pulse/internal/geo"
	"github.com/josesojo2828/proxi-pulse/internal/metrics"
	"github.com/josesojo2828/proxi-pulse/internal/proxy"
	"github.com/josesojo2828/proxi-pulse/internal/store"
	"github.com/josesojo2828/proxi-pulse/internal/web"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level()})))

	// Store
	st, err := store.New(cfg.DB.Path)
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}
	defer st.Close()

	// GeoIP
	geoLookup, err := geo.NewLookup(cfg.Geo.MMDBPath)
	if err != nil {
		slog.Warn("geoip not available (mmdb file missing?), geo disabled", "error", err)
	} else {
		defer geoLookup.Close()
	}

	var geoCache *geo.Cache
	if geoLookup != nil {
		gc, err := geo.NewCache(geoLookup, cfg.Geo.CacheSize)
		if err != nil {
			return fmt.Errorf("geo cache: %w", err)
		}
		geoCache = gc
	}

	// Visit buffer
	visitBuffer := store.NewVisitBuffer(st, 50, 5*time.Second)
	visitBuffer.Start()
	defer visitBuffer.Stop()

	// Proxy manager
	proxyManager := proxy.NewManager()
	if err := proxyManager.ReloadFromStore(st); err != nil {
		slog.Warn("failed to load sites", "error", err)
	}

	// Router
	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(60 * time.Second))
	r.Use(chiMiddleware.RealIP)

	// Health
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	// Metrics
	r.Get("/metrics", metrics.Handler().ServeHTTP)

	// API routes
	apiRouter := api.NewRouter(st, proxyManager)
	r.Mount("/api", apiRouter)

	// Web UI
	webHandler, err := web.NewHandler(st)
	if err != nil {
		return fmt.Errorf("web handler: %w", err)
	}
	r.Get("/", webHandler.Dashboard)
	r.Get("/sites", webHandler.SitesPage)
	r.Get("/visits", webHandler.VisitsPage)
	r.Get("/geo", webHandler.GeoPage)
	r.Handle("/static/*", http.FileServer(http.Dir("web/static")))

	// Proxy middleware (catch-all for unknown routes)
	proxyMiddleware := proxy.NewMiddleware(proxyManager, geoCache, visitBuffer)
	r.NotFound(proxyMiddleware.Handler(nil).ServeHTTP)

	// Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-done
	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(ctx)
}
