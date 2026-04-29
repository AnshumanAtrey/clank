package ignorant

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	amazonSigninURL = "https://www.amazon.com/ap/signin?openid.pape.max_auth_age=0&openid.return_to=https%3A%2F%2Fwww.amazon.com%2F%3F_encoding%3DUTF8%26ref_%3Dnav_ya_signin&openid.identity=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.assoc_handle=usflex&openid.mode=checkid_setup&openid.claimed_id=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0%2Fidentifier_select&openid.ns=http%3A%2F%2Fspecs.openid.net%2Fauth%2F2.0&"
	amazonPostURL   = "https://www.amazon.com/ap/signin/"
)

func amazonCheck(ctx context.Context, hc *http.Client, p parsed) Result {
	r := Result{
		Name:   "amazon",
		Domain: "amazon.com",
		Method: "signin",
	}

	getReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, amazonSigninURL, nil)
	getReq.Header.Set("User-Agent", randomUA())
	getResp, err := hc.Do(getReq)
	if err != nil {
		r.RateLimit = true
		r.Error = "get signin: " + err.Error()
		return r
	}
	defer getResp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(getResp.Body, 4<<20))
	if err != nil {
		r.RateLimit = true
		r.Error = "read get: " + err.Error()
		return r
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		r.RateLimit = true
		r.Error = "parse signin: " + err.Error()
		return r
	}

	form := url.Values{}
	doc.Find("form input").Each(func(_ int, sel *goquery.Selection) {
		name, hasName := sel.Attr("name")
		val, hasVal := sel.Attr("value")
		if hasName && hasVal {
			form.Set(name, val)
		}
	})
	if len(form) == 0 {
		r.RateLimit = true
		r.Error = "amazon signin form not found (DOM changed?)"
		return r
	}

	form.Set("email", p.cc4Phone)

	postReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, amazonPostURL,
		strings.NewReader(form.Encode()))
	postReq.Header.Set("User-Agent", randomUA())
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.Header.Set("Origin", "https://www.amazon.com")
	postReq.Header.Set("Referer", amazonSigninURL)

	postResp, err := hc.Do(postReq)
	if err != nil {
		r.RateLimit = true
		r.Error = "post signin: " + err.Error()
		return r
	}
	defer postResp.Body.Close()
	postBody, err := io.ReadAll(io.LimitReader(postResp.Body, 4<<20))
	if err != nil {
		r.RateLimit = true
		r.Error = "read post: " + err.Error()
		return r
	}

	respDoc, err := goquery.NewDocumentFromReader(strings.NewReader(string(postBody)))
	if err != nil {
		r.RateLimit = true
		r.Error = "parse response: " + err.Error()
		return r
	}

	if respDoc.Find("#auth-password-missing-alert").Length() > 0 {
		r.Exists = true
		return r
	}
	if respDoc.Find("#auth-error-message-box").Length() > 0 {
		r.Exists = false
		return r
	}
	if respDoc.Find("#authportal-main-section").Length() > 0 ||
		respDoc.Find("input[name=email]").Length() > 0 {
		r.Exists = false
		return r
	}
	r.RateLimit = true
	r.Error = "indeterminate response (Amazon DOM may have changed)"
	return r
}
