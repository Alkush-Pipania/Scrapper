package scrape

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Alkush-Pipania/Scrapper/pkg/mq"
	"github.com/Alkush-Pipania/Scrapper/pkg/redis"
	"github.com/google/uuid"
)

type service struct {
	rds  *redis.RedisStore
	mqch *mq.Publisher
}

func NewService(rds *redis.RedisStore, mqch *mq.Publisher) *service {
	return &service{
		rds:  rds,
		mqch: mqch,
	}
}

func (s *service) SubmitJob(ctx context.Context, req SubmitScrapeRequest) (string, error) {
	jobID := uuid.NewString()
	if err := s.rds.CreateJob(jobID, req.URL); err != nil {
		return "", err
	}
	msgBody, _ := json.Marshal(&MsgBody{
		URL: req.URL,
		ID:  jobID,
	})

	err := s.mqch.Publish(ctx, msgBody)
	if err != nil {
		log.Printf("failed to publish job , id : %v and error : %v", jobID, err)
		return "", err
	}
	return jobID, nil
}

func (s *service) GetJobStatus(ctx context.Context, jobID string) (*ScrapeStatusResponse, error) {
	job, err := s.rds.GetJob(jobID)
	if err != nil {
		return nil, err
	}
	return &ScrapeStatusResponse{
		ID:     job.ID,
		URL:    job.URL,
		Status: string(job.Status),
		Result: job.Result,
		Error:  job.Error,
	}, nil
}
