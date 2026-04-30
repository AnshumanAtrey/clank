package scan

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank scan <pattern-or-number> [options]
       clank scan --input <file>  [options]
       cat numbers.txt | clank scan --input -

Run 'clank deep' over every candidate produced by a pattern (or read from a
file / stdin), with rate-limit-aware sleeps so messenger lookups don't get
banned. Filters early via libphonenumber so only valid numbers reach the
expensive enrichment stages. Each result is checkpointed as it completes.

Defaults are deliberately conservative: 30 candidates, 8s sleep between
each, sequential (1 worker). For broad offline scans pass --quick which
disables messengers and is safe to parallelise.

flags:
  --region <ISO>       libphonenumber default region (e.g. IN, US)
  --max <N>            cap candidates AFTER libphonenumber filter (default 30)
  --sample <strategy>  head | stride | random (default head)
  --quick              skip messengers (Telegram, WhatsApp, ignorant)
  --no-apis            skip Tier 2 API providers
  --no-edgar           skip SEC EDGAR
  --sleep <duration>   between candidates (default 8s; 0 to disable)
  --workers <N>        parallel candidates (default 1; only safe with --quick)
  --force              allow patterns producing >100,000 combinations
  --out <path>         JSONL checkpoint (default ~/.clank/scan-<ts>.jsonl)
  --resume             skip phones already in --out file
  --top <N>            rows in summary table (default 20)
  --json               JSON Summary on stdout (machine-readable)
  --input <path>       read phones from file (one per line); use "-" for stdin
  --no-progress        suppress per-candidate progress lines

examples:
  clank scan +918115605xxx --region IN
  clank scan --max 100 --quick +918115605xxx --region IN
  clank scan --input targets.txt --quick --workers 5
  cat numbers.txt | clank scan --input - --json | jq
  clank scan --resume --out scan.jsonl +918115605xxx --region IN
