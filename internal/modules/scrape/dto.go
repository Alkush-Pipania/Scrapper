package scrape

import "errors"

type SubmitScrapeRequest struct {
	URL string `json:"url" validate:"required,url"`
}

type SubmitScrapeResponse struct {
	JobID string `json:"job_id"`
}

var ErrJobNotFound = errors.New("job not found")

type ScrapeStatusResponse struct {
	ID     string      `json:"id"`
	URL    string      `json:"url"`
	Status string      `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}
type MsgBody struct {
	id  string
	url string
}
