package imei

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/fatih/color"
)

func Command(args []string) int {
	fs := flag.NewFlagSet("imei", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "JSON output")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: clank imei <14-or-15-digit IMEI> [--json]")
		fmt.Fprintln(os.Stderr, "       clank imei --help")
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return 1
	}

	r := Parse(fs.Arg(0))

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(r)
		return 0
	}

	if r.ParseError != "" {
		fmt.Fprintln(os.Stderr, "error:", r.ParseError)
		return 1
	}

	fmt.Printf("Input        : %s\n", r.Input)
	fmt.Printf("Normalized   : %s (%d digits)\n", r.Normalized, r.Length)
	fmt.Printf("TAC          : %s\n", r.TAC)
	fmt.Printf("Serial       : %s\n", r.Serial)
	if r.Length == 15 {
		luhn := "valid"
		if !r.LuhnValid {
			luhn = color.RedString("invalid")
			if c, err := ComputeLuhnCheck(r.Normalized[:14]); err == nil {
				luhn += fmt.Sprintf(" (expected check digit: %c, got %s)", c, r.Checksum)
			}
		} else {
			luhn = color.GreenString(luhn)
		}
		fmt.Printf("Checksum     : %s  [%s]\n", r.Checksum, luhn)
	} else {
		if c, err := ComputeLuhnCheck(r.Normalized); err == nil {
			fmt.Printf("Checksum     : (none — 14-digit input. Computed Luhn: %c)\n", c)
		}
	}

	if r.Device != nil {
		fmt.Println()
		fmt.Println(color.New(color.Bold).Sprint("Device:"))
		printIfSet("  Manufacturer", r.Device.Manufacturer)
		printIfSet("  Model       ", r.Device.Model)
		printIfSet("  HW Type     ", r.Device.HWType)
		printIfSet("  OS          ", r.Device.OS)
		printIfSet("  Year        ", r.Device.Year)
	} else {
		fmt.Println()
		fmt.Println(color.YellowString("Device: TAC %s not in embedded database (Sep 2014 snapshot — newer devices unavailable).", r.TAC))
	}
	return 0
}

func printIfSet(label, value string) {
	if value != "" {
		fmt.Printf("%s: %s\n", label, value)
	}
}
