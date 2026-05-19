package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type promptSpec struct {
	name        string
	description string
	resultDesc  string
	body        string
}

var promptSpecs = []promptSpec{
	{
		name:        "news_briefing",
		description: "Concise, neutral news briefing on a topic, with citations.",
		resultDesc:  "Instructions to produce a neutral news briefing using deep_search.",
		body: `You are preparing a concise news briefing on: "{{query}}".

1. Call the ` + "`deep_search`" + ` tool with:
   - query: "{{query}} latest news"
   - count: 8
   - max_pages_to_fetch: 5
   - max_per_domain: 1
   - max_chars_per_page: 1500

2. From the returned markdown, write a 5–8 bullet neutral briefing.
   - Cite each fact inline as [n] referring to the numbered sources returned by the tool.
   - Prefer the most recent reporting; flag if sources disagree.
   - Do not add information that is not in the returned sources.
   - End with a "Sources" list mapping [n] -> URL.`,
	},
	{
		name:        "compare_sources",
		description: "Compare how multiple outlets cover the same story.",
		resultDesc:  "Instructions to compare source coverage using deep_search.",
		body: `You will compare how different outlets cover: "{{query}}".

1. Call ` + "`deep_search`" + ` with:
   - query: "{{query}}"
   - count: 12
   - max_pages_to_fetch: 6
   - max_per_domain: 1
   - max_chars_per_page: 2000

2. Produce a comparison with these sections:
   - **Common ground** — facts every source agrees on.
   - **Differences in framing** — short bullets per outlet, quoting briefly.
   - **Unique claims** — anything reported by only one source; flag it.
   - **Open questions** — what none of the sources resolve.

Cite inline as [n] using the numbered sources from the tool output. Do not invent details.`,
	},
	{
		name:        "technical_research",
		description: "Research a technical topic across docs and specs, with code-friendly excerpts.",
		resultDesc:  "Instructions to research a technical topic using deep_search.",
		body: `You are researching the technical topic: "{{query}}".

1. Call ` + "`deep_search`" + ` with:
   - query: "{{query}}"
   - count: 8
   - max_pages_to_fetch: 5
   - max_per_domain: 1
   - max_chars_per_page: 4000

2. Synthesize a developer-oriented answer:
   - Start with a 2–3 sentence summary of what the topic is.
   - Then a "Key points" list of concrete technical details (APIs, flags, semantics).
   - Quote short snippets verbatim when precise wording matters (e.g. spec language).
   - Include at least one minimal code example if the sources contain one; otherwise say so.
   - End with "Sources" listing [n] -> URL for each citation used.

Cite every non-obvious claim inline as [n]. If sources conflict, surface the conflict rather than picking silently.`,
	},
}

func registerPrompts(s *mcp.Server) {
	for _, spec := range promptSpecs {
		spec := spec
		prompt := &mcp.Prompt{
			Name:        spec.name,
			Description: spec.description,
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "query",
					Description: "The topic or question to research.",
					Required:    true,
				},
			},
		}
		s.AddPrompt(prompt, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			query := ""
			if req.Params != nil {
				query = strings.TrimSpace(req.Params.Arguments["query"])
			}
			if query == "" {
				return nil, fmt.Errorf("query argument is required")
			}
			text := strings.ReplaceAll(spec.body, "{{query}}", query)
			return &mcp.GetPromptResult{
				Description: spec.resultDesc,
				Messages: []*mcp.PromptMessage{
					{
						Role:    "user",
						Content: &mcp.TextContent{Text: text},
					},
				},
			}, nil
		})
	}
}
