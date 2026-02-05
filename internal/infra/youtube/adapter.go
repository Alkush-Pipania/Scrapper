package youtube

import (
	"context"

	engine "github.com/Alkush-Pipania/Scrapper/internal/modules/scrape/engine"
	clientpkg "github.com/Alkush-Pipania/Scrapper/pkg/youtube"
)

type Adapter struct {
	client *clientpkg.Client
}

func NewAdapter(client *clientpkg.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) IsYouTubeURL(url string) bool {
	return clientpkg.IsYouTubeURL(url)
}

func (a *Adapter) GetVideoData(ctx context.Context, url string) (*engine.VideoData, error) {
	v, err := a.client.GetVideoData(ctx, url)
	if err != nil {
		return nil, err
	}
	return &engine.VideoData{
		Title:        v.Title,
		Description:  v.Description,
		ChannelTitle: v.ChannelTitle,
		PublishedAt:  v.PublishedAt,
		ThumbnailURL: v.ThumbnailURL,
		Duration:     v.Duration,
		ViewCount:    v.ViewCount,
		LikeCount:    v.LikeCount,
		Transcript:   v.Transcript,
	}, nil
}
