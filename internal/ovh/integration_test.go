// Skipped by default — enabled with `go test -tags integration`. Hits the
// live OVH API; meant for diagnosing parse/match issues that only surface
// against real upstream data.

//go:build integration

package ovh

import (
	"context"
	"testing"
)

func TestLookup_LiveAbbevilleSampleNumber(t *testing.T) {
	resetCache()
	resp, err := Lookup(context.Background(), "+33365171234", Options{})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("response: %+v", resp)
	if resp.Zone != nil {
		t.Logf("zone: %+v", *resp.Zone)
	}
	if !resp.Found {
		t.Errorf("expected match against +33365171234 (Abbeville sample row), got Found=false")
	}
	if resp.City != "Abbeville" {
		t.Errorf("City = %q, want Abbeville", resp.City)
	}
}
