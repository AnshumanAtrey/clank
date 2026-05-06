package deep

import (
	"context"
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
