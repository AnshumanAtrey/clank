package dorks

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
)

const usage = `usage: clank dorks <phone> [options]

Generate Google-dork URLs for a phone number across 5 buckets:
  social        Facebook, LinkedIn, X/Twitter, Instagram, VK, Reddit, etc.
  disposable    TextNow, Hushed, Google Voice, Burner, Sideline
  reputation    shouldianswer.com, 800notes, tellows, mrnumber, etc.
  individuals   Whitepages, Spokeo, BeenVerified, Radaris, etc.
  general       quoted-number Google searches (no site: filter)

Each bucket × phone-format variant produces one URL. Pure string templating
— no auth, no API. Open them in a browser to see actual hits.

flags:
  --bucket <csv>     restrict to specific buckets (e.g. --bucket social,reputation)
  --region <ISO>     default region for parsing (e.g. IN, US)
  --open             open URLs in default browser (capped at 30 to avoid flooding)
  --max <N>          when --open is set, max URLs to launch (default 30)
  --json             JSON output

examples:
  clank dorks +14155552671                              # print all URLs
  clank dorks --bucket social,reputation +14155552671   # subset
  clank dorks --open --max 10 +14155552671              # browser dump, capped
  clank dorks --json +14155552671 | jq                  # structured
`

func Command(args []string) int {
	fs := flag.NewFlagSet("dorks", flag.ExitOnError)
	bucketCSV := fs.String("bucket", "", "comma-separated buckets")
	region := fs.String("region", "", "default region")
	openInBrowser := fs.Bool("open", false, "open in default browser")
	maxOpen := fs.Int("max", 30, "max URLs to open in browser")
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

	dorks, err := Generate(phone, *region, ParseBuckets(*bucketCSV))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(dorks)
		return 0
	}

	if *openInBrowser {
		return openBrowser(dorks, *maxOpen)
	}

	return printDorks(dorks)
}

func printDorks(dorks []Dork) int {
	bold := color.New(color.Bold)
	faint := color.New(color.Faint)

	currentBucket := ""
	for _, d := range dorks {
		if d.Bucket != currentBucket {
			fmt.Println()
			fmt.Println(bold.Sprintf("[%s]", d.Bucket))
			currentBucket = d.Bucket
		}
		if d.Site != "" {
			fmt.Printf("  %s\n", faint.Sprintf("%s — %s", d.Site, d.Query))
		} else {
			fmt.Printf("  %s\n", faint.Sprint(d.Query))
		}
		fmt.Printf("    %s\n", d.URL)
	}
	fmt.Println()
	fmt.Printf("%d URLs across %d buckets — pipe to | xargs open (macOS) or | xargs xdg-open (Linux) to launch all.\n",
		len(dorks), countBuckets(dorks))
	return 0
}

func openBrowser(dorks []Dork, max int) int {
	if max < 1 {
		max = 1
	}
	if len(dorks) > max {
		fmt.Fprintf(os.Stderr, "capping at %d URLs (use --max to override; full count was %d)\n", max, len(dorks))
		dorks = dorks[:max]
	}
	for _, d := range dorks {
		if err := openOne(d.URL); err != nil {
			fmt.Fprintln(os.Stderr, "open:", err)
		}
	}
	fmt.Printf("opened %d URLs in your default browser\n", len(dorks))
	return 0
}

func openOne(u string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", u).Start()
	case "linux", "freebsd", "openbsd", "netbsd":
		return exec.Command("xdg-open", u).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	default:
		return fmt.Errorf("unsupported platform %s — open URLs manually", runtime.GOOS)
	}
}

func countBuckets(dorks []Dork) int {
	seen := map[string]struct{}{}
	for _, d := range dorks {
		seen[d.Bucket] = struct{}{}
	}
	return len(seen)
}
