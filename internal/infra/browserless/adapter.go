package browserless

import (
	"context"

	engine "github.com/Alkush-Pipania/Scrapper/internal/modules/scrape/engine"
	clientpkg "github.com/Alkush-Pipania/Scrapper/pkg/browserless"
)

type Adapter struct {
	client *clientpkg.Client
}

func NewAdapter(client *clientpkg.Client) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) Scrape(ctx context.Context, targetURL string) (*engine.BrowserResult, error) {
	res, err := a.client.Scrape(ctx, targetURL)
	if err != nil {
		return nil, err
	}
	return &engine.BrowserResult{
		Title:       res.Title,
		ContentText: res.ContentText,
		Screenshot:  res.Screenshot,
	}, nil
}
