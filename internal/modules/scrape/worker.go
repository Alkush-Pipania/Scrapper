package scrape

import (
	"context"
	"encoding/json"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
	"github.com/Alkush-Pipania/Scrapper/internal/modules/scrape/engine"
	"github.com/Alkush-Pipania/Scrapper/pkg/redis"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type ScrapeWorker struct {
	store   *redis.Client
	scraper *engine.Scraper
}

func NewScrapeWorker(store *redis.Client, scraper *engine.Scraper) *ScrapeWorker {
	return &ScrapeWorker{
		store:   store,
		scraper: scraper,
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

	_ = w.store.UpdateStatus(payload.ID, domain.StatusProcessing)

	data, err := w.scraper.Scrape(ctx, payload.URL)
	if err != nil {
		log.Error().
			Err(err).
			Str("job_id", payload.ID).
			Str("url", payload.URL).
			Msg("Scrape failed")

		if storeErr := w.store.FailJob(payload.ID, err.Error()); storeErr != nil {
			log.Error().Err(storeErr).Msg("Failed to update job status to failed")
		}
		return nil
	}

	// Detailed logging for debugging scrape results
	log.Info().
		Str("job_id", payload.ID).
		Str("title", data.Title).
		Str("image_url", data.ImageURL).
		Int("content_text_len", len(data.ContentText)).
		Bool("has_description", data.Description != "").
		Str("site_name", data.SiteName).
		Msg("Scrape completed successfully")

	return w.store.UpdateResult(payload.ID, data)
}
