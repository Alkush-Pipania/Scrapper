package scraper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Alkush-Pipania/Scrapper/pkg/browserless"
	"github.com/Alkush-Pipania/Scrapper/pkg/s3"
	"github.com/Alkush-Pipania/Scrapper/pkg/youtube"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
	"github.com/rs/zerolog/log"
)

type Scraper struct {
	browser       *browserless.Client
	uploader      s3.Client
	client        *http.Client
	youtubeClient *youtube.Client
}

func New(b *browserless.Client, u s3.Client, yt *youtube.Client, h *http.Client) *Scraper {
	return &Scraper{
		client:        h,
		browser:       b,
		uploader:      u,
		youtubeClient: yt,
	}
}

// func NewWithYouTube(youtubeAPIKey string) *Scraper {
// 	s := New()
// 	if youtubeAPIKey != "" {
// 		s.youtubeClient = youtube.NewClient(youtubeAPIKey)
// 	}
// 	return s
// }

// func createHTTPClient() *http.Client {
// 	return &http.Client{
// 		Timeout: 30 * time.Second,
// 		Transport: &http.Transport{
// 			DialContext: (&net.Dialer{
// 				Timeout: 15 * time.Second,
// 			}).DialContext,
// 			ResponseHeaderTimeout: 20 * time.Second,
// 			IdleConnTimeout:       90 * time.Second,
// 			MaxIdleConns:          100,
// 			MaxIdleConnsPerHost:   10,
// 		},
// 	}
// }

func (s *Scraper) Scrape(ctx context.Context, targetURL string) (*ScrapedData, error) {
	// Check if this is a YouTube URL and we have a YouTube client configured
	if s.youtubeClient != nil && youtube.IsYouTubeURL(targetURL) {
		return s.scrapeYouTube(ctx, targetURL)
	}

	if s.browser != nil {
		data, err := s.scrapeViaBrowser(ctx, targetURL)
		if err == nil {
			return data, nil
		}
		log.Warn().Err(err).Str("url", targetURL).Msg("Browser scraping failed")
	}

	// Standard HTML scraping for non-YouTube URLs
	htmlBytes, err := s.fetchHTML(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	result := &ScrapedData{URL: targetURL}
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

		// Extract image with fallbacks and resolve relative URLs
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

	// Content validation: warn if no meaningful content was extracted
	if result.ContentText == "" && result.Title == "" {
		return result, fmt.Errorf("no meaningful content extracted from %s", targetURL)
	}

	return result, nil
}

func (s *Scraper) scrapeViaBrowser(ctx context.Context, url string) (*ScrapedData, error) {
	// Call our new package
	res, err := s.browser.Scrape(ctx, url)
	if err != nil {
		return nil, err
	}

	data := &ScrapedData{
		URL:         url,
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

func (s *Scraper) scrapeYouTube(ctx context.Context, targetURL string) (*ScrapedData, error) {
	video, err := s.youtubeClient.GetVideoData(ctx, targetURL)
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

	return &ScrapedData{
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

// extractImage tries multiple strategies to find the best image URL
func (s *Scraper) extractImage(doc *goquery.Document, baseURL string) string {
	// Priority 1: OG and Twitter meta tags
	imgURL := s.findMeta(doc, "og:image", "og:image:url", "og:image:secure_url", "twitter:image", "twitter:image:src")
	if imgURL != "" {
		return s.resolveURL(baseURL, imgURL)
	}

	// Priority 2: link rel="image_src" (some older sites)
	if href, exists := doc.Find(`link[rel="image_src"]`).Attr("href"); exists && href != "" {
		return s.resolveURL(baseURL, href)
	}

	// Priority 3: First large image in article/main content
	// Check common content containers for images
	selectors := []string{
		"article img",
		"main img",
		".post-content img",
		".entry-content img",
		".content img",
		"#content img",
	}

	for _, selector := range selectors {
		var foundImg string
		doc.Find(selector).EachWithBreak(func(i int, sel *goquery.Selection) bool {
			src := s.getImageSrc(sel)
			if src != "" && !s.isIconOrLogo(src) {
				foundImg = s.resolveURL(baseURL, src)
				return false // break
			}
			return true
		})
		if foundImg != "" {
			return foundImg
		}
	}

	// Priority 4: Any large image on the page (fallback)
	var fallbackImg string
	doc.Find("img").EachWithBreak(func(i int, sel *goquery.Selection) bool {
		src := s.getImageSrc(sel)
		if src != "" && !s.isIconOrLogo(src) {
			// Skip tiny images (likely icons/logos)
			width, _ := sel.Attr("width")
			height, _ := sel.Attr("height")
			if width != "" && (width == "1" || width == "16" || width == "32") {
				return true
			}
			if height != "" && (height == "1" || height == "16" || height == "32") {
				return true
			}
			fallbackImg = s.resolveURL(baseURL, src)
			return false
		}
		return true
	})

	return fallbackImg
}

// getImageSrc extracts the image source, handling lazy loading
func (s *Scraper) getImageSrc(sel *goquery.Selection) string {
	// Check standard src first
	if src, exists := sel.Attr("src"); exists && src != "" && !strings.HasPrefix(src, "data:") {
		return src
	}

	// Check lazy loading attributes
	lazyAttrs := []string{"data-src", "data-lazy-src", "data-original", "data-lazy", "data-srcset", "srcset"}
	for _, attr := range lazyAttrs {
		if val, exists := sel.Attr(attr); exists && val != "" {
			// For srcset, get the first URL
			if attr == "srcset" || attr == "data-srcset" {
				parts := strings.Split(val, ",")
				if len(parts) > 0 {
					firstSrc := strings.TrimSpace(strings.Split(parts[0], " ")[0])
					if firstSrc != "" {
						return firstSrc
					}
				}
			} else {
				return val
			}
		}
	}

	return ""
}

// isIconOrLogo tries to detect if an image URL is likely an icon or logo
func (s *Scraper) isIconOrLogo(src string) bool {
	lowered := strings.ToLower(src)
	iconPatterns := []string{
		"favicon", "icon", "logo", "sprite", "avatar", "badge",
		"spacer", "pixel", "tracking", "analytics", "1x1", "blank",
	}
	for _, pattern := range iconPatterns {
		if strings.Contains(lowered, pattern) {
			return true
		}
	}
	return false
}

// resolveURL converts relative URLs to absolute URLs
func (s *Scraper) resolveURL(baseURL, relativeURL string) string {
	if relativeURL == "" {
		return ""
	}

	// Already absolute
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		return relativeURL
	}

	// Protocol-relative URL
	if strings.HasPrefix(relativeURL, "//") {
		return "https:" + relativeURL
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return relativeURL
	}

	ref, err := url.Parse(relativeURL)
	if err != nil {
		return relativeURL
	}

	return base.ResolveReference(ref).String()
}

func (s *Scraper) findMeta(doc *goquery.Document, tags ...string) string {
	for _, tag := range tags {
		val := doc.Find(fmt.Sprintf("meta[property='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
		val = doc.Find(fmt.Sprintf("meta[name='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
	}
	// Special case for <title> tag
	for _, tag := range tags {
		if tag == "title" {
			return strings.TrimSpace(doc.Find("title").Text())
		}
	}
	return ""
}

func mustParseURL(u string) *url.URL {
	parsed, _ := url.Parse(u)
	return parsed
}
