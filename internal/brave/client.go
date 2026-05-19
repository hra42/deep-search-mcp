package brave

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultEndpoint = "https://api.search.brave.com/res/v1/web/search"

type Result struct {
	Title       string
	URL         string
	Description string
}

type Client struct {
	APIKey     string
	HTTPClient *http.Client
	Endpoint   string
}

func New(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		Endpoint:   defaultEndpoint,
	}
}

type rawResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func (c *Client) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if c.APIKey == "" {
		return nil, fmt.Errorf("brave: missing API key")
	}
	if count <= 0 {
		count = 5
	}
	if count > 20 {
		count = 20
	}

	q := url.Values{}
	q.Set("q", query)
	q.Set("count", strconv.Itoa(count))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Endpoint+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("brave: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("brave: status %d: %s", resp.StatusCode, string(body))
	}

	var raw rawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("brave: decode: %w", err)
	}

	out := make([]Result, 0, len(raw.Web.Results))
	for _, r := range raw.Web.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Description: r.Description})
	}
	return out, nil
}
