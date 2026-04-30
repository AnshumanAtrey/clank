package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/AnshumanAtrey/clank/internal/api"
	"github.com/AnshumanAtrey/clank/internal/audit"
	"github.com/AnshumanAtrey/clank/internal/deep"
	"github.com/AnshumanAtrey/clank/internal/dorks"
	"github.com/AnshumanAtrey/clank/internal/edgar"
	"github.com/AnshumanAtrey/clank/internal/ignorant"
	"github.com/AnshumanAtrey/clank/internal/imei"
	"github.com/AnshumanAtrey/clank/internal/local"
	"github.com/AnshumanAtrey/clank/internal/pattern"
	"github.com/AnshumanAtrey/clank/internal/scan"
	"github.com/AnshumanAtrey/clank/internal/telegram"
	"github.com/AnshumanAtrey/clank/internal/whatsapp"
	"github.com/fatih/color"
)

// version is overridden at build time via -ldflags="-X main.version=v1.2.3".
// Set by goreleaser on tagged releases; "dev" for `go install` builds.
var version = "dev"

const banner = `________  ___       ________  _________  ___  ___
|\   ____\|\  \     |\   __  \|\   ___  \|\  \|\  \
\ \  \___|\ \  \    \ \  \|\  \ \  \\ \  \ \  \/  /|_
 \ \  \    \ \  \    \ \   __  \ \  \\ \  \ \   ___  \
  \ \  \____\ \  \____\ \  \ \  \ \  \\ \  \ \  \\ \  \
   \ \_______\ \_______\ \__\ \__\ \__\\ \__\ \__\\ \__\
    \|_______|\|_______|\|__|\|__|\|__| \|__|\|__| \|__|`

const helpText = `clank — phone-number pattern + OSINT lookup CLI

Subcommands:
  clank deep <phone>                    Run ALL sources concurrently — local + APIs + Telegram + WhatsApp + ignorant + EDGAR
  clank scan <pattern>                  Run 'deep' over each pattern combo with rate-limit-aware sleeps
  clank dorks <phone>                   Generate Google-dork URLs across 5 buckets (--open to launch in browser)
  clank imei <15-digit>                 Decode IMEI (Luhn + manufacturer/model from TAC)
  clank edgar <number-or-string>        SEC EDGAR full-text filings search
  clank telegram <login|lookup|logout>  Telegram phone-to-user resolver
  clank ignorant <phone>                Phone presence on Instagram / Snapchat / Amazon
  clank whatsapp <login|lookup|logout>  Phone presence on WhatsApp (paired session)
  clank history                         Show recent clank invocations (~/.clank/history.jsonl)
  clank --version                       Print version and exit

Pattern usage:
  clank <pattern>                       Interactive menu (default)
  clank -n <pattern>                    Same as above (legacy flag)
  clank --local <pattern>               Skip menu — libphonenumber lookup
  clank --provider <name> <pattern>     Skip menu — API lookup
  clank --gen <pattern>                 Legacy: just print all combinations
  clank --json <pattern>                Output JSON (combine with mode flag)

Pattern format:
  Use digits 0-9 for known positions and x or X for unknown.
  Example: 918115605xxx  →  expands to 1000 candidates 918115605000..918115605999.

Modes:
  --local                       Pure-offline lookup (libphonenumber + MCC/MNC + spam list)
  --provider numverify          NumVerify (apilayer)              env: NUMVERIFY_KEY
  --provider veriphone          Veriphone                         env: VERIPHONE_KEY
  --provider ipqs               IPQualityScore (fraud score)      env: IPQS_KEY

Flags:
  -n <pattern>           Pattern (legacy; positional arg also works)
  -r, --region <ISO>     Default region for parsing (e.g. IN, US). Recommended for local-format input.
  --key <KEY>            Override env var for API key
  --workers <N>          Concurrent API workers (default 5)
  --show-invalid         Include invalid combinations in output (default: skip)
  --json                 Machine-readable JSON output
  --no-banner            Suppress the ASCII banner
  --force                Allow patterns >100,000 combinations
  -h, --help             Show this message

Examples:
  clank 918115605xxx                    Generate 1000, then prompt for action
  clank --local +14155552671            Single-number local OSINT
  clank --local --region IN 9181156052  Region-hinted local lookup
  clank --provider ipqs +14155552671    Single number, IPQS fraud score
  clank --gen 918115605xx               Just dump 100 combinations to stdout
  clank --json --local +14155552671     JSON output for piping to jq
`

