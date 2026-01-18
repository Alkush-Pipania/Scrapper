package scrape

import (
	"context"
	"encoding/json"

	"github.com/Alkush-Pipania/Scrapper/pkg/redis"
	"github.com/Alkush-Pipania/Scrapper/pkg/scraper"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type ScrapeWorker struct {
	store   *redis.RedisStore
	scraper *scraper.Scrapper
}

func NewScrapeWorker(store *redis.RedisStore) *ScrapeWorker {
	return &ScrapeWorker{
		store:   store,
		scraper: scraper.New(),
	}
}

func (w *ScrapeWorker) Handle(ctx context.Context, msg amqp091.Delivery) error {
	var payload MsgBody
	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal job payload")
		return nil
	}

	if payload.ID == "" || payload.URL == "" {
		log.Error().Msg("Job missing ID or URL, skipping")
		return nil
	}

	log.Info().Str("job_id", payload.ID).Str("url", payload.URL).Msg("Starting scrape")

	_ = w.store.UpdateStatus(payload.ID, "processing")

	data, err := w.scraper.Scrape(ctx, payload.URL)
	if err != nil {
		log.Error().Err(err).Str("job_id", payload.ID).Msg("Scrape failed")

		if storeErr := w.store.FailJob(payload.ID, err.Error()); storeErr != nil {
			log.Error().Err(storeErr).Msg("Failed to update job status to failed")
		}
		// Return nil so we Ack the message.
		return nil
	}
	log.Info().Str("job_id", payload.ID).Msg("Scrape successful, saving results")

	return w.store.UpdateResult(payload.ID, data)
}