`

func Command(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	region := fs.String("region", "", "default region")
	max := fs.Int("max", DefaultMax, "max candidates after filter")
	sample := fs.String("sample", string(SampleHead), "head|stride|random")
	quick := fs.Bool("quick", false, "skip messengers")
	noAPIs := fs.Bool("no-apis", false, "skip API providers")
	noEdgar := fs.Bool("no-edgar", false, "skip SEC EDGAR")
	sleep := fs.Duration("sleep", DefaultSleep, "sleep between candidates")
	workers := fs.Int("workers", DefaultWorkers, "concurrent candidates")
	force := fs.Bool("force", false, "allow >100k combinations")
	out := fs.String("out", "", "checkpoint JSONL path")
	resume := fs.Bool("resume", false, "skip phones already in --out file")
	top := fs.Int("top", DefaultTopN, "summary table size")
	jsonOut := fs.Bool("json", false, "JSON Summary on stdout")
	input := fs.String("input", "", "phones file ('-' for stdin)")
	noProgress := fs.Bool("no-progress", false, "suppress progress lines")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return 2
	}

	// --quick disables sleep unless the user explicitly set --sleep.
	sleepWasSet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "sleep" {
			sleepWasSet = true
		}
	})
	effectiveSleep := *sleep
	if *quick && !sleepWasSet {
		effectiveSleep = 0
	}

	opts := Options{
		Region:    *region,
		Max:       *max,
		Sample:    Sample(*sample),
		Quick:     *quick,
		SkipAPIs:  *noAPIs,
		SkipEdgar: *noEdgar,
		Sleep:     effectiveSleep,
		Workers:   *workers,
		Force:     *force,
		Out:       *out,
		Resume:    *resume,
	}
	if !*noProgress {
		opts.Progress = os.Stderr
	}

	if *input != "" {
		phones, err := ReadInput(*input)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		opts.Phones = phones
	} else {
		if fs.NArg() < 1 {
			fs.Usage()
			return 1
		}
		opts.Pattern = fs.Arg(0)
	}

	if opts.Workers > 1 && !opts.Quick {
		fmt.Fprintln(os.Stderr, color.YellowString(
			"warning: --workers > 1 without --quick will trigger messenger bans fast. consider --quick or --workers 1."))
	}

	ctx, cancel := signalCtx()
	defer cancel()

	sum, err := Run(ctx, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(sum)
		return 0
	}
	render(sum, *top)
	return 0
}

func signalCtx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Fprintln(os.Stderr, "\ninterrupted — flushing checkpoint and exiting cleanly")
		cancel()
	}()
	return ctx, cancel
}

func render(sum *Summary, top int) {
	bold := color.New(color.Bold)
	faint := color.New(color.Faint)

	fmt.Println()
	fmt.Println(bold.Sprintf("Scan summary"))
	fmt.Printf("  pattern         : %s\n", dash(sum.Pattern))
	fmt.Printf("  total candidates: %d\n", sum.TotalCandidates)
	fmt.Printf("  valid           : %d\n", sum.ValidAfterParse)
	fmt.Printf("  sampled to      : %d\n", sum.SampledTo)
	if sum.SkippedResume > 0 {
		fmt.Printf("  resumed (skipped already-done): %d\n", sum.SkippedResume)
	}
	fmt.Printf("  processed       : %d\n", sum.Processed)
	fmt.Printf("  took            : %s\n", sum.TookHuman)
	if sum.OutFile != "" {
		fmt.Printf("  checkpoint      : %s\n", sum.OutFile)
	}

	if len(sum.Results) == 0 {
		fmt.Println("\n(no results)")
		return
	}

	fmt.Println()
	limit := top
	if limit <= 0 || limit > len(sum.Results) {
		limit = len(sum.Results)
	}
	fmt.Println(bold.Sprintf("Top %d (by enrichment hits)", limit))

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	header := "RANK\tPHONE\tSCORE\tWHATSAPP\tTELEGRAM\tIGNORANT\tEDGAR\tAPIS\tNOTES"
	if isTTY(os.Stdout) {
		header = bold.Sprint(header)
	}
	fmt.Fprintln(tw, header)
	for i, r := range sum.Results[:limit] {
		fmt.Fprintf(tw, "%d\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\n",
			i+1,
			r.Phone,
			r.Score,
			waCol(r),
			tgCol(r),
			igCol(r),
			edgarCol(r),
			apisCol(r),
			notesCol(r, faint),
		)
	}
	tw.Flush()

	if len(sum.Results) > limit {
		fmt.Printf("\n%d more rows in checkpoint file. Re-run with --top %d to expand.\n",
			len(sum.Results)-limit, len(sum.Results))
	}
}

func waCol(r *Result) string {
	if r.Deep == nil || r.Deep.WhatsApp == nil || r.Deep.WhatsApp.Result == nil {
		return "-"
	}
	wa := r.Deep.WhatsApp.Result
	if !wa.Registered {
		return "no"
	}
	parts := []string{"yes"}
	if wa.VerifiedBusinessName != "" {
		parts = append(parts, "biz:"+truncate(wa.VerifiedBusinessName, 18))
	}
	return strings.Join(parts, " ")
}

func tgCol(r *Result) string {
	if r.Deep == nil || r.Deep.Telegram == nil || r.Deep.Telegram.Result == nil {
		return "-"
	}
	tg := r.Deep.Telegram.Result
	if !tg.Found {
		return "no"
	}
	if tg.Username != "" {
		return "@" + tg.Username
	}
	if tg.FirstName != "" || tg.LastName != "" {
		return strings.TrimSpace(tg.FirstName + " " + tg.LastName)
	}
	return fmt.Sprintf("id=%d", tg.UserID)
}

func igCol(r *Result) string {
	if r.Deep == nil || r.Deep.Ignorant == nil || len(r.Deep.Ignorant.Results) == 0 {
		return "-"
	}
	hits := []string{}
	for _, ig := range r.Deep.Ignorant.Results {
		if ig.Exists {
			hits = append(hits, ig.Name)
		}
	}
	if len(hits) == 0 {
		return "no"
	}
	return strings.Join(hits, ",")
}

func edgarCol(r *Result) string {
	if r.Deep == nil || r.Deep.Edgar == nil || r.Deep.Edgar.Result == nil {
		return "-"
	}
	if r.Deep.Edgar.Result.Total == 0 {
		return "0"
	}
	return fmt.Sprintf("%d", r.Deep.Edgar.Result.Total)
}

func apisCol(r *Result) string {
	if r.Deep == nil || len(r.Deep.APIs) == 0 {
		return "-"
	}
	hits := []string{}
	for _, a := range r.Deep.APIs {
		if a == nil || a.Result == nil || !a.Result.Valid {
			continue
		}
		if a.Result.FraudScore != nil {
			hits = append(hits, fmt.Sprintf("%s:%d", a.Provider, *a.Result.FraudScore))
		} else {
			hits = append(hits, a.Provider)
		}
	}
	if len(hits) == 0 {
		return "-"
	}
	return strings.Join(hits, ",")
}

func notesCol(r *Result, faint *color.Color) string {
	if r.Deep == nil || r.Deep.Local == nil {
		return "-"
	}
	parts := []string{}
	if r.Deep.Local.Spam {
		parts = append(parts, "SPAM")
	}
	if r.Deep.Local.Carrier != "" {
		parts = append(parts, r.Deep.Local.Carrier)
	}
	if r.Deep.Local.Region != "" {
		parts = append(parts, r.Deep.Local.Region)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "/")
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

// _ keeps tabwriter import used even if rendering refactor removes inline usage.
var _ = tabwriter.NewWriter

// _ keeps time used (Sleep references it but if Run inlines the constant later we still want it imported here).
var _ = time.Second
