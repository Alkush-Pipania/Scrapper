package browserless

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	endpoint string
	token    string
	client   *http.Client
}

type Result struct {
	Title       string
	ContentText string
	// HTML        string
	Screenshot []byte
}

func New(endpoint string, token string) *Client {
	return &Client{
		endpoint: endpoint,
		token:    token,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (b *Client) Scrape(ctx context.Context, targetURL string) (*Result, error) {
	query := `
	mutation {
	  goto(url: "%s") {
		title
		text
		html
		screenshot(fullPage: false, type: jpeg, quality: 75) {
		  base64
		}
	  }
	}`

	payload := map[string]string{
		"query": fmt.Sprintf(query, targetURL),
	}
	jsonPayload, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/graphql?token=%s", b.endpoint, b.token)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("browserless connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("browserless error code: %d", resp.StatusCode)
	}

	// Parse Response (Internal struct just for unmarshalling)
	var qlResp struct {
		Data struct {
			Goto struct {
				Title      string `json:"title"`
				Text       string `json:"text"`
				HTML       string `json:"html"`
				Screenshot struct {
					Base64 string `json:"base64"`
				} `json:"screenshot"`
			} `json:"goto"`
		} `json:"data"`
		Errors []interface{} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&qlResp); err != nil {
		return nil, err
	}

	if len(qlResp.Errors) > 0 {
		return nil, fmt.Errorf("browserql execution error: %v", qlResp.Errors)
	}

	// Decode screenshot
	imgBytes, _ := base64.StdEncoding.DecodeString(qlResp.Data.Goto.Screenshot.Base64)

	return &Result{
		Title:       qlResp.Data.Goto.Title,
		ContentText: qlResp.Data.Goto.Text,
		// HTML:        qlResp.Data.Goto.HTML,
		Screenshot: imgBytes,
	}, nil
}
