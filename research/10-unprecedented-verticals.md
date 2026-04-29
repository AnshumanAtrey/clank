# 10 — Unprecedented Phone-OSINT Verticals

This file catalogues the *less-obvious* directions a Go phone-OSINT CLI can grow into.
Mainstream tools (PhoneInfoga, Mr.Holmes, Toutatis), the lookup APIs, Truecaller methods,
local enrichment via libphonenumber, and messaging-app presence are all covered in
files `01-` through `05-`. Phone→identity pivots, GSM/SIM/IMEI tooling, and breach-data
integration live in their own files. **This file is the creative-frontier bucket** —
verticals that almost never show up in "phone OSINT" tutorials but produce real,
actionable signal.

Every vertical below has a real OSS artefact attached (library / dataset / tool) plus
a sketch of how it would integrate into `clank`. Items flagged "out of scope" are
documented for completeness and explicitly *not* recommended for direct integration.

---

## V1. Vanity-number decoding (alphabetic phone digits)

**Why it matters** — vanity numbers like `1-800-FLOWERS` or `1-866-NEW-CARS` carry
brand intent the digit form hides. If a target's phone in a leaked dataset reads
`1-844-CALL-NOW`, that's evidence of telemarketing operations, not a personal number.
libphonenumber's `PhoneNumberMatcher` deliberately *skips* alphabetic vanity input
(documented falsehood), and the Go ports inherit that gap.

**Tool** — `sreenivasanac/vanitynumber` (Python, 1 star — niche but the algorithm is
trivial: trie + BFS over a 20k-word dictionary, mapping `2→ABC, 3→DEF…`). The Ruby
gem `daddyz/phonelib` exposes `Phonelib.vanity_conversion = true`.

**Integration** — port the 80-line trie/BFS into Go (no real lib exists). Two-way
mode: `clank vanity 18007246837` →  `1-800-PAINTER`; `clank parse 1-800-FLOWERS`
auto-decodes to `+18003569377`. License: trivially clean-room rewriteable.

**Ethics** — none.

---

## V2. NPA-NXX block-level rate-center / carrier resolution

**Why it matters** — libphonenumber returns *country + region*. The North American
Numbering Plan goes much deeper: a US 10-digit number maps NPA-NXX-XXXX, where the
6-digit NPA-NXX prefix resolves to a **rate center** (city block), the **OCN**
(Operating Company Number — actual issuing carrier, mobile vs landline), and the
thousands-block holder after porting. NumVerify and Twilio Lookup expose only a
trimmed view of this. A free CSV dataset gets you 70 % of what their paid API
returns.

**Datasets / tools**
- `djbelieny/geoinfo-dataset` — every US/Canada zip + NPA/NXX exchange, lat/long
  centroid, CSV. (GitHub, public, no license listed — use as reference.)
- North American Numbering Plan Administrator (`nationalnanpa.com`) publishes
  monthly free CSVs of NPA-NXX assignments and Local Routing Number (LRN) updates.
- `local-calling-guide.com` and `telcodata.us` — scrapeable rate-center tables.

**Integration** — embed the CSV (~5–8 MB compressed) in clank, parse on first run,
expose `clank lookup +12125550100 --deep` returning rate center "New York Manhattan",
OCN "Verizon NY Inc.", original line-type "wireless block". Delta-update monthly via
HTTP if the user opts in. License: NANPA data is public domain.

**Ethics** — none. Public regulatory data.

---

## V3. ENUM / DNS-based number → URI mapping

**Why it matters** — ENUM (RFC 6116) takes an E.164 number, reverses the digits,
appends `.e164.arpa`, and asks DNS for NAPTR records that map the number to URIs
(SIP, mailto, XMPP). When a number is published in ENUM you get *the owner's
intended call-routing path* for free — sometimes with a personal SIP URI revealing
the registrant's domain.

