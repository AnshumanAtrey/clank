package deep

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"github.com/AnshumanAtrey/clank/internal/dorks"
)

const usage = `usage: clank deep <phone> [options]

Run every enrichment source for one phone number. Sources gracefully skip
when not configured (no API key, not logged in, etc.) — they never block
each other.

Sources:
  - Local        libphonenumber + MCC/MNC + spam list (instant, no key)
  - APIs         NumVerify / Veriphone / IPQS (skipped if env vars not set)
  - Telegram     phone-to-user (skipped if not logged in)
  - WhatsApp     phone-to-user (skipped if not paired)
  - ignorant     Instagram / Snapchat / Amazon presence
  - SEC EDGAR    full-text filings search
  - Dorks        Google search URLs across social/reputation/individuals/etc.

flags:
  --region <ISO>      default region for parsing (e.g. IN, US)
  --quick             skip Telegram, WhatsApp, ignorant — fast (<3s)
  --no-apis           skip free-tier API providers
  --no-edgar          skip SEC EDGAR
  --no-dorks          skip generated Google-dork URLs
  --json              JSON output
  --timeout <s>       overall timeout (default 60s)

examples:
  clank deep +14155552671
  clank deep --region IN 9181156055
  clank deep --quick +14155552671
  clank deep --json +14155552671 | jq
`

func Command(args []string) int {
	fs := flag.NewFlagSet("deep", flag.ExitOnError)
	region := fs.String("region", "", "default region")
	quick := fs.Bool("quick", false, "skip messengers")
	noAPIs := fs.Bool("no-apis", false, "skip API providers")
	noEdgar := fs.Bool("no-edgar", false, "skip SEC EDGAR")
	noDorks := fs.Bool("no-dorks", false, "skip Google-dork URL generation")
	jsonOut := fs.Bool("json", false, "JSON output")
	timeoutSec := fs.Int("timeout", 60, "overall timeout in seconds")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}
	phone := fs.Arg(0)

	res := Run(context.Background(), phone, Options{
		Region:         *region,
		SkipMessengers: *quick,
		SkipAPIs:       *noAPIs,
		SkipEdgar:      *noEdgar,
		SkipDorks:      *noDorks,
		Timeout:        time.Duration(*timeoutSec) * time.Second,
	})

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}

	render(res)
	return 0
}

func render(r *Result) {
	bold := color.New(color.Bold)
	faint := color.New(color.Faint)

	fmt.Println()
	fmt.Println(bold.Sprintf("Deep lookup — %s", r.Phone))
	if r.E164 != "" && r.E164 != r.Phone {
		fmt.Println(faint.Sprintf("  canonical: %s", r.E164))
	}
	fmt.Println()

	section(bold, "1. Local — libphonenumber + spam check")
	renderLocal(r.Local)
	fmt.Println()

	if hasAPIData(r.APIs) {
		section(bold, "2. Free-tier APIs")
		for _, a := range r.APIs {
			if a == nil || a.Skipped != "" {
				continue
			}
			renderAPI(a, faint)
		}
		fmt.Println()
	}

	if r.Telegram != nil && r.Telegram.Skipped == "" {
		section(bold, "3. Telegram")
		renderTelegram(r.Telegram, faint)
		fmt.Println()
	}

	if r.WhatsApp != nil && r.WhatsApp.Skipped == "" {
		section(bold, "4. WhatsApp")
		renderWhatsApp(r.WhatsApp, faint)
		fmt.Println()
	}

	if r.Ignorant != nil && r.Ignorant.Skipped == "" {
		section(bold, "5. Account presence (Instagram / Snapchat / Amazon)")
		renderIgnorant(r.Ignorant, faint)
		fmt.Println()
	}

	if r.Edgar != nil && r.Edgar.Skipped == "" {
		section(bold, "6. SEC EDGAR full-text")
		renderEdgar(r.Edgar, faint)
		fmt.Println()
	}

	if len(r.Dorks) > 0 {
		section(bold, "7. Pivot URLs — Google dorks")
		renderDorks(r.Dorks, faint)
		fmt.Println()
	}

	if len(r.Suggestions) > 0 {
		section(bold, "Tips to enrich future runs")
		for _, s := range r.Suggestions {
			fmt.Println("  " + faint.Sprint("• "+s))
		}
		fmt.Println()
	}

	fmt.Println(faint.Sprintf("Took %s", r.Took))
}

