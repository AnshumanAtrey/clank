package imei

import "testing"

func TestLuhn_Valid(t *testing.T) {
	// 352099001761481 — valid Luhn (sum=50, divisible by 10)
	if !LuhnCheck("352099001761481") {
		t.Error("expected valid Luhn for 352099001761481")
	}
}

func TestLuhn_Invalid(t *testing.T) {
	if LuhnCheck("352099001761482") {
		t.Error("expected invalid Luhn for 352099001761482")
	}
}

func TestComputeLuhn(t *testing.T) {
	c, err := ComputeLuhnCheck("35209900176148")
	if err != nil {
		t.Fatal(err)
	}
	if c != '1' {
		t.Errorf("got check digit %c, want 1", c)
	}
}

func TestParse_15Digit(t *testing.T) {
	r := Parse("35-2099-00-176148-1")
	if !r.StructureOK {
		t.Errorf("structure not ok: %s", r.ParseError)
	}
	if r.TAC != "35209900" {
		t.Errorf("TAC = %q, want 35209900", r.TAC)
	}
	if r.Length != 15 {
		t.Errorf("length = %d, want 15", r.Length)
	}
	if !r.LuhnValid {
		t.Errorf("expected valid Luhn for 352099001761481")
	}
}

func TestParse_TooShort(t *testing.T) {
	r := Parse("123")
	if r.ParseError == "" {
		t.Error("expected error for short input")
	}
}

func TestParse_NonDigit(t *testing.T) {
	r := Parse("3520990abc76148")
	if r.ParseError == "" {
		t.Error("expected error for non-digit")
	}
}

func TestTACCount_NonZero(t *testing.T) {
	if TACCount() < 10_000 {
		t.Errorf("expected ≥10k TAC entries, got %d", TACCount())
	}
}
