package extract

import (
	"bytes"
	"net/url"
	"strings"
	"unicode"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	readability "github.com/go-shiori/go-readability"
	"golang.org/x/net/html"
)

func HTMLToMarkdown(htmlBytes []byte, pageURL string) (title string, md string, err error) {
	u, _ := url.Parse(pageURL)
	article, rerr := readability.FromReader(bytes.NewReader(htmlBytes), u)
	if rerr == nil && strings.TrimSpace(article.Content) != "" {
		cleaned, cerr := StripImages(article.Content)
		if cerr != nil {
			cleaned = article.Content
		}
		md, err = htmltomarkdown.ConvertString(cleaned)
		return article.Title, md, err
	}
	cleaned, cerr := StripImages(string(htmlBytes))
	if cerr != nil {
		cleaned = string(htmlBytes)
	}
	md, err = htmltomarkdown.ConvertString(cleaned)
	return "", md, err
}

// StripImages removes <img>, <picture>, <figure>, and <svg> nodes from an
// HTML fragment so they don't pollute the converted markdown.
func StripImages(fragment string) (string, error) {
	doc, err := html.Parse(strings.NewReader(fragment))
	if err != nil {
		return fragment, err
	}
	stripImageNodes(doc)
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fragment, err
	}
	return buf.String(), nil
}

func stripImageNodes(n *html.Node) {
	var toRemove []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			switch c.Data {
			case "img", "picture", "figure", "svg", "source":
				toRemove = append(toRemove, c)
				continue
			}
		}
		stripImageNodes(c)
	}
	for _, c := range toRemove {
		n.RemoveChild(c)
	}
}

func TruncateRunes(s string, max int) string {
	if max <= 0 {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	cut := max
	for cut > 0 && !unicode.IsSpace(runes[cut]) {
		cut--
	}
	if cut == 0 {
		cut = max
	}
	return strings.TrimRight(string(runes[:cut]), " \t\n") + "\n\n…[truncated]"
}