// hasAPIData returns true when at least one provider produced a non-skipped
// result (success, partial result with error, etc.) — i.e. the section has
// something to render that isn't a missing-key footer hint.
func hasAPIData(apis []*APIBlock) bool {
	for _, a := range apis {
		if a != nil && a.Skipped == "" {
			return true
		}
	}
	return false
}

func section(c *color.Color, text string) {
	fmt.Println(c.Sprint(text))
}

func renderLocal(l interface{ Summary() string }) {
	if l == nil {
		fmt.Println("  (no result)")
		return
	}
	fmt.Println("  " + l.Summary())
}

func renderAPI(b *APIBlock, faint *color.Color) {
	switch {
	case b.Skipped != "":
		fmt.Printf("  %-12s %s\n", b.Provider, faint.Sprint("skipped — "+b.Skipped))
	case b.Error != "":
		fmt.Printf("  %-12s %s\n", b.Provider, color.RedString("error — "+truncate(b.Error, 80)))
	case b.Result != nil:
		r := b.Result
		parts := []string{}
		if r.Valid {
			parts = append(parts, color.GreenString("valid"))
		} else {
			parts = append(parts, color.RedString("invalid"))
		}
		if r.LineType != "" {
			parts = append(parts, r.LineType)
		}
		if r.Carrier != "" {
			parts = append(parts, r.Carrier)
		}
		if r.Country != "" {
			parts = append(parts, r.Country)
		}
		if r.Location != "" {
			parts = append(parts, r.Location)
		}
		if r.FraudScore != nil {
			fs := *r.FraudScore
			fsStr := fmt.Sprintf("fraud=%d", fs)
			switch {
			case fs >= 75:
				fsStr = color.RedString(fsStr)
			case fs >= 50:
				fsStr = color.YellowString(fsStr)
			default:
				fsStr = color.GreenString(fsStr)
			}
			parts = append(parts, fsStr)
		}
		fmt.Printf("  %-12s %s\n", b.Provider, strings.Join(parts, " · "))
	}
}

func renderTelegram(b *TelegramBlock, faint *color.Color) {
	if b.Skipped != "" {
		fmt.Println("  " + faint.Sprint("skipped — "+b.Skipped))
		return
	}
	if b.Error != "" {
		fmt.Println("  " + color.RedString("error — "+truncate(b.Error, 80)))
		return
	}
	if b.Result == nil {
		fmt.Println("  (no result)")
		return
	}
	r := b.Result
	if !r.Found {
		msg := "not registered"
		if r.Reason != "" {
			msg += " (" + r.Reason + ")"
		}
		fmt.Println("  " + faint.Sprint(msg))
		return
	}
	fmt.Printf("  %s · user_id=%d\n", color.GreenString("registered"), r.UserID)
	if r.FirstName != "" || r.LastName != "" {
		fmt.Printf("  name      : %s %s\n", r.FirstName, r.LastName)
	}
	if r.Username != "" {
		fmt.Printf("  username  : @%s\n", r.Username)
	}
	flags := []string{}
	if r.Premium {
		flags = append(flags, "premium")
	}
	if r.Verified {
		flags = append(flags, "verified")
	}
	if r.Bot {
		flags = append(flags, "bot")
	}
	if r.Restricted {
		flags = append(flags, "restricted")
	}
	if r.HasPhoto {
		flags = append(flags, "has-photo")
	}
	if len(flags) > 0 {
		fmt.Printf("  flags     : %s\n", strings.Join(flags, ", "))
	}
}

