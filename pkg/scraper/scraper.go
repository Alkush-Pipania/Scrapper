package scraper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

type Scrapper struct {
	client *http.Client
}

func New() *Scrapper {
	return &Scrapper{
		client: &http.Client{
			Timeout: 15 * time.Second, // Overall request timeout
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 10 * time.Second, // Connection timeout
				}).DialContext,
				IdleConnTimeout: 90 * time.Second,
			},
		},
	}
}

func (s *Scrapper) Scrape(ctx context.Context, targetURL string) (*ScrapedData, error) {
	// Fetch the HTML
	htmlBytes, err := s.fetchHTML(ctx, targetURL)
	if err != nil {
		return nil, err
	}

	type partialResult struct {
		data *ScrapedData
		err  error
	}

	readabilityChan := make(chan partialResult)
	metaChan := make(chan partialResult)

	go func() {
		r := bytes.NewReader(htmlBytes)
		parsed, err := readability.FromReader(r, mustParseURL(targetURL))

		res := &ScrapedData{}
		if err == nil {
			res.ContentText = parsed.TextContent
			res.ContentHTML = parsed.Content
			res.Title = parsed.Title
			res.SiteName = parsed.SiteName
			res.Author = parsed.Byline
		}
		readabilityChan <- partialResult{data: res, err: err}
	}()

	go func() {
		r := bytes.NewReader(htmlBytes)
		doc, err := goquery.NewDocumentFromReader(r)

		res := &ScrapedData{}
		if err == nil {
			res.Title = s.findMeta(doc, "og:title", "twitter:title", "title")
			res.Description = s.findMeta(doc, "og:description", "twitter:description", "description")
			res.ImageURL = s.findMeta(doc, "og:image", "twitter:image")
			res.SiteName = s.findMeta(doc, "og:site_name", "application-name")
		}
		metaChan <- partialResult{data: res, err: err}
	}()

	result := &ScrapedData{URL: targetURL}

	// We wait for 2 results.
	for range 2 {
		select {
		case res := <-readabilityChan:
			if res.err == nil {
				result.ContentText = res.data.ContentText
				result.ContentHTML = res.data.ContentHTML
				result.Author = res.data.Author
				if res.data.Title != "" {
					result.Title = res.data.Title
				}
			}

		case res := <-metaChan:
			if res.err == nil {
				if res.data.ImageURL != "" {
					result.ImageURL = res.data.ImageURL
				}
				if res.data.Description != "" {
					result.Description = res.data.Description
				}
				if result.Title == "" {
					result.Title = res.data.Title
				}
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return result, nil

}

func (s *Scrapper) fetchHTML(ctx context.Context, urlStr string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	// Pretend to be a real browser to avoid 403s
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := s.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to fetch: %d", resp.StatusCode)
	}
	limitedBody := io.LimitReader(resp.Body, 50*1024*1024)
	return io.ReadAll(limitedBody)
}

func (s *Scrapper) findMeta(doc *goquery.Document, tags ...string) string {
	for _, tag := range tags {
		// check property="..."
		val := doc.Find(fmt.Sprintf("meta[property='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
		// check name="..."
		val = doc.Find(fmt.Sprintf("meta[name='%s']", tag)).AttrOr("content", "")
		if val != "" {
			return val
		}
	}
	// Special case for <title> tag
	for _, tag := range tags {
		if tag == "title" {
			return doc.Find("title").Text()
		}
	}
	return ""
}

func mustParseURL(u string) *url.URL {
	parsed, _ := url.Parse(u)
	return parsed
}
