package scrape

import domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"

type SubmitScrapeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type SubmitScrapeResponse struct {
	JobID string `json:"job_id"`
}

var ErrJobNotFound = domain.ErrJobNotFound

type ScrapeStatusResponse struct {
	ID     string      `json:"id"`
	URL    string      `json:"url"`
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}
type MsgBody struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}
