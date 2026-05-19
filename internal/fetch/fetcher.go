package fetch

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const DefaultUserAgent = "deep-search-mcp/0.1 (+https://github.com/hra42/deep-search-mcp)"

type Fetcher struct {
	HTTPClient       *http.Client
	UserAgent        string
	IgnoreRobotsTxt  bool

	robotsMu    sync.Mutex
	robotsCache map[string]*robotsRules
}

type Response struct {
	URL         string
	Status      int
	ContentType string
	Body        []byte
}

func New(userAgent string, proxyURL string, ignoreRobots bool) (*Fetcher, error) {
	transport := &http.Transport{}
	if proxyURL != "" {
		u, err := url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url: %w", err)
		}
		transport.Proxy = http.ProxyURL(u)
	}
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}
	return &Fetcher{
		HTTPClient: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		UserAgent:       userAgent,
		IgnoreRobotsTxt: ignoreRobots,
		robotsCache:     make(map[string]*robotsRules),
	}, nil
}

func (f *Fetcher) Get(ctx context.Context, target string) (*Response, error) {
	if !f.IgnoreRobotsTxt {
		allowed, err := f.robotsAllowed(ctx, target)
		if err == nil && !allowed {
			return nil, fmt.Errorf("blocked by robots.txt")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", f.UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}
	return &Response{
		URL:         target,
		Status:      resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Body:        body,
	}, nil
}

func (r *Response) IsHTML() bool {
	ct := strings.ToLower(r.ContentType)
	return strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml")
}
