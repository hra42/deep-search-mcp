package extract

import (
	"strings"
	"testing"
)

func TestStripImagesRemovesImageNodes(t *testing.T) {
	in := `<div><p>Hello</p><img src="a.png"/><picture><source srcset="x"/><img src="b.png"/></picture><figure><img src="c.png"/><figcaption>cap</figcaption></figure><svg><circle/></svg><p>world</p></div>`
	out, err := StripImages(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, banned := range []string{"<img", "<picture", "<figure", "<svg", "<source"} {
		if strings.Contains(out, banned) {
			t.Errorf("expected %q to be stripped, got: %s", banned, out)
		}
	}
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "world") {
		t.Errorf("expected text to be preserved, got: %s", out)
	}
}

func TestHTMLToMarkdownDropsImages(t *testing.T) {
	page := `<!doctype html><html><head><title>T</title></head><body><article><h1>Headline</h1><p>Lead paragraph with enough words to satisfy readability heuristics so this article gets selected as the main content of the page.</p><img src="hero.jpg" alt="hero"/><p>Second paragraph also with substantial length to keep the readability score high enough to be considered article content.</p></article></body></html>`
	_, md, err := HTMLToMarkdown([]byte(page), "https://example.com/x")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(md, "hero.jpg") || strings.Contains(md, "![") {
		t.Fatalf("expected no image markdown, got: %s", md)
	}
	if !strings.Contains(md, "Headline") || !strings.Contains(md, "Lead paragraph") {
		t.Fatalf("expected article text preserved, got: %s", md)
	}
}
