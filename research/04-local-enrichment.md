# 04 — Local Enrichment (No-API-Key Mode)

What we can extract from a phone number **purely offline** for `clank`. Everything here ships as Go dependencies or static data files — no signups, no rate limits, no network calls at runtime.

Scope: research-only. We pick what to ship in section 7.

---

## 1. libphonenumber via `nyaruka/phonenumbers`

The Go port of Google's libphonenumber. Maintained, MIT-licensed, ~1.5k stars, latest release `v1.7.2`. Single dependency that gives us 80% of what a phone-OSINT CLI needs.

```
go.mod:  github.com/nyaruka/phonenumbers v1.7.2
```

Important architectural note from reading the repo: there are **no separate sub-packages** for carrier, geocoder, and timezone. The Java/Python originals split these out, but the Go port exposes everything as top-level functions on the `phonenumbers` package (e.g., `phonenumbers.GetCarrierForNumber`, not `carrier.GetNameForNumber`). The `gen/prefix_to_*_bin.go` files are internal data, not importable. Plan accordingly.

### 1a. API surface a CLI cares about

**Parsing / validation**
- `Parse(input, defaultRegion string) (*PhoneNumber, error)`
- `ParseAndKeepRawInput(input, defaultRegion string) (*PhoneNumber, error)`
- `IsValidNumber(*PhoneNumber) bool`
- `IsValidNumberForRegion(*PhoneNumber, regionCode string) bool`
- `IsPossibleNumber(*PhoneNumber) bool`
- `IsPossibleNumberWithReason(*PhoneNumber) ValidationResult`

**Country / region**
- `GetRegionCodeForNumber(*PhoneNumber) string`             // e.g., "US"
- `GetRegionCodeForCountryCode(int) string`
- `GetCountryCodeForRegion(string) int`                     // "US" -> 1
- `GetRegionCodesForCountryCode(int) []string`              // 1 -> [US, CA, ...]
- `GetSupportedRegions() map[string]bool`
- `GetSupportedCallingCodes() map[int]bool`

**Line type**
- `GetNumberType(*PhoneNumber) PhoneNumberType`

`PhoneNumberType` constants: `FIXED_LINE`, `MOBILE`, `FIXED_LINE_OR_MOBILE`, `TOLL_FREE`, `PREMIUM_RATE`, `SHARED_COST`, `VOIP`, `PERSONAL_NUMBER`, `PAGER`, `UAN`, `VOICEMAIL`, `UNKNOWN`.

**Carrier (best-effort, original carrier only)**
- `GetCarrierForNumber(*PhoneNumber, lang string) (string, error)`

**Geo description**
- `GetGeocodingForNumber(*PhoneNumber, lang string) (string, error)` — city/region/country in `lang`

**Timezone**
- `GetTimezonesForNumber(*PhoneNumber) ([]string, error)` — IANA tz names

**Formatting**
- `Format(*PhoneNumber, PhoneNumberFormat) string` — `E164`, `INTERNATIONAL`, `NATIONAL`, `RFC3966`
- `FormatInOriginalFormat`, `FormatOutOfCountryCallingNumber`, `FormatNumberForMobileDialing`

**Helpers**
- `GetNationalSignificantNumber(*PhoneNumber) string`
- `GetLengthOfNationalDestinationCode(*PhoneNumber) int`
- `GetLengthOfGeographicalAreaCode(*PhoneNumber) int`
- `ConvertAlphaCharactersInNumber(string) string` — `1-800-FLOWERS` → `1-800-3569377`
- `IsAlphaNumber(string) bool`
- `IsValidShortNumber(*PhoneNumber) bool`
- `IsEmergencyNumber(string, regionCode string) bool`
- `ConnectsToEmergencyNumber(string, regionCode string) bool`

### 1b. End-to-end snippet — every enrichment libphonenumber gives us

