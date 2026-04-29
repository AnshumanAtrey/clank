# Phone-OSINT Tooling Survey (April 2026)

**Goal:** catalog actively-maintained, open-source phone-number reconnaissance tools that run locally, then pick which techniques to clone into `clank` (Go CLI, currently a pattern-fill generator).

**Scope rules:**
- Only tools whose primary input is a phone number, or that are commonly chained into phone OSINT.
- Pure social-media-username-only tools (Sherlock, Osintgram, Toutatis) are referenced where they share techniques but not deeply profiled.
- Last commit dates verified via GitHub API on 2026-04-29.
- API/keys: catalog only public APIs and reverse-engineered endpoints already used by mainstream OSINT tools. No leaked keys, no auth bypasses.

---

## 1. sundowndev/phoneinfoga

- **GitHub:** https://github.com/sundowndev/phoneinfoga
- **Language / Stars / Last commit:** Go / 16.3k / 2026-01-06
- **License:** GPL-3.0
- **Status note:** README says "stable but unmaintained" — but it still got a commit in 2026 and remains the de facto reference implementation.

**What it does.** Multi-stage phone-number reconnaissance framework. Takes an E.164 number, runs a chain of "scanners" (local + remote), and produces JSON / console output. Ships a CLI, REST API, and web UI.

**Concrete techniques.**
- **Local scanner:** parses with `nyaruka/phonenumbers` (Go port of Google libphonenumber) → validity, E.164, national format, country, carrier hint.
- **Numverify scanner:** queries `apilayer.com/marketplace/numverify-api` (`NUMVERIFY_API_KEY` env) → carrier, line type, location.
- **OVH scanner:** hits OVH Telecom REST API (`api.ovh.com/console/#/telephony/number/detailedZones`) for FR/BE/UK/ES/CH numbers → number-range owner, city, zip.
- **Google Search dorks:** programmatically generates dorks via `sundowndev/dorkgen` across five buckets — social media, disposable-number providers (hs3x.com, receive-sms-now.com, smslisten.com, freesmscode.com…), reputation sites, individual people-search, general.
- **Google CSE scanner:** uses Google Custom Search JSON API (`google.golang.org/api/customsearch/v1`) to actually run the dorks server-side and return ranked URLs.

**Install:** `brew install phoneinfoga` / `go install github.com/sundowndev/phoneinfoga/v2/cmd/phoneinfoga@latest` / `docker pull sundowndev/phoneinfoga`.

**Standout for clank:**
1. Modular **Scanner interface** (`Name`, `Description`, `DryRun`, `Run`) — clone this directly; lets us add scanners independently.
2. The dork-bucket taxonomy + URL generator — high-value, zero auth needed.

---

## 2. phoneintel/phoneintel

- **GitHub:** https://github.com/phoneintel/phoneintel
- **Language / Stars / Last commit:** Python / 116 / 2024-09-11
- **License:** GPL-3.0

**What it does.** Phone-number lookup CLI with both bulk-from-file and inline-extraction-from-text modes. Hits a wider set of free public sites than phoneinfoga.

**Concrete techniques.**
- Uses `phonenumbers` (Python libphonenumber) for parse / region / carrier / timezone.
- **Tellows scrape:** pulls the public Tellows page for the number → score, call type, complaint counts.
- **SpamCalls.net scrape:** spam-risk score, last-activity date, latest report.
- **c-qui.fr scrape:** French-only, returns carrier and request count.
- **Neutrino API integration:** optional commercial API key.
- **OpenStreetMap deep-link generator** (`--map`): builds a URL pre-centred on the area-code's lat/long.
- **Dork generator with categories:** `social_networks | forums | classifieds | ecommerce | news | blogs | job_sites | pastes | reputation | phone_directories | people_search | all` — broader than phoneinfoga's 5.

**Install:** `pip install phoneintel`.

**Standout for clank:**
1. Twelve dork categories — phoneinfoga only has five.
2. Bulk-from-file + inline-from-text-string ingestion modes — easy to add to clank's existing arg parser.

---

## 3. megadose/ignorant

- **GitHub:** https://github.com/megadose/ignorant
- **Language / Stars / Last commit:** Python / 1.6k / 2024-07-27
- **License:** GPL-3.0

