package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"

	domain "github.com/Alkush-Pipania/Scrapper/internal/domain/scrape"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

func (s *Scraper) scrapeStatic(ctx context.Context, targetURL string) (*domain.ScrapedData, error) {
	htmlBytes, err := s.fetchHTML(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	result := &domain.ScrapedData{URL: targetURL}
	var mu sync.Mutex
	var wg sync.WaitGroup
	var readabilityErr, metaErr error

	wg.Add(2)

	// Goroutine 1: Parse with go-readability for main content
	go func() {
		defer wg.Done()

		r := bytes.NewReader(htmlBytes)
		parsed, err := readability.FromReader(r, mustParseURL(targetURL))
		if err != nil {
			readabilityErr = err
			return
		}

		mu.Lock()
		defer mu.Unlock()
		result.ContentText = parsed.TextContent
		result.Author = parsed.Byline
		if parsed.Title != "" {
			result.Title = parsed.Title
		}
		if parsed.SiteName != "" {
			result.SiteName = parsed.SiteName
		}
	}()

	// Goroutine 2: Parse with goquery for meta tags and fallback images
	go func() {
		defer wg.Done()

		r := bytes.NewReader(htmlBytes)
		doc, err := goquery.NewDocumentFromReader(r)
		if err != nil {
			metaErr = err
			return
		}

		mu.Lock()
		defer mu.Unlock()

		result.ImageURL = s.extractImage(doc, targetURL)
		result.Description = s.findMeta(doc, "og:description", "twitter:description", "description")

		if result.Title == "" {
			result.Title = s.findMeta(doc, "og:title", "twitter:title", "title")
		}
		if result.SiteName == "" {
			result.SiteName = s.findMeta(doc, "og:site_name", "application-name")
		}
	}()

	// Wait for both goroutines with context cancellation support
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Fail only if both parsers failed
	if readabilityErr != nil && metaErr != nil {
		return nil, fmt.Errorf("both parsers failed: readability=%v, meta=%v", readabilityErr, metaErr)
	}

	if result.ContentText == "" && result.Title == "" {
		return result, fmt.Errorf("no meaningful content extracted from %s", targetURL)
	}

	return result, nil
}

func (s *Scraper) fetchHTML(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Set headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to fetch (status %d): %s", resp.StatusCode, urlStr)
	}
	limitedBody := io.LimitReader(resp.Body, 50*1024*1024)
	return io.ReadAll(limitedBody)
}

func mustParseURL(u string) *url.URL {
	parsed, _ := url.Parse(u)
	return parsed
}