```go
package main

import (
    "fmt"

    pn "github.com/nyaruka/phonenumbers"
)

func enrich(raw, defaultRegion string) {
    num, err := pn.Parse(raw, defaultRegion)
    if err != nil {
        fmt.Printf("parse error: %v\n", err)
        return
    }

    fmt.Println("input:           ", raw)
    fmt.Println("possible:        ", pn.IsPossibleNumber(num))
    fmt.Println("valid:           ", pn.IsValidNumber(num))
    fmt.Println("country code:    ", num.GetCountryCode())
    fmt.Println("national number: ", num.GetNationalNumber())
    fmt.Println("region:          ", pn.GetRegionCodeForNumber(num))
    fmt.Println("type:            ", typeName(pn.GetNumberType(num)))

    fmt.Println("E.164:           ", pn.Format(num, pn.E164))
    fmt.Println("international:   ", pn.Format(num, pn.INTERNATIONAL))
    fmt.Println("national:        ", pn.Format(num, pn.NATIONAL))
    fmt.Println("rfc3966:         ", pn.Format(num, pn.RFC3966))

    if c, err := pn.GetCarrierForNumber(num, "en"); err == nil && c != "" {
        fmt.Println("carrier (orig):  ", c)
    }
    if g, err := pn.GetGeocodingForNumber(num, "en"); err == nil && g != "" {
        fmt.Println("geo:             ", g)
    }
    if tzs, err := pn.GetTimezonesForNumber(num); err == nil {
        fmt.Println("timezones:       ", tzs)
    }
}

func typeName(t pn.PhoneNumberType) string {
    return map[pn.PhoneNumberType]string{
        pn.FIXED_LINE:           "FIXED_LINE",
        pn.MOBILE:                "MOBILE",
        pn.FIXED_LINE_OR_MOBILE:  "FIXED_LINE_OR_MOBILE",
        pn.TOLL_FREE:             "TOLL_FREE",
        pn.PREMIUM_RATE:          "PREMIUM_RATE",
        pn.SHARED_COST:           "SHARED_COST",
        pn.VOIP:                  "VOIP",
        pn.PERSONAL_NUMBER:       "PERSONAL_NUMBER",
        pn.PAGER:                 "PAGER",
        pn.UAN:                   "UAN",
        pn.VOICEMAIL:             "VOICEMAIL",
        pn.UNKNOWN:               "UNKNOWN",
    }[t]
}

func main() { enrich("+14155552671", "US") }
```

### 1c. Caveats

- `GetCarrierForNumber` returns the **originally-allocated** carrier. After number portability the answer may be wrong (see §3).
- `GetGeocodingForNumber` is a coarse string ("Mountain View, CA"), not coordinates.
- `GetTimezonesForNumber` returns an `[]string` of IANA names; for ranges that span multiple zones (e.g., country-level) you get all of them.
- `FIXED_LINE_OR_MOBILE` shows up in the US/Canada because allocation rules don't separate them — don't over-promise "this is a mobile."

---

## 2. MCC / MNC datasets

For deeper carrier intel than libphonenumber's prefix-to-original-carrier map, we can ship a static MCC+MNC table. Two solid sources:

### 2a. `musalbas/mcc-mnc-table` (CSV/JSON/XML, daily-updated)

Raw URL: `https://raw.githubusercontent.com/musalbas/mcc-mnc-table/master/mcc-mnc-table.csv`

Header + sample rows:
```
MCC,MCC (int),MNC,MNC (int),ISO,Country,Country Code,Network
289,649,88,2191,ge,Abkhazia,7,A-Mobile
412,1042,01,31,af,Afghanistan,93,AWCC
```

License: scraped from mcc-mnc.com. Not perfectly clean, but redistributable de facto and good enough for OSINT.

### 2b. `pbakondy/mcc-mnc-list` (richer JSON, Wikipedia-sourced)

Raw URL: `https://raw.githubusercontent.com/pbakondy/mcc-mnc-list/master/mcc-mnc-list.json`

Schema:
```json
{
  "type": "National",
  "countryName": "Hungary",
  "countryCode": "HU",
  "mcc": "216",
  "mnc": "30",
  "brand": "T-Mobile",
  "operator": "Magyar Telekom Plc",
  "status": "Operational",
  "bands": "GSM 900 / GSM 1800 / UMTS 2100 / LTE 800 / LTE 1800 / LTE 2600",
  "notes": "Former WESTEL, ..."
}
```

