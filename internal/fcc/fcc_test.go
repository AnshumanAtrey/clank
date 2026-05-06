package fcc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToUSHyphenated_E164(t *testing.T) {
	got, err := toUSHyphenated("+14155552671")
	if err != nil {
		t.Fatal(err)
	}
	if got != "415-555-2671" {
		t.Errorf("got %q, want 415-555-2671", got)
	}
}

func TestToUSHyphenated_National(t *testing.T) {
	got, err := toUSHyphenated("4155552671")
	if err != nil {
		t.Fatal(err)
	}
	if got != "415-555-2671" {
		t.Errorf("got %q, want 415-555-2671", got)
	}
}

func TestToUSHyphenated_NonUSReturnsErrNonUS(t *testing.T) {
	if _, err := toUSHyphenated("+919181156055"); !errors.Is(err, ErrNonUS) {
		t.Errorf("expected ErrNonUS for Indian number, got %v", err)
	}
}

func TestToUSHyphenated_GarbageInput(t *testing.T) {
	if _, err := toUSHyphenated("not-a-phone"); err == nil {
		t.Error("expected parse error, got nil")
	}
}

func TestSearch_AggregatesAndOrdersByDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("caller_id_number"); got != "203-760-1637" {
			t.Errorf("Socrata called with caller_id_number=%q, want 203-760-1637", got)
		}
		if got := r.URL.Query().Get("$order"); got != "issue_date DESC" {
			t.Errorf("missing $order param, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]Complaint{
			{ID: "1", IssueDate: "2024-03-15T00:00:00.000", Issue: "Unwanted Calls", State: "CT", CallerID: "203-760-1637"},
			{ID: "2", IssueDate: "2023-08-01T00:00:00.000", Issue: "Unwanted Calls", State: "NY", CallerID: "203-760-1637"},
			{ID: "3", IssueDate: "2024-01-10T00:00:00.000", Issue: "Robocalls", State: "CT", CallerID: "203-760-1637"},
		})
	}))
	t.Cleanup(srv.Close)

	resp, err := Search(context.Background(), "+12037601637", Options{Endpoint: srv.URL, Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Count != 3 {
		t.Errorf("Count = %d, want 3", resp.Count)
	}
	if resp.LatestDate != "2024-03-15T00:00:00.000" {
		t.Errorf("LatestDate = %q, want 2024-03-15...", resp.LatestDate)
	}
	if resp.EarliestDate != "2023-08-01T00:00:00.000" {
		t.Errorf("EarliestDate = %q, want 2023-08-01...", resp.EarliestDate)
	}
	// Sorted, deduped
	if len(resp.Issues) != 2 || resp.Issues[0] != "Robocalls" || resp.Issues[1] != "Unwanted Calls" {
		t.Errorf("Issues = %v, want [Robocalls Unwanted Calls]", resp.Issues)
	}
	if len(resp.States) != 2 || resp.States[0] != "CT" || resp.States[1] != "NY" {
		t.Errorf("States = %v, want [CT NY]", resp.States)
	}
}

func TestSearch_NonUSReturnsErrNonUS(t *testing.T) {
	// Endpoint should never be called for non-US input — fail loud if it is.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("endpoint must not be called for non-US input")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	_, err := Search(context.Background(), "+919181156055", Options{Endpoint: srv.URL})
	if !errors.Is(err, ErrNonUS) {
		t.Errorf("expected ErrNonUS, got %v", err)
	}
}

func TestSearch_HTTPErrorSurfaced(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(w, "upstream timeout")
	}))
	t.Cleanup(srv.Close)

	_, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL})
	if err == nil {
		t.Fatal("expected http error, got nil")
	}
}

func TestSearch_EmptyResultBecomesZeroCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(srv.Close)

	resp, err := Search(context.Background(), "+14155552671", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Count != 0 {
		t.Errorf("Count = %d, want 0", resp.Count)
	}
	if len(resp.Issues) != 0 || len(resp.States) != 0 {
		t.Error("expected empty Issues/States on zero hits")
	}
}
