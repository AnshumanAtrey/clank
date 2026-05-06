package github

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank github <phone> [--limit 10] [--json]

Search GitHub commit messages for the given phone number across multiple
format variations (E.164, national, hyphenated, digits-only). Useful when
the target leaked their number in a public repo's commit log — happens
more often than you'd think.

flags:
  --limit <N>        max hits to return (1-100; default 10)
  --json             JSON output

env:
  GITHUB_TOKEN       optional PAT (raises rate limit from 10/min unauth
                     to 30/min authenticated; classic or fine-grained both
                     work, no scopes required for public commit search).

examples:
  clank github +14155552671
  clank github --json +14155552671 | jq
`

func Command(args []string) int {
	fs := flag.NewFlagSet("github", flag.ExitOnError)
	limit := fs.Int("limit", 10, "max hits to fetch")
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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, err := Search(ctx, phone, Options{Limit: *limit})
	if err != nil {
		if errors.Is(err, ErrRateLimited) {
			fmt.Fprintln(os.Stderr, "error:", err)
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

	fmt.Printf("GitHub commit search for %s\n", resp.Query)
	fmt.Println(faint.Sprint("queried variants: " + joinSep(resp.Variants, " · ")))
	if resp.Returned == 0 {
		fmt.Println(faint.Sprint("(no commit messages reference this number)"))
		return 0
	}

	fmt.Printf("%d unique commit(s)  · GitHub reported %d total\n\n", resp.Returned, resp.Total)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "DATE\tREPO\tAUTHOR\tMESSAGE"
	if isTTY(os.Stdout) {
		header = bold.Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, h := range resp.Hits {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			dash(shortDate(h.CommitDate)),
			dash(truncate(h.Repo, 32)),
			dash(authorLabel(h)),
			dash(truncate(h.MessageHead, 64)),
		)
	}
	tw.Flush()

	if isTTY(os.Stdout) {
		fmt.Println()
		fmt.Println(faint.Sprint("Commit URLs:"))
		for _, h := range resp.Hits {
			if h.URL != "" {
				fmt.Printf("  %s\n", h.URL)
			}
		}
	}
	return 0
}

func authorLabel(h Hit) string {
	if h.AuthorLogin != "" {
		return "@" + h.AuthorLogin
	}
	if h.AuthorEmail != "" {
		return h.AuthorEmail
	}
	return h.AuthorName
}

func shortDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func joinSep(xs []string, sep string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += sep
		}
		out += x
	}
	return out
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
