package ignorant

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	snapchatHome = "https://accounts.snapchat.com"
	snapchatVal  = "https://accounts.snapchat.com/accounts/validate_phone_number"
)

func snapchatCheck(ctx context.Context, hc *http.Client, p parsed) Result {
	r := Result{
		Name:   "snapchat",
		Domain: "snapchat.com",
		Method: "validate_phone_number",
	}

	homeReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, snapchatHome, nil)
	commonHeaders(homeReq)
	homeResp, err := hc.Do(homeReq)
	if err != nil {
		r.RateLimit = true
		r.Error = "home: " + err.Error()
		return r
	}
	homeResp.Body.Close()

	var xsrf string
	for _, c := range homeResp.Cookies() {
		if c.Name == "xsrf_token" {
			xsrf = c.Value
			break
		}
	}
	if xsrf == "" {
		hu, _ := url.Parse(snapchatHome)
		for _, c := range hc.Jar.Cookies(hu) {
			if c.Name == "xsrf_token" {
				xsrf = c.Value
				break
			}
		}
	}
	if xsrf == "" {
		r.RateLimit = true
		r.Error = "method broken — Snapchat redesigned web auth (~2024); CSRF now JS-side, no cookie"
		return r
	}

	region := p.regionISO
	if region == "" {
		r.RateLimit = true
		r.Error = "could not derive ISO region from phone"
		return r
	}

	form := url.Values{}
	form.Set("phone_country_code", region)
	form.Set("phone_number", p.nationalNum)
	form.Set("xsrf_token", xsrf)

	postReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, snapchatVal,
		strings.NewReader(form.Encode()))
	commonHeaders(postReq)
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	postReq.Header.Set("Origin", snapchatHome)
	postReq.Header.Set("Referer", snapchatHome+"/")

	resp, err := hc.Do(postReq)
	if err != nil {
		r.RateLimit = true
		r.Error = "validate: " + err.Error()
		return r
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode == 429 {
		r.RateLimit = true
		r.FrequentRateLimit = true
		r.Error = "rate-limited (429)"
		return r
	}

	var rep struct {
		StatusCode string `json:"status_code"`
	}
	if err := json.Unmarshal(body, &rep); err != nil {
		r.RateLimit = true
		r.Error = "decode: " + snippet(body)
		return r
	}

	switch rep.StatusCode {
	case "TAKEN_NUMBER":
		r.Exists = true
	case "OK":
		r.Exists = false
	default:
		r.RateLimit = true
		r.Error = "snap says: " + rep.StatusCode
	}
	return r
}

func commonHeaders(req *http.Request) {
	req.Header.Set("User-Agent", randomUA())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en,en-US;q=0.5")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-GPC", "1")
}

func snippet(b []byte) string {
	if len(b) > 200 {
		return string(b[:200]) + "..."
	}
	return string(b)
}
