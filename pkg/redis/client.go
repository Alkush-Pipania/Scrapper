package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
	"github.com/redis/go-redis/v9"
)

var (
	ErrKeyNotFound = redis.Nil
)

type Client struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(redisURL string) (*Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	// Timeouts
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second

	// Pool tuning
	opt.PoolSize = 10
	opt.MinIdleConns = 5

	// Connection lifecycle
	opt.ConnMaxLifetime = 2 * time.Minute
	opt.ConnMaxIdleTime = 30 * time.Second

	rdb := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Client{rdb: rdb, ttl: 24 * time.Hour}, nil //todo: make it change according to time
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

func (r *Client) key(id string) string {
	return "job:" + id
}

func (r *Client) save(job *domain.Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}
	return r.rdb.Set(context.Background(), r.key(job.ID), data, r.ttl).Err()
}

func (r *Client) CreateJob(id, url string) error {
	job := &domain.Job{
		ID:     id,
		URL:    url,
		Status: domain.StatusPending,
	}
	return r.save(job)
}

func (r *Client) UpdateStatus(id string, status domain.JobStatus) error {
	job, err := r.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = status
	return r.save(job)
}

func (r *Client) GetJob(id string) (*domain.Job, error) {
	ctx := context.Background()

	val, err := r.rdb.Get(ctx, r.key(id)).Result()
	if err != nil {
		return nil, domain.ErrJobNotFound
	}

	var job domain.Job
	if err := json.Unmarshal([]byte(val), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *Client) FailJob(id string, errMsg string) error {
	job, err := r.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = domain.StatusFailed
	job.Error = errMsg
	return r.save(job)
}

func (r *Client) UpdateResult(id string, data *domain.ScrapedData) error {
	job, err := r.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = domain.StatusCompleted
	job.Result = data
	return r.save(job)
}