**What it does.** Asks "is this phone number registered on site X?" without alerting the target — by abusing public registration / login flows. Pure existence checks, no PII returned.

**Concrete techniques.**
- **Instagram:** POSTs `i.instagram.com/api/v1/users/lookup/` with HMAC-SHA256-signed body using `IG_SIG_KEY` (the public mobile-app key); response `"No users found"` ⇒ not registered.
- **Snapchat:** POSTs to `accounts.snapchat.com/accounts/validate_phone_number` with a fresh `xsrf_token` cookie; status `TAKEN_NUMBER` vs `OK` distinguishes registered vs free.
- **Amazon:** scrapes the OpenID sign-in page form, posts the phone number as `email`; HTML element `#auth-password-missing-alert` ⇒ registered.
- All checks are async via `httpx + trio`.
- Each module returns `{name, domain, method, rateLimit, exists}` — clean, machine-readable.

**Install:** `pip3 install ignorant`.

**Standout for clank:**
1. The **per-site existence-check** pattern — adapter with a uniform `{exists, rateLimit}` return is straightforward to port to Go HTTP clients.
2. The Snapchat/Instagram exact request bodies are documented in code — no reverse engineering required.

---

## 4. martinvigo/email2phonenumber

- **GitHub:** https://github.com/martinvigo/email2phonenumber
- **Language / Stars / Last commit:** Python / 2.6k / 2024-07-26
- **License:** MIT

**What it does.** Inverse OSINT: given an email, derives the target's phone number by abusing password-reset masked hints. Three modes: `scrape`, `generate`, `bruteforce`.

**Concrete techniques.**
- **scrape:** triggers password-reset on Ebay / Lastpass / Amazon / Twitter; parses the masked hint (e.g. `***-***-12-34`) to learn last digits.
- **generate:** builds a candidate number list from a country's official Phone Numbering Plan + a mask string like `555XXX1234`. Directly relevant to clank — this is exactly what clank already does today, just with country-plan validity filtering on top.
- **bruteforce:** iterates the candidate list, posting password-reset on a different site that returns the masked email; correlates back to the target email.

**Install:** `git clone … && pip3 install beautifulsoup4 requests`.

**Standout for clank:**
1. **Country-numbering-plan-aware mask expansion** — clank currently expands `xxx` blindly; filtering by the destination country's valid prefix ranges would cut output dramatically.
2. Author's newer "Phonerator" tool (referenced in README) does this at scale — worth studying.

---

## 5. dsonbaker/email2whatsapp

- **GitHub:** https://github.com/dsonbaker/email2whatsapp
- **Language / Stars / Last commit:** Go / 182 / 2025-10-31
- **License:** MIT

**What it does.** Email-to-WhatsApp pipeline written in Go. Scrapes leaked digits from password-reset flows (Magalu, PayPal, PagBank, Mercado Livre, Rappi, Uber), generates valid-number candidates, then verifies each against WhatsApp.

**Concrete techniques.**
- Uses `tulir/whatsmeow` (multidevice WhatsApp protocol library) — connects via QR-pairing, then issues `IsOnWhatsApp` queries for each candidate. Rate-limited but no API key needed.
- Saves WhatsApp profile photos to `./numberphone/profile/` for visual matching.
- Brute-force module per provider: Mercado Livre, PayPal, Twitter, Google, Microsoft, Uber — each leaks different shapes of email hint.

**Install:** `go install -v github.com/dsonbaker/email2whatsapp@latest`.

**Standout for clank:**
1. **WhatsApp existence + profile-photo fetch** via whatsmeow — clank is Go, this drops in cleanly.
2. The "verify candidates against WhatsApp" step is exactly the missing OSINT layer for clank's current generated-numbers output.

---

## 6. Lucksi/Mr.Holmes

- **GitHub:** https://github.com/Lucksi/Mr.Holmes
- **Language / Stars / Last commit:** Python / 3.3k / 2026-02-21
- **License:** GPL-3.0

**What it does.** Multi-target OSINT framework (domain, username, phone) with proxy rotation, GUI/CLI, and WhoIS-XML-API integration.

