package audit

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
)

const usage = `usage: clank history [options]

Show what you've looked up recently. Each clank invocation appends one
line to ~/.clank/history.jsonl (phone-shaped first arg only — never API
keys or full args).

flags:
  --tail <N>     show the most recent N entries (default 20)
  --grep <s>     filter by substring (matches Cmd or Phone)
  --json         JSON output
  --path         print history file path and exit
  --clear        delete the history file (with confirmation)

examples:
  clank history
  clank history --tail 50
  clank history --grep +91
  clank history --grep telegram
  clank history --json | jq

set CLANK_NO_AUDIT=1 in your environment to disable history logging.
`

func Command(args []string) int {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	tail := fs.Int("tail", 20, "show last N entries")
	grep := fs.String("grep", "", "filter substring")
	jsonOut := fs.Bool("json", false, "JSON output")
	pathOnly := fs.Bool("path", false, "print path and exit")
	clear := fs.Bool("clear", false, "delete the history file")
	fs.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *pathOnly {
		fmt.Println(Path())
		return 0
	}
	if *clear {
		fmt.Print("delete the entire clank history? [y/N] ")
		var ans string
		_, _ = fmt.Scanln(&ans)
		if ans != "y" && ans != "yes" && ans != "Y" {
			fmt.Println("cancelled")
			return 0
		}
		if err := os.Remove(Path()); err != nil && !os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
		fmt.Println("history cleared")
		return 0
	}

	entries, err := Read(*tail, *grep)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	if len(entries) == 0 {
		fmt.Println("(no history)")
		return 0
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(entries)
		return 0
	}

	bold := color.New(color.Bold)
	fmt.Println(bold.Sprintf("%-19s  %-10s  %-25s  %-7s  %s",
		"WHEN", "CMD", "PHONE", "TOOK", "STATUS"))
	for _, e := range entries {
		fmt.Println(FormatLine(e))
	}
	return 0
}
