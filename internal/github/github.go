// Package github queries GitHub's commit-search API for phone numbers leaked
// in commit messages. People accidentally include phone numbers in commit
// messages all the time ("call +1-415-555-2671 for prod outage", "fixes
// support@x.com 415-555-2671 ticket", etc.); the GitHub commit search
// indexes those messages and is queryable without authentication.
//
// Limitations:
//   - GitHub search only indexes commit messages, NOT diff/file content.
//     For phones-in-source-code, the /search/code endpoint is required and
//     that endpoint mandates authentication; we therefore stay with commit
//     search for the zero-config default.
//   - Unauthenticated rate limit is 10 requests / minute (per IP). A user-
//     supplied GITHUB_TOKEN raises that to 30 / minute.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	pn "github.com/nyaruka/phonenumbers"
)

const (
	endpoint        = "https://api.github.com/search/commits"
	defaultUA       = "clank-osint build@atrey.dev"
	envTokenVar     = "GITHUB_TOKEN"
	envUserAgentVar = "CLANK_GITHUB_UA"
)

// Hit is one commit returned by the search.
type Hit struct {
	SHA         string `json:"sha"`
	Repo        string `json:"repo"`
	URL         string `json:"url"` // human-friendly github.com URL
	AuthorName  string `json:"author_name,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	AuthorLogin string `json:"author_login,omitempty"`
	CommitDate  string `json:"commit_date,omitempty"`
	MessageHead string `json:"message_head,omitempty"`
	Variant     string `json:"variant,omitempty"` // which phone-format variation matched
}

// Response is the de-duped, summarized result for one phone number.
type Response struct {
	Query    string `json:"query"`
	Total    int    `json:"total"`
	Returned int    `json:"returned"`
	Variants []string `json:"variants,omitempty"` // formats actually queried
	Hits     []Hit  `json:"hits,omitempty"`
}

// Options configures Search.
type Options struct {
	Limit    int
	Endpoint string // override for tests
}

// Search runs GitHub commit-search for the given phone, querying multiple
// format variations (E.164, national, hyphenated, digits-only) and de-duping
// by commit SHA.
func Search(ctx context.Context, e164 string, opts Options) (*Response, error) {
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	ep := opts.Endpoint
	if ep == "" {
		ep = endpoint
	}

	variants, err := phoneVariants(e164)
	if err != nil {
		return nil, err
	}

	hc := &http.Client{Timeout: 10 * time.Second}
	out := &Response{Query: e164, Variants: variants}

	seenSHA := map[string]bool{}
	totalAcrossVariants := 0
	perVariantCap := limit
	if perVariantCap < 3 {
		perVariantCap = 3
	}

	for _, v := range variants {
		hits, total, err := searchOne(ctx, hc, ep, v, perVariantCap)
		if err != nil {
			return nil, err
		}
		totalAcrossVariants += total
		for _, h := range hits {
			if seenSHA[h.SHA] {
				continue
			}
			seenSHA[h.SHA] = true
			h.Variant = v
			out.Hits = append(out.Hits, h)
			if len(out.Hits) >= limit {
				break
			}
		}
		if len(out.Hits) >= limit {
			break
		}
	}
	out.Total = totalAcrossVariants
	out.Returned = len(out.Hits)
	return out, nil
}

// phoneVariants derives the format strings to search for. We only emit forms
// people actually paste into commit messages — E.164, hyphenated US,
// digits-only, and the national-format string from libphonenumber.
func phoneVariants(input string) ([]string, error) {
	num, err := pn.Parse(input, "")
	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", input, err)
	}
	cc := num.GetCountryCode()
	natnum := num.GetNationalNumber()

	e164 := pn.Format(num, pn.E164)
	natl := pn.Format(num, pn.NATIONAL)
	intl := pn.Format(num, pn.INTERNATIONAL)
	digits := fmt.Sprintf("%d%d", cc, natnum)

	candidates := []string{e164, natl, intl, digits}
	seen := map[string]bool{}
	out := make([]string, 0, len(candidates))
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out, nil
}

type ghCommitItem struct {
	SHA    string `json:"sha"`
	HTMLURL string `json:"html_url"`
	Commit struct {
		Author struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Date  string `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

type ghSearchResp struct {
	TotalCount int            `json:"total_count"`
	Items      []ghCommitItem `json:"items"`
}

func searchOne(ctx context.Context, hc *http.Client, ep, query string, perPage int) ([]Hit, int, error) {
	q := url.Values{}
	q.Set("q", `"`+query+`"`)
	q.Set("per_page", fmt.Sprintf("%d", perPage))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ep+"?"+q.Encode(), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent())
	if tok := strings.TrimSpace(os.Getenv(envTokenVar)); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("github request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == 429 {
		// Rate-limited (most common 4xx without auth). Surface as a
		// dedicated error so callers can render an actionable hint.
		return nil, 0, ErrRateLimited
	}
	if resp.StatusCode >= 400 {
		return nil, 0, fmt.Errorf("github http %d: %s", resp.StatusCode, snippet(body))
	}

	var raw ghSearchResp
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, 0, fmt.Errorf("github decode: %w", err)
	}

	hits := make([]Hit, 0, len(raw.Items))
	for _, it := range raw.Items {
		h := Hit{
			SHA:         it.SHA,
			Repo:        it.Repository.FullName,
			URL:         it.HTMLURL,
			AuthorName:  it.Commit.Author.Name,
			AuthorEmail: it.Commit.Author.Email,
			AuthorLogin: it.Author.Login,
			CommitDate:  it.Commit.Author.Date,
			MessageHead: firstLine(it.Commit.Message),
		}
		hits = append(hits, h)
	}
	return hits, raw.TotalCount, nil
}

// ErrRateLimited indicates a GitHub 403/429 — typically because the
// unauthenticated 10-req/min cap was hit. Caller should suggest the user set
// GITHUB_TOKEN to raise the cap.
var ErrRateLimited = fmt.Errorf("github rate-limited (set GITHUB_TOKEN to raise the cap)")

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	// 120-byte cap including the ellipsis. "…" is 3 bytes in UTF-8.
	if len(s) > 120 {
		s = s[:117] + "…"
	}
	return s
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