**Concrete techniques.**
- For phone: libphonenumber parse + a built-in **rotating proxy manager** (`Proxies/` folder) so dorks/scraping don't get IP-blocked.
- Google dorks across multiple categories (similar bucket idea to phoneinfoga).
- Configurable per-feature via `Configuration/Configuration.ini`.
- GUI auth (login.json) for multi-user deployments.

**Install:** `git clone https://github.com/Lucksi/Mr.Holmes && bash install.sh`.

**Standout for clank:**
1. **Rotating-proxy support** — phone OSINT scrapes get rate-limited fast. A simple proxy-list reader + round-robin would harden every other scanner.
2. INI-file config pattern for declaring API keys / proxy lists / enabled scanners.

---

## 7. spider863644/PhoneNumber-OSINT

- **GitHub:** https://github.com/spider863644/PhoneNumber-OSINT
- **Language / Stars / Last commit:** Python / 646 / 2025-10-21
- **License:** MIT

**What it does.** Termux-friendly interactive menu CLI for basic phone-number metadata + bulk-extract-from-file.

**Concrete techniques.**
- Pure `phonenumbers` (libphonenumber) — `geocoder.description_for_number`, `carrier.name_for_number`, `timezone.time_zones_for_number`.
- Regex extracts numbers from a pasted text blob or uploaded file.
- Validator mode (yes/no, no enrichment).

**Install:** `git clone … && pip install -r requirements.txt`.

**Standout for clank:**
1. **Number-extraction-from-text-blob** with regex — useful as a chained mode (`clank --extract dump.txt`).
2. Confirms libphonenumber is the universal foundation — clank should adopt `nyaruka/phonenumbers` early.

---

## 8. TermuxHackz/X-osint

- **GitHub:** https://github.com/TermuxHackz/X-osint
- **Language / Stars / Last commit:** Python / 2.2k / 2026-02-20
- **License:** GPL-3.0

**What it does.** Kitchen-sink OSINT framework (IP, email, phone, VIN, subdomains, CVE search, etc.). Phone module is a thin libphonenumber wrapper, but the framework architecture is interesting.

**Concrete techniques.**
- Phone: country, region, timezone, carrier via libphonenumber.
- Email: HIBP / breach lookup, reverse-search.
- Image: EXIF metadata extraction.
- Bash + Python hybrid; modular `setup.sh` per platform (Termux / Linux / Mac).

**Install:** `git clone https://github.com/TermuxHackz/X-osint && bash setup.sh`.

**Standout for clank:**
1. Single-CLI multi-target architecture — useful UX precedent if clank later expands beyond phone.
2. The "platform-specific installer" pattern is good for distribution beyond `go install`.

---

## 9. N0rz3/Inspector

- **GitHub:** https://github.com/N0rz3/Inspector
- **Language / Stars / Last commit:** Python / 188 / 2023-06-17 (older but clean reference)
- **License:** GPL-3.0

**What it does.** France-focused phone-number OSINT — wraps libphonenumber, adds reputation lookup and integrates `ignorant` modules in one run.

**Concrete techniques.**
- libphonenumber for validity / type (mobile/line/surcharged).
- Calls `ignorant`'s amazon + instagram modules inline.
- **free-lookup.net format generator:** prints all common formats (national, international, E.164, `tel:` URI).
- **Reputation lookup:** scrapes a French reputation aggregator for caller-ID strings (e.g. "Carrefour Express 0666666666").

**Install:** `git clone https://github.com/N0rz3/Inspector && pip install -r requirements.txt`.

**Standout for clank:**
1. **Multi-format printer** (E.164 / national / international / `tel:` URI / dash-separated) — trivial to add, high UX value.
2. Pattern of *embedding* other tools (ignorant) instead of duplicating — clank can shell out to `ignorant` early, port natively later.

---

## 10. sumithemmadi/truecallerpy

- **GitHub:** https://github.com/sumithemmadi/truecallerpy
- **Language / Stars / Last commit:** Python / 154 / 2024-05-04
- **License:** MIT

**What it does.** Python CLI/library for the unofficial Truecaller mobile API. Returns name, email, tags associated with a number — Truecaller's database is the largest crowdsourced caller-ID dataset in the world.

