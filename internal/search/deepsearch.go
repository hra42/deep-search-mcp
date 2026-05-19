package search

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/publicsuffix"

	"github.com/hra42/deep-search-mcp/internal/brave"
	"github.com/hra42/deep-search-mcp/internal/extract"
	"github.com/hra42/deep-search-mcp/internal/fetch"
)

type Options struct {
	Query           string
	Count           int
	MaxPagesToFetch int
	MaxCharsPerPage int
	// MaxPerDomain caps how many results may share the same host. 0 = no cap.
	MaxPerDomain int
}

type Service struct {
	Brave   *brave.Client
	Fetcher *fetch.Fetcher
}

type pageResult struct {
	title   string
	url     string
	excerpt string
	note    string
}

func (s *Service) DeepSearch(ctx context.Context, opts Options) (string, error) {
	if strings.TrimSpace(opts.Query) == "" {
		return "", fmt.Errorf("query is required")
	}
	if opts.Count <= 0 {
		opts.Count = 5
	}
	if opts.MaxPagesToFetch <= 0 {
		opts.MaxPagesToFetch = opts.Count
	}
	if opts.MaxPagesToFetch > opts.Count {
		opts.MaxPagesToFetch = opts.Count
	}
	if opts.MaxCharsPerPage <= 0 {
		opts.MaxCharsPerPage = 4000
	}

	results, err := s.Brave.Search(ctx, opts.Query, opts.Count)
	if err != nil {
		return "", fmt.Errorf("brave search: %w", err)
	}

	if opts.MaxPerDomain > 0 {
		results = capPerDomain(results, opts.MaxPerDomain)
	}

	n := opts.MaxPagesToFetch
	if n > len(results) {
		n = len(results)
	}

	pages := make([]pageResult, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r := results[i]
			pages[i] = pageResult{title: r.Title, url: r.URL}

			fctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			resp, ferr := s.Fetcher.Get(fctx, r.URL)
			if ferr != nil {
				pages[i].note = fmt.Sprintf("fetch failed: %v", ferr)
				pages[i].excerpt = r.Description
				return
			}
			if resp.Status/100 != 2 {
				pages[i].note = fmt.Sprintf("HTTP %d", resp.Status)
				pages[i].excerpt = r.Description
				return
			}
			if !resp.IsHTML() {
				pages[i].note = fmt.Sprintf("non-HTML content-type: %s", resp.ContentType)
				pages[i].excerpt = r.Description
				return
			}

			title, md, eerr := extract.HTMLToMarkdown(resp.Body, r.URL)
			if eerr != nil {
				pages[i].note = fmt.Sprintf("extract failed: %v", eerr)
				pages[i].excerpt = r.Description
				return
			}
			if title != "" {
				pages[i].title = title
			}
			pages[i].excerpt = extract.TruncateRunes(strings.TrimSpace(md), opts.MaxCharsPerPage)
		}(i)
	}
	wg.Wait()

	return assemble(opts.Query, pages), nil
}

func capPerDomain(results []brave.Result, max int) []brave.Result {
	if max <= 0 {
		return results
	}
	counts := make(map[string]int)
	out := make([]brave.Result, 0, len(results))
	for _, r := range results {
		key := domainKey(r.URL)
		if key == "" {
			out = append(out, r)
			continue
		}
		if counts[key] >= max {
			continue
		}
		counts[key]++
		out = append(out, r)
	}
	return out
}

// domainKey returns the registrable domain (eTLD+1) for a URL, lowercased.
// Falls back to the bare host if publicsuffix can't resolve.
func domainKey(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	etld1, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return host
	}
	return etld1
}

func assemble(query string, pages []pageResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Deep search: %s\n\n", query)
	for i, p := range pages {
		title := p.title
		if title == "" {
			title = p.url
		}
		fmt.Fprintf(&b, "## %d. %s\n", i+1, title)
		fmt.Fprintf(&b, "Source: %s\n\n", p.url)
		if p.note != "" {
			fmt.Fprintf(&b, "_Note: %s_\n\n", p.note)
		}
		if strings.TrimSpace(p.excerpt) != "" {
			b.WriteString(p.excerpt)
			b.WriteString("\n\n")
		}
	}
	return b.String()
}
