package demo

import (
	"html/template"
	"io/fs"
	"net/http"
)

type Handler struct {
	stripePublishableKey string
	indexTemplate        *template.Template
	staticHandler        http.Handler
}

func NewHandler(stripePublishableKey string) (*Handler, error) {
	tmpl, err := template.ParseFS(assets, "templates/index.html")
	if err != nil {
		return nil, err
	}

	staticFS, err := fs.Sub(assets, "static")
	if err != nil {
		return nil, err
	}

	return &Handler{
		stripePublishableKey: stripePublishableKey,
		indexTemplate:        tmpl,
		staticHandler:        http.StripPrefix("/demo/static/", http.FileServer(http.FS(staticFS))),
	}, nil
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/demo", h.serveIndex)
	mux.Handle("/demo/static/", h.staticHandler)
}

func (h *Handler) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/demo" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		StripePublishableKey string
	}{
		StripePublishableKey: h.stripePublishableKey,
	}

	if err := h.indexTemplate.Execute(w, data); err != nil {
		http.Error(w, "failed to render demo page", http.StatusInternalServerError)
		return
	}
}
