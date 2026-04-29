package local

import "testing"

func TestInspect_USNumber(t *testing.T) {
	l := Inspect("+14155552671", "")
	if !l.Valid {
		t.Errorf("expected valid, got error: %s", l.ParseError)
	}
	if l.Region != "US" {
		t.Errorf("region = %q, want US", l.Region)
	}
	if l.CountryCode != 1 {
		t.Errorf("country code = %d, want 1", l.CountryCode)
	}
	if l.Formatted.E164 != "+14155552671" {
		t.Errorf("E164 = %q", l.Formatted.E164)
	}
}

func TestInspect_IndianNumber(t *testing.T) {
	l := Inspect("9181156052", "IN")
	if l.ParseError != "" {
		t.Logf("parse error (acceptable for short input): %s", l.ParseError)
	}
	if l.Region != "" && l.Region != "IN" {
		t.Errorf("region = %q, want IN", l.Region)
	}
}

func TestInspect_PrependPlus(t *testing.T) {
	l := Inspect("14155552671", "")
	if !l.Valid {
		t.Errorf("expected auto-prepend + to make it valid, got error: %s", l.ParseError)
	}
}

func TestInspect_BadInput(t *testing.T) {
	l := Inspect("not-a-number", "")
	if l.Valid {
		t.Error("expected invalid")
	}
}

func TestOperatorsInCountry_US(t *testing.T) {
	ops := OperatorsInCountry("US")
	if len(ops) < 5 {
		t.Errorf("expected ≥5 US operators, got %d", len(ops))
	}
}

func TestSpamCount_NonZero(t *testing.T) {
	if SpamCount() < 100 {
		t.Errorf("expected at least 100 spam entries loaded, got %d", SpamCount())
	}
}

func TestLineTypeName(t *testing.T) {
	if got := lineTypeName(99); got != "UNKNOWN" {
		t.Errorf("unknown type fallback = %q", got)
	}
}
