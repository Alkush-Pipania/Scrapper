package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	secret string
	client *http.Client
}

func New(secret string) *Client {
	return &Client{
		secret: secret,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

type turnstileResponse struct {
	Success bool     `json:"success"`
	Errors  []string `json:"error-codes"`
}

func (c *Client) Verify(ctx context.Context, token, ip string) error {
	if c.secret == "" {
		return nil
	}

	verifyURL := "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	req, err := http.NewRequestWithContext(ctx, "POST", verifyURL, nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("secret", c.secret)
	q.Add("response", token)
	q.Add("remoteip", ip)
	req.URL.RawQuery = q.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to turnstile: %w", err)
	}
	defer resp.Body.Close()

	var result turnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("validation failed: %v", result.Errors)
	}

	return nil
}
