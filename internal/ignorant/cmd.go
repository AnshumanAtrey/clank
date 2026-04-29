package ignorant

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank ignorant <phone> [--only instagram,snapchat,amazon] [--json]

Phone-to-account presence checks across Instagram, Snapchat, and Amazon.
Pure HTTP probes — no auth, no API key. Ports megadose/ignorant to native Go.

flags:
  --only <csv>     restrict to a subset (e.g. --only instagram,snapchat)
  --json           machine-readable output

examples:
  clank ignorant +14155552671
  clank ignorant --only instagram +919181156055
  clank ignorant --json +14155552671
`

func Command(args []string) int {
	fs := flag.NewFlagSet("ignorant", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	only := fs.String("only", "", "comma-separated subset of modules")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}
	phone := fs.Arg(0)

	var onlySlice []string
	if *only != "" {
		for _, p := range strings.Split(*only, ",") {
			if t := strings.TrimSpace(p); t != "" {
				onlySlice = append(onlySlice, t)
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	results, err := Run(ctx, phone, onlySlice)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(results)
		return 0
	}

	render(results, phone)
	return 0
}

func render(results []Result, phone string) {
	fmt.Printf("Account presence for %s\n\n", phone)
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "PLATFORM\tDOMAIN\tEXISTS\tNOTES"
	if isTTY(os.Stdout) {
		header = color.New(color.Bold).Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, r := range results {
		exists := dash(yesNo(r.Exists))
		if isTTY(os.Stdout) {
			if r.RateLimit {
				exists = color.YellowString("?")
			} else if r.Exists {
				exists = color.GreenString("✓")
			} else {
				exists = color.New(color.Faint).Sprint("✗")
			}
		} else if r.RateLimit {
			exists = "?"
		} else if r.Exists {
			exists = "yes"
		} else {
			exists = "no"
		}
		notes := ""
		if r.Error != "" {
			notes = "ERR: " + truncate(r.Error, 60)
		} else if r.RateLimit {
			notes = "rate-limited / inconclusive"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", r.Name, r.Domain, exists, dash(notes))
	}
	tw.Flush()
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
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
