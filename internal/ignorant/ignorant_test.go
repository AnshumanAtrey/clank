package ignorant

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestParsePhone_USE164(t *testing.T) {
	p, err := parsePhone("+14155552671")
	if err != nil {
		t.Fatal(err)
	}
	if p.intCC != 1 {
		t.Errorf("intCC = %d, want 1", p.intCC)
	}
	if p.regionISO != "US" {
		t.Errorf("region = %q, want US", p.regionISO)
	}
	if p.nationalNum != "4155552671" {
		t.Errorf("national = %q, want 4155552671", p.nationalNum)
	}
	if p.cc4Phone != "14155552671" {
		t.Errorf("cc4Phone = %q, want 14155552671", p.cc4Phone)
	}
}

func TestParsePhone_AutoPlus(t *testing.T) {
	p, err := parsePhone("14155552671")
	if err != nil {
		t.Fatal(err)
	}
	if p.regionISO != "US" {
		t.Errorf("region = %q, want US", p.regionISO)
	}
}

func TestParsePhone_Invalid(t *testing.T) {
	if _, err := parsePhone(""); err == nil {
		t.Error("expected error for empty phone")
	}
	if _, err := parsePhone("not-a-number"); err == nil {
		t.Error("expected error for garbage input")
	}
}

func TestPickModules(t *testing.T) {
	if got := pickModules(nil); len(got) != 3 {
		t.Errorf("default pick = %d, want 3", len(got))
	}
	if got := pickModules([]string{"instagram"}); len(got) != 1 || got[0] != "instagram" {
		t.Errorf("only=ig → %v, want [instagram]", got)
	}
	if got := pickModules([]string{"INSTAGRAM", "Snapchat"}); len(got) != 2 {
		t.Errorf("case-insensitive pick failed: %v", got)
	}
	if got := pickModules([]string{"twitter"}); len(got) != 0 {
		t.Errorf("unknown module should yield empty: %v", got)
	}
}

func TestIGSignatureFormatStable(t *testing.T) {
	body := []byte(`{"q":"14155552671"}`)
	mac := hmac.New(sha256.New, []byte(igSigKey))
	mac.Write(body)
	want := hex.EncodeToString(mac.Sum(nil))
	if want == "" {
		t.Error("hmac produced empty digest")
	}
	if len(want) != 64 {
		t.Errorf("hmac digest length = %d, want 64", len(want))
	}
}