type opts struct {
	patternFlag string
	region      string
	provider    string
	key         string
	workers     int
	gen         bool
	localOnly   bool
	jsonOut     bool
	noBanner    bool
	force       bool
	showInvalid bool
	help        bool
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Println("clank", version)
			return
		case "telegram":
			os.Exit(audit.Wrap("telegram", os.Args[2:], telegram.Command))
		case "imei":
			os.Exit(audit.Wrap("imei", os.Args[2:], imei.Command))
		case "edgar":
			os.Exit(audit.Wrap("edgar", os.Args[2:], edgar.Command))
		case "ignorant":
			os.Exit(audit.Wrap("ignorant", os.Args[2:], ignorant.Command))
		case "whatsapp":
			os.Exit(audit.Wrap("whatsapp", os.Args[2:], whatsapp.Command))
		case "deep":
			os.Exit(audit.Wrap("deep", os.Args[2:], deep.Command))
		case "scan":
			os.Exit(audit.Wrap("scan", os.Args[2:], scan.Command))
		case "dorks":
			os.Exit(audit.Wrap("dorks", os.Args[2:], dorks.Command))
		case "history":
			os.Exit(audit.Command(os.Args[2:]))
		}
	}

	o := parseFlags()
	if o.help {
		fmt.Print(helpText)
		return
	}

	pat := o.patternFlag
	if pat == "" && flag.NArg() > 0 {
		pat = flag.Arg(0)
	}
	if pat == "" {
		fmt.Fprintln(os.Stderr, "error: no pattern provided")
		fmt.Fprintln(os.Stderr, "run 'clank -h' for usage")
		os.Exit(1)
	}

	combos, err := pattern.Generate(pat, o.force)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if !o.noBanner && !o.jsonOut {
		fmt.Println(banner)
		fmt.Printf("\nGenerated %d candidate%s for pattern %s\n\n", len(combos), plural(len(combos)), pat)
	}

	switch {
	case o.gen:
		printAll(combos, o.jsonOut)
	case o.localOnly:
		runLocal(combos, o.region, o.jsonOut, o.showInvalid)
	case o.provider != "":
		runAPI(combos, o.region, o.provider, o.key, o.workers, o.jsonOut, o.showInvalid)
	default:
		if len(combos) == 1 {
			runLocal(combos, o.region, o.jsonOut, o.showInvalid)
			return
		}
		runMenu(combos, o.region, o.workers, o.showInvalid)
	}
}

func parseFlags() opts {
	var o opts
	flag.StringVar(&o.patternFlag, "n", "", "phone number pattern")
	flag.StringVar(&o.region, "region", "", "default region for parsing (e.g. IN, US)")
	flag.StringVar(&o.region, "r", "", "default region for parsing (e.g. IN, US)")
	flag.StringVar(&o.provider, "provider", "", "API provider: numverify | veriphone | ipqs")
	flag.StringVar(&o.key, "key", "", "API key (overrides env var)")
	flag.IntVar(&o.workers, "workers", 5, "concurrent API workers")
	flag.BoolVar(&o.gen, "gen", false, "legacy: just print all combinations")
	flag.BoolVar(&o.localOnly, "local", false, "skip menu, run local lookup")
	flag.BoolVar(&o.jsonOut, "json", false, "output JSON")
	flag.BoolVar(&o.noBanner, "no-banner", false, "skip ASCII banner")
	flag.BoolVar(&o.force, "force", false, "allow patterns >100,000 combinations")
	flag.BoolVar(&o.showInvalid, "show-invalid", false, "include invalid combinations in output")
	flag.BoolVar(&o.help, "h", false, "show help")
	flag.BoolVar(&o.help, "help", false, "show help")
	flag.Usage = func() { fmt.Print(helpText) }
	flag.Parse()
	if o.workers < 1 {
		o.workers = 1
	}
	return o
}

