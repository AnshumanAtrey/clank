package ignorant

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"
	"time"

	pn "github.com/nyaruka/phonenumbers"
)

type Result struct {
	Name              string `json:"name"`
	Domain            string `json:"domain"`
	Method            string `json:"method"`
	FrequentRateLimit bool   `json:"frequent_rate_limit"`
	RateLimit         bool   `json:"rate_limit"`
	Exists            bool   `json:"exists"`
	Error             string `json:"error,omitempty"`
}

type parsed struct {
	intCC       int
	regionISO   string
	nationalNum string
	cc4Phone    string
}

type checker func(ctx context.Context, hc *http.Client, p parsed) Result

var registry = map[string]checker{
	"instagram": instagramCheck,
	"snapchat":  snapchatCheck,
	"amazon":    amazonCheck,
}

func parsePhone(input string) (parsed, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return parsed{}, errors.New("empty phone")
	}
	if !strings.HasPrefix(s, "+") {
		s = "+" + s
	}
	num, err := pn.Parse(s, "")
	if err != nil {
		return parsed{}, fmt.Errorf("parse %q: %w", input, err)
	}
	if !pn.IsValidNumber(num) {
		return parsed{}, fmt.Errorf("not a valid phone number: %s", input)
	}
	return parsed{
		intCC:       int(num.GetCountryCode()),
		regionISO:   pn.GetRegionCodeForNumber(num),
		nationalNum: fmt.Sprintf("%d", num.GetNationalNumber()),
		cc4Phone:    fmt.Sprintf("%d%d", num.GetCountryCode(), num.GetNationalNumber()),
	}, nil
}

func newClient(timeout time.Duration) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Timeout: timeout,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
}

func Run(ctx context.Context, phone string, only []string) ([]Result, error) {
	p, err := parsePhone(phone)
	if err != nil {
		return nil, err
	}

	picks := pickModules(only)
	if len(picks) == 0 {
		return nil, fmt.Errorf("no matching modules in %v (available: instagram, snapchat, amazon)", only)
	}

	results := make([]Result, len(picks))
	var wg sync.WaitGroup
	for i, name := range picks {
		i, name := i, name
		fn := registry[name]
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = fn(ctx, newClient(15*time.Second), p)
		}()
	}
	wg.Wait()
	return results, nil
}

func pickModules(only []string) []string {
	all := []string{"instagram", "snapchat", "amazon"}
	if len(only) == 0 {
		return all
	}
	want := map[string]bool{}
	for _, o := range only {
		want[strings.ToLower(strings.TrimSpace(o))] = true
	}
	var out []string
	for _, n := range all {
		if want[n] {
			out = append(out, n)
		}
	}
	return out
}

var chromeUA = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
}

func randomUA() string {
	return chromeUA[rand.IntN(len(chromeUA))]
}
