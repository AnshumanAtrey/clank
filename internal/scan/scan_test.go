package scan

import (
	"reflect"
	"testing"

	"github.com/AnshumanAtrey/clank/internal/api"
	"github.com/AnshumanAtrey/clank/internal/deep"
	"github.com/AnshumanAtrey/clank/internal/edgar"
	"github.com/AnshumanAtrey/clank/internal/ignorant"
	"github.com/AnshumanAtrey/clank/internal/local"
	"github.com/AnshumanAtrey/clank/internal/telegram"
	"github.com/AnshumanAtrey/clank/internal/whatsapp"
)

func TestSampleHead(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e"}
	got := sampleCandidates(in, 3, SampleHead)
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("head: got %v, want %v", got, want)
	}
}

func TestSampleStride(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	got := sampleCandidates(in, 3, SampleStride)
	if len(got) > 3 {
		t.Errorf("stride: expected ≤3 items, got %d", len(got))
	}
}

func TestSampleRandom_NoOOB(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e"}
	got := sampleCandidates(in, 3, SampleRandom)
	if len(got) != 3 {
		t.Errorf("random: got %d items, want 3", len(got))
	}
	seen := map[string]bool{}
	for _, x := range got {
		if seen[x] {
			t.Errorf("random: duplicate %q", x)
		}
		seen[x] = true
	}
}

func TestSampleAllUnderMax(t *testing.T) {
	in := []string{"a", "b"}
	for _, s := range []Sample{SampleHead, SampleStride, SampleRandom} {
		got := sampleCandidates(in, 5, s)
		if !reflect.DeepEqual(got, in) {
			t.Errorf("%s with max>len: got %v, want all %v", s, got, in)
		}
	}
}

func TestScoreResult_Empty(t *testing.T) {
	if scoreResult(nil) != 0 {
		t.Error("nil should score 0")
	}
	if scoreResult(&deep.Result{}) != 0 {
		t.Error("empty result should score 0")
	}
}

func TestScoreResult_Stack(t *testing.T) {
	d := &deep.Result{
		Local: &local.Lookup{Valid: true},
		APIs: []*deep.APIBlock{
			{Result: &api.Result{Valid: true}},
			{Result: &api.Result{Valid: false}},
		},
		Telegram: &deep.TelegramBlock{Result: &telegram.Lookup{Found: true, Username: "u"}},
		WhatsApp: &deep.WhatsAppBlock{Result: &whatsapp.Result{Registered: true, About: "hi"}},
		Ignorant: &deep.IgnorantBlock{Results: []ignorant.Result{
			{Name: "instagram", Exists: true},
			{Name: "amazon", Exists: false},
		}},
		Edgar: &deep.EdgarBlock{Result: &edgar.Response{Total: 3}},
	}
	got := scoreResult(d)
	// 1 (local) + 1 (api) + 2+1 (tg+username) + 2+1 (wa+about) + 1 (ig) + 2 (edgar) = 11
	if got != 11 {
		t.Errorf("composite score: got %d, want 11", got)
	}
}

func TestScoreResult_SpamPenalty(t *testing.T) {
	d := &deep.Result{
		Local: &local.Lookup{Valid: true, Spam: true},
	}
	if got := scoreResult(d); got != -1 {
		t.Errorf("spam penalty: got %d, want -1", got)
	}
}

func TestBuildCandidates_Phones(t *testing.T) {
	got, err := buildCandidates(Options{Phones: []string{"+14155552671", "+14155552672", "+14155552671"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("dedup failed: got %v", got)
	}
}

func TestBuildCandidates_PatternFromX(t *testing.T) {
	got, err := buildCandidates(Options{Pattern: "1415555267x"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 10 {
		t.Errorf("expected 10 combos for one x, got %d", len(got))
	}
}

func TestBuildCandidates_Empty(t *testing.T) {
	if _, err := buildCandidates(Options{}); err == nil {
		t.Error("expected error for empty input")
	}
}
