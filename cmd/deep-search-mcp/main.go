package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/hra42/deep-search-mcp/internal/brave"
	"github.com/hra42/deep-search-mcp/internal/fetch"
	"github.com/hra42/deep-search-mcp/internal/search"
	"github.com/hra42/deep-search-mcp/internal/server"
)

const version = "0.1.0"

func main() {
	var (
		transport    = flag.String("transport", "stdio", "Transport: stdio or http")
		listen       = flag.String("listen", ":8080", "Bind address for http transport")
		userAgent    = flag.String("user-agent", "", "Custom User-Agent for page fetches")
		proxyURL     = flag.String("proxy-url", "", "HTTP/HTTPS proxy URL")
		ignoreRobots = flag.Bool("ignore-robots-txt", false, "Ignore robots.txt for page fetches")
	)
	flag.Parse()

	apiKey := os.Getenv("BRAVE_API_KEY")
	if apiKey == "" {
		log.Fatal("BRAVE_API_KEY environment variable is required")
	}

	fetcher, err := fetch.New(*userAgent, *proxyURL, *ignoreRobots)
	if err != nil {
		log.Fatalf("init fetcher: %v", err)
	}
	svc := &search.Service{
		Brave:   brave.New(apiKey),
		Fetcher: fetcher,
	}
	mcpServer := server.New("deep-search-mcp", version, svc)

	switch strings.ToLower(*transport) {
	case "stdio":
		if err := mcpServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("stdio server: %v", err)
		}
	case "http":
		token := os.Getenv("DEEP_SEARCH_MCP_TOKEN")
		if token == "" {
			log.Fatal("DEEP_SEARCH_MCP_TOKEN is required for http transport")
		}
		mcpHandler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return mcpServer },
			nil,
		)
		mux := http.NewServeMux()
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		mux.Handle("/", bearerAuth(token, mcpHandler))
		log.Printf("deep-search-mcp listening on %s (http)", *listen)
		if err := http.ListenAndServe(*listen, mux); err != nil {
			log.Fatalf("http server: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown transport: %s\n", *transport)
		os.Exit(2)
	}
}

func bearerAuth(token string, next http.Handler) http.Handler {
	expected := "Bearer " + token
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != expected {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
