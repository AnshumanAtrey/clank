![Clank Logo](./clank-preview-image.png)

**Clank** is a command-line tool for phone-number reconnaissance. Give it a partial number with `x` placeholders for unknown digits — it expands every combination, then runs OSINT enrichment locally (libphonenumber + carrier maps + spam blacklists) or via a free-tier API of your choice.

```
$ clank 918115605xxx
[banner]

Generated 1000 candidates for pattern 918115605xxx

What would you like to do?
  [1] Local lookup       — instant, no API key (libphonenumber + spam check)
  [2] Free API lookup    — needs your API key
  [3] Print all combos   — original behaviour
  [q] Quit

> 1
Local lookup — 1000 valid / 1000 total
NUMBER         VALID  TYPE    REGION  CARRIER      GEO    TZ             SPAM
+918115605000  ✓      MOBILE  IN      BSNL MOBILE  India  Asia/Calcutta  -
+918115605001  ✓      MOBILE  IN      BSNL MOBILE  India  Asia/Calcutta  -
…
```

## Features

- **Pattern combinator** — fill `x`-placeholders 0..9 to generate up to 100,000 candidates from a partial number.
- **Local OSINT (no API key)** — Google's libphonenumber via `nyaruka/phonenumbers` gives validity, region, line type, carrier-of-record, geo, timezone, plus E.164/INTERNATIONAL/NATIONAL/RFC3966 formats.
- **MCC/MNC operator list** — embedded `pbakondy/mcc-mnc-list`; shows every operational network in the detected country.
- **Spam check** — embedded `Oros42/phone-blacklist` + `jwoertink/blocked-numbers`; flags known reported numbers.
- **Free-tier API lookup** — pluggable providers (NumVerify, Veriphone, IPQualityScore) with worker pool, quota guard, and clean key handling.
- **JSON output** — `--json` for piping to `jq` / scripting.
- **Single static binary** — all data files embedded with `go:embed`. No runtime downloads, no install scripts.

## Install

```bash
go install github.com/AnshumanAtrey/clank@latest
```

Or build from source:

```bash
git clone https://github.com/AnshumanAtrey/clank.git
cd clank
go build -o clank .
```

## Usage

```bash
clank <pattern>                       # interactive menu (default)
clank --local <pattern>               # skip menu — pure-offline lookup
clank --provider numverify <pattern>  # skip menu — call NumVerify
clank --gen <pattern>                 # legacy: just print all combinations
clank --json --local <pattern>        # JSON output for scripting
```

### Pattern format

- Digits `0-9` for known positions
- `x` or `X` for unknown digits
- Optional leading `+` for E.164 international form
- Examples: `918115605xxx`, `+1415xxxxxxx`, `+44xxxxxxxxxx`

### Region hint

For local-format input without a country code, pass `--region`:

```bash
clank --local --region IN 9181156052      # treat as Indian number
clank --local --region US 4155552671      # treat as US number
```

### API providers

Set the relevant env var (or pass `--key`), then run with `--provider <name>`:

| Provider          | Env var          | Free tier   | Standout field      |
|-------------------|------------------|-------------|---------------------|
| NumVerify         | `NUMVERIFY_KEY`  | 100/mo      | mainstream          |
| Veriphone         | `VERIPHONE_KEY`  | 1,000/mo    | 249 countries       |
| IPQualityScore    | `IPQS_KEY`       | 1,000/mo    | `fraud_score` 0-100 |

Sign up:
- IPQualityScore — https://www.ipqualityscore.com/create-account
- Veriphone — https://veriphone.io/signup
- NumVerify — https://numverify.com/signup/free

```bash
export IPQS_KEY=your_key_here
clank --provider ipqs --region US 4155552671
```

Before any API run, clank shows a confirmation prompt with the call count so it never silently burns your monthly quota:

```
About to call ipqs 47 times. Continue? [y/N]
```

API calls run in a worker pool (default 5 concurrent, override with `--workers N`).

## Project structure

```
clank/
├── main.go                                # CLI entry, menu, output rendering
├── internal/
│   ├── pattern/                           # combinator
│   ├── local/                             # libphonenumber wrapper, MCC/MNC, spam list
│   │   └── data/                          # embedded JSON / CSV
│   └── api/                               # NumVerify / Veriphone / IPQS providers
└── research/                              # OSINT research notes (5 files)
```

## Subcommands

### `clank imei <15-digit>` — IMEI decoder + TAC database

