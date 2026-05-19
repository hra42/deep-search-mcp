# deep-search-mcp

[![CI](https://github.com/hra42/deep-search-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/hra42/deep-search-mcp/actions/workflows/ci.yml)

An MCP server that combines Brave Web Search with in-process page fetching and readability extraction. Exposes a single high-level `deep_search` tool: given a query, it searches Brave, fetches the top results in parallel, extracts their main content as markdown, and returns one aggregated document with citations. Also ships three prompt templates that wrap the tool with sensible defaults.

Built on the official [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk). Content extraction uses [`go-shiori/go-readability`](https://github.com/go-shiori/go-readability) + [`JohannesKaufmann/html-to-markdown`](https://github.com/JohannesKaufmann/html-to-markdown). Sibling project of [go-web-fetch-mcp](https://github.com/hra42/go-web-fetch-mcp).

## Install

```sh
go install github.com/hra42/deep-search-mcp/cmd/deep-search-mcp@latest
```

Or build from source: `make build` → `./bin/deep-search-mcp`.

### Desktop install (.mcpb bundle)

Each tagged release also ships a cross-platform [MCPB bundle](https://github.com/modelcontextprotocol/mcpb) — a `.mcpb` ZIP that contains prebuilt binaries for darwin / linux / windows (amd64 + arm64) and a manifest. Download the latest `deep-search-mcp-<version>.mcpb` from [Releases](https://github.com/hra42/deep-search-mcp/releases) and open it in Claude Desktop; you'll be prompted for your Brave API key on install.

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
| `query` | string | — | Required. Natural-language search query. |
| `count` | number | 5 | Brave results to consider (max 20). |
| `max_pages_to_fetch` | number | = `count` | Top-N results to actually fetch and extract. |
| `max_chars_per_page` | number | 4000 | Per-page markdown truncation (rune-safe, word boundary). |
| `max_per_domain` | number | 0 | Cap results per registrable domain (eTLD+1). `0` = no cap. |

Returns a single markdown document:

```
# Deep search: <query>

## 1. <title>
Source: <url>

<extracted excerpt>

## 2. <title>
...
```

Pages are fetched in parallel and reassembled in Brave's original order. Fetch failures still produce a numbered entry with the URL and a note, so citation indexes remain stable. Images (`<img>`, `<picture>`, `<figure>`, `<svg>`) are stripped after readability so excerpts stay clean text.

## Prompts

Three prompt templates ship alongside the tool. Each takes a required `query` argument and returns a user-role message instructing the model to call `deep_search` with tuned defaults and synthesize the result.

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

The container's `ENTRYPOINT` hard-codes `--transport=http --listen=:8080`. Append extra flags after the image name to override.

Endpoints:
- `POST /mcp` — MCP Streamable HTTP transport; requires `Authorization: Bearer $DEEP_SEARCH_MCP_TOKEN`. Responses are Server-Sent Events.
- `GET /healthz` — unauthenticated liveness probe, returns `200 ok`.

**TLS:** the server speaks plain HTTP. Terminate TLS in front of it (Caddy, Traefik, Cloudflare, your load balancer of choice) — do not expose the container directly to the internet.

### docker compose with Traefik

A `compose.yaml` is included for a Traefik-fronted deployment. Drop a `.env` with `BRAVE_API_KEY` and `DEEP_SEARCH_MCP_TOKEN` next to it, then:

```sh
docker compose up -d --build
```

Adjust the `Host(...)` label and external `traefik` network name to match your setup.

## Client configuration

### Claude Code — stdio (local binary)

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

### Claude Code — remote HTTP

```json
{
  "mcpServers": {
    "deep-search": {
      "url": "https://deep-search-mcp.example.com/mcp",
      "headers": { "Authorization": "Bearer <DEEP_SEARCH_MCP_TOKEN>" }
    }
  }
}
```

### Testing with the MCP Inspector CLI

```sh
# stdio
BRAVE_API_KEY=... npx -y @modelcontextprotocol/inspector --cli ./bin/deep-search-mcp \
  --method tools/call --tool-name deep_search --tool-arg query="..."

# remote HTTP
npx -y @modelcontextprotocol/inspector --cli https://deep-search-mcp.example.com/mcp \
  --transport http --header "Authorization: Bearer $TOKEN" \
  --method tools/call --tool-name deep_search --tool-arg query="..."
```

## Development

```sh
make build   # compile to bin/deep-search-mcp
make test    # go test ./...
make vet     # go vet ./...
```
