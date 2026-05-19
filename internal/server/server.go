package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpsrv "github.com/mark3labs/mcp-go/server"

	"github.com/hra42/deep-search-mcp/internal/search"
)

func New(name, version string, svc *search.Service) *mcpsrv.MCPServer {
	s := mcpsrv.NewMCPServer(name, version, mcpsrv.WithPromptCapabilities(false))

	tool := mcp.NewTool("deep_search",
		mcp.WithDescription("Search the web via Brave, fetch and extract the top results, and return aggregated markdown with citations."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Natural-language search query."),
		),
		mcp.WithNumber("count",
			mcp.Description("Number of Brave results to consider (default 5, max 20)."),
		),
		mcp.WithNumber("max_pages_to_fetch",
			mcp.Description("How many of the top results to fetch and extract (default = count)."),
		),
		mcp.WithNumber("max_chars_per_page",
			mcp.Description("Maximum characters of extracted markdown to keep per page (default 4000)."),
		),
		mcp.WithNumber("max_per_domain",
			mcp.Description("Cap how many results may share the same host (0 = no cap)."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := req.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		opts := search.Options{Query: query}
		args := req.GetArguments()
		if v, ok := args["count"].(float64); ok {
			opts.Count = int(v)
		}
		if v, ok := args["max_pages_to_fetch"].(float64); ok {
			opts.MaxPagesToFetch = int(v)
		}
		if v, ok := args["max_chars_per_page"].(float64); ok {
			opts.MaxCharsPerPage = int(v)
		}
		if v, ok := args["max_per_domain"].(float64); ok {
			opts.MaxPerDomain = int(v)
		}

		md, derr := svc.DeepSearch(ctx, opts)
		if derr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("deep_search failed: %v", derr)), nil
		}
		return mcp.NewToolResultText(md), nil
	})

	registerPrompts(s)

	return s
}