**Reality check (2026)** — `e164.org` and `freenum.org` are both **dead** since
2016. The ITU-rooted `e164.arpa` zone is sparsely populated; only a handful of
countries (DE, AT, NL via NREN-style trees) actually publish records. Coverage
< 5 % globally. But for the cases where it *does* hit, the signal is uniquely
high-quality (it's authoritative and self-published).

**Library / approach** — no maintained Go ENUM library. Roll-your-own: take
`+15551234567` → reverse `7.6.5.4.3.2.1.5.5.5.1.e164.arpa` → DNS NAPTR query via
`miekg/dns`. ~30 lines of Go. Also try `nrenum.net` zone for academic hits.

**Integration** — `clank enum +49301234567` returns NAPTR records and the SIP URI
if any. Cheap, low-noise, low-hit-rate but signal-rich.

**Ethics** — none. Public DNS.

---

## V4. Disposable-VoIP block detection (free, beyond the paid APIs)

**Why it matters** — IPQS, Twilio, and Telnyx all charge to flag a number as
TextNow/Google-Voice/Hushed/Burner. The actual list of allocated CLEC blocks for
these providers is **public**. FCC Form 477 and the NANPA monthly utilization
reports list the OCNs of every wireless reseller. Cross-referencing OCN → service
brand (TextNow uses Bandwidth.com OCNs, Google Voice uses 484-series resellers,
Hushed uses Voxbone/Bandwidth, Burner uses Twilio resellers) gives you free
disposable-flag classification.

**Datasets / tools**
- `iliayar/textnow-numbers-list` and similar GitHub gists — community-maintained
  prefix lists for TextNow.
- `EnableSecurity/sipvicious` ships an OCN→brand mapping for VoIP fingerprinting.
- The PhoneInfoga `local` scanner already does carrier→VoIP heuristics — it's
  basic but extendable.

**Integration** — bundle a JSON map `OCN → {brand, type:disposable|carrier|reseller}`
in clank. After NPA-NXX lookup, classify line-type. Flag `disposable=true` when
OCN hits the disposable allow-list. Update quarterly.

**Ethics** — none.

---

## V5. CallerID Name (CNAM) free / community sources

**Why it matters** — CNAM is the 15-char text shown on a ringing phone ("AMAZON",
"DR JANE SMITH"). Carriers populate it in a per-number CNAM dip when the call lands.
The data is dirty but extremely revealing — names, business affiliations, sometimes
city. Paid CNAM dips run $0.005–$0.01. Free dips exist if you know where.

**Sources**
- **OpenCNAM free tier** — 60 lookups/hour, no key needed for "tier 0" data.
- **FreePBX `cidlookup` module** (`FreePBX/cidlookup`) — supports OpenCNAM, MySQL,
  HTTP, ENUM, and self-hosted phonebook back-ends. License: GPL-2.0.
- **SIPSTACK WHOIS** — community-supported AI-driven CNAM with no auth required.
  (Quality vs OpenCNAM is mixed — useful as a fallback.)
- **freecallerregistry.com** — anyone can self-publish CNAM for their own number;
  surprisingly populated for SMB/landlords.

**Integration** — add a `clank cnam` subcommand that fans out across OpenCNAM
(rate-limited) → SIPSTACK → freecallerregistry → returns the union of names with
provenance flags. Pure HTTP, no creds.

**Ethics** — public look-up, no telecom auth, fine.

---

## V6. STIR/SHAKEN attestation analysis (passive)

**Why it matters** — every US originated call since June 2021 carries a STIR/SHAKEN
SIP `Identity` header with an A/B/C attestation. If you ever capture a SIP INVITE
from a target number (e.g. as the called party on your own honeypot, or in a pcap
during authorized pentesting), the JWT in the Identity header tells you the
*attesting carrier* and their confidence the originating number was legitimate. A
"C" attestation on a number claiming to be a Fortune-500 IT support line is a near
certain spoof flag.

**Library** — `signalwire/libstirshaken` (C, 36 stars, MIT). Provides full STI-AS
(signing) and STI-VS (verification). Go bindings via cgo, or call out to the
`stirshaken` CLI it ships. There's also `secsipid` (Kamailio-related, Go/C) for
pure Go signing/verification.

**Integration** — out of scope for *outbound* clank usage (you can't STIR-verify
a number without an actual inbound call). But **document this as a downstream
"verify-on-receipt" companion**: `clank stir verify-jwt < pcap.txt` parses an
Identity header and returns attestation level + cert chain. Useful for IR teams
correlating clank intel with phishing-call captures.

**Ethics** — purely passive analysis of headers you legitimately received. Fine.

---

## V7. Reverse WHOIS — phone in registrant data

**Why it matters** — a phone number that appears as a domain registrant phone in
WHOIS is gold: every domain that person/company owns falls out. ICANN privacy
masking removed most of the easy hits post-2018, but historic WHOIS (2005-2018) is
still indexed by archive.org / WhoisXMLAPI / Domaintools and surfaces phone fields
freely.

**Tools**
- `harleo/knockknock` — Go, MIT, **190 stars**, last commit recent. Wraps
  ViewDNS.info reverse-whois (free for ~250 results/query, no key needed). Native
  Go means easy embed.
- `YashKarthik/reverse_WHOIS_lookup` — Python wrapper, also ViewDNS.
- `whoxy.com` — pay-per-query API ($1 per 100 lookups), supports phone search.

**Integration** — `clank whois-rev +14155551212` calls ViewDNS via knockknock-style
HTTP. List domains; attach to clank report. Combine with archive.org Wayback CDX
for *historic* WHOIS retrieval (free, no key). When a number looks personal in
libphonenumber, the absence of WHOIS hits is itself a signal.

**Ethics** — public ICANN data. Fine.

---

## V8. SEC EDGAR / corporate-filings phone scraping

**Why it matters** — every US public-company 10-K / 10-Q / 8-K and form D contains
the registrant's contact phone in the cover page. This is one of the cleanest
"company → phone" pivots and works in reverse: *given* a phone, EDGAR full-text
search finds every filing referencing it. Pretty common for criminals to register
shell companies that file Form D for fundraises and leave a real number.

**Tools**
- `dgunning/edgartools` (Python) — gives you `Filing.text` and structured cover
  pages.
- `sec-edgar/sec-edgar` — bulk download CLI.
- EDGAR full-text search: `https://efts.sec.gov/LATEST/search-index?q="+1+415+555+0123"&forms=10-K`
  — public REST endpoint, no key, JSON response.

**Integration** — `clank edgar +14155550123` hits the EFTS endpoint (free, no
auth), returns each filing where the digits appear, plus the filer CIK. ~50 LOC
in Go.

**Ethics** — public regulatory disclosures. Fine.

---

## V9. Wayback Machine / historic-Web phone resolution

**Why it matters** — a phone number scrubbed from a current website still lives
in archive.org. CDX API lets you query every archived URL whose HTML contains
your number, free, no key, no rate limit beyond ~10 req/sec. Effective against
real-estate scams (rotating shop fronts), removed personal pages, and pre-GDPR
WHOIS as Nixintel documents.

**Endpoint**
- CDX: `https://web.archive.org/cdx/search/cdx?url=*&matchType=domain&filter=mimetype:text/html&output=json&q=<number>`
  is *not* exactly right — CDX indexes URLs, not bodies. The trick is to
  combine Google's `site:web.archive.org "+1 555 0100"` (still indexed) plus
  archive.org's *Search Inside Books / Documents* full-text and the
  Wayback URL Search API.

**Tools** — no maintained Go lib. `tomnomnom/waybackurls` is the closest spirit
(URLs not body content). Easy ~100-line Go wrapper.

**Integration** — `clank wayback +14155550100` runs Google `site:web.archive.org`
+ a CDX URL-pattern guess, returns archived pages mentioning the number with
snapshot timestamps. Great for "this number used to be on shadyops.com circa
2019".

**Ethics** — public archived material. Fine.

---

## V10. Gravatar (and look-alike) hash-tray pivot — phone via email

**Why it matters** — phone-OSINT often produces an email candidate (via password
reset, Google address-book leak, or social-graph guessing). Gravatar exposes a
public profile if the user MD5-ed any email and uploaded an avatar — **and** the
profile JSON sometimes includes contact methods, including phone numbers and SMS
URLs. `balestek/hashtray` (66 stars, GPL-3.0, recent commits — last release
July 2025) handles both directions: email → profile, and hash → email candidate
search.

**Integration** — when clank's email-pivot module returns an email, fan it
through hashtray's logic (or port the ~150 lines into Go). The phone field on
Gravatar is rare but the *associated* WordPress.com / freelance profile data is
load-bearing for identity verification.

**Ethics** — entirely public Gravatar API. Fine.

---

## V11. Cell-tower geolocation (MCC/MNC/LAC/CID) — but with the right caveats

**Why it matters** — separate from phone-number enrichment, *cell-tower IDs*
extracted from MMS metadata, GSMA logs, or device exports map to lat/long via
OpenCelliD. This isn't directly "phone number → location" but works adjacent:
a leaked iCloud backup might pair the user's number to their last-seen cell-ID.

**Reality** — Mozilla Location Services (MLS) was retired in July 2024. The
remaining open game in town is **OpenCelliD** (`opencellid.org`), 50M+ towers
under CC-BY-SA, free API key (~1k queries/day). `tramseyer/cell-geolocation` is
a self-hosted Go-friendly wrapper.

**Integration** — `clank cell --mcc 310 --mnc 410 --lac 12345 --cid 67890`
returns approximate lat/long. Optional, off by default; requires user-supplied
MCC/MNC.

**Ethics** — fine; published infrastructure data. Not "tracking" — it's
mapping a cell-ID the user already has.

---

## V12. Phone-number entropy / sequential-allocation classification

**Why it matters** — bot-rented numbers come in **sequential blocks** from VoIP
brokers. A genuine personal number has a roughly random subscriber portion
within a city's NPA-NXX. If you analyse a list of 10k breach phones and see
1000 numbers in a tight `+1-555-XXX-7000` … `+1-555-XXX-7999` band assigned to
a single OCN issued in the last 18 months, that's a bot farm.

**Tools** — no off-the-shelf classifier exists. The wordlist generator
`toxydose/pnwgen` (72 stars, Python, no license) is the *opposite* primitive —
it generates sequential lists. Inverting it: given a set of numbers, fit a
sequential-density score per NPA-NXX × time-of-allocation bucket, flag dense
clusters.

**Integration** — `clank entropy <file.csv>` scores each number 0–100 on
sequential-allocation likelihood given OCN issuance dates and NPA-NXX
neighbours. Pure Go, ~200 LOC after you have the NPA-NXX map (V2).

**Use case** — fraud-prevention teams catching SIM-rental SMS-OTP farms. Novel
because no commercial tool packages this.

**Ethics** — purely statistical inference on data the user already has.

---

## V13. Robocall / IVR honeypot dataset cross-reference

**Why it matters** — `wspr-ncsu/robocall-audio-dataset` (15 stars, CC-BY-ND,
April 2024) is **1,432 real robocall recordings** harvested from FTC's
Project Point of No Entry. Each recording has FTC cease-and-desist PDFs
attached → which name **the calling number, the campaign, and sometimes the
shell company**. That's a free database of "this E.164 number is a
documented robocaller" that nobody integrates into OSINT tools.

**Integration** — extract the ~1.4k phone-number → campaign mappings into a
JSON file shipped with clank. `clank lookup +18004031120` returns "FTC
documented robocaller, campaign: medicare-fraud-2023, see
robocall-audio-dataset/calls/0421.txt". Tiny embed, high signal.

**Ethics** — public FTC enforcement data. Fine.

---

## V14. Pastebin / paste-site phone-number leak monitoring

**Why it matters** — phone numbers leak in pastebin dumps (combolists, cracked
DBs, OPS reports) constantly. Most "have I been pwned" services skip phones.

**Tools**
- `dibsy/pastehunter` — Python, automated pastebin pull + regex match. Active.
- `d-Rickyy-b/pastepwn` — modular paste-scraper with custom analyzers.
- `loseys/Oblivion` — combined Google-dork + pastebin + leak checker (data
  is third-party; quality varies).

**Integration** — `clank pastebin +14155550100` runs a regex against
pastebin.com search + Google `site:pastebin.com "+14155550100"`. No keys
needed for read-only search. Cache hits in `~/.clank/paste-leaks.db`.

**Ethics** — public paste contents only. Don't host the dumped data, just link
to the URL. Fine.

---

## V15. Discord / WeChat / Snapchat / TikTok phone-presence (the "long tail")

**Why it matters** — file 05 covers WhatsApp/Telegram/Signal. The long-tail
messengers each have their own quirks:
- **Discord** — phone is *optional* on signup but, if linked, the password-reset
  flow leaks "an SMS was sent to ••• ••• 47" (last-2 digits) — exactly
  martinvigo's `email2phonenumber` technique extended.
- **Snapchat** — "Forgot password" via phone leaks last-2 digits identically.
- **TikTok** — same pattern; also exposes country.
- **WeChat** — phone-only signup; can be probed via the international web sign-up
  flow (`weixin.qq.com/d/...`) but heavily anti-bot rate-limited.
- **Threema** — *opt-in* directory matching (so a hash of the number can be
  queried); the protocol is documented at `threema-ch/app-remote-protocol`.
- **Session / Wire** — no phone tied; skip.

**Tools / refs**
- `martinvigo/email2phonenumber` — original technique, easy to extend per
  platform.
- `husseinmuhaisen/DiscordOSINT` — broader Discord OSINT; phone-recovery flow
  documented.

**Integration** — `clank pwreset <platform> +14155550100` returns the masked
last-digits the platform discloses. Compare across platforms to confirm which
mask matches → identifies which accounts the number is tied to without a
direct presence query. Slow (rate-limited) but effective.

**Ethics** — gray area. Password-reset oracles are public-but-frowned-upon;
this is exactly what `email2phonenumber` already does and the OSINT community
generally treats as defensive. **Flag in clank as `--invasive` mode**, off
by default, with a banner reminding users to scope to authorized targets.

---

## V16. SIP / VoIP scanning context (out-of-scope, documented for completeness)

**Why it matters** — phone numbers terminate on SIP endpoints. If you're doing
*authorized* infrastructure security testing (your own PBX, your client's SBC),
SIP scanning surfaces extension lists, weak passwords, and open RTP sockets
exposed by phone-number routing.

**Tools**
- `EnableSecurity/sipvicious` — the canonical SIP audit toolkit.
- `Pepelux/sippts` — **562 stars, GPL-3.0, last release v4.1.2 Nov 2024**.
  Fingerprinting, enum, brute, sniff. Most actively maintained option in 2026.
- `qeeqbox/honeypots` (963 stars, AGPL-3.0) — multi-protocol, includes a SIP
  honeypot for *receiving* attack traffic, not scanning.

**Integration into clank** — **don't.** Wardialing and SIP scanning are
attacker-tooling outside the ethical zone of a phone-OSINT CLI. Mention in
clank's `man` page that for authorized SIP testing, users should reach for
sippts directly. Hard line.

---

## V17. VoIP honeypot / robocall honeypot ecosystem (out-of-scope)

**Why it matters** — running honeypots is the *source* for V13's data. Honeypot
projects like `honeynet/whisperpot` (22 stars, MIT) and `0xNslabs/sip-honeypot`
(4 stars, MIT) collect attack telemetry with timestamps, source-IPs, and SIP
user-agents. T-Pot (`telekom-security/tpotce`) bundles 20+ honeypots including
SIP. Useful for fraud-prevention teams who *receive* malicious calls and want
to publish their findings back into clank-style enrichment.

**Integration** — `clank` consumes the *output* of these systems (V13). It
should not deploy honeypots itself. Document as "feed me your honeypot logs
and I'll cross-correlate" via a JSON ingestion command:
`clank ingest --honeypot whisperpot.json`.

**Ethics** — fine for the *consumer*; honeypot operators bear their own
disclosure / legal obligations.

---

## V18. Mobile Number Portability (MNP) status check via free HLR

**Why it matters** — once a number ports, libphonenumber's "carrier" answer is
wrong. The current carrier (post-port) is found only via an MNP / HLR query.
Most paid services charge $0.005/lookup; a couple of free tiers exist for
casual investigation.

**Tools**
- `hlr-lookups.com` free trial: 100 free lookups (signup required; not API-keyed
  for production).
- `ipqualityscore.com/free-hlr-lookup` — 5 free lookups/day, no auth.
- `vmgltd/hlr-lookup-api-ruby-sdk` — official Ruby SDK; pattern is HTTP+API-key,
  trivially port to Go.
- For some EU countries: NRENum / national MNP DB exposed via ENUM (V3 overlap).

**Integration** — `clank hlr +14155550100 --provider ipqs` returns current
carrier post-port + active/inactive status. Free tier sufficient for one-off
investigations; warn user when rate limit hit.

**Ethics** — HLR is industry-standard query; free tiers explicitly allow
investigative use.

---

## V19. STIR/SHAKEN cert / OCN provenance lookup via Policy Administrator

**Why it matters** — separate from V6 (verify a JWT), the **STI-PA registry**
(operated by iconectiv for the FCC) lists every authorized signing carrier with
their OCN. Public look-ups confirm whether an OCN claiming to attest a number
is even allowed to. This is a free integrity signal absent from every existing
OSINT tool.

**Source** — `authenticate-api.iconectiv.com/download/v1/spc/spc-token` returns
a signed JSON list of authorized SP codes. No auth for read.

**Integration** — `clank stir-pa-check OCN-1234` returns "authorized" or
"revoked / not listed". Embed the SPC token nightly. ~50 LOC.

**Ethics** — public regulatory data.

---

## V20. Phone-number → username / handle generation

**Why it matters** — many users create handles incorporating their phone
fragments (`john7890`, `jane.0421`). Sherlock-style enumeration with a phone-
derived candidate set hits surprisingly often.

**Tools** — `sherlock-project/sherlock` (450+ platforms, no API keys) is the
fan-out engine; a small Go function generates handle candidates from the
phone subscriber digits + common naming patterns.

**Integration** — `clank handles +14155557890` → generates `[7890, x7890,
user7890, 7890user, 415x7890, …]` → spawns Sherlock subprocess (or reimplements
the HTTP-probe loop natively in Go, ~300 LOC). Highest-yield for low-effort
OSINT pivots.

**Ethics** — pure public-profile probing. Fine.

---

# Creativity-ranked top 10 for clank

(Most novel × most impactful × highest signal-per-LOC.)

1. **V13 — Robocall-audio-dataset cross-reference.** 1,432 documented
   robocaller numbers from the FTC, free, ships as a JSON. Nobody else
   integrates this. Wow factor + small embed cost.
2. **V12 — Phone-number entropy / sequential-allocation scoring.** Genuinely
   novel: turns the wordlist-generator primitive on its head. Fraud teams
   would pay for this; it's just statistics.
3. **V2 — NPA-NXX rate-center / OCN deep parsing.** Closes the "libphonenumber
   only knows region" gap with public NANPA data. Foundation for V4 and V12.
4. **V8 — SEC EDGAR full-text phone search.** Free public REST endpoint, no
   tool integrates this for phone-OSINT, surfaces shell-company filings cleanly.
5. **V3 — ENUM DNS lookup.** Sparse coverage but uniquely authoritative when
   it hits; <50 LOC of Go.
6. **V20 — Phone → handle generation + Sherlock fan-out.** High-yield pivot
   that closes the loop between phone and social-media presence.
7. **V19 — STIR/SHAKEN STI-PA OCN provenance.** Trust signal nobody else exposes.
8. **V7 — Reverse WHOIS via knockknock-pattern (free ViewDNS).** Native Go
   library exists; trivially embed.
9. **V14 — Pastebin leak monitoring.** Standard but underused for phones.
10. **V5 — Free CNAM dip cascade (OpenCNAM → SIPSTACK → freecallerregistry).**
    Genuinely free name-on-caller-ID without paid APIs.

The remaining items (V1 vanity, V6 STIR-verify, V10 Gravatar pivot, V11 cell-
tower, V15 messenger pwreset oracles, V18 MNP/HLR) are solid second-tier adds.
V16 (SIP scanning) and V17 (running honeypots) are explicitly **out of scope**
for clank's authorized-defensive-OSINT positioning — document, don't integrate.
