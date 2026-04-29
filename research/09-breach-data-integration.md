# 09 — Breach Data, Reverse Lookup, and Spam-Tag Integration for `clank`

Scope: defensible OSINT signals for a phone number — has it appeared in public breach indexes, government complaint data, crowdsourced spam reports, or carrier reputation APIs. Stolen-credential dumps and ToS-violating scrape targets are explicitly excluded; see "Off-limits" at the bottom.

Verified April 2026.

---

## 1. Have I Been Pwned (HIBP) and k-anonymity

### 1.1 Phone-number lookup status (2026)

HIBP's website search box still accepts an email, username, or phone number, and the "breached account" API at `api.haveibeenpwned.com/api/v3/breachedaccount/{account}` will accept a phone number as the `{account}` value. **However, Troy Hunt has publicly stated no further non-email data will be loaded** — phone-number coverage is frozen at the small handful of breaches (Facebook 2019, Acuity, Mob Land, etc.) that already had phone fields. Treat phone-number HIBP as an "occasionally hits" signal, not a primary source.

- URL: `https://haveibeenpwned.com/api/v3/`
- Auth: paid API key, $3.95/month "Pwned 1" tier (10 RPM) up through enterprise. `hibp-api-key` header.
- Response: JSON array of breach objects (`{Name, Domain, BreachDate, ...}`) or 404 if clean.
- License: free at-rest, paid at-query.
- Maintenance: actively maintained; phone data effectively frozen.
- Go integration: plain HTTPS GET, parse JSON; respect `retry-after` on 429.

### 1.2 Pwned Passwords (k-anonymity reference)

Free, no-auth API. Client SHA-1s the password, sends the first 5 hex chars of the hash to `https://api.pwnedpasswords.com/range/{prefix}`, server returns ~300–800 suffix:count pairs; client checks if its full hash's suffix is in the response. Server never sees the full hash. Pure HTTPS, plaintext response, trivially embeddable in Go (single `http.Get` + line scanner).

### 1.3 Phone-number k-anonymity equivalents

**There is no production HIBP-like k-anonymity service for phone numbers in 2026.** Searches surface marketing pages from `databreach.com`, `forestalsecurity.com`, `idstrong.com`, etc., but all require full phone-number submission to the server, which is the privacy property k-anonymity exists to avoid. **Recommendation for `clank`: implement a local k-anonymity-style API ourselves later** — hash phone E.164 with SHA-256, ship a prefix-indexed bloom-or-suffix list of known-spam numbers. Defer; not a 2026 ship target.

---

## 2. Public breach indexers

| Service | Phone search? | Auth | Free? | Verdict |
|---|---|---|---|---|
| **HIBP** | Frozen subset | Paid key | $3.95/mo | Ship — see §1 |
| **DeHashed** | Yes (full PII) | Paid key | $5.20/mo trial → $30+ | **Off-limits** — distributes plaintext credentials |
| **LeakCheck.io** | Yes | Paid key | $9.99 starter | **Off-limits** — same |
| **Snusbase** | Yes | Paid key | $14/mo+ | **Off-limits** — same |
| **BreachDirectory.com / .org** | No (email/user/IP only) | RapidAPI key, free 50/mo | Free tier exists | Useful for cross-PII pivot, not phone |
| **Leak-Lookup** | No phone | Paid | Limited free | Email/user/IP only |
| **IntelTechniques (Bazzell)** | Tools page only | n/a | Mostly free | Browser bookmarklets, not an API |
| **databreach.com** | Yes | None | Free web UI | No public API; ToS forbids automated queries |

The clean-source line is bright: HIBP and BreachDirectory have voluntary breach-notification semantics; the others traffic in plaintext credentials and are off-limits for `clank` regardless of license. We do not integrate them.

---

## 3. Reverse phone lookup community sites

