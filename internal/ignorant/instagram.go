package ignorant

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	igEndpoint = "https://i.instagram.com/api/v1/users/lookup/"
	igSigKey   = "e6358aeede676184b9fe702b30f4fd35e71744605e39d2181a34cede076b3c33"
	igSigVer   = "4"
)

func instagramCheck(ctx context.Context, hc *http.Client, p parsed) Result {
	r := Result{
		Name:   "instagram",
		Domain: "instagram.com",
		Method: "users.lookup",
	}

	payload := map[string]string{
		"login_attempt_count": "0",
		"directly_sign_in":    "true",
		"source":              "default",
		"q":                   p.cc4Phone,
		"ig_sig_key_version":  igSigVer,
	}
	body, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(igSigKey))
	mac.Write(body)
	signed := hex.EncodeToString(mac.Sum(nil))

	form := url.Values{}
	form.Set("ig_sig_key_version", igSigVer)
	form.Set("signed_body", signed+"."+string(body))

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, igEndpoint,
		strings.NewReader(form.Encode()))
	req.Header.Set("Accept-Language", "en-US")
	req.Header.Set("User-Agent", "Instagram 101.0.0.15.120")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-FB-HTTP-Engine", "Liger")
	req.Header.Set("Connection", "close")

	resp, err := hc.Do(req)
	if err != nil {
		r.RateLimit = true
		r.Error = "request: " + err.Error()
		return r
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode == 429 {
		r.RateLimit = true
		r.Error = "rate-limited (429)"
		return r
	}

	var rep struct {
		Message  string `json:"message"`
		Status   string `json:"status"`
		ErrorMsg string `json:"error_message"`
	}
	if err := json.Unmarshal(raw, &rep); err != nil {
		r.RateLimit = true
		r.Error = "decode: " + err.Error()
		return r
	}

	if rep.Message == "No users found" {
		r.Exists = false
		return r
	}
	if rep.ErrorMsg != "" || strings.EqualFold(rep.Status, "fail") {
		r.RateLimit = true
		r.Error = "ig says: " + first(rep.ErrorMsg, rep.Message)
		return r
	}
	r.Exists = true
	return r
}

func first(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}
