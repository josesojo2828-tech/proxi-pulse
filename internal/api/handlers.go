package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/josesojo2828-tech/proxi-pulse/internal/proxy"
	"github.com/josesojo2828-tech/proxi-pulse/internal/store"
)

// apiResponse is the standard JSON envelope for all API responses.
type apiResponse struct {
	OK    bool        `json:"ok"`
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

// handler holds dependencies for all API route handlers.
type handler struct {
	store        *store.Store
	proxyManager *proxy.Manager
}

// writeJSON serializes the response as JSON and writes it to the client.
func writeJSON(w http.ResponseWriter, status int, resp apiResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("json encode", "error", err)
	}
}

// writeError is a convenience wrapper for error responses.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, apiResponse{OK: false, Error: msg})
}

// parseInt64Param extracts an int64 URL parameter from a chi route.
func parseInt64Param(r *http.Request, name string) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, name), 10, 64)
}

// ListSites handles GET /api/sites
func (h *handler) ListSites(w http.ResponseWriter, r *http.Request) {
	sites, err := h.store.ListSites()
	if err != nil {
		slog.Error("list sites", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to list sites")
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: sites})
}

// CreateSite handles POST /api/sites
// Body: {"domain": "...", "upstream": "..."}
func (h *handler) CreateSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain   string `json:"domain"`
		Upstream string `json:"upstream"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Domain == "" || req.Upstream == "" {
		writeError(w, http.StatusBadRequest, "domain and upstream are required")
		return
	}

	site, err := h.store.CreateSite(req.Domain, req.Upstream)
	if err != nil {
		if errors.Is(err, store.ErrDuplicateDomain) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("create site", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to create site")
		return
	}

	// Register the proxy
	if err := h.proxyManager.AddOrUpdate(req.Domain, req.Upstream); err != nil {
		slog.Error("proxy add", "domain", req.Domain, "error", err)
		// Site was created but proxy registration failed — still report success
		// since the proxy can be reloaded later.
	}

	writeJSON(w, http.StatusCreated, apiResponse{OK: true, Data: site})
}

// GetSite handles GET /api/sites/{id}
func (h *handler) GetSite(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid site id")
		return
	}

	site, err := h.store.GetSite(id)
	if err != nil {
		if errors.Is(err, store.ErrSiteNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		slog.Error("get site", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get site")
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: site})
}

// UpdateSite handles PUT /api/sites/{id}
// Body: {"domain": "...", "upstream": "..."}
func (h *handler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid site id")
		return
	}

	var req struct {
		Domain   string `json:"domain"`
		Upstream string `json:"upstream"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	site, err := h.store.UpdateSite(id, req.Domain, req.Upstream)
	if err != nil {
		if errors.Is(err, store.ErrSiteNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		if errors.Is(err, store.ErrDuplicateDomain) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		slog.Error("update site", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to update site")
		return
	}

	// Update the proxy
	if err := h.proxyManager.AddOrUpdate(req.Domain, req.Upstream); err != nil {
		slog.Error("proxy update", "domain", req.Domain, "error", err)
	}

	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: site})
}

// DeleteSite handles DELETE /api/sites/{id}
func (h *handler) DeleteSite(w http.ResponseWriter, r *http.Request) {
	id, err := parseInt64Param(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid site id")
		return
	}

	// Get the site first to know the domain for proxy removal
	site, err := h.store.GetSite(id)
	if err != nil {
		if errors.Is(err, store.ErrSiteNotFound) {
			writeError(w, http.StatusNotFound, "site not found")
			return
		}
		slog.Error("get site for delete", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete site")
		return
	}

	if err := h.store.DeleteSite(id); err != nil {
		slog.Error("delete site", "id", id, "error", err)
		writeError(w, http.StatusInternalServerError, "failed to delete site")
		return
	}

	// Remove the proxy
	h.proxyManager.Remove(site.Domain)

	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: map[string]string{"status": "deleted"}})
}

// StatsOverview handles GET /api/stats/overview
func (h *handler) StatsOverview(w http.ResponseWriter, r *http.Request) {
	sites, err := h.store.ListSites()
	if err != nil {
		slog.Error("stats overview: list sites", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	totalSites := len(sites)
	activeSites := 0
	for _, s := range sites {
		if s.Active {
			activeSites++
		}
	}

	totalVisits, err := h.store.GetTotalVisits()
	if err != nil {
		slog.Error("stats overview: total visits", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	visitsToday, err := h.store.GetVisitsToday()
	if err != nil {
		slog.Error("stats overview: visits today", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get stats")
		return
	}

	data := map[string]interface{}{
		"total_sites":  totalSites,
		"active_sites": activeSites,
		"total_visits": totalVisits,
		"visits_today": visitsToday,
	}

	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: data})
}

// StatsVisits handles GET /api/stats/visits
// Query params: page (default 1), limit (default 20), site_id (optional)
func (h *handler) StatsVisits(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	var siteID *int64
	if sid := r.URL.Query().Get("site_id"); sid != "" {
		if v, err := strconv.ParseInt(sid, 10, 64); err == nil {
			siteID = &v
		}
	}

	visits, total, err := h.store.GetVisits(page, limit, siteID)
	if err != nil {
		slog.Error("stats visits", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get visits")
		return
	}

	data := map[string]interface{}{
		"visits": visits,
		"total":  total,
		"page":   page,
		"limit":  limit,
	}
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: data})
}

// StatsCountries handles GET /api/stats/countries
func (h *handler) StatsCountries(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetVisitsByCountry()
	if err != nil {
		slog.Error("stats countries", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to get country stats")
		return
	}
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Data: stats})
}
