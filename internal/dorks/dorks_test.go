package dorks

import (
	"strings"
	"testing"
)

func TestGenerate_AllBuckets(t *testing.T) {
	dorks, err := Generate("+14155552671", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(dorks) < 30 {
		t.Errorf("expected ≥30 dorks for full bucket set, got %d", len(dorks))
	}
	seen := map[string]bool{}
	for _, d := range dorks {
		seen[d.Bucket] = true
	}
	for _, b := range []string{"social", "disposable", "reputation", "individuals", "general"} {
		if !seen[b] {
			t.Errorf("missing bucket %q in output", b)
		}
	}
}

func TestGenerate_BucketFilter(t *testing.T) {
	dorks, err := Generate("+14155552671", "", []Bucket{Social})
	if err != nil {
		t.Fatal(err)
	}
	if len(dorks) == 0 {
		t.Fatal("expected social dorks, got none")
	}
	for _, d := range dorks {
		if d.Bucket != "social" {
			t.Errorf("expected social bucket, got %q", d.Bucket)
		}
	}
}

func TestGenerate_AutoPlus(t *testing.T) {
	dorks, err := Generate("14155552671", "", []Bucket{General})
	if err != nil {
		t.Fatal(err)
	}
	if len(dorks) == 0 {
		t.Fatal("auto-plus parsing failed")
	}
}

func TestGenerate_RegionHint(t *testing.T) {
	dorks, err := Generate("9181156055", "IN", []Bucket{General})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range dorks {
		if strings.Contains(d.Query, "+91") || strings.Contains(d.Query, "919181156055") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected E.164 +91... variant in queries")
	}
}

func TestGenerate_BadInput(t *testing.T) {
	if _, err := Generate("not-a-number", "", nil); err == nil {
		t.Error("expected error for garbage input")
	}
}

func TestParseBuckets(t *testing.T) {
	if got := ParseBuckets(""); got != nil {
		t.Errorf("empty input → %v, want nil", got)
	}
	if got := ParseBuckets("social,reputation"); len(got) != 2 {
		t.Errorf("expected 2 buckets, got %v", got)
	}
	if got := ParseBuckets("social,SOCIAL,reputation"); len(got) != 2 {
		t.Errorf("dedup case-insensitive failed: %v", got)
	}
	if got := ParseBuckets("nonexistent"); len(got) != 0 {
		t.Errorf("unknown bucket should yield empty: %v", got)
	}
}

func TestPhoneVariations_Dedup(t *testing.T) {
	vars, err := phoneVariations("+14155552671", "")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, v := range vars {
		if seen[v] {
			t.Errorf("duplicate variant %q", v)
		}
		seen[v] = true
	}
}
