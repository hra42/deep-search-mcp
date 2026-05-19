package search

import (
	"testing"

	"github.com/hra42/deep-search-mcp/internal/brave"
)

func TestCapPerDomain(t *testing.T) {
	in := []brave.Result{
		{URL: "https://www.bbc.com/news/articles/1"},
		{URL: "https://news.bbc.com/articles/2"},
		{URL: "https://en.wikipedia.org/wiki/X"},
		{URL: "https://bbc.com/news/articles/3"},
		{URL: "https://www.anthropic.com/news/y"},
		{URL: "https://www.bbc.co.uk/news/articles/4"},
	}
	out := capPerDomain(in, 1)

	got := make([]string, len(out))
	for i, r := range out {
		got[i] = r.URL
	}
	want := []string{
		"https://www.bbc.com/news/articles/1",
		"https://en.wikipedia.org/wiki/X",
		"https://www.anthropic.com/news/y",
		"https://www.bbc.co.uk/news/articles/4",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("position %d: got %s, want %s", i, got[i], w)
		}
	}
}

func TestCapPerDomainEtldPlusOne(t *testing.T) {
	// All three are eTLD+1 = bbc.co.uk; with cap=1 only the first survives.
	in := []brave.Result{
		{URL: "https://www.bbc.co.uk/a"},
		{URL: "https://news.bbc.co.uk/b"},
		{URL: "https://bbc.co.uk/c"},
	}
	out := capPerDomain(in, 1)
	if len(out) != 1 || out[0].URL != "https://www.bbc.co.uk/a" {
		t.Fatalf("expected single bbc.co.uk entry, got %+v", out)
	}
}

func TestCapPerDomainZeroNoOp(t *testing.T) {
	in := []brave.Result{{URL: "https://a.example"}, {URL: "https://a.example"}}
	if got := capPerDomain(in, 0); len(got) != 2 {
		t.Fatalf("expected passthrough with cap=0, got %d", len(got))
	}
}
