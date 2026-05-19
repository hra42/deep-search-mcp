package search

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hra42/deep-search-mcp/internal/brave"
	"github.com/hra42/deep-search-mcp/internal/fetch"
)

func TestDeepSearchAssemblesMarkdown(t *testing.T) {
	page := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><title>Hello Page</title></head><body><article><h1>Hello</h1><p>This is the body of an article about widgets. ` + strings.Repeat("word ", 50) + `</p></article></body></html>`))
	}))
	defer page.Close()

	braveSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"web":{"results":[{"title":"T1","url":"` + page.URL + `","description":"snippet"}]}}`))
	}))
	defer braveSrv.Close()

	bc := brave.New("k")
	bc.Endpoint = braveSrv.URL

	f, err := fetch.New("test-ua", "", true)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{Brave: bc, Fetcher: f}

	md, err := svc.DeepSearch(context.Background(), Options{Query: "widgets", Count: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "# Deep search: widgets") {
		t.Fatalf("missing header: %s", md)
	}
	if !strings.Contains(md, "Source: "+page.URL) {
		t.Fatalf("missing source line: %s", md)
	}
	if !strings.Contains(md, "## 1.") {
		t.Fatalf("missing numbered section: %s", md)
	}
}

func TestDeepSearchFailedFetchStillCited(t *testing.T) {
	braveSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"web":{"results":[{"title":"Dead","url":"http://127.0.0.1:1/missing","description":"fallback snippet"}]}}`))
	}))
	defer braveSrv.Close()

	bc := brave.New("k")
	bc.Endpoint = braveSrv.URL

	f, err := fetch.New("", "", true)
	if err != nil {
		t.Fatal(err)
	}
	svc := &Service{Brave: bc, Fetcher: f}

	md, err := svc.DeepSearch(context.Background(), Options{Query: "x", Count: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(md, "## 1.") || !strings.Contains(md, "Source: http://127.0.0.1:1/missing") {
		t.Fatalf("expected citation despite fetch failure: %s", md)
	}
	if !strings.Contains(md, "fallback snippet") {
		t.Fatalf("expected fallback to Brave snippet: %s", md)
	}
}
