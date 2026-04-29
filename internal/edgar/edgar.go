package edgar

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
)

const (
	endpoint        = "https://efts.sec.gov/LATEST/search-index"
	defaultUA       = "clank-osint anonymous@clank.local"
	envUserAgentVar = "CLANK_EDGAR_UA"
)

func userAgent() string {
	if v := strings.TrimSpace(os.Getenv(envUserAgentVar)); v != "" {
		return v
	}
	return defaultUA
}

type Hit struct {
	ID           string   `json:"id"`
	Score        float64  `json:"score"`
	Form         string   `json:"form"`
	FileDate     string   `json:"file_date"`
	DisplayNames []string `json:"display_names"`
	CIKs         []string `json:"ciks"`
	Accession    string   `json:"accession"`
	BizLocations []string `json:"biz_locations,omitempty"`
	BizStates    []string `json:"biz_states,omitempty"`
	IncStates    []string `json:"inc_states,omitempty"`
	URL          string   `json:"url"`
}

type Response struct {
	Query    string `json:"query"`
	Total    int    `json:"total"`
	Returned int    `json:"returned"`
	Forms    string `json:"forms,omitempty"`
	Hits     []Hit  `json:"hits"`
}

type efts struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source struct {
				Form         string   `json:"form"`
				FileDate     string   `json:"file_date"`
				DisplayNames []string `json:"display_names"`
				CIKs         []string `json:"ciks"`
				Accession    string   `json:"adsh"`
				BizLocations []string `json:"biz_locations"`
				BizStates    []string `json:"biz_states"`
				IncStates    []string `json:"inc_states"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type Options struct {
	Forms []string
	Hits  int
}

func Search(ctx context.Context, query string, opts Options) (*Response, error) {
	hits := opts.Hits
	if hits <= 0 || hits > 100 {
		hits = 10
	}
	q := url.Values{}
	q.Set("q", `"`+query+`"`)
	if len(opts.Forms) > 0 {
		q.Set("forms", strings.Join(opts.Forms, ","))
	}
	q.Set("hits", fmt.Sprintf("%d", hits))

	u := endpoint + "?" + q.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", userAgent())
	req.Header.Set("Accept", "application/json")

	hc := &http.Client{Timeout: 20 * time.Second}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("edgar request: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("edgar http %d: %s", resp.StatusCode, snippet(body))
	}

	var raw efts
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("edgar decode: %w", err)
	}

	out := &Response{
		Query:    query,
		Total:    raw.Hits.Total.Value,
		Returned: len(raw.Hits.Hits),
		Forms:    strings.Join(opts.Forms, ","),
		Hits:     make([]Hit, 0, len(raw.Hits.Hits)),
	}
	for _, h := range raw.Hits.Hits {
		hh := Hit{
			ID:           h.ID,
			Score:        h.Score,
			Form:         h.Source.Form,
			FileDate:     h.Source.FileDate,
			DisplayNames: h.Source.DisplayNames,
			CIKs:         h.Source.CIKs,
			Accession:    h.Source.Accession,
			BizLocations: h.Source.BizLocations,
			BizStates:    h.Source.BizStates,
			IncStates:    h.Source.IncStates,
			URL:          filingURL(h.Source.Accession, h.ID, h.Source.CIKs),
		}
		out.Hits = append(out.Hits, hh)
	}
	return out, nil
}

func filingURL(accession, id string, ciks []string) string {
	if accession == "" || len(ciks) == 0 {
		return ""
	}
	cik := strings.TrimLeft(ciks[0], "0")
	if cik == "" {
		cik = "0"
	}
	noDashes := strings.ReplaceAll(accession, "-", "")
	doc := id
	if i := strings.Index(doc, ":"); i != -1 {
		doc = doc[i+1:]
	}
	if doc == "" {
		doc = noDashes + "-index.htm"
	}
	return fmt.Sprintf("https://www.sec.gov/Archives/edgar/data/%s/%s/%s",
		cik, noDashes, doc)
}

func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}

func QueryForms(formsCSV string) []string {
	formsCSV = strings.TrimSpace(formsCSV)
	if formsCSV == "" {
		return nil
	}
	parts := strings.Split(formsCSV, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