Has `status` (Operational/Not operational), `bands`, brand vs. operator. Better for OSINT. JSON only; no CSV.

### 2c. ITU Operational Bulletin

Authoritative source for country-code allocations and MNC delegations: <https://www.itu.int/pub/T-SP-OB>. PDFs only, no machine-readable bulk feed — we use it as ground-truth for sanity-checking the GitHub mirrors, not as a runtime dependency.

### 2d. Loading in Go

Vendor the JSON at build time (`go:embed`), parse once at startup:

```go
package mccmnc

import (
    _ "embed"
    "encoding/json"
)

//go:embed data/mcc-mnc-list.json
var raw []byte

type Record struct {
    Type, CountryName, CountryCode, MCC, MNC string
    Brand, Operator, Status, Bands, Notes    string
}

var byMCCMNC map[string]Record

func init() {
    var rs []Record
    _ = json.Unmarshal(raw, &rs)
    byMCCMNC = make(map[string]Record, len(rs))
    for _, r := range rs {
        byMCCMNC[r.MCC+"-"+r.MNC] = r
    }
}

func Lookup(mcc, mnc string) (Record, bool) {
    r, ok := byMCCMNC[mcc+"-"+mnc]
    return r, ok
}
```

**Important gap:** MCC+MNC keys an HLR record, not a phone number. From an E.164 number alone we can't derive MCC+MNC offline — we'd need an HLR lookup (paid network call) or a portability database. So this dataset is useful when the user *also* has IMSI/SIM data, or for the country-side bonus info. For pure-number input it's mainly a richer "what carriers exist in this country" reference.

---

## 3. Number portability — known gap

In countries with mobile number portability the original-prefix → carrier mapping that libphonenumber and MCC/MNC tables ship is wrong for any ported number. Real percentages of numbers that have ported, by country:

- **US (LRN)**: NPAC manages ~1B numbers; access is restricted to authorized carriers and select law-enforcement agencies via PortDataSource. **No public bulk feed.**
- **UK (NPD)**: portability info via Ofcom is *block-level* (range-holder/operator who currently routes each block), not number-level. CSV files: <https://www.ofcom.org.uk/phones-and-broadband/phone-numbers/numbering-data> — ZIP, ~3 MB, weekly updates. Useful but tells us "this 1000-block currently routes via X" not "+44 7…" status.
- **India (MNP)**: TRAI publishes aggregate stats, no public per-number lookup.

**Action for clank:** When we report carrier, label it clearly as "original-allocated carrier; portability not detected." Document the limitation in `clank --help`. We cannot fix this offline without scraping HLR data we shouldn't have.

The `nyaruka/phonenumbers` source tree includes a `GetSafeCarrierDisplayNameForNumber`-style guard — when a region supports portability, it conservatively returns empty for known mobile ranges. We should surface that in the UI ("carrier hidden — portability region").

---

## 4. Type-of-number heuristics

libphonenumber already classifies into the 12 `PhoneNumberType` values listed in §1a. That covers premium-rate, toll-free, shared-cost, VoIP, pager, voicemail, UAN — without any extra dataset.

What's missing from libphonenumber:

