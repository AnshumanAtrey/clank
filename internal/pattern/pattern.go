package pattern

import (
	"errors"
	"fmt"
	"strings"
)

const MaxCombinations = 100_000

var (
	ErrInvalid   = errors.New("pattern must contain only digits and 'x' placeholders")
	ErrEmpty     = errors.New("pattern cannot be empty")
	ErrTooLarge  = errors.New("pattern would generate too many combinations (max 100,000) — use --force to override")
	ErrNoCountry = errors.New("pattern has no leading digits — at least one known digit required")
)

func IsValid(pattern string) bool {
	if pattern == "" {
		return false
	}
	body := strings.TrimPrefix(pattern, "+")
	if body == "" {
		return false
	}
	for _, c := range body {
		if c != 'x' && c != 'X' && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

func CountCombinations(pattern string) int {
	placeholders := strings.Count(strings.ToLower(pattern), "x")
	total := 1
	for i := 0; i < placeholders; i++ {
		total *= 10
	}
	return total
}

func Generate(pattern string, force bool) ([]string, error) {
	if pattern == "" {
		return nil, ErrEmpty
	}
	if !IsValid(pattern) {
		return nil, ErrInvalid
	}
	prefix := ""
	body := pattern
	if strings.HasPrefix(body, "+") {
		prefix = "+"
		body = body[1:]
	}
	body = strings.ToLower(body)
	placeholders := strings.Count(body, "x")

	if placeholders == 0 {
		return []string{prefix + body}, nil
	}

	total := CountCombinations(body)
	if total > MaxCombinations && !force {
		return nil, fmt.Errorf("%w (would generate %d)", ErrTooLarge, total)
	}

	out := make([]string, 0, total)
	for i := 0; i < total; i++ {
		fill := fmt.Sprintf("%0*d", placeholders, i)
		s := body
		for _, d := range fill {
			s = strings.Replace(s, "x", string(d), 1)
		}
		out = append(out, prefix+s)
	}
	return out, nil
}
