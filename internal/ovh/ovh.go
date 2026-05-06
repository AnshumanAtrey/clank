// Package ovh queries OVH's free, unauthenticated telephony zones API to
// resolve a phone number to its number-range zone (city, ZIP, geographic /
// non-geographic type). Coverage: FR, BE, CH, CZ, DE, ES, FI, IT, NL, PL,
// PT, TN, UK — the European countries where OVH operates landline DID blocks.
//
// The API returns one row per allocated 10k-block (number pattern like
// "036517xxxx"). We pull the full per-country list (typically 200-500 rows),
// then do longest-prefix match against the input number's E.164 form sans
// "+". Mobile prefixes are not in the dataset (06/07 in FR, 07 in UK, etc.)
// so the source effectively complements libphonenumber's coarse geocoder
// for landlines but adds nothing for mobiles — that's fine; mobile-flagged
// inputs return Found=false rather than an error.
package ovh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	pn "github.com/nyaruka/phonenumbers"
)

const (
	endpoint        = "https://api.ovh.com/1.0/telephony/number/detailedZones"
	defaultUA       = "clank-osint build@atrey.dev"
	envUserAgentVar = "CLANK_OVH_UA"
)

// supportedCountries maps libphonenumber region codes to OVH country params.
// libphonenumber emits "GB" for the UK; OVH uses "uk".
var supportedCountries = map[string]string{
	"FR": "fr", "BE": "be", "CH": "ch", "CZ": "cz", "DE": "de",
	"ES": "es", "FI": "fi", "IT": "it", "NL": "nl", "PL": "pl",
	"PT": "pt", "TN": "tn", "GB": "uk", "UK": "uk",
}

// ErrUnsupportedRegion indicates the input phone's region is not one of
// OVH's covered countries.
var ErrUnsupportedRegion = errors.New("OVH covers FR/BE/CH/CZ/DE/ES/FI/IT/NL/PL/PT/TN/UK only")

// Zone is one row from the OVH detailedZones response. JSON tags match the
// OVH API verbatim (camelCase) — keeping snake_case here would silently fail
// on live data while round-tripping fine through tests.
type Zone struct {
	City                string   `json:"city,omitempty"`
	ZipCode             string   `json:"zipCode,omitempty"`
	Country             string   `json:"country,omitempty"`
	Prefix              int      `json:"prefix,omitempty"`
	Number              string   `json:"number,omitempty"`
	InternationalNumber string   `json:"internationalNumber,omitempty"`
	Type                string   `json:"type,omitempty"`
	ZneList             []string `json:"zneList,omitempty"`
}

// Response is the per-number summary clank's deep orchestrator consumes.
type Response struct {
	Query  string `json:"query"`
	Region string `json:"region"`
	Found  bool   `json:"found"`
	City   string `json:"city,omitempty"`
	Zip    string `json:"zip,omitempty"`
	Zone   *Zone  `json:"zone,omitempty"`
}

// Options configures Lookup.
type Options struct {
	Endpoint string // override for tests; defaults to live OVH URL
}

// Lookup fetches the per-country zone list (cached for the life of the
// process) and returns the longest-prefix match against the input.
func Lookup(ctx context.Context, e164 string, opts Options) (*Response, error) {
	num, err := pn.Parse(e164, "")
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", e164, err)
	}
	region := pn.GetRegionCodeForNumber(num)
	ovhCountry, ok := supportedCountries[region]
	if !ok {
		return nil, ErrUnsupportedRegion
	}

	ep := opts.Endpoint
	if ep == "" {
		ep = endpoint
	}

	zones, err := zonesFor(ctx, ep, ovhCountry)
	if err != nil {
		return nil, err
	}

	// E.164 without "+" — matches OVH's `internationalNumber` after stripping
	// its leading "00" IDD.
	target := pn.Format(num, pn.E164)
	target = strings.TrimPrefix(target, "+")

	best := bestZoneMatch(zones, target)
	out := &Response{Query: e164, Region: region}
	if best == nil {
		return out, nil
	}
	out.Found = true
	out.City = best.City
	out.Zip = best.ZipCode
	out.Zone = best
	return out, nil
}

// bestZoneMatch picks the zone whose internationalNumber's non-`x` prefix is
// the longest one that prefixes the target (E.164 without the leading +).
func bestZoneMatch(zones []Zone, target string) *Zone {
	var best *Zone
	bestLen := -1
	for i := range zones {
		z := &zones[i]
		// internationalNumber pattern is "<00><cc><nat>xxxx". Strip the IDD
		// prefix and trailing x's to get the matchable prefix.
		pat := strings.TrimPrefix(z.InternationalNumber, "00")
		pat = strings.TrimRight(pat, "xX")
		if pat == "" {
			continue
		}
		if !strings.HasPrefix(target, pat) {
			continue
		}
		if len(pat) > bestLen {
			best = z
			bestLen = len(pat)
		}
	}
	return best
}

// zoneCache memoizes per-country zones for the life of the process. The
// build plan calls for a global per-source disk cache in v0.2.0-rc1's
// transport layer; until that lands, an in-process map keeps repeat lookups
// in the same `clank deep` invocation cheap.
var (
	zoneCacheMu sync.Mutex
	zoneCache   = map[string][]Zone{}
)

// resetCache is exposed for tests so each test starts from a clean cache.
func resetCache() {
	zoneCacheMu.Lock()
	zoneCache = map[string][]Zone{}
	zoneCacheMu.Unlock()
}

func zonesFor(ctx context.Context, ep, country string) ([]Zone, error) {
	zoneCacheMu.Lock()
	if cached, ok := zoneCache[ep+"|"+country]; ok {
		zoneCacheMu.Unlock()
		return cached, nil
	}
	zoneCacheMu.Unlock()

	q := url.Values{}
	q.Set("country", country)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent())

	hc := &http.Client{Timeout: 15 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ovh request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("ovh http %d: %s", resp.StatusCode, snippet(body))
	}

	var zones []Zone
	if err := json.Unmarshal(body, &zones); err != nil {
		return nil, fmt.Errorf("ovh decode: %w", err)
	}

	// Pre-sort by descending matchable-prefix length so callers iterating
	// once already see the longest-match candidates first. (We still rely on
	// bestZoneMatch's bestLen accounting for correctness.)
	sort.Slice(zones, func(i, j int) bool {
		return matchableLen(zones[i].InternationalNumber) > matchableLen(zones[j].InternationalNumber)
	})

	zoneCacheMu.Lock()
	zoneCache[ep+"|"+country] = zones
	zoneCacheMu.Unlock()
	return zones, nil
}

func matchableLen(internationalNumber string) int {
	pat := strings.TrimPrefix(internationalNumber, "00")
	pat = strings.TrimRight(pat, "xX")
	return len(pat)
}

func userAgent() string {
	if v := strings.TrimSpace(os.Getenv(envUserAgentVar)); v != "" {
		return v
	}
	return defaultUA
}

func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}
