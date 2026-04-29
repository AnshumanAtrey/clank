package local

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed data/mcc-mnc-list.json
var mccmncRaw []byte

type Operator struct {
	Type        string `json:"type"`
	CountryName string `json:"countryName"`
	CountryCode string `json:"countryCode"`
	MCC         string `json:"mcc"`
	MNC         string `json:"mnc"`
	Brand       string `json:"brand,omitempty"`
	Operator    string `json:"operator,omitempty"`
	Status      string `json:"status,omitempty"`
	Bands       string `json:"bands,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

var (
	allOperators []Operator
	byISO        map[string][]Operator
)

func init() {
	if err := json.Unmarshal(mccmncRaw, &allOperators); err != nil {
		return
	}
	byISO = make(map[string][]Operator, 256)
	for _, op := range allOperators {
		key := strings.ToUpper(strings.SplitN(op.CountryCode, "-", 2)[0])
		if key == "" {
			continue
		}
		byISO[key] = append(byISO[key], op)
	}
}

func OperatorsInCountry(iso string) []Operator {
	if iso == "" {
		return nil
	}
	out := byISO[strings.ToUpper(iso)]
	ops := make([]Operator, 0, len(out))
	for _, o := range out {
		if o.Status == "" || strings.EqualFold(o.Status, "Operational") {
			ops = append(ops, o)
		}
	}
	return ops
}

func AllOperators() []Operator {
	return allOperators
}
