package whatsapp

import (
	"encoding/json"
	"strings"
	"testing"
)

// Compile-check tests. These don't talk to WhatsApp — they verify that the
// package's exported surface is wired up correctly. Live integration testing
// has to happen on a real paired account.

func TestStripPlus(t *testing.T) {
	cases := map[string]string{
		"+14155552671": "14155552671",
		"14155552671":  "14155552671",
		"":             "",
		"+":            "",
	}
	for in, want := range cases {
		if got := stripPlus(in); got != want {
			t.Errorf("stripPlus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDSN_HasMandatoryPragmas(t *testing.T) {
	d := dsn("/tmp/whatsapp.db")
	for _, pragma := range []string{"foreign_keys(1)", "busy_timeout(5000)", "journal_mode(wal)"} {
		if !strings.Contains(d, pragma) {
			t.Errorf("DSN missing pragma %q: %s", pragma, d)
		}
	}
	if !strings.HasPrefix(d, "file:/tmp/whatsapp.db") {
		t.Errorf("DSN should start with file:<path>, got: %s", d)
	}
}

func TestResultJSON_OmitEmpty(t *testing.T) {
	r := Result{Query: "+14155552671", Registered: false}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, field := range []string{"jid", "lid", "verified_business_name", "about"} {
		if strings.Contains(got, field) {
			t.Errorf("empty %s leaked into JSON: %s", field, got)
		}
	}
	if !strings.Contains(got, `"registered":false`) {
		t.Errorf("registered missing from output: %s", got)
	}
}

func TestResultJSON_FullShape(t *testing.T) {
	r := Result{
		Query:             "+14155552671",
		JID:               "14155552671@s.whatsapp.net",
		Registered:        true,
		About:             "hello",
		PictureID:         "abc",
		ProfilePictureURL: "https://pps.whatsapp.net/...",
		DeviceCount:       3,
	}
	b, _ := json.Marshal(r)
	got := string(b)
	for _, want := range []string{`"jid"`, `"about":"hello"`, `"device_count":3`, `"profile_picture_url"`} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in output: %s", want, got)
		}
	}
}
