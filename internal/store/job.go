package store

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
	Result interface{} `json:"result,omitempty"` // The scraped data will go here
	Error  string      `json:"error,omitempty"`
}

type JobRepository interface {
	CreateJob(id string, url string) error
	UpdateStatus(id string, status JobStatus) error
	UpdateResult(id string, result interface{}) error
	FailJob(id string, errMsg string) error
	GetJob(id string) (*Job, error)
}
