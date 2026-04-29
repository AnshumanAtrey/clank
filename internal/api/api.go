package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Provider interface {
	Name() string
	Lookup(ctx context.Context, e164 string) (Result, error)
}

type Result struct {
	Provider    string          `json:"provider"`
	Number      string          `json:"number"`
	Valid       bool            `json:"valid"`
	Country     string          `json:"country,omitempty"`
	CountryCode string          `json:"country_code,omitempty"`
	Carrier     string          `json:"carrier,omitempty"`
	LineType    string          `json:"line_type,omitempty"`
	Location    string          `json:"location,omitempty"`
	CallerName  string          `json:"caller_name,omitempty"`
	FraudScore  *int            `json:"fraud_score,omitempty"`
	Active      *bool           `json:"active,omitempty"`
	Risky       *bool           `json:"risky,omitempty"`
	Raw         json.RawMessage `json:"raw,omitempty"`
	Error       string          `json:"error,omitempty"`
}

func New(provider, key string) (Provider, error) {
	switch strings.ToLower(provider) {
	case "numverify", "nv":
		return &numverify{key: key, hc: defaultClient()}, nil
	case "veriphone", "vp":
		return &veriphone{key: key, hc: defaultClient()}, nil
	case "ipqs", "ipqualityscore":
		return &ipqs{key: key, hc: defaultClient()}, nil
	}
	return nil, fmt.Errorf("unknown provider %q (try: numverify | veriphone | ipqs)", provider)
}

func defaultClient() *http.Client {
	return &http.Client{Timeout: 15 * time.Second}
}

func LookupBatch(ctx context.Context, p Provider, numbers []string, workers int) []Result {
	if workers < 1 {
		workers = 1
	}
	if workers > len(numbers) {
		workers = len(numbers)
	}
	out := make([]Result, len(numbers))
	in := make(chan int, len(numbers))
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range in {
				r, err := p.Lookup(ctx, numbers[idx])
				if err != nil {
					r.Error = err.Error()
				}
				if r.Number == "" {
					r.Number = numbers[idx]
				}
				if r.Provider == "" {
					r.Provider = p.Name()
				}
				out[idx] = r
			}
		}()
	}
	for i := range numbers {
		in <- i
	}
	close(in)
	wg.Wait()
	return out
}

func httpJSON(ctx context.Context, hc *http.Client, req *http.Request, into interface{}) (json.RawMessage, error) {
	req = req.WithContext(ctx)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return body, fmt.Errorf("auth failed (%d) — check your API key", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return body, fmt.Errorf("rate-limited (429) — slow down or wait for quota reset")
	}
	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("http %d: %s", resp.StatusCode, snippet(body))
	}
	if into != nil {
		if err := json.Unmarshal(body, into); err != nil {
			return body, fmt.Errorf("decode: %w (body: %s)", err, snippet(body))
		}
	}
	return body, nil
}

func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}

type numverify struct {
	key string
	hc  *http.Client
}

func (n *numverify) Name() string { return "numverify" }

func (n *numverify) Lookup(ctx context.Context, e164 string) (Result, error) {
	u := "https://api.apilayer.com/number_verification/validate?number=" + url.QueryEscape(strings.TrimPrefix(e164, "+"))
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("apikey", n.key)
	var raw struct {
		Valid               bool   `json:"valid"`
		Number              string `json:"number"`
		LocalFormat         string `json:"local_format"`
		InternationalFormat string `json:"international_format"`
		CountryPrefix       string `json:"country_prefix"`
		CountryCode         string `json:"country_code"`
		CountryName         string `json:"country_name"`
		Location            string `json:"location"`
		Carrier             string `json:"carrier"`
		LineType            string `json:"line_type"`
		Message             string `json:"message"`
	}
	body, err := httpJSON(ctx, n.hc, req, &raw)
	r := Result{Provider: "numverify", Number: e164, Raw: body}
	if err != nil {
		return r, err
	}
	if raw.Message != "" && !raw.Valid {
		return r, fmt.Errorf("%s", raw.Message)
	}
	r.Valid = raw.Valid
	r.Country = raw.CountryName
	r.CountryCode = raw.CountryCode
	r.Carrier = raw.Carrier
	r.LineType = raw.LineType
	r.Location = raw.Location
	return r, nil
}

type veriphone struct {
	key string
	hc  *http.Client
}

func (v *veriphone) Name() string { return "veriphone" }

func (v *veriphone) Lookup(ctx context.Context, e164 string) (Result, error) {
	u := "https://api.veriphone.io/v2/verify?phone=" + url.QueryEscape(e164)
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("Authorization", "Bearer "+v.key)
	var raw struct {
		Status      string `json:"status"`
		Phone       string `json:"phone"`
		PhoneValid  bool   `json:"phone_valid"`
		PhoneType   string `json:"phone_type"`
		PhoneRegion string `json:"phone_region"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		Carrier     string `json:"carrier"`
		E164        string `json:"e164"`
	}
	body, err := httpJSON(ctx, v.hc, req, &raw)
	r := Result{Provider: "veriphone", Number: e164, Raw: body}
	if err != nil {
		return r, err
	}
	if raw.Status != "" && !strings.EqualFold(raw.Status, "success") && raw.E164 == "" {
		return r, fmt.Errorf("status=%s", raw.Status)
	}
	r.Valid = raw.PhoneValid
	r.LineType = raw.PhoneType
	r.Country = raw.Country
	r.CountryCode = raw.CountryCode
	r.Carrier = raw.Carrier
	r.Location = raw.PhoneRegion
	return r, nil
}

type ipqs struct {
	key string
	hc  *http.Client
}

func (i *ipqs) Name() string { return "ipqs" }

func (i *ipqs) Lookup(ctx context.Context, e164 string) (Result, error) {
	u := "https://www.ipqualityscore.com/api/json/phone/" + url.PathEscape(i.key) + "/" + url.PathEscape(strings.TrimPrefix(e164, "+"))
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	var raw struct {
		Success     bool   `json:"success"`
		Message     string `json:"message"`
		Valid       bool   `json:"valid"`
		FraudScore  int    `json:"fraud_score"`
		Active      bool   `json:"active"`
		Risky       bool   `json:"risky"`
		VOIP        bool   `json:"VOIP"`
		Prepaid     bool   `json:"prepaid"`
		Carrier     string `json:"carrier"`
		LineType    string `json:"line_type"`
		Country     string `json:"country"`
		Region      string `json:"region"`
		City        string `json:"city"`
		Timezone    string `json:"timezone"`
		DoNotCall   bool   `json:"do_not_call"`
		RecentAbuse bool   `json:"recent_abuse"`
	}
	body, err := httpJSON(ctx, i.hc, req, &raw)
	r := Result{Provider: "ipqs", Number: e164, Raw: body}
	if err != nil {
		return r, err
	}
	if !raw.Success {
		if raw.Message != "" {
			return r, fmt.Errorf("%s", raw.Message)
		}
	}
	r.Valid = raw.Valid
	score := raw.FraudScore
	r.FraudScore = &score
	active := raw.Active
	r.Active = &active
	risky := raw.Risky
	r.Risky = &risky
	r.LineType = raw.LineType
	r.Carrier = raw.Carrier
	r.Country = raw.Country
	loc := strings.TrimSpace(raw.City + ", " + raw.Region)
	if loc != "," {
		r.Location = strings.Trim(loc, ", ")
	}
	return r, nil
}
