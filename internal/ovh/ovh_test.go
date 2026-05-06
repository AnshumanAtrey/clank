package ovh

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeFRZones mirrors the real OVH detailedZones response shape. Three zones
// with overlapping prefixes so we can prove longest-match-wins.
var fakeFRZones = []Zone{
	{
		City: "Paris", ZipCode: "75001", Country: "fr",
		Prefix: 33, Number: "01xxxxxxxx", InternationalNumber: "003301xxxxxxxx",
		Type: "geographic",
	},
	{
		City: "Abbeville", ZipCode: "80100", Country: "fr",
		Prefix: 33, Number: "036517xxxx", InternationalNumber: "003336517xxxx",
		Type: "geographic",
	},
	{
		City: "Chartres", ZipCode: "28000", Country: "fr",
		Prefix: 33, Number: "023440xxxx", InternationalNumber: "003323440xxxx",
		Type: "geographic",
	},
}

func TestLookup_ExactPrefixMatch(t *testing.T) {
	resetCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("country"); got != "fr" {
			t.Errorf("expected country=fr, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeFRZones)
	}))
	t.Cleanup(srv.Close)

	resp, err := Lookup(context.Background(), "+33365171234", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatalf("expected Found=true, got %+v", resp)
	}
	if resp.City != "Abbeville" {
		t.Errorf("City = %q, want Abbeville", resp.City)
	}
	if resp.Zip != "80100" {
		t.Errorf("Zip = %q, want 80100", resp.Zip)
	}
	if resp.Region != "FR" {
		t.Errorf("Region = %q, want FR", resp.Region)
	}
}

func TestLookup_LongestPrefixWins(t *testing.T) {
	resetCache()
	// OVH's internationalNumber drops the FR national leading 0, so a Paris
	// "01xx-xx-xx-xx" range serializes as "00331xxxxxxxx" (13 chars) and a
	// hypothetical narrower district "0112-xx-xx-xx" as "0033112xxxxxx".
	// The narrower one must win on a number that matches both.
	zones := []Zone{
		{City: "Paris", InternationalNumber: "00331xxxxxxxx", Type: "geographic", Country: "fr"},
		{City: "ParisDistrict5", InternationalNumber: "0033112xxxxxx", Type: "geographic", Country: "fr"},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(zones)
	}))
	t.Cleanup(srv.Close)

	resp, err := Lookup(context.Background(), "+33112345678", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if resp.City != "ParisDistrict5" {
		t.Errorf("longest-match failed: City = %q, want ParisDistrict5 (resp=%+v)", resp.City, resp)
	}
}

func TestLookup_UnsupportedRegion(t *testing.T) {
	resetCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Errorf("endpoint must not be called for unsupported region")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	if _, err := Lookup(context.Background(), "+14155552671", Options{Endpoint: srv.URL}); !errors.Is(err, ErrUnsupportedRegion) {
		t.Errorf("expected ErrUnsupportedRegion for US number, got %v", err)
	}
	if _, err := Lookup(context.Background(), "+919181156055", Options{Endpoint: srv.URL}); !errors.Is(err, ErrUnsupportedRegion) {
		t.Errorf("expected ErrUnsupportedRegion for IN number, got %v", err)
	}
}

func TestLookup_GBMapsToUKQuery(t *testing.T) {
	resetCache()
	saw := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		saw = r.URL.Query().Get("country")
		w.Header().Set("Content-Type", "application/json")
		// Return one matching zone for the UK number we'll query
		_ = json.NewEncoder(w).Encode([]Zone{
			{City: "London", ZipCode: "EC1A", Country: "uk", InternationalNumber: "00442071xxxxxx"},
		})
	}))
	t.Cleanup(srv.Close)

	if _, err := Lookup(context.Background(), "+442071234567", Options{Endpoint: srv.URL}); err != nil {
		t.Fatal(err)
	}
	if saw != "uk" {
		t.Errorf("GB region must query country=uk, got %q", saw)
	}
}

func TestLookup_NoMatch_ReturnsFoundFalse(t *testing.T) {
	resetCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Only one zone, doesn't cover our query prefix
		_ = json.NewEncoder(w).Encode([]Zone{
			{City: "Marseille", InternationalNumber: "0033491xxxxxx"},
		})
	}))
	t.Cleanup(srv.Close)

	resp, err := Lookup(context.Background(), "+33365171234", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Found {
		t.Errorf("expected Found=false when no zone matches, got %+v", resp)
	}
}

func TestLookup_ZonesCachedAcrossCalls(t *testing.T) {
	resetCache()
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(fakeFRZones)
	}))
	t.Cleanup(srv.Close)

	// First call populates cache; second hits cache; total upstream calls = 1.
	for i := 0; i < 2; i++ {
		if _, err := Lookup(context.Background(), "+33365171234", Options{Endpoint: srv.URL}); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Errorf("expected 1 upstream call (cached on second), got %d", calls)
	}
}

// Regression test for a real bug: the Zone struct used snake_case JSON tags
// while the OVH API returns camelCase. Round-tripping through the same struct
// in tests masked the mismatch; live calls returned empty fields and no
// matches. This test serves verbatim camelCase JSON (matching the OVH API
// response) and asserts the source still resolves correctly.
func TestLookup_AcceptsCamelCaseJSONFromUpstream(t *testing.T) {
	resetCache()
	const camelCase = `[{
		"city": "Abbeville",
		"askedCity": null,
		"zipCode": "80100",
		"matchingCriteria": null,
		"zneList": ["80001"],
		"type": "geographic",
		"country": "fr",
		"prefix": 33,
		"number": "036517xxxx",
		"internationalNumber": "003336517xxxx"
	}]`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(camelCase))
	}))
	t.Cleanup(srv.Close)

	resp, err := Lookup(context.Background(), "+33365171234", Options{Endpoint: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Found {
		t.Fatalf("camelCase upstream JSON failed to populate Zone — Found=false (resp=%+v)", resp)
	}
	if resp.City != "Abbeville" || resp.Zip != "80100" {
		t.Errorf("got %+v, want Abbeville / 80100", resp)
	}
}

func TestLookup_HTTPErrorSurfaced(t *testing.T) {
	resetCache()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	_, err := Lookup(context.Background(), "+33365171234", Options{Endpoint: srv.URL})
	if err == nil {
		t.Fatal("expected http error, got nil")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("expected error to mention 502, got %v", err)
	}
}
