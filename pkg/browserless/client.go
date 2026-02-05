package browserless

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	mutation Scrape {
	  goto(url: "%s") {
		status
		time
	  }
	  pageText: text {
		text
	  }
	  pageTitle: text(selector: "title") {
		text
	  }
	  shot: screenshot(fullPage: false, type: jpeg, quality: 75) {
		base64
	  }
	}`

	payload := map[string]string{
		"query": fmt.Sprintf(query, targetURL),
	}
	jsonPayload, _ := json.Marshal(payload)

	base := strings.TrimRight(b.endpoint, "/")
	url := fmt.Sprintf("%s/chromium/bql?token=%s", base, b.token)
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
				Status int     `json:"status"`
				Time   float64 `json:"time"`
			} `json:"goto"`
			PageText struct {
				Text string `json:"text"`
			} `json:"pageText"`
			PageTitle struct {
				Text string `json:"text"`
			} `json:"pageTitle"`
			Shot struct {
				Base64 string `json:"base64"`
			} `json:"shot"`
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
	imgBytes, _ := base64.StdEncoding.DecodeString(qlResp.Data.Shot.Base64)

	return &Result{
		Title:       strings.TrimSpace(qlResp.Data.PageTitle.Text),
		ContentText: qlResp.Data.PageText.Text,
		Screenshot:  imgBytes,
	}, nil
}
