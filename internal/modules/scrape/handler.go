package scrape

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Alkush-Pipania/Scrapper/pkg/turnstile"
	"github.com/go-chi/chi/v5"
)

type Service interface {
	SubmitJob(context.Context, SubmitScrapeRequest) (string, error)
	GetJobStatus(context.Context, string) (*ScrapeStatusResponse, error)
}

type Handler struct {
	service   Service
	turnstile *turnstile.Client
}

func NewHandler(service Service, turnstile *turnstile.Client) *Handler {
	return &Handler{
		service:   service,
		turnstile: turnstile,
	}
}

func (h *Handler) SubmitScrape(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("X-Turnstile-Token")
	if token == "" {
		http.Error(w, "Missing turnstile token", http.StatusUnauthorized)
		return
	}

	if err := h.turnstile.Verify(r.Context(), token, r.RemoteAddr); err != nil {
		http.Error(w, "Invalid turnstile token", http.StatusUnauthorized)
		return
	}

	var req SubmitScrapeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	jobID, err := h.service.SubmitJob(r.Context(), req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(SubmitScrapeResponse{JobID: jobID})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "id")

	resp, err := h.service.GetJobStatus(r.Context(), jobID)
	if err != nil {
		if err == ErrJobNotFound {
			http.Error(w, "Job not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(resp)
}
