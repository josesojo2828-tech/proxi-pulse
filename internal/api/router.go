package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/josesojo2828/proxi-pulse/internal/proxy"
	"github.com/josesojo2828/proxi-pulse/internal/store"
)

// NewRouter creates a chi router with all API routes.
func NewRouter(st *store.Store, pm *proxy.Manager) chi.Router {
	r := chi.NewRouter()
	h := &handler{store: st, proxyManager: pm}

	r.Get("/sites", h.ListSites)
	r.Post("/sites", h.CreateSite)
	r.Get("/sites/{id}", h.GetSite)
	r.Put("/sites/{id}", h.UpdateSite)
	r.Delete("/sites/{id}", h.DeleteSite)

	r.Get("/stats/overview", h.StatsOverview)
	r.Get("/stats/visits", h.StatsVisits)
	r.Get("/stats/countries", h.StatsCountries)

	return r
}
