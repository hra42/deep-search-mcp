# deep-search-mcp

An MCP server that combines Brave Web Search with in-process page fetching and readability extraction. Exposes a single high-level `deep_search` tool: given a query, it searches Brave, fetches the top results in parallel, extracts their main content as markdown, and returns one aggregated document with citations.

Architecturally a sibling of [go-web-fetch-mcp](https://github.com/hra42/go-web-fetch-mcp) — same Go stack (`mark3labs/mcp-go`, `go-shiori/go-readability`, `JohannesKaufmann/html-to-markdown`) and the same stdio + HTTP transports.

## Install

```sh
go install github.com/hra42/deep-search-mcp/cmd/deep-search-mcp@latest
```

Or build from source: `make build` → `./bin/deep-search-mcp`.

## Configure

- `BRAVE_API_KEY` (required) — your Brave Search API key, sent as `X-Subscription-Token`.
- `DEEP_SEARCH_MCP_TOKEN` — required when using `--transport=http`; clients must send `Authorization: Bearer <token>`.

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--transport` | `stdio` | `stdio` or `http` |
| `--listen` | `:8080` | Bind address for HTTP transport |
| `--user-agent` | (built-in) | Override User-Agent for page fetches |
| `--proxy-url` | — | HTTP/HTTPS proxy URL |
| `--ignore-robots-txt` | `false` | Skip robots.txt checks |

## Tool

### `deep_search`

| Param | Type | Default | Notes |
|-------|------|---------|-------|
| `query` | string | — | Required |
| `count` | number | 5 | Brave results to consider (max 20) |
| `max_pages_to_fetch` | number | = `count` | Top-N to fetch and extract |
| `max_chars_per_page` | number | 4000 | Per-page markdown truncation |

Returns a single markdown document:

```
# Deep search: <query>

## 1. <title>
Source: <url>

<extracted excerpt>

## 2. <title>
...
```

Fetch failures still produce a numbered entry with the URL and a note, so citation indexes remain stable.

## Prompts

Three prompt templates ship alongside the tool. Each takes a single required `query` argument and returns a user-role message instructing the model to call `deep_search` with sensible defaults and synthesize the result.

| Prompt | Purpose |
|--------|---------|
| `news_briefing` | Neutral 5–8 bullet briefing with inline citations. |
| `compare_sources` | Cross-outlet comparison: common ground, framing differences, unique claims. |
| `technical_research` | Developer-focused synthesis with quoted spec language and code excerpts. |

## Container

A multi-stage Dockerfile is included. The runtime image is `distroless/static-debian12:nonroot` — static binary, no shell, runs as a non-root UID.

Build and run:

```sh
docker build -t deep-search-mcp .

docker run --rm -p 8080:8080 \
  -e BRAVE_API_KEY=... \
  -e DEEP_SEARCH_MCP_TOKEN=... \
  deep-search-mcp
```

The container's `ENTRYPOINT` hard-codes `--transport=http --listen=:8080`. Override with `docker run ... deep-search-mcp --transport=http --listen=:9000 --ignore-robots-txt` if needed.

Endpoints:
- `POST /mcp` — MCP Streamable HTTP transport; requires `Authorization: Bearer $DEEP_SEARCH_MCP_TOKEN`.
- `GET /healthz` — unauthenticated liveness probe, returns `200 ok`.

**TLS:** the server speaks plain HTTP. Terminate TLS in front of it (Caddy, Traefik, Cloudflare, your load balancer of choice) — do not expose the container directly to the internet.

## Claude Code config example

```json
{
  "mcpServers": {
    "deep-search": {
      "command": "deep-search-mcp",
      "env": { "BRAVE_API_KEY": "..." }
    }
  }
}
```
