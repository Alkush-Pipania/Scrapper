package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	youtubeAPIBase  = "https://www.googleapis.com/youtube/v3"
	transcriptProxy = "https://www.youtube.com/api/timedtext"
)

var videoIDPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/|youtube\.com\/v\/|youtube\.com\/shorts\/)([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`^([a-zA-Z0-9_-]{11})$`),
}

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 15 * time.Second,
				}).DialContext,
				ResponseHeaderTimeout: 20 * time.Second,
				IdleConnTimeout:       90 * time.Second,
			},
		},
	}
}

func IsYouTubeURL(rawURL string) bool {
	lower := strings.ToLower(rawURL)
	return strings.Contains(lower, "youtube.com") || strings.Contains(lower, "youtu.be")
}

func ExtractVideoID(rawURL string) string {
	for _, pattern := range videoIDPatterns {
		matches := pattern.FindStringSubmatch(rawURL)
		if len(matches) >= 2 {
			return matches[1]
		}
	}
	return ""
}

func (c *Client) GetVideoData(ctx context.Context, rawURL string) (*VideoData, error) {
	videoID := ExtractVideoID(rawURL)
	if videoID == "" {
		return nil, fmt.Errorf("could not extract video ID from URL: %s", rawURL)
	}

	video, err := c.fetchVideoDetails(ctx, videoID)
	if err != nil {
		return nil, err
	}

	transcript, _ := c.fetchTranscript(ctx, videoID)
	video.Transcript = transcript

	return video, nil
}

func (c *Client) fetchVideoDetails(ctx context.Context, videoID string) (*VideoData, error) {
	endpoint := fmt.Sprintf("%s/videos?part=snippet,statistics,contentDetails&id=%s&key=%s",
		youtubeAPIBase, videoID, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result videoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	item := result.Items[0]
	thumbnail := item.Snippet.Thumbnails.MaxRes.URL
	if thumbnail == "" {
		thumbnail = item.Snippet.Thumbnails.High.URL
	}
	if thumbnail == "" {
		thumbnail = item.Snippet.Thumbnails.Medium.URL
	}
	if thumbnail == "" {
		thumbnail = item.Snippet.Thumbnails.Default.URL
	}

	return &VideoData{
		VideoID:      videoID,
		Title:        item.Snippet.Title,
		Description:  item.Snippet.Description,
		ChannelTitle: item.Snippet.ChannelTitle,
		ChannelID:    item.Snippet.ChannelID,
		PublishedAt:  item.Snippet.PublishedAt,
		ThumbnailURL: thumbnail,
		Duration:     item.Details.Duration,
		ViewCount:    item.Stats.ViewCount,
		LikeCount:    item.Stats.LikeCount,
	}, nil
}

func (c *Client) fetchTranscript(ctx context.Context, videoID string) (string, error) {
	langs := []string{"en", "en-US", "en-GB", ""}

	for _, lang := range langs {
		params := url.Values{
			"v":    {videoID},
			"lang": {lang},
			"fmt":  {"srv3"},
		}

		endpoint := fmt.Sprintf("%s?%s", transcriptProxy, params.Encode())

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			continue
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err == nil && len(body) > 0 {
				transcript := parseTranscriptXML(string(body))
				if transcript != "" {
					return transcript, nil
				}
			}
		}
		resp.Body.Close()
	}

	return "", fmt.Errorf("no transcript available for video: %s", videoID)
}

func parseTranscriptXML(xmlData string) string {
	textPattern := regexp.MustCompile(`<text[^>]*>([^<]*)</text>`)
	matches := textPattern.FindAllStringSubmatch(xmlData, -1)

	var parts []string
	for _, match := range matches {
		if len(match) >= 2 {
			text := decodeHTMLEntities(match[1])
			text = strings.TrimSpace(text)
			if text != "" {
				parts = append(parts, text)
			}
		}
	}

	return strings.Join(parts, " ")
}

func decodeHTMLEntities(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&apos;": "'",
		"&#39;":  "'",
		"&nbsp;": " ",
	}
	for entity, char := range replacements {
		s = strings.ReplaceAll(s, entity, char)
	}
	newlinePattern := regexp.MustCompile(`&#xa;|&#xA;|\n`)
	s = newlinePattern.ReplaceAllString(s, " ")
	return s
}
