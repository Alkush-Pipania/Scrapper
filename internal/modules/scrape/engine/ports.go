package engine

import "context"

type BrowserResult struct {
	Title       string
	ContentText string
	Screenshot  []byte
}

type Browser interface {
	Scrape(ctx context.Context, targetURL string) (*BrowserResult, error)
}

type Uploader interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (string, error)
}

type YouTube interface {
	IsYouTubeURL(url string) bool
	GetVideoData(ctx context.Context, url string) (*VideoData, error)
}

type VideoData struct {
	Title        string
	Description  string
	ChannelTitle string
	PublishedAt  string
	ThumbnailURL string
	Duration     string
	ViewCount    string
	LikeCount    string
	Transcript   string
}
