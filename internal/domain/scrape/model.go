package scrape

import "errors"

type ScrapedData struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	SiteName    string `json:"site_name"`

	ContentText string `json:"content_text"`
	Author      string `json:"author"`
	PublishedAt string `json:"published_at"`
}

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID     string      `json:"id"`
	URL    string      `json:"url"`
	Status JobStatus   `json:"status"`
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

var ErrJobNotFound = errors.New("job not found")
