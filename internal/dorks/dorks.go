// Package dorks generates Google search URLs to surface a phone number across
// social media, disposable-number providers, reputation sites, people-search,
// and general web. Pure string templating — no auth, no network. Mirrors
// PhoneInfoga's 5-bucket dork taxonomy.
package dorks

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	pn "github.com/nyaruka/phonenumbers"
)

type Dork struct {
	Bucket string `json:"bucket"`
	Site   string `json:"site,omitempty"`
	Query  string `json:"query"`
	URL    string `json:"url"`
}

type Bucket string

const (
	Social      Bucket = "social"
	Disposable  Bucket = "disposable"
	Reputation  Bucket = "reputation"
	Individuals Bucket = "individuals"
	General     Bucket = "general"
)

func AllBuckets() []Bucket {
	return []Bucket{Social, Disposable, Reputation, Individuals, General}
}

var sites = map[Bucket][]string{
	Social: {
		"facebook.com",
		"linkedin.com",
		"twitter.com",
		"x.com",
		"instagram.com",
		"vk.com",
		"reddit.com",
		"github.com",
		"medium.com",
	},
	Disposable: {
		"textnow.com",
		"textfree.us",
		"hushed.com",
		"burnerapp.com",
		"voice.google.com",
		"sideline.com",
	},
	Reputation: {
		"shouldianswer.com",
		"800notes.com",
		"tellows.com",
		"mrnumber.com",
		"whocallsme.com",
		"unknownphone.com",
		"truecaller.com",
		"sync.me",
	},
	Individuals: {
		"whitepages.com",
		"411.com",
		"anywho.com",
		"spokeo.com",
		"beenverified.com",
		"radaris.com",
		"truepeoplesearch.com",
		"fastpeoplesearch.com",
	},
	General: {
		// general bucket gets one Google search per phone variation, no site:
	},
}

// Generate produces all dork URLs for a phone number across the requested
// buckets (or all buckets if none provided). The phone is parsed via
// libphonenumber to derive multiple format variants (E.164, INTERNATIONAL,
// NATIONAL, digits-only) so the resulting Google quoted-string queries hit
// any way the number might appear in a target page.
func Generate(input, defaultRegion string, buckets []Bucket) ([]Dork, error) {
	variations, err := phoneVariations(input, defaultRegion)
	if err != nil {
		return nil, err
	}
	if len(buckets) == 0 {
		buckets = AllBuckets()
	}

	out := make([]Dork, 0, 64)
	for _, b := range buckets {
		switch b {
		case General:
			for _, v := range variations {
				q := `"` + v + `"`
				out = append(out, Dork{
					Bucket: string(General),
					Query:  q,
					URL:    googleSearch(q),
				})
			}
		default:
			for _, site := range sites[b] {
				for _, v := range variations {
					q := fmt.Sprintf(`site:%s "%s"`, site, v)
					out = append(out, Dork{
						Bucket: string(b),
						Site:   site,
						Query:  q,
						URL:    googleSearch(q),
					})
				}
			}
		}
	}
	return out, nil
}

// phoneVariations returns the unique format variants a Google query may need to
// match against pages that mention the number in different forms.
func phoneVariations(input, defaultRegion string) ([]string, error) {
	if defaultRegion == "" && !strings.HasPrefix(input, "+") {
		input = "+" + input
	}
	num, err := pn.Parse(input, defaultRegion)
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", input, err)
	}
	e164 := pn.Format(num, pn.E164)
	intl := pn.Format(num, pn.INTERNATIONAL)
	natl := pn.Format(num, pn.NATIONAL)

	cc := num.GetCountryCode()
	natnum := num.GetNationalNumber()

	digitsOnly := fmt.Sprintf("%d%d", cc, natnum)
	natlDigits := fmt.Sprintf("%d", natnum)

	candidates := []string{
		e164,
		strings.TrimPrefix(e164, "+"),
		intl,
		natl,
		digitsOnly,
		natlDigits,
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		seen[c] = struct{}{}
		out = append(out, c)
	}
	sort.Strings(out)
	return out, nil
}

func googleSearch(q string) string {
	v := url.Values{}
	v.Set("q", q)
	return "https://www.google.com/search?" + v.Encode()
}

func ParseBuckets(csv string) []Bucket {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}
	known := map[string]Bucket{
		"social":      Social,
		"disposable":  Disposable,
		"reputation":  Reputation,
		"individuals": Individuals,
		"general":     General,
	}
	out := []Bucket{}
	seen := map[Bucket]struct{}{}
	for _, p := range strings.Split(csv, ",") {
		key := strings.ToLower(strings.TrimSpace(p))
		b, ok := known[key]
		if !ok {
			continue
		}
		if _, dup := seen[b]; dup {
			continue
		}
		seen[b] = struct{}{}
		out = append(out, b)
	}
	return out
}
