package engine

import (
	"context"
	"fmt"
	"strings"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
)

func (s *Scraper) scrapeYouTube(ctx context.Context, targetURL string) (*domain.ScrapedData, error) {
	video, err := s.youtube.GetVideoData(ctx, targetURL)
	if err != nil {
		return nil, fmt.Errorf("youtube scrape failed: %w", err)
	}

	// Combine all YouTube data into content_text
	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("Title: %s\n\n", video.Title))
	contentBuilder.WriteString(fmt.Sprintf("Channel: %s\n", video.ChannelTitle))
	contentBuilder.WriteString(fmt.Sprintf("Published: %s\n", video.PublishedAt))
	if video.Duration != "" {
		contentBuilder.WriteString(fmt.Sprintf("Duration: %s\n", video.Duration))
	}
	if video.ViewCount != "" {
		contentBuilder.WriteString(fmt.Sprintf("Views: %s\n", video.ViewCount))
	}
	if video.LikeCount != "" {
		contentBuilder.WriteString(fmt.Sprintf("Likes: %s\n", video.LikeCount))
	}
	contentBuilder.WriteString(fmt.Sprintf("\nDescription:\n%s\n", video.Description))
	if video.Transcript != "" {
		contentBuilder.WriteString(fmt.Sprintf("\nTranscript:\n%s", video.Transcript))
	}

	return &domain.ScrapedData{
		URL:         targetURL,
		Title:       video.Title,
		Description: video.Description,
		ImageURL:    video.ThumbnailURL,
		SiteName:    "YouTube",
		ContentText: contentBuilder.String(),
		Author:      video.ChannelTitle,
		PublishedAt: video.PublishedAt,
	}, nil
}
