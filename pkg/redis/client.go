package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewRedis(addr string) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})
	return &RedisStore{
		redisClient: rdb,
		ttl:         24 * time.Hour,
	}
}

func (r *RedisStore) key(id string) string {
	return "job:" + id
}

func (r *RedisStore) save(job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}
	return r.redisClient.Set(context.Background(), r.key(job.ID), data, r.ttl).Err()
}

func (r *RedisStore) CreateJob(id, url string) error {
	job := &Job{
		ID:     id,
		URL:    url,
		Status: StatusPending,
	}
	return r.save(job)
}

func (r *RedisStore) UpdateStatus(id string, status JobStatus) error {
	job, err := r.GetJob(id)
	if err != nil {
		return err
	}
	job.Status = status
	return r.save(job)
}

func (r *RedisStore) GetJob(id string) (*Job, error) {
	ctx := context.Background()

	val, err := r.redisClient.Get(ctx, r.key(id)).Result()
	if err != nil {
		return nil, ErrJobNotFound
	}

	var job Job
	if err := json.Unmarshal([]byte(val), &job); err != nil {
		return nil, err
	}
	return &job, nil
}