| Site | OSS client | Status 2026 | Notes |
|---|---|---|---|
| WhitePages / 411 / AnyWho | None alive | Heavy CF + JS challenges | Old PyPI scrapers (`whitepages-scraper`) all 403 since 2023 |
| Sync.me | None | Mobile-app gated | Requires phonebook upload — privacy-toxic |
| TrueCaller | `nvzard/truecaller-unofficial-api` (Python), `apify/truecaller-scraper` | Bearer-token reverse-engineered API still works but ToS-prohibited | Covered in `03-truecaller-methods.md`; do not call from `clank` per Truecaller ToS |
| Eyecon / Bharatcaller | None | Mobile-app gated | No public API |
| `stevemurr/reverse-phone-lookup` (Go) | Yes | Last commit 2018 | Scrapes `nomorobo.com` only; abandoned |
| `FOGSEC/PhoneInfoga` (Python) | Yes | Active | Aggregates Google dorks + numverify; we already cover its scanners in `01` |
| `phoneintel/phoneintel` | Yes | Active | Modern PhoneInfoga-style aggregator |

**Verdict:** No clean OSS reverse-lookup scraper for the US/global community sites is alive in 2026. We should not add a stale scraper. Truecaller-class apps require uploading the user's phonebook to enable reverse-lookup and so are categorically off-limits for `clank`.

---

## 4. Spam / scam crowdsourced reports

### 4.1 shouldianswer.com / .net / .co.uk

- robots.txt: blanket `Disallow: /` for GPTBot, AhrefsBot, SemrushBot, etc. — generic crawlers not explicitly named are not disallowed, but the site's intent is restrictive.
- No public API. No OSS client in 2026.
- Trustpilot reviews report email-spam-after-signup behavior — privacy hygiene is poor.
- **Verdict:** scrape-only, low-trust. Skip.

### 4.2 800notes.com

- robots.txt: bans `GPTBot`, `AhrefsBot`, etc., wholesale; no per-path `Disallow` for `/number/{n}`.
- No public API. Scraping is technically tolerated for non-banned UAs but commercial use is implicitly disallowed in their ToS.
- Useful US data, but legally grey for us.
- **Verdict:** out of scope. Document existence; do not integrate.

### 4.3 tellows.com — **the cleanest in this category**

- Commercial API: `https://www.tellows.com/s/about-en/tellows-api-partnership-program`
- Three products: **Scorelist** (CSV blacklist export, customizable by country/score/comments), **Live API** (real-time JSON lookup), **Reputation API**.
- robots.txt explicitly disallows `/phoneid/` for unauthorized crawlers — using the API is the *required* path.
- Pricing: paid only; €99–€299/yr range per `shop.tellows.de` (private API key) — no free tier. CSV download is also a paid product.
- Coverage: 50+ countries, strong DACH region.
- Format: CSV (Scorelist) and JSON (Live).
- Go integration: HTTPS GET with `partner=&api_key=` query params; standard `encoding/csv` for the blacklist.
- Active maintenance: yes.
- **Verdict:** Best paid integration in this category. Defer until paying users justify it; ship as an optional `--tellows-key` flag.

### 4.4 mrnumber.com

- Owned by Hiya. No public API. Site is now mostly a redirect to Hiya's app.
- **Verdict:** dead for OSINT.

### 4.5 nomorobo.com

- Lookup form scrape only. Used to power `stevemurr/reverse-phone-lookup` but anti-bot has tightened.
- No public API for individual lookups (B2B carrier integration only).
- **Verdict:** skip.

---

## 5. Government / regulator open data — the cleanest tier

### 5.1 FCC Consumer Complaints (US) — **ship this**

- Dataset: `Consumer Complaints Data - Unwanted Calls`
- URL: `https://opendata.fcc.gov/Consumer/Consumer-Complaints-Data-Unwanted-Calls/vakf-fz8e`
- Auth: none for low-volume; SODA app-token recommended for >1k req/hr.
- API: Socrata SODA v2/OData v4 — `https://opendata.fcc.gov/resource/vakf-fz8e.json?$where=...&$limit=...`
- Update cadence: nightly.
- License: public domain (US Federal).
- Fields include: `caller_id_number`, `ticket_created`, `time_issue`, `state`, `method`, `type_of_call_or_messge`, `advertiser_business_subject` — *the calling number is included*.
- Volume: ~5M records since 2014, growing weekly.
- Go integration: SODA queries are plain JSON over HTTPS, e.g. `?caller_id_number=8005551234`. Use `encoding/json`. Optional Socrata `X-App-Token` header if rate-limited.
- **Bonus:** archive a slice locally as a `clank` data file for offline first-hit lookups.

### 5.2 FTC DNC Reported Calls (US) — **ship this**

