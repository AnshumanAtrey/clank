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

**Pre-built binaries** for Linux, macOS, and Windows (amd64 + arm64) are published to [GitHub Releases](https://github.com/AnshumanAtrey/clank/releases) on every tag — no Go toolchain required. Grab the archive for your platform, extract, and put `clank` on your `PATH`.

**Homebrew** (after the tap is set up — see *Maintainer notes* below):

```bash
brew install AnshumanAtrey/tap/clank
```

**Go install**:

```bash
go install github.com/AnshumanAtrey/clank@latest
```

**Build from source**:

```bash
git clone https://github.com/AnshumanAtrey/clank.git
cd clank
go build -o clank .
```

Verify the install:

```bash
clank --version
```

### Maintainer notes — Homebrew tap setup (one-time)

To enable `brew install AnshumanAtrey/tap/clank`, the maintainer needs to:

1. Create a public empty repo `AnshumanAtrey/homebrew-tap`.
2. Generate a Personal Access Token with `repo` scope.
3. Add the token as `HOMEBREW_TAP_GITHUB_TOKEN` in this repo's Actions secrets.
4. Tag a release: `git tag v0.1.0 && git push origin v0.1.0`.

GoReleaser (configured in `.goreleaser.yaml`) will then build cross-platform binaries, attach them to the GitHub Release, and auto-publish a Ruby formula to the tap repo. Subsequent `git tag vX.Y.Z` pushes update everything automatically.

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

### `clank scan <pattern>` — bulk `deep` over every candidate

The OG `+918115605xxx` workflow. Take a pattern (or list of phones), generate every candidate, filter early via libphonenumber to drop invalid numbers, then run `clank deep` over each survivor with rate-limit-aware sleeps so messenger lookups don't trigger bans. Each completed result is checkpointed to JSONL so a crash mid-scan never loses work.

```bash
clank scan +918115605xxx --region IN              # default: 30 candidates, 8s sleep, full deep
clank scan --max 100 --quick +918115605xxx        # offline + APIs + EDGAR; no messengers; no sleep
clank scan --workers 5 --quick +918115605xxx      # parallel quick scan
clank scan --input targets.txt --quick            # read phones from file
cat numbers.txt | clank scan --input - --json     # stdin → JSON output
clank scan --resume --out scan.jsonl +918115605xxx  # skip phones already done
clank scan --sample stride --max 50 +91811xxxxxx  # spread samples across the range
```

Defaults are deliberately conservative:
- `--max 30` — 30 candidates is safe under WhatsApp's ~50/min ban threshold even at 0s sleep.
- `--sleep 8s` — total ~6/min, well under all messenger limits.
- `--workers 1` — sequential, deterministic, predictable rate.

`--quick` flips into safe-bulk mode: skip messengers entirely (no ban risk), default sleep to 0, recommend `--workers 5+` for throughput. Use it for offline-only sweeps over hundreds of candidates.

Results are scored by enrichment hits (WhatsApp + Telegram count 2-3 each, ignorant/EDGAR/API valid counts 1-2 each, spam-flag subtracts 2) and ranked descending. Top-N shown in the human table; full set in `--json` and the JSONL checkpoint.

### `clank dorks <phone>` — Google-dork URL generator (PhoneInfoga's killer feature)

Generates Google search URLs across 5 buckets — one per platform/site, multiplied by every phone-format variant — to surface a number's footprint. No auth, no API, pure string templating. Roughly 30-160 URLs depending on which buckets you keep.

```bash
clank dorks +14155552671                              # print all (~160 URLs)
clank dorks --bucket social,reputation +14155552671   # subset
clank dorks --open --max 10 +14155552671              # launch first 10 in browser
clank dorks --json +14155552671 | jq                  # structured
```

Buckets:
- **social** — Facebook, LinkedIn, X/Twitter, Instagram, VK, Reddit, GitHub, Medium
- **disposable** — TextNow, TextFree, Hushed, Burner, Google Voice, Sideline
- **reputation** — shouldianswer.com, 800notes, tellows, mrnumber, whocallsme, Truecaller, Sync.me
- **individuals** — Whitepages, 411, AnyWho, Spokeo, BeenVerified, Radaris, TruePeopleSearch
- **general** — quoted-number Google searches, no `site:` filter

Each bucket × phone-variant combo emits one URL. Phone variants come from libphonenumber: E.164 (`+14155552671`), international (`+1 415-555-2671`), national (`(415) 555-2671`), digits-only (`14155552671`, `4155552671`).

### `clank history` — see what you've looked up

Every clank invocation appends one line to `~/.clank/history.jsonl` (only the first phone-shaped positional arg — never API keys or full args). Read it back with:

```bash
clank history                # last 20 entries
clank history --tail 50      # last 50
clank history --grep +91     # filter by substring (matches Cmd or Phone)
clank history --json | jq    # structured
clank history --path         # print path to history file
clank history --clear        # wipe history (with confirmation)
```

Set `CLANK_NO_AUDIT=1` in your env to disable history logging entirely.

### `clank deep <phone>` — run every source at once

The orchestrator. Takes one phone number, fans out to all six enrichment sources concurrently, and prints a unified report. Sources gracefully skip when not configured (no API key, not logged in, etc.) — they never block each other.

```bash
clank deep +14155552671                 # everything
clank deep --region IN 9181156055       # auto-prepend +91
clank deep --quick +14155552671         # skip messengers — fast (<3s)
clank deep --no-apis +14155552671       # skip Tier 2 even if keys are set
clank deep --json +14155552671 | jq     # structured output for scripting
```

Sources run by default:
- **Local** — libphonenumber + MCC/MNC + spam list (instant, no key)
- **Free APIs** — NumVerify, Veriphone, IPQS (skipped per-provider if its env var is unset)
- **Telegram** — phone-to-user (skipped if `~/.clank/telegram.session` missing or `TG_APP_ID` unset)
- **WhatsApp** — phone-to-user (skipped if `~/.clank/whatsapp.db` missing or unpaired)
- **ignorant** — Instagram / Snapchat / Amazon presence (no auth, but rate-limited)
- **SEC EDGAR** — full-text filings search (free, no auth)

Total runtime depends on which sources are configured. With nothing set up: ~1.5s (just Local + EDGAR + ignorant). With Telegram + WhatsApp paired and API keys set: ~5-10s.

```
Deep lookup — +14155552671

1. Local — libphonenumber + spam check
  +14155552671  FIXED_OR_MOBILE  US  San Francisco, CA

2. Free-tier APIs
  numverify    skipped — NUMVERIFY_KEY not set
  veriphone    valid · MOBILE · AT&T Mobility · United States
  ipqs         valid · MOBILE · AT&T · fraud=12

3. Telegram
  registered · user_id=123456789
  name      : Jane Doe
  username  : @jane_d

4. WhatsApp
  registered
  jid          : 14155552671@s.whatsapp.net
  about        : Hey there! I'm using WhatsApp.
  devices      : 2

5. Account presence (Instagram / Snapchat / Amazon)
  instagram  registered
  snapchat   ? — method broken (Snapchat redesigned web auth ~2024)
  amazon     ? rate-limited / inconclusive

6. SEC EDGAR full-text
  0 hits

Took 6.2s
```

### `clank imei <15-digit>` — IMEI decoder + TAC database

Pure offline. Validates Luhn checksum and resolves the 8-digit Type Allocation Code to manufacturer/model + year via an embedded database from [`MoazEb/tac-database`](https://github.com/MoazEb/tac-database) — **254,993 device records, March 2026 snapshot, MIT-licensed**. Covers everything from 2010-era Alcatels through iPhone 15 Pro Max and Galaxy S24+. Newer devices and obscure IoT modules may still be missing — clearly labelled when not found.

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

Optional: set `CLANK_EDGAR_UA="Your Name your-email@example.com"` to override the User-Agent and comply with [SEC's user-agent policy](https://www.sec.gov/os/accessing-edgar-data) when running at any volume. Default: `clank-osint build@atrey.dev`.

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

### `clank whatsapp <login | lookup | logout | reset>` — WhatsApp phone-to-user

Native Go via [`go.mau.fi/whatsmeow`](https://github.com/tulir/whatsmeow) — the same MTProto-of-WhatsApp client that powers Beeper's `mautrix-whatsapp` bridge. Returns registered boolean + canonical JID + business verified-name + "About" text + profile-pic URL + device count.

**Setup** (one-time):

1. `clank whatsapp login` — renders a QR code in your terminal. Open WhatsApp on your phone → Settings → Linked Devices → Link a Device → scan.
2. Session is persisted to `~/.clank/whatsapp.db` (pure-Go SQLite via `modernc.org/sqlite` — no C compiler needed).

**Lookup**:

```bash
clank whatsapp lookup +14155552671              # human-readable
clank whatsapp lookup --json +14155552671       # JSON
clank whatsapp logout                            # unlink server-side, delete local DB
clank whatsapp reset                             # delete local DB only (when session is invalid)
```

**2026 reality** (per research file `11-whatsapp-whatsmeow.md`):
- Your phone must stay online within 14 days or WhatsApp unlinks all companions and you'll need to re-pair.
- OSINT-style usage at >50 lookups/min triggers `events.TemporaryBan` (server-side, hard to predict thresholds). Use a burner WhatsApp account if you plan to query at volume.
- WhatsApp silently drops invalid numbers (whatsmeow issue #1086) — clank reconciles by `Query` field and surfaces "no response — likely invalid format" instead of false-negative-as-not-registered.
- `events.LoggedOut` (e.g. you signed out from your phone) auto-deletes the local device row; the next lookup will say "not paired" so you can re-pair.

## Roadmap

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

## Author

Built by **Anshuman Atrey** — software engineer, hackathon winner, builder of weird tools.

- Website: <https://atrey.dev>
- LinkedIn: <https://linkedin.com/in/anshumanatrey/>
- Email: <build@atrey.dev>

Bugs, ideas, war stories from your own scans — drop me a line. PRs welcome.

## License

MIT — © Anshuman Atrey · <https://atrey.dev>
