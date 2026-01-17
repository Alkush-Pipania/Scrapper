package app

import (
	"net/http"

	"github.com/Alkush-Pipania/Scrapper/internal/modules/scrape"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(container *Container) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Group(func(r chi.Router) {
		r.Route("/api/v1", func(v1Route chi.Router) {
			v1Route.Mount("/scrape", scrape.Routes(container.ScrapeHandler))
		})
	})

	return r
}