Pure offline. Validates Luhn checksum and resolves the 8-digit Type Allocation Code to manufacturer/model via an embedded database (`jpanganiban/imei`, ~25,900 device records, Sep 2014 snapshot). Newer devices won't be in the DB — clearly labelled when missing.

```bash
clank imei 352099001761481                  # → Alcatel OneTouch 332 (Luhn ✓)
clank imei 35209900176148                   # 14-digit input — computes the check digit
clank imei --json 352099001761481           # full JSON for scripting
```

### `clank edgar <query>` — SEC EDGAR full-text search

Free, no API key. Hits the SEC's full-text search engine (`efts.sec.gov`). Pivots a phone number, address, or any string to every 10-K/10-Q/Form-D/Form-144 filing referencing it. Surfaces business associations and insider activity.

```bash
clank edgar --hits 5 "(415) 421-7900"          # → 453 Williams Sonoma filings
clank edgar --forms 10-K --hits 10 "Apple Park Way"  # filter by form type
clank edgar --json --hits 20 "<your-phone>"         # JSON for jq
```

Optional: set `CLANK_EDGAR_UA="Your Name your-email@example.com"` to comply with [SEC's user-agent policy](https://www.sec.gov/os/accessing-edgar-data) when running at any volume.

### `clank telegram <subcommand>` — Telegram phone-to-user

Resolves a phone number to a Telegram user (id, name, username, premium/verified flags) via MTProto's `contacts.resolvePhone`. Built on `gotd/td`. Maintained by the same authors who power [Bellingcat's investigations](https://github.com/bellingcat/telegram-phone-number-checker).

**Setup** (one-time):

1. Get an `app_id` and `app_hash` from <https://my.telegram.org/apps>
2. `export TG_APP_ID=12345`
3. `export TG_APP_HASH=abc...`
4. `clank telegram login` — enters phone → SMS code → optional 2FA password. Session is saved to `~/.clank/telegram.session`.

**Lookup**:

```bash
clank telegram lookup +14155552671          # human-readable
clank telegram lookup --json +14155552671   # JSON
clank telegram logout                        # clears session
```

Returns: user ID, first/last name, username, profile-photo presence, premium / verified / bot / restricted flags. Honours Telegram's privacy settings — users who hide phone-search will return `not registered`. Subject to `FLOOD_WAIT_X` rate-limits at scale.

### `clank ignorant <phone>` — Phone presence on Instagram / Snapchat / Amazon

Pure HTTP probes — no auth, no API key. Native-Go port of [`megadose/ignorant`](https://github.com/megadose/ignorant). Three concurrent goroutines, each hitting a platform's signup / login endpoint to determine whether the phone is registered.

```bash
clank ignorant +14155552671                            # all 3 checks
clank ignorant --only instagram,amazon +14155552671    # subset
clank ignorant --json +14155552671                     # JSON output
```

**2026 reality:**
- **Instagram**: endpoint still answers; aggressive 429 throttling per IP. Rotate IPs/VPN for higher throughput.
- **Snapchat**: web auth was redesigned; CSRF token moved from cookie to JS-side. Returns "method broken" until a new technique is found.
- **Amazon**: works inconsistently due to anti-bot DOM drift. Returns "yes / no / indeterminate" honestly.

This matches research file `07-phone-to-identity.md` — we ship the endpoints we found, surface their real-world failure modes rather than pretending they always work.

## Roadmap

- **WhatsApp presence** via `tulir/whatsmeow` (paired QR session — boolean + JID + business name + status text + profile pic + device count)
- Truecaller-import flow for users with a paired Android device
- Additional providers: Twilio Lookup v2, Telnyx, SignalWire, HLR-Lookups
- `clank dorks` — Google-dork URL generator (PhoneInfoga's 5-bucket taxonomy)
- IMEI TAC database refresh (the embedded snapshot is from 2014)
- Snapchat CSRF rediscovery — the JS-side bearer flow needs tracing

See `research/` for the full landscape — 10 markdown files mapping the OSINT-phone-tool ecosystem and unprecedented verticals.

## Limitations

- `GetCarrierForNumber` returns the **originally-allocated** carrier. After number portability the answer may be wrong; libphonenumber suppresses the carrier name in known-portability regions (US/UK/IN mobile). When that happens, clank prints `-` rather than guessing.
- Google-dork generation, messaging-app presence, and live HLR are roadmap items, not in v1.

## Disclaimer

This tool is for legitimate use — fraud investigation, journalism, security research, knowing what's online about your own number. Comply with applicable laws and the providers' Terms of Service.

## License

MIT
