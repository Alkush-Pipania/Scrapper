package redis

import "errors"

type Job struct {
	ID     string      `json:"id"`
	URL    string      `json:"url"`
	Status JobStatus   `json:"status"`
	Result interface{} `json:"result,omitempty"` // The scraped data will go here
	Error  string      `json:"error,omitempty"`
}

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

var ErrJobNotFound = errors.New("job not found")