func renderWhatsApp(b *WhatsAppBlock, faint *color.Color) {
	if b.Skipped != "" {
		fmt.Println("  " + faint.Sprint("skipped — "+b.Skipped))
		return
	}
	if b.Error != "" {
		fmt.Println("  " + color.RedString("error — "+truncate(b.Error, 100)))
		return
	}
	if b.Result == nil {
		fmt.Println("  (no result)")
		return
	}
	r := b.Result
	if !r.Registered {
		msg := "not registered"
		if r.Reason != "" {
			msg += " (" + r.Reason + ")"
		}
		fmt.Println("  " + faint.Sprint(msg))
		return
	}
	fmt.Println("  " + color.GreenString("registered"))
	if r.JID != "" {
		fmt.Printf("  jid          : %s\n", r.JID)
	}
	if r.VerifiedBusinessName != "" {
		fmt.Printf("  business     : %s\n", r.VerifiedBusinessName)
	}
	if r.About != "" {
		fmt.Printf("  about        : %s\n", r.About)
	}
	if r.DeviceCount > 0 {
		fmt.Printf("  devices      : %d\n", r.DeviceCount)
	}
	if r.ProfilePictureURL != "" {
		fmt.Printf("  pic url      : %s\n", r.ProfilePictureURL)
	}
}

func renderIgnorant(b *IgnorantBlock, faint *color.Color) {
	if b.Skipped != "" {
		fmt.Println("  " + faint.Sprint("skipped — "+b.Skipped))
		return
	}
	if b.Error != "" {
		fmt.Println("  " + color.RedString("error — "+truncate(b.Error, 80)))
		return
	}
	if len(b.Results) == 0 {
		fmt.Println("  (no results)")
		return
	}
	for _, r := range b.Results {
		var status string
		switch {
		case r.RateLimit && r.Error != "":
			status = faint.Sprint("? — " + truncate(r.Error, 60))
		case r.RateLimit:
			status = color.YellowString("? rate-limited / inconclusive")
		case r.Exists:
			status = color.GreenString("registered")
		default:
			status = faint.Sprint("not registered")
		}
		fmt.Printf("  %-10s %s\n", r.Name, status)
	}
}

func renderEdgar(b *EdgarBlock, faint *color.Color) {
	if b.Skipped != "" {
		fmt.Println("  " + faint.Sprint("skipped — "+b.Skipped))
		return
	}
	if b.Error != "" {
		fmt.Println("  " + color.RedString("error — "+truncate(b.Error, 80)))
		return
	}
	if b.Result == nil {
		fmt.Println("  (no result)")
		return
	}
	r := b.Result
	if r.Total == 0 {
		fmt.Println("  " + faint.Sprint("0 hits"))
		return
	}
	fmt.Printf("  %d total hits, showing top %d:\n", r.Total, r.Returned)
	for _, h := range r.Hits {
		name := "(unknown)"
		if len(h.DisplayNames) > 0 {
			name = truncate(h.DisplayNames[0], 60)
		}
		fmt.Printf("    %s  %-6s  %s\n", h.FileDate, h.Form, name)
	}
}

// renderDorks prints one URL per (bucket, site) pair, preferring the E.164
// (`+`-prefixed) variation when present so the displayed URL is the most
// canonical form. Full per-variation list stays available in JSON output.
func renderDorks(d []dorks.Dork, faint *color.Color) {
	type key struct{ bucket, site string }
	chosen := map[key]dorks.Dork{}
	order := []key{}
	for _, dk := range d {
		k := key{dk.Bucket, dk.Site}
		cur, exists := chosen[k]
		if !exists {
			chosen[k] = dk
			order = append(order, k)
			continue
		}
		if !strings.Contains(cur.Query, "+") && strings.Contains(dk.Query, "+") {
			chosen[k] = dk
		}
	}

	currentBucket := ""
	for _, k := range order {
		dk := chosen[k]
		if dk.Bucket != currentBucket {
			fmt.Println("  " + faint.Sprintf("[%s]", dk.Bucket))
			currentBucket = dk.Bucket
		}
		label := dk.Site
		if label == "" {
			label = "google"
		}
		fmt.Printf("    %-22s %s\n", label, dk.URL)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return s[:n-1] + "…"
}
