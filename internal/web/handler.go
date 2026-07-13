package web

import (
	"html/template"
	"io/fs"
	"net/http"

	webpkg "github.com/josesojo2828/proxi-pulse/web"

	"github.com/josesojo2828/proxi-pulse/internal/store"
)

// Handler serves the dashboard UI pages.
type Handler struct {
	store    *store.Store
	staticFS http.FileSystem
	baseTmpl *template.Template
}

// NewHandler creates a web handler, parsing templates from the embedded FS.
func NewHandler(s *store.Store) (*Handler, error) {
	// Parse the base layout template
	base, err := template.ParseFS(webpkg.FS, "templates/layout.html")
	if err != nil {
		return nil, err
	}

	staticFS, err := fs.Sub(webpkg.FS, "static")
	if err != nil {
		return nil, err
	}

	return &Handler{
		store:    s,
		staticFS: http.FS(staticFS),
		baseTmpl: base,
	}, nil
}

// Static returns an HTTP handler for embedded static files (CSS, JS, etc.).
func (h *Handler) Static() http.Handler {
	return http.FileServer(h.staticFS)
}

// Dashboard renders the main overview page.
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dashboard.html")
}

// SitesPage renders the sites CRUD page.
func (h *Handler) SitesPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "sites.html")
}

// VisitsPage renders the paginated visits page.
func (h *Handler) VisitsPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "visits.html")
}

// GeoPage renders the geographic traffic map page.
func (h *Handler) GeoPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "geo.html")
}

// render clones the base layout and parses the page template on top,
// so each page gets its own "content" definition without conflicts.
func (h *Handler) render(w http.ResponseWriter, pageFile string) {
	// Clone base layout so we don't modify the shared template
	clone, err := h.baseTmpl.Clone()
	if err != nil {
		http.Error(w, "template clone error", http.StatusInternalServerError)
		return
	}

	// Parse the page-specific template into the clone
	clone, err = clone.ParseFS(webpkg.FS, "templates/"+pageFile)
	if err != nil {
		http.Error(w, "template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := clone.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, "template execute error: "+err.Error(), http.StatusInternalServerError)
	}
}
