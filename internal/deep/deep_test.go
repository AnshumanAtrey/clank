package deep

import (
	"context"
	"strings"
	"testing"
)

// offlineOpts skips every block that would touch the network, leaving only the
// pure-offline `local` enrichment + `dorks` generation.
func offlineOpts() Options {
	return Options{
		SkipMessengers: true,
		SkipAPIs:       true,
		SkipEdgar:      true,
	}
}

func TestRun_DorksPopulatedByDefault(t *testing.T) {
	res := Run(context.Background(), "+14155552671", offlineOpts())
	if len(res.Dorks) == 0 {
		t.Fatal("expected Dorks to be populated by default")
	}
	// Every documented bucket must appear at least once so JSON consumers can
	// rely on the schema.
	seen := map[string]bool{}
	for _, d := range res.Dorks {
		seen[d.Bucket] = true
	}
	for _, b := range []string{"social", "disposable", "reputation", "individuals", "general"} {
		if !seen[b] {
			t.Errorf("missing bucket %q in Dorks", b)
		}
	}
}

func TestRun_SkipDorksRespected(t *testing.T) {
	opts := offlineOpts()
	opts.SkipDorks = true
	res := Run(context.Background(), "+14155552671", opts)
	if len(res.Dorks) != 0 {
		t.Errorf("expected 0 Dorks with SkipDorks=true, got %d", len(res.Dorks))
	}
}

func TestRun_DorksToleratesUnparseableInput(t *testing.T) {
	// Garbage input must not panic the orchestrator — Dorks.Generate will
	// return an error and we silently leave Dorks empty.
	res := Run(context.Background(), "definitely-not-a-phone", offlineOpts())
	if len(res.Dorks) != 0 {
		t.Errorf("expected 0 Dorks for unparseable input, got %d", len(res.Dorks))
	}
}

func TestCollectSuggestions_EveryActionIsCallable(t *testing.T) {
	// Synthesize a Result where every optional block is in its skipped state.
	// The synthesized state mirrors what Run() produces on a fresh install
	// with no env vars and no telegram/whatsapp pairing.
	r := &Result{
		APIs: []*APIBlock{
			{Provider: "numverify", Skipped: "NUMVERIFY_KEY not set"},
			{Provider: "veriphone", Skipped: "VERIPHONE_KEY not set"},
			{Provider: "ipqs", Skipped: "IPQS_KEY not set"},
		},
		Telegram: &TelegramBlock{Skipped: "TG_APP_ID / TG_APP_HASH not set in env"},
		WhatsApp: &WhatsAppBlock{Skipped: "not paired — run `clank whatsapp login`"},
	}
	suggs := collectSuggestions(r)
	if len(suggs) != 5 {
		t.Fatalf("expected 5 suggestions for fully-skipped blocks, got %d: %v", len(suggs), suggs)
	}
	// Every suggestion must be user-actionable: either an env var to set or
	// a clank subcommand to run. If neither verb appears, the hint is just
	// noise.
	for _, s := range suggs {
		if !strings.Contains(s, "set ") && !strings.Contains(s, "run ") {
			t.Errorf("suggestion not actionable (no set/run verb): %q", s)
		}
	}
}

func TestRun_NoSuggestionsWhenUserOptsOutOfEverything(t *testing.T) {
	// When the user sets every Skip* flag, blocks aren't even invoked, so
	// the Suggestions footer must stay empty — we don't badger them about
	// sources they explicitly chose to skip.
	res := Run(context.Background(), "+14155552671", offlineOpts())
	if len(res.Suggestions) != 0 {
		t.Errorf("expected 0 suggestions when user opts out of every source, got %v", res.Suggestions)
	}
}
