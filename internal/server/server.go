package server

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hra42/deep-search-mcp/internal/search"
)

type DeepSearchInput struct {
	Query           string `json:"query" jsonschema:"natural-language search query"`
	Count           int    `json:"count,omitempty" jsonschema:"number of Brave results to consider (default 5, max 20)"`
	MaxPagesToFetch int    `json:"max_pages_to_fetch,omitempty" jsonschema:"how many of the top results to fetch and extract (default = count)"`
	MaxCharsPerPage int    `json:"max_chars_per_page,omitempty" jsonschema:"maximum characters of extracted markdown to keep per page (default 4000)"`
	MaxPerDomain    int    `json:"max_per_domain,omitempty" jsonschema:"cap how many results may share the same registrable domain (0 = no cap)"`
}

func New(name, version string, svc *search.Service) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)

	mcp.AddTool(s,
		&mcp.Tool{
			Name:        "deep_search",
			Description: "Search the web via Brave, fetch and extract the top results, and return aggregated markdown with citations.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, in DeepSearchInput) (*mcp.CallToolResult, any, error) {
			opts := search.Options{
				Query:           in.Query,
				Count:           in.Count,
				MaxPagesToFetch: in.MaxPagesToFetch,
				MaxCharsPerPage: in.MaxCharsPerPage,
				MaxPerDomain:    in.MaxPerDomain,
			}
			md, err := svc.DeepSearch(ctx, opts)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("deep_search failed: %v", err)}},
				}, nil, nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: md}},
			}, nil, nil
		},
	)

	registerPrompts(s)

	return s
}
