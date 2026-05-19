package fetch

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type robotsRules struct {
	disallows []string
}

func (f *Fetcher) robotsAllowed(ctx context.Context, target string) (bool, error) {
	u, err := url.Parse(target)
	if err != nil {
		return true, err
	}
	host := u.Scheme + "://" + u.Host

	f.robotsMu.Lock()
	rules, ok := f.robotsCache[host]
	f.robotsMu.Unlock()

	if !ok {
		rules = fetchRobots(ctx, f.HTTPClient, f.UserAgent, host)
		f.robotsMu.Lock()
		f.robotsCache[host] = rules
		f.robotsMu.Unlock()
	}

	path := u.Path
	if path == "" {
		path = "/"
	}
	for _, d := range rules.disallows {
		if d != "" && strings.HasPrefix(path, d) {
			return false, nil
		}
	}
	return true, nil
}

func fetchRobots(ctx context.Context, client *http.Client, ua, host string) *robotsRules {
	r := &robotsRules{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host+"/robots.txt", nil)
	if err != nil {
		return r
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return r
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return r
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return r
	}
	parseRobots(string(body), r)
	return r
}

func parseRobots(body string, r *robotsRules) {
	var inStar bool
	for _, raw := range strings.Split(body, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "user-agent":
			inStar = val == "*"
		case "disallow":
			if inStar && val != "" {
				r.disallows = append(r.disallows, val)
			}
		}
	}
}
