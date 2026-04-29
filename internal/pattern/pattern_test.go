package pattern

import "testing"

func TestIsValid(t *testing.T) {
	cases := map[string]bool{
		"918115605xxx": true,
		"123":          true,
		"xxx":          true,
		"918XXX":       true,
		"":             false,
		"918abc":       false,
		"+":            false,
		"+919999":      true,
		"+918115xxx":   true,
		"918 115":      false,
	}
	for in, want := range cases {
		if got := IsValid(in); got != want {
			t.Errorf("IsValid(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestGenerate_NoPlaceholders(t *testing.T) {
	out, err := Generate("9181156052", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0] != "9181156052" {
		t.Errorf("got %v, want [9181156052]", out)
	}
}

func TestGenerate_OnePlaceholder(t *testing.T) {
	out, err := Generate("123x", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 10 {
		t.Errorf("len = %d, want 10", len(out))
	}
	if out[0] != "1230" || out[9] != "1239" {
		t.Errorf("first/last = %s/%s, want 1230/1239", out[0], out[9])
	}
}

func TestGenerate_ThreePlaceholders(t *testing.T) {
	out, err := Generate("918115605xxx", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1000 {
		t.Errorf("len = %d, want 1000", len(out))
	}
	if out[0] != "918115605000" || out[999] != "918115605999" {
		t.Errorf("first/last = %s/%s", out[0], out[999])
	}
}

func TestGenerate_TooLarge(t *testing.T) {
	_, err := Generate("xxxxxxx", false) // 10^7 = 10M
	if err == nil {
		t.Error("expected ErrTooLarge")
	}
}

func TestGenerate_Invalid(t *testing.T) {
	if _, err := Generate("918abc", false); err == nil {
		t.Error("expected ErrInvalid for non-digit chars")
	}
	if _, err := Generate("", false); err == nil {
		t.Error("expected ErrEmpty for empty input")
	}
}

func TestCountCombinations(t *testing.T) {
	cases := map[string]int{
		"123":          1,
		"123x":         10,
		"918115605xxx": 1000,
		"xxxxx":        100_000,
	}
	for in, want := range cases {
		if got := CountCombinations(in); got != want {
			t.Errorf("Count(%q) = %d, want %d", in, got, want)
		}
	}
}
