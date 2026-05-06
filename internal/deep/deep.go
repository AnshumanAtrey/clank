// Package deep orchestrates every clank enrichment source for a single phone
// number. Each source runs in its own goroutine and gracefully degrades when
// not configured (no API key, not logged in, etc.) — never blocks the others.
package deep

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AnshumanAtrey/clank/internal/api"
	"github.com/AnshumanAtrey/clank/internal/dorks"
	"github.com/AnshumanAtrey/clank/internal/edgar"
	"github.com/AnshumanAtrey/clank/internal/fcc"
	clankgithub "github.com/AnshumanAtrey/clank/internal/github"
	"github.com/AnshumanAtrey/clank/internal/ignorant"
	"github.com/AnshumanAtrey/clank/internal/local"
	"github.com/AnshumanAtrey/clank/internal/ovh"
	"github.com/AnshumanAtrey/clank/internal/telegram"
	"github.com/AnshumanAtrey/clank/internal/whatsapp"
)

// Result is the merged report from every source.
type Result struct {
	Phone       string         `json:"phone"`
	Region      string         `json:"default_region,omitempty"`
	E164        string         `json:"e164,omitempty"`
	Local       *local.Lookup  `json:"local,omitempty"`
	APIs        []*APIBlock    `json:"apis,omitempty"`
	Telegram    *TelegramBlock `json:"telegram,omitempty"`
	WhatsApp    *WhatsAppBlock `json:"whatsapp,omitempty"`
	Ignorant    *IgnorantBlock `json:"ignorant,omitempty"`
	Edgar       *EdgarBlock    `json:"edgar,omitempty"`
	FCC         *FCCBlock      `json:"fcc,omitempty"`
	OVH         *OVHBlock      `json:"ovh,omitempty"`
	GitHub      *GitHubBlock   `json:"github,omitempty"`
	Dorks       []dorks.Dork   `json:"dorks,omitempty"`
	Suggestions []string       `json:"suggestions,omitempty"`
	Took        string         `json:"took"`
}

type APIBlock struct {
	Provider string      `json:"provider"`
	Result   *api.Result `json:"result,omitempty"`
	Skipped  string      `json:"skipped,omitempty"`
	Error    string      `json:"error,omitempty"`
}

