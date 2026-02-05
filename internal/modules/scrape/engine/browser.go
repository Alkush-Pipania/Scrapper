package engine

import (
	"context"
	"fmt"
	"time"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
	"github.com/rs/zerolog/log"
)

func (s *Scraper) scrapeViaBrowser(ctx context.Context, targetURL string) (*domain.ScrapedData, error) {
	res, err := s.browser.Scrape(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	data := &domain.ScrapedData{
		URL:         targetURL,
		Title:       res.Title,
		ContentText: res.ContentText,
		SiteName:    "Web",
	}

	// Handle Screenshot Upload via Interface
	if len(res.Screenshot) > 0 {
		fileName := fmt.Sprintf("screenshots/%d.jpg", time.Now().UnixNano())
		imgURL, err := s.uploader.Upload(ctx, fileName, res.Screenshot, "image/jpeg")
		if err == nil {
			data.ImageURL = imgURL
		} else {
			log.Error().Err(err).Msg("Failed to upload screenshot")
		}
	}

	return data, nil
}
