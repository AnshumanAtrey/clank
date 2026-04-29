package local

import (
	"errors"
	"fmt"

	pn "github.com/nyaruka/phonenumbers"
)

type Lookup struct {
	Input       string     `json:"input"`
	E164        string     `json:"e164,omitempty"`
	Valid       bool       `json:"valid"`
	Possible    bool       `json:"possible"`
	Region      string     `json:"region,omitempty"`
	CountryCode int        `json:"country_code,omitempty"`
	NationalNum uint64     `json:"national_number,omitempty"`
	LineType    string     `json:"line_type,omitempty"`
	Carrier     string     `json:"carrier,omitempty"`
	Geo         string     `json:"geo,omitempty"`
	Timezones   []string   `json:"timezones,omitempty"`
	Formatted   Formats    `json:"formatted"`
	Spam        bool       `json:"spam"`
	SpamReason  string     `json:"spam_reason,omitempty"`
	Operators   []Operator `json:"operators_in_country,omitempty"`
	ParseError  string     `json:"parse_error,omitempty"`
}

type Formats struct {
	E164          string `json:"e164,omitempty"`
	International string `json:"international,omitempty"`
	National      string `json:"national,omitempty"`
	RFC3966       string `json:"rfc3966,omitempty"`
}

func Inspect(input, defaultRegion string) Lookup {
	out := Lookup{Input: input}

	parsed := input
	if defaultRegion == "" && len(input) > 0 && input[0] != '+' {
		parsed = "+" + input
	}

	num, err := pn.Parse(parsed, defaultRegion)
	if err != nil {
		out.ParseError = err.Error()
		return out
	}

	out.Possible = pn.IsPossibleNumber(num)
	out.Valid = pn.IsValidNumber(num)
	out.Region = pn.GetRegionCodeForNumber(num)
	out.CountryCode = int(num.GetCountryCode())
	out.NationalNum = num.GetNationalNumber()
	out.LineType = lineTypeName(pn.GetNumberType(num))

	if c, e := pn.GetCarrierForNumber(num, "en"); e == nil && c != "" {
		out.Carrier = c
	}
	if g, e := pn.GetGeocodingForNumber(num, "en"); e == nil && g != "" {
		out.Geo = g
	}
	if tz, e := pn.GetTimezonesForNumber(num); e == nil {
		out.Timezones = tz
	}

	out.Formatted = Formats{
		E164:          pn.Format(num, pn.E164),
		International: pn.Format(num, pn.INTERNATIONAL),
		National:      pn.Format(num, pn.NATIONAL),
		RFC3966:       pn.Format(num, pn.RFC3966),
	}
	out.E164 = out.Formatted.E164

	if out.E164 != "" {
		if reason, hit := lookupSpam(out.E164); hit {
			out.Spam = true
			out.SpamReason = reason
		}
	}
	if out.Region != "" {
		out.Operators = OperatorsInCountry(out.Region)
	}
	return out
}

func InspectMany(inputs []string, defaultRegion string) []Lookup {
	out := make([]Lookup, len(inputs))
	for i, s := range inputs {
		out[i] = Inspect(s, defaultRegion)
	}
	return out
}

var ErrNoLib = errors.New("phonenumbers library not initialised")

func lineTypeName(t pn.PhoneNumberType) string {
	switch t {
	case pn.MOBILE:
		return "MOBILE"
	case pn.FIXED_LINE:
		return "FIXED_LINE"
	case pn.FIXED_LINE_OR_MOBILE:
		return "FIXED_OR_MOBILE"
	case pn.TOLL_FREE:
		return "TOLL_FREE"
	case pn.PREMIUM_RATE:
		return "PREMIUM_RATE"
	case pn.SHARED_COST:
		return "SHARED_COST"
	case pn.VOIP:
		return "VOIP"
	case pn.PERSONAL_NUMBER:
		return "PERSONAL"
	case pn.PAGER:
		return "PAGER"
	case pn.UAN:
		return "UAN"
	case pn.VOICEMAIL:
		return "VOICEMAIL"
	default:
		return "UNKNOWN"
	}
}

func (l Lookup) Summary() string {
	if l.ParseError != "" {
		return fmt.Sprintf("%s — parse error", l.Input)
	}
	if !l.Valid {
		return fmt.Sprintf("%s — invalid", l.Input)
	}
	parts := []string{l.E164, l.LineType, l.Region}
	if l.Carrier != "" {
		parts = append(parts, l.Carrier)
	}
	if l.Geo != "" {
		parts = append(parts, l.Geo)
	}
	s := ""
	for i, p := range parts {
		if i > 0 {
			s += "  "
		}
		s += p
	}
	if l.Spam {
		s += "  ⚠ SPAM"
	}
	return s
}
