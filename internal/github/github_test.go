package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPhoneVariants_DedupesAndOrders(t *testing.T) {
	got, err := phoneVariants("+14155552671")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 3 {
		t.Errorf("expected at least 3 variants, got %v", got)
	}
	seen := map[string]bool{}
	for _, v := range got {
		if seen[v] {
			t.Errorf("duplicate variant %q", v)
		}
		seen[v] = true
	}
	// Sanity-check that the canonical E.164 form is present.
	if !seen["+14155552671"] {
		t.Errorf("expected +14155552671 in variants, got %v", got)
	}
}

func TestSearch_DedupesBySHAAcrossVariants(t *testing.T) {
	hitsServed := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Same SHA returned for every variant query — we should only see it
		// once in the deduped output.
		hitsServed++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"total_count": 1,
			"items": [{
				"sha": "deadbeef0000000000000000000000000000beef",
				"html_url": "https://github.com/example/repo/commit/deadbeef",
				"commit": {
					"author": {"name": "Test User", "email": "test@example.com", "date": "2024-01-01T00:00:00Z"},
					"message": "fix call to +1-415-555-2671\n\nMore detail."
				},
				"author": {"login": "testuser"},
				"repository": {"full_name": "example/repo"}
			}]
		}`))
	}))
	t.Cleanup(srv.Close)

	resp, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if hitsServed < 2 {
		t.Errorf("expected ≥2 variant queries to fan out, got %d", hitsServed)
	}
	if resp.Returned != 1 {
		t.Errorf("expected 1 dedup'd hit, got %d (Hits=%+v)", resp.Returned, resp.Hits)
	}
	if resp.Hits[0].MessageHead != "fix call to +1-415-555-2671" {
		t.Errorf("MessageHead = %q, want first line of message", resp.Hits[0].MessageHead)
	}
	if resp.Hits[0].URL != "https://github.com/example/repo/commit/deadbeef" {
		t.Errorf("URL = %q, want github.com URL", resp.Hits[0].URL)
	}
}

func TestSearch_RateLimitedSurfacedAsTypedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintln(w, `{"message":"rate limit exceeded"}`)
	}))
	t.Cleanup(srv.Close)

	_, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL, Limit: 5})
	if !errors.Is(err, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestSearch_TokenSentWhenEnvSet(t *testing.T) {
	t.Setenv(envTokenVar, "ghp_testtoken123")
	saw := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count": 0, "items": []}`))
	}))
	t.Cleanup(srv.Close)

	_, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(saw, "Bearer ghp_testtoken123") {
		t.Errorf("Authorization header = %q, want Bearer ghp_testtoken123", saw)
	}
}

func TestSearch_NoTokenSendsNoAuthHeader(t *testing.T) {
	t.Setenv(envTokenVar, "")
	saw := "<unset>"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count": 0, "items": []}`))
	}))
	t.Cleanup(srv.Close)

	_, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if saw != "" {
		t.Errorf("Authorization header = %q, want empty", saw)
	}
}

func TestSearch_EmptyResultIsZeroReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(ghSearchResp{TotalCount: 0, Items: nil})
	}))
	t.Cleanup(srv.Close)

	resp, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Returned != 0 {
		t.Errorf("expected 0 hits, got %d", resp.Returned)
	}
}

func TestFirstLine_TruncatesLong(t *testing.T) {
	s := strings.Repeat("a", 200)
	out := firstLine(s)
	if len(out) > 120 {
		t.Errorf("firstLine produced %d chars, want ≤120", len(out))
	}
	if !strings.HasSuffix(out, "…") {
		t.Errorf("firstLine should end with ellipsis when truncated, got %q", out[len(out)-3:])
	}
}
