package fcc

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank fcc <us-phone> [--limit 25] [--json]

Search the FCC's "Consumer Complaints — Unwanted Calls" public dataset
(opendata.fcc.gov, ~5M rows since 2014, nightly updates) for spam-call
complaints filed against the given phone number.

US numbers only — non-NANP inputs return an explanatory error.

flags:
  --limit <N>        max complaints to fetch (1-100; default 25)
  --json             JSON output

examples:
  clank fcc +12037601637
  clank fcc 415-555-2671
  clank fcc --json +12037601637 | jq

env:
  FCC_APP_TOKEN      optional Socrata app token for higher rate limits
                     (free at https://opendata.fcc.gov/profile/app_tokens)
`

func Command(args []string) int {
	fs := flag.NewFlagSet("fcc", flag.ExitOnError)
	limit := fs.Int("limit", 25, "max complaints to fetch")
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}

	phone := fs.Arg(0)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := Search(ctx, phone, Options{Limit: *limit})
	if err != nil {
		if errors.Is(err, ErrNonUS) {
			fmt.Fprintf(os.Stderr, "error: %s\n", err)
			return 2
		}
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		return 0
	}

	bold := color.New(color.Bold)
	faint := color.New(color.Faint)

	fmt.Printf("FCC complaints for %s\n", resp.Query)
	if resp.Count == 0 {
		fmt.Println(faint.Sprint("(no complaints filed against this number)"))
		return 0
	}

	fmt.Printf("%d %s, %s through %s\n\n",
		resp.Count, plural(resp.Count, "complaint"),
		shortDate(resp.EarliestDate), shortDate(resp.LatestDate))

	if len(resp.Issues) > 0 {
		fmt.Println(bold.Sprint("Issues:  ") + strings.Join(resp.Issues, ", "))
	}
	if len(resp.States) > 0 {
		fmt.Println(bold.Sprint("States:  ") + strings.Join(resp.States, ", "))
	}
	fmt.Println()

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "DATE\tISSUE\tMETHOD\tCALL TYPE\tSTATE\tZIP"
	if isTTY(os.Stdout) {
		header = bold.Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, h := range resp.Hits {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			dash(shortDate(h.IssueDate)),
			dash(truncate(h.Issue, 24)),
			dash(h.Method),
			dash(truncate(h.CallType, 24)),
			dash(h.State),
			dash(h.Zip),
		)
	}
	tw.Flush()
	return 0
}

func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
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
