package brave

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchParsesResults(t *testing.T) {
	body := `{"web":{"results":[
		{"title":"A","url":"https://a.example","description":"alpha"},
		{"title":"B","url":"https://b.example","description":"beta"}
	]}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Subscription-Token"); got != "k" {
			t.Errorf("missing api key header: %q", got)
		}
		if r.URL.Query().Get("q") != "hello" {
			t.Errorf("unexpected query: %q", r.URL.Query().Get("q"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := New("k")
	c.Endpoint = srv.URL

	res, err := c.Search(context.Background(), "hello", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 || res[0].URL != "https://a.example" || res[1].Title != "B" {
		t.Fatalf("unexpected results: %+v", res)
	}
}

func TestSearchErrorOnNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	c := New("k")
	c.Endpoint = srv.URL
	if _, err := c.Search(context.Background(), "hi", 3); err == nil {
		t.Fatal("expected error on 403")
	}
}
