package imei

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"
)

//go:embed data/tac.csv
var tacRaw []byte

type Device struct {
	TAC          string `json:"tac"`
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
	HWType       string `json:"hw_type,omitempty"`
	OS           string `json:"os,omitempty"`
	Year         string `json:"year,omitempty"`
}

type Result struct {
	Input       string  `json:"input"`
	Normalized  string  `json:"normalized"`
	Length      int     `json:"length"`
	LuhnValid   bool    `json:"luhn_valid"`
	StructureOK bool    `json:"structure_ok"`
	TAC         string  `json:"tac"`
	Serial      string  `json:"serial,omitempty"`
	Checksum    string  `json:"checksum,omitempty"`
	Device      *Device `json:"device,omitempty"`
	ParseError  string  `json:"parse_error,omitempty"`
}

var (
	ErrTooShort = errors.New("imei too short — need 14 (no checksum) or 15 digits")
	ErrTooLong  = errors.New("imei too long — max 15 digits")
	ErrNonDigit = errors.New("imei must contain only digits")
)

var tacIndex map[string]Device

func init() {
	tacIndex = make(map[string]Device, 30_000)
	r := csv.NewReader(bufio.NewReader(bytes.NewReader(tacRaw)))
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		if len(row) < 3 {
			continue
		}
		tac := strings.TrimSpace(row[0])
		if tac == "" {
			continue
		}
		d := Device{TAC: tac}
		d.Manufacturer = get(row, 1)
		d.Model = get(row, 2)
		d.HWType = get(row, 3)
		d.OS = get(row, 4)
		d.Year = get(row, 5)
		tacIndex[tac] = d
	}
}

func get(row []string, i int) string {
	if i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func TACCount() int { return len(tacIndex) }

func Parse(input string) Result {
	r := Result{Input: input}
	digits := strings.Map(func(c rune) rune {
		if c >= '0' && c <= '9' {
			return c
		}
		if c == ' ' || c == '-' || c == '/' {
			return -1
		}
		return c
	}, input)
	for _, c := range digits {
		if c < '0' || c > '9' {
			r.ParseError = ErrNonDigit.Error()
			return r
		}
	}
	r.Normalized = digits
	r.Length = len(digits)

	switch r.Length {
	case 14:
		r.StructureOK = true
		r.TAC = digits[:8]
		r.Serial = digits[8:14]
		r.Checksum = ""
		r.LuhnValid = false
	case 15:
		r.StructureOK = true
		r.TAC = digits[:8]
		r.Serial = digits[8:14]
		r.Checksum = digits[14:15]
		r.LuhnValid = LuhnCheck(digits)
	default:
		if r.Length < 14 {
			r.ParseError = ErrTooShort.Error()
		} else {
			r.ParseError = ErrTooLong.Error()
		}
		return r
	}

	if d, ok := tacIndex[r.TAC]; ok {
		dev := d
		r.Device = &dev
	}
	return r
}

func LuhnCheck(digits string) bool {
	if len(digits) == 0 {
		return false
	}
	sum := 0
	odd := true
	for i := len(digits) - 1; i >= 0; i-- {
		c := int(digits[i] - '0')
		if c < 0 || c > 9 {
			return false
		}
		if odd {
			sum += c
		} else {
			c *= 2
			if c > 9 {
				c -= 9
			}
			sum += c
		}
		odd = !odd
	}
	return sum%10 == 0
}

func ComputeLuhnCheck(digits14 string) (byte, error) {
	if len(digits14) != 14 {
		return 0, fmt.Errorf("need 14 digits, got %d", len(digits14))
	}
	sum := 0
	odd := false
	for i := len(digits14) - 1; i >= 0; i-- {
		c := int(digits14[i] - '0')
		if c < 0 || c > 9 {
			return 0, ErrNonDigit
		}
		if odd {
			sum += c
		} else {
			c *= 2
			if c > 9 {
				c -= 9
			}
			sum += c
		}
		odd = !odd
	}
	check := (10 - sum%10) % 10
	return byte('0' + check), nil
}