type TelegramBlock struct {
	Result  *telegram.Lookup `json:"result,omitempty"`
	Skipped string           `json:"skipped,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type WhatsAppBlock struct {
	Result  *whatsapp.Result `json:"result,omitempty"`
	Skipped string           `json:"skipped,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type IgnorantBlock struct {
	Results []ignorant.Result `json:"results,omitempty"`
	Skipped string            `json:"skipped,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type EdgarBlock struct {
	Result  *edgar.Response `json:"result,omitempty"`
	Skipped string          `json:"skipped,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type FCCBlock struct {
	Result  *fcc.Response `json:"result,omitempty"`
	Skipped string        `json:"skipped,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type GitHubBlock struct {
	Result  *clankgithub.Response `json:"result,omitempty"`
	Skipped string                `json:"skipped,omitempty"`
	Error   string                `json:"error,omitempty"`
}

type OVHBlock struct {
	Result  *ovh.Response `json:"result,omitempty"`
	Skipped string        `json:"skipped,omitempty"`
	Error   string        `json:"error,omitempty"`
}

type Options struct {
	Region         string
	SkipMessengers bool
	SkipEdgar      bool
	SkipAPIs       bool
	SkipFCC        bool
	SkipOVH        bool
	SkipGitHub     bool
	SkipDorks      bool
	Timeout        time.Duration
}

// DefaultTimeout is the parent ceiling applied when caller does not supply
// one. Tightened from the original 60 s in clank ≤ v0.1.0: zero-config runs
// finish in ~2 s, so 30 s leaves comfortable headroom for paired Telegram /
// WhatsApp IQ round-trips while still bounding worst-case waits when a slow
// source hangs. Per-source timeouts arrive with the v0.2.0-rc1 transport
// rework.
const DefaultTimeout = 30 * time.Second

func Run(ctx context.Context, input string, opts Options) *Result {
	start := time.Now()
	if opts.Timeout == 0 {
		opts.Timeout = DefaultTimeout
	}

	runCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	l := local.Inspect(input, opts.Region)
	res := &Result{Phone: input, Region: opts.Region, Local: &l, E164: l.E164}

	canonical := l.E164
	if canonical == "" {
		canonical = input
	}

	var wg sync.WaitGroup

	if !opts.SkipAPIs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.APIs = runAPIs(runCtx, canonical)
		}()
	}

	if !opts.SkipMessengers {
		wg.Add(3)
		go func() {
			defer wg.Done()
			res.Telegram = runTelegram(runCtx, canonical)
		}()
		go func() {
			defer wg.Done()
			res.WhatsApp = runWhatsApp(runCtx, canonical)
		}()
		go func() {
			defer wg.Done()
			res.Ignorant = runIgnorant(runCtx, canonical)
		}()
	}

	if !opts.SkipEdgar {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.Edgar = runEdgar(runCtx, canonical)
		}()
	}

	if !opts.SkipFCC {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.FCC = runFCC(runCtx, canonical)
		}()
	}

	if !opts.SkipGitHub {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.GitHub = runGitHub(runCtx, canonical)
		}()
	}

	if !opts.SkipOVH {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res.OVH = runOVH(runCtx, canonical)
		}()
	}

	wg.Wait()

	if !opts.SkipDorks {
		// Pure string templating — no network — but parse failures are non-fatal.
		if d, err := dorks.Generate(input, opts.Region, nil); err == nil {
			res.Dorks = d
		}
	}

	res.Suggestions = collectSuggestions(res)

	res.Took = time.Since(start).Round(time.Millisecond).String()
	return res
}

// collectSuggestions inspects skipped sources and produces user-actionable
// hints — keys to set, subcommands to run — so the UI can surface them as a
// single footer instead of cluttering the report with N "skipped" lines.
func collectSuggestions(r *Result) []string {
	var out []string
	for _, a := range r.APIs {
		if a == nil || a.Skipped == "" {
			continue
		}
		switch a.Provider {
		case "numverify":
			out = append(out, "set NUMVERIFY_KEY for free validation + carrier (100/mo free)")
		case "veriphone":
			out = append(out, "set VERIPHONE_KEY for line type + carrier (1k/mo free)")
		case "ipqs":
			out = append(out, "set IPQS_KEY for fraud score + VOIP/abuse flags (1k/mo free)")
		}
	}
	if r.Telegram != nil && r.Telegram.Skipped != "" {
		if strings.Contains(r.Telegram.Skipped, "TG_APP_ID") {
			out = append(out, "set TG_APP_ID + TG_APP_HASH (free at my.telegram.org/apps) and run `clank telegram login` for phone → username + name")
		} else {
			out = append(out, "run `clank telegram login` for phone → username + name attribution")
		}
	}
	if r.WhatsApp != nil && r.WhatsApp.Skipped != "" {
		out = append(out, "run `clank whatsapp login` for phone → display name + business + about + device count")
	}
	if r.GitHub != nil && r.GitHub.Skipped != "" && strings.Contains(r.GitHub.Skipped, "GITHUB_TOKEN") {
		out = append(out, "set GITHUB_TOKEN to raise GitHub rate limit from 10/min to 30/min (any classic or fine-grained PAT)")
	}
	return out
}

func runAPIs(ctx context.Context, phone string) []*APIBlock {
	type spec struct {
		name string
		env  string
	}
	specs := []spec{
		{"numverify", "NUMVERIFY_KEY"},
		{"veriphone", "VERIPHONE_KEY"},
		{"ipqs", "IPQS_KEY"},
	}

	out := make([]*APIBlock, len(specs))
	var wg sync.WaitGroup
	for i, s := range specs {
		i, s := i, s
		wg.Add(1)
		go func() {
			defer wg.Done()
			b := &APIBlock{Provider: s.name}
			key := os.Getenv(s.env)
			if key == "" {
				b.Skipped = fmt.Sprintf("%s not set", s.env)
				out[i] = b
				return
			}
			prov, err := api.New(s.name, key)
			if err != nil {
				b.Error = err.Error()
				out[i] = b
				return
			}
			r, err := prov.Lookup(ctx, phone)
			if err != nil {
				b.Error = err.Error()
			}
			b.Result = &r
			out[i] = b
		}()
	}
	wg.Wait()
	return out
}

func runTelegram(ctx context.Context, phone string) *TelegramBlock {
	home, err := os.UserHomeDir()
	if err != nil {
		return &TelegramBlock{Error: "home dir: " + err.Error()}
	}
	if _, err := os.Stat(filepath.Join(home, ".clank", "telegram.session")); os.IsNotExist(err) {
		return &TelegramBlock{Skipped: "not logged in — run `clank telegram login`"}
	}
	if os.Getenv("TG_APP_ID") == "" || os.Getenv("TG_APP_HASH") == "" {
		return &TelegramBlock{Skipped: "TG_APP_ID / TG_APP_HASH not set in env"}
	}
	r, err := telegram.ResolvePhone(ctx, phone)
	b := &TelegramBlock{Result: &r}
	if err != nil && r.Reason == "" {
		b.Error = err.Error()
	}
	return b
}

func runWhatsApp(ctx context.Context, phone string) *WhatsAppBlock {
	home, err := os.UserHomeDir()
	if err != nil {
		return &WhatsAppBlock{Error: "home dir: " + err.Error()}
	}
	if _, err := os.Stat(filepath.Join(home, ".clank", "whatsapp.db")); os.IsNotExist(err) {
		return &WhatsAppBlock{Skipped: "not paired — run `clank whatsapp login`"}
	}
	h, err := whatsapp.Open(ctx, false)
	if err != nil {
		// Differentiate "no session yet" from real connection errors.
		if strings.Contains(err.Error(), "not paired") {
			return &WhatsAppBlock{Skipped: "not paired — run `clank whatsapp login`"}
		}
		return &WhatsAppBlock{Error: err.Error()}
	}
	defer h.Close()
	r, err := h.Lookup(ctx, phone)
	b := &WhatsAppBlock{Result: r}
	if err != nil {
		b.Error = err.Error()
	}
	return b
}

func runIgnorant(ctx context.Context, phone string) *IgnorantBlock {
	results, err := ignorant.Run(ctx, phone, nil)
	if err != nil {
		return &IgnorantBlock{Error: err.Error()}
	}
	return &IgnorantBlock{Results: results}
}

func runEdgar(ctx context.Context, phone string) *EdgarBlock {
	resp, err := edgar.Search(ctx, phone, edgar.Options{Hits: 5})
	if err != nil {
		return &EdgarBlock{Error: err.Error()}
	}
	return &EdgarBlock{Result: resp}
}

func runFCC(ctx context.Context, phone string) *FCCBlock {
	resp, err := fcc.Search(ctx, phone, fcc.Options{Limit: 25})
	if err != nil {
		// US-only dataset — surface as a skip rather than an error so non-US
		// users see a clean report instead of a red error line.
		if errors.Is(err, fcc.ErrNonUS) {
			return &FCCBlock{Skipped: "FCC dataset is US-only"}
		}
		return &FCCBlock{Error: err.Error()}
	}
	return &FCCBlock{Result: resp}
}

func runOVH(ctx context.Context, phone string) *OVHBlock {
	resp, err := ovh.Lookup(ctx, phone, ovh.Options{})
	if err != nil {
		// EU-only — surface as skip so non-supported regions get a clean
		// report rather than a red error.
		if errors.Is(err, ovh.ErrUnsupportedRegion) {
			return &OVHBlock{Skipped: "OVH covers EU landline ranges only"}
		}
		return &OVHBlock{Error: err.Error()}
	}
	return &OVHBlock{Result: resp}
}

func runGitHub(ctx context.Context, phone string) *GitHubBlock {
	resp, err := clankgithub.Search(ctx, phone, clankgithub.Options{Limit: 5})
	if err != nil {
		// Rate-limit is the most likely failure mode in zero-config use; tag
		// it specifically so the renderer can suggest GITHUB_TOKEN.
		if errors.Is(err, clankgithub.ErrRateLimited) {
			return &GitHubBlock{Skipped: "GitHub rate-limited (10 req/min unauth) — set GITHUB_TOKEN"}
		}
		return &GitHubBlock{Error: err.Error()}
	}
	return &GitHubBlock{Result: resp}
}