func runMenu(combos []string, region string, workers int, showInvalid bool) {
	fmt.Println("What would you like to do?")
	fmt.Println("  [1] Local lookup       — instant, no API key (libphonenumber + spam check)")
	fmt.Println("  [2] Free API lookup    — needs your API key")
	fmt.Println("  [3] Print all combos   — original behaviour")
	fmt.Println("  [q] Quit")
	choice := readLine("\n> ")
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "1":
		runLocal(combos, region, false, showInvalid)
	case "2":
		runAPIMenu(combos, region, workers, showInvalid)
	case "3":
		printAll(combos, false)
	case "q", "":
		fmt.Println("bye")
	default:
		fmt.Println("invalid choice")
	}
}

func runAPIMenu(combos []string, region string, workers int, showInvalid bool) {
	fmt.Println("\nChoose provider:")
	fmt.Println("  [a] NumVerify       (env: NUMVERIFY_KEY)")
	fmt.Println("  [b] Veriphone       (env: VERIPHONE_KEY)")
	fmt.Println("  [c] IPQualityScore  (env: IPQS_KEY)  ← only one with fraud_score")
	fmt.Println("  [b]ack")
	choice := readLine("\n> ")
	var providerKey string
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "a", "numverify":
		providerKey = "numverify"
	case "veriphone":
		providerKey = "veriphone"
	case "c", "ipqs", "ipqualityscore":
		providerKey = "ipqs"
	case "back", "":
		return
	default:
		fmt.Println("invalid choice")
		return
	}
	runAPI(combos, region, providerKey, "", workers, false, showInvalid)
}

func runLocal(combos []string, region string, jsonOut, showInvalid bool) {
	results := local.InspectMany(combos, region)

	valid := make([]local.Lookup, 0, len(results))
	for _, r := range results {
		if r.Valid || showInvalid {
			valid = append(valid, r)
		}
	}

	if jsonOut {
		writeJSON(valid)
		return
	}

	fmt.Printf("Local lookup — %d valid / %d total\n", countValid(results), len(results))
	if len(valid) == 0 {
		fmt.Println("(no results — try --show-invalid or --region <ISO>)")
		return
	}
	renderLocalTable(os.Stdout, valid)
	if !showInvalid && countValid(results) < len(results) {
		fmt.Printf("\n(%d invalid candidates hidden — pass --show-invalid to see them)\n", len(results)-countValid(results))
	}
	if op := pickFirstWithOperators(valid); len(op) > 0 {
		fmt.Println()
		fmt.Println(color.New(color.Bold).Sprint("Operators in country:"))
		shown := dedupeOperators(op)
		const cap = 12
		for i, o := range shown {
			if i >= cap {
				fmt.Printf("  …and %d more (run `clank --json --local <number>` to see them all)\n", len(shown)-cap)
				break
			}
			line := fmt.Sprintf("  • %s", firstNonEmpty(o.Brand, o.Operator))
			if o.Operator != "" && o.Operator != o.Brand {
				line += fmt.Sprintf(" (%s)", o.Operator)
			}
			if o.Bands != "" {
				line += "  — " + truncate(o.Bands, 50)
			}
			fmt.Println(line)
		}
	}
}

func dedupeOperators(ops []local.Operator) []local.Operator {
	seen := map[string]struct{}{}
	out := make([]local.Operator, 0, len(ops))
	for _, o := range ops {
		key := o.Brand + "|" + o.Operator
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, o)
	}
	return out
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return "(unknown)"
}

func runAPI(combos []string, region, providerKey, keyOverride string, workers int, jsonOut, showInvalid bool) {
	pre := local.InspectMany(combos, region)
	candidates := make([]string, 0, len(combos))
	for _, p := range pre {
		if p.Valid {
			candidates = append(candidates, p.E164)
		} else if showInvalid && p.E164 != "" {
			candidates = append(candidates, p.E164)
		}
	}
	if len(candidates) == 0 {
		fmt.Fprintln(os.Stderr, "no valid numbers in pattern — pass --region <ISO> or --show-invalid")
		os.Exit(1)
	}

	key := keyOverride
	if key == "" {
		key = keyForProvider(providerKey)
	}
	if key == "" {
		fmt.Fprintf(os.Stderr, "error: no API key for %q — set %s or pass --key\n", providerKey, envVarFor(providerKey))
		os.Exit(1)
	}

	prov, err := api.New(providerKey, key)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if !jsonOut {
		fmt.Printf("About to call %s %d time%s. Continue? [y/N] ", prov.Name(), len(candidates), plural(len(candidates)))
		ans := strings.ToLower(strings.TrimSpace(readLine("")))
		if ans != "y" && ans != "yes" {
			fmt.Println("cancelled")
			return
		}
		fmt.Printf("Looking up %d numbers via %s (workers=%d)...\n", len(candidates), prov.Name(), workers)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60+5*len(candidates))*time.Second)
	defer cancel()
	results := api.LookupBatch(ctx, prov, candidates, workers)

	if jsonOut {
		writeJSON(results)
		return
	}
	renderAPITable(os.Stdout, results)
}