1. **Per-country premium-rate sub-buckets**: it tells you "PREMIUM_RATE" but not "this is an adult-content premium rate vs. a directory-services one". For UK we'd need to layer Ofcom's Part B/Part C CSVs (<https://www.ofcom.org.uk/phones-and-broadband/phone-numbers/numbering-data>) to differentiate.
2. **Newly allocated ranges**: metadata is regenerated from Google's XML on every release. If a country opens a new toll-free block today, we won't know until the next `nyaruka/phonenumbers` cut. Mitigation: our `clank metadata-update` subcommand runs `go run github.com/nyaruka/phonenumbers/cmd/buildmetadata` to fetch latest XML.
3. **Vanity / alpha numbers**: handled via `IsAlphaNumber` + `ConvertAlphaCharactersInNumber`. Already covered.
4. **Short codes / emergency**: `IsValidShortNumber`, `IsEmergencyNumber`, `ConnectsToEmergencyNumber`. Already covered.

For a v1 ship, libphonenumber alone is enough for type detection. Ofcom CSV layering is a "v2 nice-to-have" for the UK power-user.

---

## 5. Crowdsourced spam / scam lists

What's actually shippable as public-domain CSV/JSON:

- **`Oros42/phone-blacklist`** — Unlicense (public domain). CSV separated by `spams.csv` and `scams.csv`. International E.164 with ISO-prefixed short codes. <https://github.com/Oros42/phone-blacklist> — small, biased toward FR (originated for Free Mobile users), but legit and redistributable.
- **`jwoertink/blocked-numbers`** — `list.csv` of US robocall numbers. Small, MIT.
- **FTC DNC reported-calls dataset** — official FTC daily CSV of consumer-reported unwanted calls including originating number, robocall flag, subject, city/state. <https://www.ftc.gov/policy-notices/open-government/data-sets/do-not-call-data>. **This is the gold-standard public source for US.** Updates daily by noon ET. Public-domain US-government work.
- **GitHub topics `spam-numbers` / `scam-numbers`** — mostly tiny country-specific lists. Worth scanning at build time, low signal individually.

**Off-limits — explicitly do NOT ship or link:**
- The 2021 Facebook scrape (533M phone numbers paired to user profiles). Stolen PII.
- Truecaller's full crowdsourced spam DB. Their public-tag API is theirs; bulk re-use violates ToS and is mostly scraped PII anyway.
- Any "leaked carrier customer database." Lawsuit waiting.

`clank` should refuse to ingest a list it doesn't have a license for. We'll add a `clank list ingest --acknowledge-license` flag and ship only the FTC + Oros42 sets in-tree.

### Loading FTC CSV in Go

```go
package spam

import (
    _ "embed"
    "encoding/csv"
    "strings"
)

//go:embed data/ftc-dnc-recent.csv
var raw string

var bad map[string]struct{}

func init() {
    bad = map[string]struct{}{}
    r := csv.NewReader(strings.NewReader(raw))
    rows, _ := r.ReadAll()
    for _, row := range rows[1:] { // skip header
        bad[normalize(row[0])] = struct{}{}
    }
}

func normalize(s string) string { /* strip non-digits, prepend country */ return s }

func IsReported(e164 string) bool { _, ok := bad[e164]; return ok }
```

We refresh the embedded CSV in CI weekly via a small fetch script.

---

## 6. Country-pattern E.164 prefix tree

How libphonenumber resolves a leading-digit string to a region:

1. Strip non-digits and any `+` / IDD prefix (`00`, `011`, etc.).
2. Walk a trie of known country calling codes (length 1, 2, or 3 digits — `1`, `7`, `20`, `27`, `30`...`998`, `1242`, `1809`, `1876`...).
3. Multiple regions can share a country code (NANP `+1` → US, CA, plus a long list of Caribbean territories). When this happens, libphonenumber consults per-region `leading_digits` patterns and `national_prefix` rules from the metadata, then picks the unique region whose prefix pattern matches.
4. The trie + per-region patterns are baked into `metadata_bin.go`; we never have to reimplement this.

If we ever need to roll our own (e.g., for a fast pre-validation before calling Parse), `nyaruka/phonenumbers` exposes `GetRegionCodesForCountryCode(int) []string` and `GetSupportedCallingCodes() map[int]bool`. With those we can build a tiny in-memory trie ourselves at startup. The `biter777/countries` Go package has ITU-T E.164 IDD codes alongside ISO-3166 if we want a fully independent fallback table (no libphonenumber dependency for the prefix step).

```go
import "github.com/nyaruka/phonenumbers"

// Naive longest-prefix country-code lookup.
func splitCC(digits string) (cc int, rest string) {
    for n := 3; n >= 1; n-- {
        if len(digits) < n {
            continue
        }
        var v int
        fmt.Sscanf(digits[:n], "%d", &v)
        if phonenumbers.GetRegionCodeForCountryCode(v) != "ZZ" {
            return v, digits[n:]
        }
    }
    return 0, digits
}
```

`"ZZ"` is libphonenumber's sentinel for "unknown region."

---

## 7. Pure-offline feature roadmap for `clank`

What we ship in **one PR**, with zero external API and zero signups:

1. **`clank lookup <number>`** — accepts E.164 or local format with `--region` flag.
2. From `nyaruka/phonenumbers` alone, output:
   - validity (`possible`, `valid`)
   - country / region / calling code
   - line type (one of the 12 `PhoneNumberType` constants, human-readable)
   - original carrier guess (with explicit "may be ported" note in mobile portability regions)
   - geographic description string
   - IANA timezones (we can also localize the user's clock difference vs. the number's tz)
   - formatted variants: E.164, INTERNATIONAL, NATIONAL, RFC3966
3. **MCC/MNC bonus block** — when `--country` info is requested, embed `pbakondy/mcc-mnc-list` and show all known operators in that country with `status: Operational`. Useful for "who could this be?" intel.
4. **Spam check** — embed `Oros42/phone-blacklist` + a snapshot of the FTC DNC last-90-days CSV. If number matches: print a `reported: true` flag with the source name and last-seen date. CI refreshes the snapshot weekly.
5. **`clank metadata-update`** subcommand — wraps `go run github.com/nyaruka/phonenumbers/cmd/buildmetadata` so power users can pull latest libphonenumber XML without waiting for a `clank` release.
6. **Honest defaults** — every output that depends on guessable data (carrier, geo, type for ported numbers) gets a `confidence: low|medium|high` field. We never lie about portability.
7. **Output formats** — `--json` (machine), default human, `--quiet` (just the bottom-line type+region for piping).

What we explicitly *don't* ship in v1 (deferred or out-of-scope):

- HLR live lookup (paid, network).
- LRN current-carrier resolution (no public dataset; restricted to NPAC users).
- Truecaller/Hiya-style identity tagging (their data, scraped PII risk).
- Anything sourced from the 2021 Facebook leak or similar dumps.

Total binary size estimate with embedded data: libphonenumber ~5 MB metadata + MCC/MNC JSON ~500 KB + spam CSVs ~1 MB. Well under 10 MB for a Go binary — fine.

### `go.mod` additions for this PR

```
require (
    github.com/nyaruka/phonenumbers v1.7.2
)
```

That's it. MCC/MNC + spam lists ship as `go:embed` static files; no extra Go deps.

---

## Sources

- [nyaruka/phonenumbers GitHub](https://github.com/nyaruka/phonenumbers)
- [nyaruka/phonenumbers GoDoc on pkg.go.dev](https://pkg.go.dev/github.com/nyaruka/phonenumbers)
- [phonenumbers.go source on main](https://github.com/nyaruka/phonenumbers/blob/main/phonenumbers.go)
- [musalbas/mcc-mnc-table](https://github.com/musalbas/mcc-mnc-table)
- [pbakondy/mcc-mnc-list](https://github.com/pbakondy/mcc-mnc-list)
- [P1sec/MCC_MNC](https://github.com/P1sec/MCC_MNC)
- [ITU-T E.164 list of assigned country codes](https://www.itu.int/pub/t-sp-e.164d)
- [NPAC (US LRN administrator)](https://www.npac.com/)
- [Ofcom UK numbering data downloads](https://www.ofcom.org.uk/phones-and-broadband/phone-numbers/numbering-data)
- [TRAI Mobile Number Portability page](http://www.trai.gov.in/telecom/mobile-number-portability)
- [FTC Do Not Call reported-calls dataset](https://www.ftc.gov/policy-notices/open-government/data-sets/do-not-call-data)
- [Oros42/phone-blacklist (Unlicense)](https://github.com/Oros42/phone-blacklist)
- [jwoertink/blocked-numbers](https://github.com/jwoertink/blocked-numbers)
- [biter777/countries (Go ISO/E.164 codes)](https://github.com/biter777/countries)
- [Google libphonenumber FAQ (prefix matching)](https://github.com/google/libphonenumber/blob/master/FAQ.md)