**Concrete techniques.**
- One-time login: phone + OTP → bearer token (`installationId`) saved to `~/.truecallerpy.json`.
- `truecallerpy -s +1234567890` issues an authenticated GET to the mobile search endpoint.
- Bulk mode reads numbers from file, one per line.
- Returns the full Truecaller JSON: name, addresses, emails, tags, spam-score.

**Install:** `pip install truecallerpy`.

**Standout for clank:**
1. **Truecaller is the single most valuable phone OSINT data source** — the auth flow is well documented here and would be ~150 lines of Go.
2. The **persistent-token-on-disk** pattern is clean and reusable for any other auth'd source clank adds.

---

## 11. HackUnderway/whatslookup

- **GitHub:** https://github.com/HackUnderway/whatslookup
- **Language / Stars / Last commit:** Python / 124 / 2025-09-12
- **License:** MIT

**What it does.** Six-endpoint WhatsApp OSINT CLI built on the RapidAPI "WhatsApp OSINT" community API. Endpoints: `about`, `base64` (profile pic), `business`, `devices`, `doublecheck` (existence), `privacy`.

**Concrete techniques.**
- Loads `RAPIDAPI_KEY` from `.env`.
- Each endpoint = one `requests.get()` to the RapidAPI proxy.
- Saves profile pics as `.jpg`.
- Validates phone format before query.
- Detects "no profile photo / hidden" state — useful signal.

**Install:** `git clone https://github.com/HackUnderway/whatslookup && pip install -r requirements.txt`.

**Standout for clank:**
1. The **6 WhatsApp signals** (status text, profile pic, business-account flag, linked-device count, privacy settings, exists-check) define a target schema for clank's WhatsApp output.
2. RapidAPI integration is one HTTP call — cheaper than wrangling whatsmeow if user has a key.

---

## 12. laramies/theHarvester

- **GitHub:** https://github.com/laramies/theHarvester
- **Language / Stars / Last commit:** Python / 16.1k / 2026-04-28
- **License:** AGPL-3.0
- **Phone-relevance:** Indirect — domain/email harvester, but its ingest patterns and 60+ passive modules are the gold standard for OSINT plumbing.

**What it does.** Pulls names, emails, IPs, subdomains, URLs from 60+ passive sources (Baidu, Brave, Censys, crt.sh, BeVigil, criminalip, etc.). Phone numbers surface in Bevigil mobile-app scans.

**Concrete techniques.**
- Pluggable async-module architecture; each source = one Python file with a `search()` async method.
- Passive vs active separation — never queries the target directly.
- `uv`-based modern dep management.
- Built-in rate-limit + retry middleware.

**Install:** `git clone … && uv sync && uv run theHarvester`.

**Standout for clank:**
1. The **passive-module registry pattern** — each source declares its name, deps, env-key. Clean and idiomatic to port to Go.
2. Bevigil's mobile-app OSINT API is a phone-adjacent source clank could query directly.

---

## 13. p1ngul1n0/blackbird

- **GitHub:** https://github.com/p1ngul1n0/blackbird
- **Language / Stars / Last commit:** Python / 6.0k / 2025-07-13
- **License:** none stated

**What it does.** Username + email account search across hundreds of social networks (now also accepts phone-adjacent identifiers). Often paired with phoneinfoga in workflows.

