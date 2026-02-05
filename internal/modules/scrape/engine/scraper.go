package engine

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
	"github.com/rs/zerolog/log"
)

type Scraper struct {
	browser  Browser
	uploader Uploader
	client   *http.Client
	youtube  YouTube
}

func New(b Browser, u Uploader, yt YouTube, h *http.Client) *Scraper {
	return &Scraper{
		client:   h,
		browser:  b,
		uploader: u,
		youtube:  yt,
	}
}

func (s *Scraper) Scrape(ctx context.Context, targetURL string) (*domain.ScrapedData, error) {
	// YouTube short-circuit
	if s.youtube != nil && s.youtube.IsYouTubeURL(targetURL) {
		return s.scrapeYouTube(ctx, targetURL)
	}

	// 1) Static scrape first
	staticData, staticErr := s.scrapeStatic(ctx, targetURL)
	var eval staticEval

	if staticErr == nil {
		eval = s.evaluateStatic(staticData)
		if eval.ok {
			// Try to fill missing image with browser screenshot (only if needed)
			if eval.needsScreenshot && s.browser != nil {
				if bData, bErr := s.scrapeViaBrowser(ctx, targetURL); bErr == nil && bData.ImageURL != "" {
					staticData.ImageURL = bData.ImageURL
				} else if bErr != nil {
					log.Warn().Err(bErr).Str("url", targetURL).Msg("Browser screenshot failed")
				}
			}
			return staticData, nil
		}

		log.Warn().Str("url", targetURL).Msgf("Static scrape insufficient: %s", eval.reason)
	} else {
		log.Warn().Err(staticErr).Str("url", targetURL).Msg("Static scrape failed")
	}

	// 2) Fallback to browserless
	if s.browser != nil {
		bData, bErr := s.scrapeViaBrowser(ctx, targetURL)
		if bErr == nil {
			return bData, nil
		}
		log.Warn().Err(bErr).Str("url", targetURL).Msg("Browser scraping failed")
	}

	// 3) If browser also failed, return best available result
	if staticErr != nil {
		return nil, staticErr
	}

	// If we got any usable data, return it even if "insufficient"
	if staticData != nil && (strings.TrimSpace(staticData.Title) != "" ||
		strings.TrimSpace(staticData.ContentText) != "" ||
		strings.TrimSpace(staticData.Description) != "") {
		return staticData, nil
	}

	if eval.reason == "" {
		eval.reason = "no usable content"
	}
	return nil, fmt.Errorf("static scrape insufficient: %s", eval.reason)
}
