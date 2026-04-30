package audit

import "testing"

func TestLooksLikePhone(t *testing.T) {
	cases := map[string]bool{
		"+14155552671":                      true,
		"14155552671":                       true,
		"+919181156055":                     true,
		"918115605xxx":                      true,
		"(415) 555-2671":                    true,
		"":                                  false,
		"abc":                               false,
		"--json":                            false,
		"--key":                             false,
		"a-very-long-non-phone-string-here": false,
		"123":                               false, // too short
	}
	for in, want := range cases {
		if got := looksLikePhone(in); got != want {
			t.Errorf("looksLikePhone(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestFirstPhoneArg(t *testing.T) {
	cases := []struct {
		args []string
		want string
	}{
		{[]string{"--json", "+14155552671"}, "+14155552671"},
		{[]string{"+14155552671"}, "+14155552671"},
		{[]string{"--region", "IN", "9181156055"}, "9181156055"},
		{[]string{"--key", "secret123", "+14155552671"}, "+14155552671"},
		{[]string{"--json"}, ""},
		{nil, ""},
	}
	for _, c := range cases {
		if got := FirstPhoneArg(c.args); got != c.want {
			t.Errorf("FirstPhoneArg(%v) = %q, want %q", c.args, got, c.want)
		}
	}
}