- URL: `https://www.ftc.gov/policy-notices/open-government/data-sets/do-not-call-data`
- API: `https://api.ftc.gov/v0/dnc-complaints` (no key required for low volume; key available free at `api.data.gov`).
- Update: every weekday around noon ET; weekend data on Mondays.
- Format: JSON via API, also weekly CSV dumps.
- Includes: originating phone number, call date/time, consumer state, subject, robocall flag.
- License: public domain (US Federal).
- Already noted in `04-local-enrichment.md`; add programmatic fetch path.
- Go integration: `GET https://api.ftc.gov/v0/dnc-complaints/?api_key=KEY&CompanyPhoneNumber=...` returns JSON.

### 5.3 Ofcom (UK)

- Publishes quarterly *aggregate* complaint volumes per provider as CSV at `https://www.ofcom.org.uk/research-statistics-and-data`.
- **Aggregate-only** — no per-number data. Useless for clank.
- **Verdict:** skip.

### 5.4 TRAI (India)

- TRAI DND 3.0 app (consumer-facing) and the National DNC Registry are *check-only* — no public dump.
- The "DND scrubbed" lists are sold through licensed telemarketing intermediaries.
- **Verdict:** no clean public dataset; skip.

### 5.5 CRTC (Canada)

- DNCL operator forwards raw complaints to CRTC; only annual aggregate stats published.
- No per-number open data.
- **Verdict:** skip.

### 5.6 Bundesnetzagentur (Germany) — **bonus, ship as embedded list**

- URL: `https://www.bundesnetzagentur.de/DE/Vportal/TK/Aerger/Aktuelles/Maßnahmen/start_RM.html`
- Annual *Maßnahmenliste* PDFs list every number on which the regulator has imposed a measure (Spam-Messenger, Ping-Calls, etc.) for the last 10 years.
- Format: PDF — extract once via `pdftotext`, ship as embedded CSV.
- Strong signal for German numbers; small (low thousands of entries) but high precision.
- License: official public publication.
- Go integration: parse PDF offline → embed CSV → `embed.FS` lookup at runtime.

---

## 6. Caller-ID / spam-tag datasets (downloadable, embeddable)

### 6.1 `Oros42/phone-blacklist` — **ship this**

- URL: `https://github.com/Oros42/phone-blacklist`
- License: Unlicense (public domain dedication).
- Format: single `blacklist.csv`, E.164-style, multi-country.
- Size: small (~few hundred entries — *not* a primary source by itself).
- Last update: irregular but accepts PRs in 2025.
- Go integration: embed via `//go:embed blacklist.csv` and load into a map at init. Zero runtime cost.

### 6.2 `wspr-ncsu/robocall-audio-dataset`

- FTC Project Point-of-No-Entry recordings + metadata CSV.
- Numbers are present in metadata but it's a research-grade audio set, not a live blocklist.
- **Verdict:** novelty, skip for clank's CLI flow.

### 6.3 `CallControl/CallControlClient`

- Closed proprietary platform with a thin OSS client. Requires a paid key.
- **Verdict:** skip; tellows is a better paid fit (§4.3).

### 6.4 Bundesnetzagentur PDF → CSV

- See §5.6. Ship as embedded data file.

### 6.5 GitHub topic crawls

- `spam-numbers`, `phone-blacklist`, `robocall` topics surface mostly toy lists or duplicates of Oros42. None ≥10k entries with active 2025/2026 commits found.

---

## 7. Phone reputation APIs (free / freemium)

| API | Free tier | Phone-specific signals | Notes |
|---|---|---|---|
| **IPQualityScore** (already integrated) | 5,000/mo free | `fraud_score`, `recent_abuse`, `VOIP`, `risky` | Best-in-class, keep |
| **APIVoid Phone Reputation** | Pay-as-you-go credits, small free trial | risk score, line type, carrier, country | Closest peer to IPQS; document but skip in v1 |
| **Veriphone** | 1,000/mo free | type, carrier, country | Validation-grade, no spam score |
| **NumVerify (apilayer)** | 100/mo free | carrier, line_type, location | Validation only, no reputation |
| **Twilio Lookup v2** | No free; $0.008 line-type, $0.005 carrier, free format-validate | line_type, carrier, SIM-swap | Premium; skip until enterprise tier |
| **Trestle Phone Intelligence** | Paid only | reassigned-number, spam score | Strong but US-centric, paid |

