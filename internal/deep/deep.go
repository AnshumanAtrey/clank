// Package deep orchestrates every clank enrichment source for a single phone
// number. Each source runs in its own goroutine and gracefully degrades when
// not configured (no API key, not logged in, etc.) — never blocks the others.
package deep

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AnshumanAtrey/clank/internal/api"
	"github.com/AnshumanAtrey/clank/internal/edgar"
	"github.com/AnshumanAtrey/clank/internal/ignorant"
	"github.com/AnshumanAtrey/clank/internal/local"
	"github.com/AnshumanAtrey/clank/internal/telegram"
	"github.com/AnshumanAtrey/clank/internal/whatsapp"
)

// Result is the merged report from every source.
type Result struct {
	Phone    string         `json:"phone"`
	Region   string         `json:"default_region,omitempty"`
	E164     string         `json:"e164,omitempty"`
	Local    *local.Lookup  `json:"local,omitempty"`
	APIs     []*APIBlock    `json:"apis,omitempty"`
	Telegram *TelegramBlock `json:"telegram,omitempty"`
	WhatsApp *WhatsAppBlock `json:"whatsapp,omitempty"`
	Ignorant *IgnorantBlock `json:"ignorant,omitempty"`
	Edgar    *EdgarBlock    `json:"edgar,omitempty"`
	Took     string         `json:"took"`
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

type Options struct {
	Region         string
	SkipMessengers bool
	SkipEdgar      bool
	SkipAPIs       bool
	Timeout        time.Duration
}

func Run(ctx context.Context, input string, opts Options) *Result {
	start := time.Now()
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
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

	wg.Wait()
	res.Took = time.Since(start).Round(time.Millisecond).String()
	return res
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
