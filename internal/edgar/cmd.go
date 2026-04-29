package edgar

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

func Command(args []string) int {
	fs := flag.NewFlagSet("edgar", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	formsArg := fs.String("forms", "", "comma-separated form types (e.g. 10-K,10-Q,8-K,D)")
	hits := fs.Int("hits", 10, "max hits to fetch (1-100)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: clank edgar <number-or-string> [--forms 10-K,8-K] [--hits 10] [--json]")
		fmt.Fprintln(os.Stderr, "       searches SEC EDGAR full-text for the given number/string")
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}

	query := strings.Join(fs.Args(), " ")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := Search(ctx, query, Options{Forms: QueryForms(*formsArg), Hits: *hits})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(resp)
		return 0
	}

	fmt.Printf("EDGAR full-text search for %q", query)
	if resp.Forms != "" {
		fmt.Printf("  (forms: %s)", resp.Forms)
	}
	fmt.Printf("\n%d total hits, showing %d.\n\n", resp.Total, resp.Returned)
	if resp.Returned == 0 {
		fmt.Println("(no filings reference this query — try broader formatting or remove --forms)")
		return 0
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "DATE\tFORM\tCOMPANY\tLOCATION\tACCESSION"
	if isTTY(os.Stdout) {
		header = color.New(color.Bold).Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for _, h := range resp.Hits {
		name := joinNonEmpty(h.DisplayNames, " | ", 60)
		loc := joinNonEmpty(h.BizLocations, ", ", 30)
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			dash(h.FileDate), dash(h.Form), name, dash(loc), dash(h.Accession))
	}
	tw.Flush()

	if isTTY(os.Stdout) {
		fmt.Println()
		fmt.Println(color.New(color.Faint).Sprint("Filing URLs:"))
		for _, h := range resp.Hits {
			if h.URL != "" {
				fmt.Printf("  %s  %s\n", h.Accession, h.URL)
			}
		}
	}
	return 0
}

func joinNonEmpty(xs []string, sep string, max int) string {
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	s := strings.Join(out, sep)
	if max > 0 && len(s) > max {
		s = s[:max-1] + "…"
	}
	return s
}

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func isTTY(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
