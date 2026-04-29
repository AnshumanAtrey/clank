package telegram

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank telegram <subcommand>

subcommands:
  login                authenticate with Telegram (saves session to ~/.clank/telegram.session)
  lookup <phone>       resolve a phone number to a Telegram user
  logout               remove the saved session

env vars (required for login + lookup):
  TG_APP_ID            from https://my.telegram.org/apps
  TG_APP_HASH          same page

flags (lookup):
  --json               machine-readable output

example:
  export TG_APP_ID=12345
  export TG_APP_HASH=abc...
  clank telegram login
  clank telegram lookup +14155552671
`

func Command(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}
	switch args[0] {
	case "login":
		return runLogin(args[1:])
	case "lookup":
		return runLookup(args[1:])
	case "logout":
		return runLogout(args[1:])
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n%s", args[0], usage)
		return 1
	}
}

func runLogin(_ []string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if err := Login(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "login failed:", err)
		return 1
	}
	return 0
}

func runLogout(_ []string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := Logout(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "logout failed:", err)
		return 1
	}
	return 0
}

func runLookup(args []string) int {
	fs := flag.NewFlagSet("telegram lookup", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: clank telegram lookup <phone>")
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := ResolvePhone(ctx, fs.Arg(0))
	if err != nil && res.Reason == "" {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}

	fmt.Printf("Telegram lookup for %s\n", res.Phone)
	if !res.Found {
		fmt.Println(color.YellowString("Not registered or hidden by privacy"))
		if res.Reason != "" {
			fmt.Println("  reason:", res.Reason)
		}
		return 0
	}
	fmt.Println(color.GreenString("Registered ✓"))
	fmt.Printf("  user id    : %d\n", res.UserID)
	if res.FirstName != "" || res.LastName != "" {
		fmt.Printf("  name       : %s %s\n", res.FirstName, res.LastName)
	}
	if res.Username != "" {
		fmt.Printf("  username   : @%s\n", res.Username)
	}
	if res.HasPhoto {
		fmt.Println("  photo      : yes")
	}
	if res.Premium {
		fmt.Println("  premium    : yes")
	}
	if res.Verified {
		fmt.Println("  verified   : yes")
	}
	if res.Bot {
		fmt.Println("  bot        : yes")
	}
	if res.Restricted {
		fmt.Println("  restricted : yes")
	}
	return 0
}
