package web

import (
	"html/template"
	"io/fs"
	"net/http"

	webpkg "github.com/josesojo2828/proxi-pulse/web"

	"github.com/josesojo2828/proxi-pulse/internal/store"
)

type Handler struct {
	store    *store.Store
	staticFS http.FileSystem
	baseTmpl *template.Template
}

func NewHandler(s *store.Store) (*Handler, error) {
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

func (h *Handler) Static() http.Handler {
	return http.FileServer(h.staticFS)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	h.render(w, "dashboard.html")
}

func (h *Handler) SitesPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "sites.html")
}

func (h *Handler) VisitsPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "visits.html")
}

func (h *Handler) render(w http.ResponseWriter, pageFile string) {
	clone, err := h.baseTmpl.Clone()
	if err != nil {
		http.Error(w, "template clone error", http.StatusInternalServerError)
		return
	}

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