**Concrete Go path for APIVoid (the second-best after IPQS):**
```
GET https://endpoint.apivoid.com/phonerep/v1/pay-as-you-go/?key=KEY&phone=14155551212
```
Returns JSON with `risk_score` (0–100), `is_voip`, `is_invalid`, etc. Same shape as IPQS — easy to add as a fallback provider.

---

## 8. Public paste-site / archive grep tools

### 8.1 PasteHunter (`kevthehermit/PasteHunter`)

- Python, YARA-rules driven. Polls Pastebin Pro feed (paid Pastebin sub required) + scrape.
- Useful for *finding* breach pastes, not for live phone lookups.
- **Verdict:** out of scope for clank's per-query lookup model.

### 8.2 sniff-paste, Scavenger

- Older Pastebin scrapers. Same shape as PasteHunter; not per-number queryable.
- **Verdict:** skip.

### 8.3 Wayback Machine API

- URL: `https://archive.org/wayback/available?url=whitepages.com/phone/1-415-555-1212`
- Returns the closest snapshot URL.
- Free, no auth, generous rate limits (a few req/sec).
- **Useful for clank as a "cold case" pivot:** when WhitePages 403s a live request, fetch the most recent archived snapshot. Snapshots from 2018–2022 retain the old indexable format.
- License: archive.org ToS — non-commercial usage OK for OSINT.
- Go integration: `GET https://archive.org/wayback/available?url=...` → JSON `{closest:{url, timestamp}}` → fetch that URL, parse with `golang.org/x/net/html`.

---

## Top 5 to integrate, ranked by (legal-cleanliness × value)

1. **FCC Consumer Complaints Data — Unwanted Calls (Socrata API)** — public-domain US data, includes the complained-about phone number, nightly updates, no auth, well-documented. Highest signal-to-effort ratio. *Ship in v1.*
2. **FTC Do Not Call Reported Calls API** — same legal footing (US public domain), daily updates, plain JSON via `api.data.gov` key. Complementary to FCC. *Ship in v1.*
3. **Wayback Machine pivot for stale reverse-lookup pages** — free, no auth, lets clank surface cached WhitePages/411 entries without scraping the live sites. *Ship in v1.*
4. **`Oros42/phone-blacklist` embedded CSV + Bundesnetzagentur extracted CSV** — public-domain seed data, zero-runtime-cost local hits. *Ship in v1 as `embed.FS` data files.*
5. **APIVoid Phone Reputation (optional)** — opt-in via flag, mirrors our IPQS shape; gives users a fallback when IPQS quota is exhausted. *Ship in v1.1 behind `--apivoid-key`.*

Tellows is the best paid spam-tag source but stays behind a paid flag (`--tellows-key`) until we have demand. HIBP phone search is a low-yield bonus — wire it in only after the above five.

---

## Off-limits — do NOT integrate

These look tempting but are categorically excluded:

- **DeHashed, LeakCheck, Snusbase, leak-lookup, breachforums-replacements, "data breach search engines"** — distribute plaintext stolen credentials. Querying them on behalf of users is OK ethically, but `clank` will *not* embed these as integrations because the ecosystem normalizes consuming PII without subject consent.
- **Truecaller / Eyecon / Bharatcaller / Sync.me unofficial APIs** — every one of these requires uploading the user's address book or impersonating a logged-in mobile client. ToS-violating and user-PII-exfiltrating. Reference only — never call.
- **The Facebook 533M scrape, Truecaller 2019 leak, Acuity 2024 leak as raw datasets** — even though some indexers serve them, distributing or re-querying these inside clank is a non-starter. We will only consume them via a consented breach-notification surface (HIBP).
- **Ransomware-actor leak forums, "combolists", credential-stuffing dumps** — never.
- **WhitePages/411/AnyWho live scraping** — current anti-bot makes it brittle, ToS forbids automated access. Use Wayback Machine for stale snapshots only.
- **shouldianswer.com / 800notes.com / mrnumber.com scraping** — robots.txt is permissive enough that a polite scraper would technically work, but ToS forbids commercial automated access and the data quality is poor enough that it's not worth the legal exposure.

Hard rule for clank: every phone-number signal we ship must trace to (a) a government open dataset, (b) a service the queried subject has implicitly consented to (HIBP-style notification), or (c) a paid commercial API where the user supplies their own key. Anything else stays out.