**Concrete techniques.**
- JSON-driven site list (`data.json`) — same architecture as Sherlock and WhatsMyName, swappable.
- Async HTTP via `httpx`.
- "Profile-data" parsing (extracts found username's bio/photo from each match).
- Self-updating site list from a remote URL.

**Install:** `git clone https://github.com/p1ngul1n0/blackbird && pip install -r requirements.txt`.

**Standout for clank:**
1. **JSON site-definition file** — separating site rules from code lets non-Go contributors add scanners.
2. **Self-update from remote JSON** — site detection patterns rot fast; auto-pulling a community list keeps clank fresh without releases.

---

## 14. The-Osint-Toolbox/Telephone-OSINT

- **GitHub:** https://github.com/The-Osint-Toolbox/Telephone-OSINT
- **Language / Stars / Last commit:** Markdown-only / 488 / 2025-10-05

**What it does.** Curated link list, not a tool — but it's the single best up-to-date reference for free web-based phone-OSINT services (free-lookup.net, hlrlookup.com, ipqualityscore, hcaptcha-bypass-friendly endpoints, regional carrier-portability lookups, etc.).

**Use as input to clank:** mine this list to seed the dork-targets and per-region scanner ideas for clank's roadmap.

---

## 15. nyaruka/phonenumbers (foundation library, not a tool)

- **GitHub:** https://github.com/nyaruka/phonenumbers
- **Language / Stars / Last commit:** Go / 1.5k / 2026-04-27
- **License:** MIT

Go port of Google's libphonenumber. Not an OSINT tool itself but is the **dependency every Go phone tool uses** (phoneinfoga, email2whatsapp, stevemurr/reverse-phone-lookup). Adding `import "github.com/nyaruka/phonenumbers"` is the single highest-leverage line clank can add: instant validation, E.164 / national / international formatting, country detection, line-type classification (mobile vs landline vs VoIP vs toll-free), carrier hint, and timezone — all offline and free.

**Install (as Go lib):** `go get github.com/nyaruka/phonenumbers`.

---

## Top 5 features to port to clank

Ranked by **(ease of implementation) × (value)** — easiest + highest-impact first.

### 1. Adopt `nyaruka/phonenumbers` for parse/validate/format/classify
**Effort:** ~1 hour. **Value:** Foundation. Without this, clank can't even tell that `91-9999-9999` and `+919999999999` are the same number, or that `+1-800-FLOWERS` is toll-free. Every other tool in the survey uses libphonenumber. After integration, clank's pattern-fill output can be filtered to only valid numbers — likely cutting `xxx`-expansion noise by 90%+ for most country codes.

### 2. Google-dork URL generator with phoneinfoga's bucket taxonomy
**Effort:** ~2 hours. **Value:** Massive. Generates clickable Google search URLs scoped to social media / disposable-providers / reputation / individuals / general. Zero auth, zero rate limit (the user clicks through), and instantly transforms clank from a number generator into an OSINT starter. The dork strings are public in `phoneinfoga/lib/remote/googlesearch_scanner.go` — literally a translation exercise.

### 3. WhatsApp existence check via `tulir/whatsmeow`
**Effort:** ~1 day (QR-pair flow + IsOnWhatsApp loop). **Value:** This is the step that turns clank's pattern-generated 1000 candidates into a filtered shortlist of "real people who might be the target." The pattern in `dsonbaker/email2whatsapp` is already Go and MIT-licensed — directly portable. Adds the single most distinguishing signal in modern phone OSINT.

### 4. Multi-format printer + "extract numbers from text" mode
**Effort:** ~30 minutes total. **Value:** Compounds with #1. After parsing, also output: E.164, national, international, `tel:` URI, dash-separated, dot-separated. Plus accept `--extract <file>` to pull all valid numbers from a free-text dump (phone-osint, copy-pasted leak, etc.). These are tiny features with outsized usability gain.

### 5. Pluggable scanner interface + JSON site-definition file
**Effort:** ~3 hours for the interface, +1 hour per scanner. **Value:** Long-term. Define `type Scanner interface { Name() string; DryRun() error; Run(Number) (any, error) }` exactly like phoneinfoga, plus a `sites.json` for cheap HTTP-based scanners (a la Sherlock / blackbird / WhatsMyName). Lets contributors add new sources without touching Go code. This is the architectural decision that determines whether clank scales past v0.2.

---

**Honorable mentions skipped from full profile:** sherlock-project/sherlock (username search, not phone), Datalux/Osintgram + megadose/toutatis (Instagram-only), AzeemIdrisi/PhoneSploit-Pro (Android ADB exploitation, off-topic), smicallef/spiderfoot (huge framework, phone is one tiny module), xadhrit/terra (Twitter+Instagram, not phone), kalmux1/PhoneOsint (thin wrapper around truecallerpy), GulsHanyadav788/NumIntense (libphonenumber-only re-implementation), stevemurr/reverse-phone-lookup (0 stars, niche), jasperan/whatsapp-osint (Selenium scraper for online status, not number lookup).