func printAll(combos []string, jsonOut bool) {
	if jsonOut {
		writeJSON(combos)
		return
	}
	for _, c := range combos {
		fmt.Println(c)
	}
}

func renderLocalTable(w *os.File, ls []local.Lookup) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	header := "NUMBER\tVALID\tTYPE\tREGION\tCARRIER\tGEO\tTZ\tSPAM"
	if isTTY(w) {
		header = color.New(color.Bold).Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, l := range ls {
		valid := "✓"
		if !l.Valid {
			valid = "✗"
		}
		spam := "-"
		if l.Spam {
			spam = "⚠"
		}
		if isTTY(w) {
			if l.Valid {
				valid = color.GreenString(valid)
			} else {
				valid = color.RedString(valid)
			}
			if l.Spam {
				spam = color.YellowString(spam)
			}
		}
		tz := ""
		if len(l.Timezones) > 0 {
			tz = l.Timezones[0]
		}
		num := l.E164
		if num == "" {
			num = l.Input
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			num, valid, dash(l.LineType), dash(l.Region), dash(truncate(l.Carrier, 20)), dash(truncate(l.Geo, 20)), dash(tz), spam)
	}
	tw.Flush()
}

func renderAPITable(w *os.File, rs []api.Result) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	header := "NUMBER\tVALID\tTYPE\tCARRIER\tCOUNTRY\tLOCATION\tFRAUD\tNOTES"
	if isTTY(w) {
		header = color.New(color.Bold).Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, r := range rs {
		valid := "✓"
		if !r.Valid {
			valid = "✗"
		}
		fraud := "-"
		if r.FraudScore != nil {
			fraud = fmt.Sprintf("%d", *r.FraudScore)
			if isTTY(w) {
				switch {
				case *r.FraudScore >= 75:
					fraud = color.RedString(fraud)
				case *r.FraudScore >= 50:
					fraud = color.YellowString(fraud)
				default:
					fraud = color.GreenString(fraud)
				}
			}
		}
		notes := ""
		if r.Error != "" {
			notes = "ERR: " + truncate(r.Error, 40)
			if isTTY(w) {
				notes = color.RedString(notes)
			}
		} else if r.Risky != nil && *r.Risky {
			notes = "risky"
		}
		if isTTY(w) {
			if r.Valid {
				valid = color.GreenString(valid)
			} else {
				valid = color.RedString(valid)
			}
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			r.Number, valid, dash(r.LineType), dash(truncate(r.Carrier, 20)),
			dash(r.Country), dash(truncate(r.Location, 20)), fraud, dash(notes))
	}
	tw.Flush()
}

func keyForProvider(p string) string {
	switch strings.ToLower(p) {
	case "numverify":
		return os.Getenv("NUMVERIFY_KEY")
	case "veriphone":
		return os.Getenv("VERIPHONE_KEY")
	case "ipqs":
		return os.Getenv("IPQS_KEY")
	}
	return ""
}

func envVarFor(p string) string {
	switch strings.ToLower(p) {
	case "numverify":
		return "NUMVERIFY_KEY"
	case "veriphone":
		return "VERIPHONE_KEY"
	case "ipqs":
		return "IPQS_KEY"
	}
	return "<provider>_KEY"
}

func writeJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func readLine(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return line
}

func countValid(ls []local.Lookup) int {
	n := 0
	for _, l := range ls {
		if l.Valid {
			n++
		}
	}
	return n
}

func pickFirstWithOperators(ls []local.Lookup) []local.Operator {
	for _, l := range ls {
		if len(l.Operators) > 0 {
			return l.Operators
		}
	}
	return nil
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
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

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
