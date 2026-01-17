package scrape

import "github.com/go-chi/chi/v5"

func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Post("/scrape", h.SubmitScrape)
	r.Get("/scrape/{id}", h.GetStatus)
	return r
}
