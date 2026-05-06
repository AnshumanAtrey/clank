// Package fcc queries the FCC's public "Consumer Complaints — Unwanted Calls"
// dataset (Socrata SODA endpoint, public domain, nightly updates, ~5M rows
// since 2014). No auth required for low-volume queries; an optional
// FCC_APP_TOKEN env var raises the per-hour rate ceiling. US-only by data —
// non-NANP inputs return ErrNonUS and the deep orchestrator surfaces them as
// inline skips.
package fcc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	pn "github.com/nyaruka/phonenumbers"
)

const (
	endpoint        = "https://opendata.fcc.gov/resource/vakf-fz8e.json"
	defaultUA       = "clank-osint build@atrey.dev"
	envAppTokenVar  = "FCC_APP_TOKEN"
	envUserAgentVar = "CLANK_FCC_UA"
)

// ErrNonUS is returned when the input cannot be normalized to a NANP (+1)
// 10-digit number — the FCC dataset only stores US complaints.
var ErrNonUS = fmt.Errorf("FCC dataset covers US numbers only")

// Complaint is one row from the dataset, normalized.
type Complaint struct {
	ID            string `json:"id"`
	IssueDate     string `json:"issue_date,omitempty"`
	IssueTime     string `json:"issue_time,omitempty"`
	IssueType     string `json:"issue_type,omitempty"`
	Method        string `json:"method,omitempty"`
	Issue         string `json:"issue,omitempty"`
	CallerID      string `json:"caller_id_number,omitempty"`
	AdvertiserNum string `json:"advertiser_business_phone_number,omitempty"`
	// CallType maps to the dataset's `type_of_call_or_messge` field — the typo
	// is in the upstream schema, retained verbatim so consumers diagnosing JSON
	// payloads see the same key.
	CallType string `json:"type_of_call_or_messge,omitempty"`
	State    string `json:"state,omitempty"`
	Zip      string `json:"zip,omitempty"`
}

// Response summarizes FCC complaints for a phone number.
type Response struct {
	Query        string      `json:"query"`
	Count        int         `json:"count"`
	Issues       []string    `json:"issues,omitempty"`
	States       []string    `json:"states,omitempty"`
	LatestDate   string      `json:"latest_date,omitempty"`
	EarliestDate string      `json:"earliest_date,omitempty"`
	Hits         []Complaint `json:"hits,omitempty"`
}

// Options configures Search.
type Options struct {
	Limit    int
	Endpoint string // override for tests; defaults to the live Socrata URL
}

// Search queries the FCC's Socrata endpoint for unwanted-call complaints
// matching the given phone number.
func Search(ctx context.Context, e164 string, opts Options) (*Response, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	ep := opts.Endpoint
	if ep == "" {
		ep = endpoint
	}

	hyphenated, err := toUSHyphenated(e164)
	if err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Set("caller_id_number", hyphenated)
	q.Set("$order", "issue_date DESC")
	q.Set("$limit", fmt.Sprintf("%d", limit))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep+"?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "application/json")
	if token := strings.TrimSpace(os.Getenv(envAppTokenVar)); token != "" {
		req.Header.Set("X-App-Token", token)
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fcc request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("fcc http %d: %s", resp.StatusCode, snippet(body))
	}

	var rows []Complaint
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, fmt.Errorf("fcc decode: %w", err)
	}

	return aggregate(hyphenated, rows), nil
}

// toUSHyphenated converts an E.164 / national-format US input to the
// AAA-NNN-NNNN string the FCC dataset stores.
func toUSHyphenated(input string) (string, error) {
	num, err := pn.Parse(input, "US")
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", input, err)
	}
	if num.GetCountryCode() != 1 {
		return "", ErrNonUS
	}
	nat := fmt.Sprintf("%d", num.GetNationalNumber())
	if len(nat) != 10 {
		return "", ErrNonUS
	}
	return nat[:3] + "-" + nat[3:6] + "-" + nat[6:], nil
}

func aggregate(query string, rows []Complaint) *Response {
	issues := map[string]struct{}{}
	states := map[string]struct{}{}
	var latest, earliest string
	for _, r := range rows {
		if r.Issue != "" {
			issues[r.Issue] = struct{}{}
		}
		if r.State != "" {
			states[r.State] = struct{}{}
		}
		if r.IssueDate != "" {
			if latest == "" || r.IssueDate > latest {
				latest = r.IssueDate
			}
			if earliest == "" || r.IssueDate < earliest {
				earliest = r.IssueDate
			}
		}
	}
	return &Response{
		Query:        query,
		Count:        len(rows),
		Issues:       keysSorted(issues),
		States:       keysSorted(states),
		LatestDate:   latest,
		EarliestDate: earliest,
		Hits:         rows,
	}
}

func keysSorted(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
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
