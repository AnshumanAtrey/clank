package whatsapp

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

const usage = `usage: clank whatsapp <subcommand>

subcommands:
  login                 pair with WhatsApp (renders QR — scan with phone)
  lookup <phone>        check if a phone is on WhatsApp, fetch profile info
  logout                unlink the session server-side and delete ~/.clank/whatsapp.db
  reset                 delete ~/.clank/whatsapp.db without contacting WhatsApp
                        (use when the session is already invalid)

flags (lookup):
  --json                JSON output

example:
  clank whatsapp login
  clank whatsapp lookup +14155552671
  clank whatsapp logout

session is persisted in ~/.clank/whatsapp.db (modernc.org/sqlite, no CGo).
phone must stay online and connected within 14 days, or WhatsApp unlinks
all companions and you'll need to re-pair.
`

func Command(args []string) int {
	if len(args) < 1 {
		fmt.Fprint(os.Stderr, usage)
		return 1
	}
	switch args[0] {
	case "login":
		return runLogin()
	case "lookup":
		return runLookup(args[1:])
	case "logout":
		return runLogout()
	case "reset":
		return runReset()
	case "-h", "--help", "help":
		fmt.Print(usage)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n%s", args[0], usage)
		return 1
	}
}

func runLogin() int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	h, err := Open(ctx, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, "login failed:", err)
		return 1
	}
	defer h.Close()
	fmt.Println("WhatsApp paired and connected. Session stored in ~/.clank/whatsapp.db.")
	return 0
}

func runLookup(args []string) int {
	fs := flag.NewFlagSet("whatsapp lookup", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: clank whatsapp lookup <phone>")
		return 1
	}
	phone := fs.Arg(0)
	if !strings.HasPrefix(phone, "+") {
		phone = "+" + phone
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	h, err := Open(ctx, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	defer h.Close()

	res, err := h.Lookup(ctx, phone)
	if err != nil {
		fmt.Fprintln(os.Stderr, "lookup failed:", err)
		return 1
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(res)
		return 0
	}

	fmt.Printf("WhatsApp lookup for %s\n", res.Query)
	if res.Reason != "" {
		fmt.Println(color.YellowString(res.Reason))
	}
	if !res.Registered {
		fmt.Println(color.New(color.Faint).Sprint("Not registered"))
		return 0
	}
	fmt.Println(color.GreenString("Registered ✓"))
	if res.JID != "" {
		fmt.Printf("  jid              : %s\n", res.JID)
	}
	if res.LID != "" {
		fmt.Printf("  lid              : %s\n", res.LID)
	}
	if res.VerifiedBusinessName != "" {
		fmt.Printf("  business name    : %s\n", res.VerifiedBusinessName)
	}
	if res.About != "" {
		fmt.Printf("  about            : %s\n", res.About)
	}
	if res.DeviceCount > 0 {
		fmt.Printf("  device count     : %d\n", res.DeviceCount)
	}
	if res.PictureID != "" {
		fmt.Printf("  picture id       : %s\n", res.PictureID)
	}
	if res.ProfilePictureURL != "" {
		fmt.Printf("  profile picture  : %s\n", res.ProfilePictureURL)
		fmt.Println(color.New(color.Faint).Sprint("    (URL is short-lived — download promptly)"))
	}
	return 0
}

func runLogout() int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	h, err := Open(ctx, false)
	if err != nil {
		// already not paired — fall through to local cleanup
		path, rerr := Reset()
		if rerr != nil {
			fmt.Fprintln(os.Stderr, "logout (local-only):", rerr)
			return 1
		}
		fmt.Println("session was not paired — cleaned local file at", path)
		return 0
	}
	defer h.Close()
	path, lerr := h.Logout(ctx)
	if lerr != nil {
		fmt.Fprintln(os.Stderr, "server logout returned:", lerr)
		fmt.Println("local session deleted anyway:", path)
		return 0
	}
	fmt.Println("unlinked from WhatsApp; deleted", path)
	return 0
}

func runReset() int {
	path, err := Reset()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reset failed:", err)
		return 1
	}
	fmt.Println("removed local session at", path)
	return 0
}
