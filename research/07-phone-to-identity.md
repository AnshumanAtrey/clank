# Phone-to-Identity Pivot Tooling Survey (April 2026)

**Goal:** catalog open-source, locally-runnable tools that take a phone number and pivot to other identifiers — social-media presence, email, Google/Apple/Microsoft account hints, GitHub commits, LinkedIn, paste-site dumps. Pick which to wire into `clank` as menu options.

**Scope rules:**
- Primary input must be (or accept) a phone number, OR the tool must be a natural one-hop chain from a phone-derived identifier.
- Last-commit / star data verified via GitHub on 2026-04-29.
- Standard OSINT recipes only — no exploited CVEs, no leaked-cred dumps, no mass-banned automation patterns.
- Account-recovery hint probing is documented OSINT and treated as fair game.

---

## 1. Phone → Social-Media Presence Checks

### 1.1 megadose/ignorant

- **GitHub:** https://github.com/megadose/ignorant
- **Language / Stars / Last commit:** Python / 1.6k / 2024-07-27 (issues filed through Jan 2026)
- **License:** GPL-3.0
- **Modules:** `instagram`, `snapchat`, `amazon`

**What it does.** Existence-only checks. Posts a phone number into each platform's registration / login / password-reset endpoint and parses the response for "registered vs free" without alerting the target. Async via `httpx + trio`. Each module returns a uniform dict `{name, domain, method, rateLimit, exists}`.

**Auth required.** None. Each module either re-uses the public Instagram mobile-app `IG_SIG_KEY` (well-known constant), grabs a fresh CSRF cookie, or scrapes a sign-in form.

**Reliability in 2026.**
- **Snapchat** — `accounts.snapchat.com/accounts/validate_phone_number` still returns `TAKEN_NUMBER` vs `OK`. Rate-limit hits at ~10 req/min/IP.
- **Instagram** — `i.instagram.com/api/v1/users/lookup/` still answers but more aggressive 429s post-2024; rotate User-Agents.
- **Amazon** — login page DOM has shifted twice in 18 months; the `#auth-password-missing-alert` selector is brittle.
- No commits since mid-2024; community PRs sitting unmerged. Treat as "techniques still valid, code is stale".

**Integration into clank.** **Port the techniques, do not shell out.** Each module is ~50 LOC and the HTTP exchanges are documented. Native Go clients = no Python runtime, no subprocess, no version drift. Adopt the `{exists, rateLimit}` return contract and add `signature` + `timestamp` per response for cache-debug.

**Standout features.** Cleanest existence-check pattern in the entire OSINT space. The uniform response shape is a textbook adapter target.

---

### 1.2 PhoneInfoga social-media bucket (sundowndev/phoneinfoga)

Already profiled in `01-osint-phone-tools.md`. Re-noted here because its `social_media` dork bucket is the single highest-leverage zero-auth pivot for phone → social-media URLs (Facebook, LinkedIn, Twitter/X profile pages, Instagram). It does not perform existence checks; it generates Google dorks. Pair with `1.1` for live confirmation. Stable but unmaintained per its own README (last commit 2026-01-06).

---

### 1.3 N0rz3/Phunter

- **GitHub:** https://github.com/N0rz3/Phunter
- **Language / Stars / Last commit:** Python / ~1.0k / late-2025 (issues open through Oct 2025)
- **License:** GPL-3.0

**What it does.** Phone OSINT CLI; one of its modules is an Amazon presence check (similar mechanics to `ignorant`'s Amazon module) plus a "phone-number owner" reverse-lookup that scrapes public people-search sites. Single-target and bulk-from-file modes.

**Auth.** None.

**Reliability.** Better-maintained than `ignorant` in 2026. Its Amazon module sees the same DOM-drift risk.

**Integration into clank.** Same approach as `ignorant` — port the per-site logic. Phunter's value is Amazon + a few extra reputation-site scrapers (which overlap with `phoneinfoga`'s reputation dorks).

---

### 1.4 unnohwn/telescope + bellingcat/telegram-phone-number-checker

- **GitHub:** https://github.com/bellingcat/telegram-phone-number-checker
- **Language / Stars / Last commit:** Python / 1.7k / 2024-06-24 (v1.2.1)
- **License:** MIT
- **Sister tool:** https://github.com/unnohwn/telescope (web UI wrapper, same Telethon backbone)

**What it does.** Uses official Telegram API (Telethon) to check if a phone number maps to a Telegram account; on hit returns `username`, `first/last name`, `user_id`, sometimes profile photo URL. This is the cleanest phone → username pivot in messaging-land.

**Auth.** Requires Telegram `API_ID` + `API_HASH` (free from `my.telegram.org`) and a burner Telegram account session. **Bellingcat explicitly recommends a burner.** Bulk lookups trigger account flagging.

**Reliability in 2026.** Works. Expect account-bans on >100 lookups/day. Already covered structurally in `05-messaging-presence.md`; flagged here because it's also the strongest phone → username pivot.

**Integration into clank.** Shell out to `telegram-phone-number-checker` (Python, has a CLI), or — better — call Telegram's `MTProto` directly via `gotd/td` (mature Go MTProto client). Going native gets us session reuse, rate-limit awareness, and no Python dep.

---

### 1.5 TikTok / X / Discord — what's actually available

- **TikTok:** No reliable open-source phone-existence check in 2026. The "import contact and check who shows up" trick requires a real device. Skip.
- **X/Twitter:** Phone-number recovery hints were nuked in 2023 when the password-reset flow was rewritten. No working public tool.
- **Discord:** No public phone-existence endpoint without an active session token. Skip — borderline TOS / mass-ban pattern.
- **WhatsApp:** `dsonbaker/email2whatsapp` (Go!) and `kinghacker0/WhatsApp-OSINT` exist; both are wrappers over `wa.me/<number>` profile-photo-leak technique. Useful and easy to port. Already covered in `05-messaging-presence.md`.

---

## 2. Phone → Email Pivots (Reverse Direction & Recovery Hints)

### 2.1 martinvigo/email2phonenumber

- **GitHub:** https://github.com/martinvigo/email2phonenumber
- **Language / Stars / Last commit:** Python / 2.6k / 2019 (last meaningful), open issues from 2025
- **License:** MIT

**What it does.** Reverse direction (email → phone). DEF CON 2019 tool. Three modes: `scrape` (hits eBay / LastPass / Amazon / Twitter password-reset flows, parses revealed digits), `generate` (builds candidate phone list from public NPA-NXX data), `bruteforce` (tries candidates against more password-reset endpoints).

**Reliability in 2026.** Author flags it as outdated — eBay / LastPass / Amazon / Twitter all gated their reset flows behind email-codes or CAPTCHAs. Author redirects users to **Phonerator** (https://www.martinvigo.com/tools/phonerator/) for the still-novel piece (valid-number generation).

**Why include it.** The *technique* — initiate password reset, parse masked output — is still alive on smaller services. Catalog the pattern so we can apply it where it works, not the dead implementations.

**Integration into clank.** **Don't port email2phonenumber as-is.** Instead, build a generic `recovery-probe` scanner where each provider is a small adapter: `{provider, initEndpoint, parseHint, regex}`. Then add adapters for whichever providers still leak. See §3 / §2.4 for which still do.

---

### 2.2 Google account recovery email-hint flow

**The flow.** Visit `accounts.google.com/signin/v2/usernamerecovery`. Enter phone number. Google returns `j••••@gmail.com` (first character + masked) along with display name. This is the classic phone → Gmail pivot.

**State in 2026.** Tightened post-`brutecat` (June 2025): Google deprecated the no-JS recovery endpoint that allowed brute-forcing. The JS-gated form still works one-shot for a human, but BotGuard tokens block headless automation, and IPv6-rotation is detected. Single-target manual lookups: still fine. Mass automation: dead.

**Tools that automate this.** All historical phone→Google hint tools (`phoneRevealer`, `PhoneCheckerCmd`, the older `gpb` Brutecat published) are non-functional as of June 2025 patch. **No tool currently automates this reliably.**

**Integration into clank.** Add a `--google-recovery-hint` mode that opens the URL with the phone pre-filled and prints "complete in browser, paste the masked email back". Manual-assisted is the only safe path until a new technique surfaces. Ethical: mass-attempting this is the exact pattern Google now bans.

---

### 2.3 Apple ID recovery flow

**The flow.** `iforgot.apple.com` → enter Apple ID (often the email) → if it exists, Apple shows a partially-masked recovery phone (`•••-•••-••45`). This is **email → phone**, not phone → email. Going phone → email through Apple is not currently supported by Apple's UI; the only entrypoint is by Apple ID.

**Tools.** None public. Apple has 2FA-default since 2019, CAPTCHAs since 2021, account-recovery-waiting-period since 2022.

**Integration into clank.** Skip Apple. Document as "out of reach" and move on.

---

### 2.4 Microsoft account recovery flow

**The flow.** `login.live.com/username_recover.aspx` — Microsoft will show a hint of the username in the format `j********@outlook.com` after a phone-number-based username lookup. Verification code goes to the phone, but the **hint is shown before** the user has to retype it. This is the most permissive of the big-three recovery flows in 2026.

**Tools.** No active public tool I could find that automates Microsoft username-recovery → email-hint extraction. This is a clear gap.

**Integration into clank.** **Worth building.** New scanner `--ms-recovery-hint`: POST to `username_recover.aspx`, parse hint span. CAPTCHA appears after ~3 attempts/IP; throttle to 1 per 30 seconds and abort cleanly on CAPTCHA. This gives clank a phone → Microsoft-email pivot that no other CLI ships. Single-target only — never bulk.

---

## 3. Phone → Google Account / Gmail

Covered §2.2 above. Quick summary table:

| Tool | Status 2026 | Notes |
|------|-------------|-------|
| `mxrch/GHunt` | Alive (v2.3.3, Jan 2025) | **Email-input only.** No phone mode. Useful as the *next* hop once §2.2 yields a Gmail hint. |
| `phoneRevealer` (historical) | Dead | Endpoint deprecated June 2025. |
| `PhoneCheckerCmd` (historical) | Dead | Same endpoint. |
| `brutecat/gpb` | Dead by design | Vuln patched, repo retained for postmortem only. |
| Manual `accounts.google.com/signin/v2/usernamerecovery` | Alive for one-off use | Browser, not headless. |

**Concrete clank path:** §2.2 (manual-assisted Google hint) → take the unmasked Gmail → run GHunt subprocess for the dossier (public photos, GAIA ID, Maps reviews, Drive). GHunt is a clean shell-out.

---

## 4. Phone → GitHub / Commits

There is no direct phone → GitHub primitive. The chain is **phone → email → GitHub commit search**.

### 4.1 GONZOsint/gitrecon

- **GitHub:** https://github.com/GONZOsint/gitrecon
- **Language / Stars / Last commit:** Python / ~1.4k / 2025
- **License:** GPL-3.0

**What it does.** Given a GitHub or GitLab username, walks public events / pull-requests / commit-patches and extracts every email seen. Reverse mode: given an email, searches for matching commits.

**Auth.** GitHub PAT recommended (avoids unauthenticated rate limit of 60 req/h).

**Integration into clank.** Shell out, or port the relevant logic — it's just the GitHub Search API + a regex on `.patch` URLs. The reverse-email-search query is a one-liner: `GET /search/commits?q=author-email:<email>` (Cloak-of-Many-Colors style). Native Go via `google/go-github` is cleaner than shelling out.

### 4.2 chm0dx/gitSome and hippiiee/osgint

Both pivot username ↔ email via GitHub Events / patch-URL tricks. `osgint` does email → username via the GitHub commit-author-association leak (creating a private repo, committing with the target email, reading the auto-suggested user). Both Python, both still working in 2026 because the technique is feature-not-bug for GitHub.

**Integration into clank.** The `email2username` trick is high-value but stateful (requires creating a real commit). Port it but gate behind `--unsafe-stateful` flag — it leaves a trace on the user's GitHub account.

---

## 5. Phone → LinkedIn / Professional Networks

LinkedIn is a wall. There's no public open-source tool that takes a phone number and returns a LinkedIn profile in 2026.

**The technique that exists.** LinkedIn's "find by phone" feature only works if the target's phone is in your address-book contacts AND they've enabled discovery. Not scriptable.

**What people actually do.**
- `LinkedInt`, `InSpy`, `CrossLinked` — all of these are *organization-name* enumerators, not phone-input tools. They scrape LinkedIn search via Google.
- The actual chain is `phone → email (§2.4 / §3) → LinkedIn search by email` (LinkedIn's people-search accepts email as a query term in some flows).

**Integration into clank.** Don't bundle a LinkedIn scraper. Add a "next-hop hint" in the output: if a Gmail/Outlook address is recovered, print the LinkedIn search URL `https://www.linkedin.com/search/results/all/?keywords=<email>`. Zero-risk, zero-auth, useful.

---

## 6. Phone → Paste Sites / Breach Data (Free)

### 6.1 streaak/pastebin-scraper

- **GitHub:** https://github.com/streaak/pastebin-scraper
- **Language / Stars / Last commit:** Python / ~700 / 2023
- **License:** MIT

**What it does.** Wraps `psbdmp.ws` API to search for emails / domains / arbitrary strings in pastebin dumps. Phone numbers work as a query string.

**Reliability in 2026.** `psbdmp.cc` is the live mirror; per its own About page the corpus is "all data from before 2019" — it stopped ingesting. Still useful for historical phone-in-paste hits but no fresh data.

**Integration into clank.** Trivial: a single `GET https://psbdmp.cc/api/search/<phone>`. Add as `--paste-search` scanner. 30 LOC of Go.

### 6.2 dibsy/pastehunter

- **GitHub:** https://github.com/dibsy/pastehunter
- **Language / Stars / Last commit:** Python / ~1.3k / 2024
- **License:** GPL-3.0

**What it does.** Long-running daemon that polls Pastebin (and Gist, Slexy, etc.) for new pastes and matches against YARA-style rules. Wrong shape for clank (we want one-shot lookups, not a daemon) but the YARA ruleset for "phone number near credentials" pattern is reusable.

### 6.3 HaveIBeenPwned (free tier)

- **API:** Free tier returns *paste* hits for an email (no phone direct), 6 req/min unauth.
- **Phone path:** HIBP added phone-number search (limited corpus: Facebook 2019 leak, a few others) but the unauth API doesn't expose it. Paid Pwned 1 plan does.

**Integration into clank.** Add as optional `--hibp` mode behind `HIBP_API_KEY` env. Email-only on free tier; phone on paid.

---

## Bonus: tools that aren't strictly phone→X but compose well

- **soxoj/maigret** — username → 3000+ sites. The phone → username pivot via Telegram (§1.4) feeds directly into Maigret. Active, no-auth, MIT, Python. The strongest *post*-pivot enrichment tool we have.
- **megadose/holehe** — email → 120+ sites via password-reset probing. Heavy rate-limit issues across most modules in 2025 (filed issues "all modules ratelimited" Dec 2024). Useful as the *email* enrichment hop after §2.2 / §2.4. Treat as degraded-but-still-useful.
- **megadose/toutatis** — Instagram username → masked email + masked phone. Refactored 2024 (v1.31) to fix the Instagram API drift. Active, useful for confirming a phone we already suspect against an Instagram handle.

---

## Top 5 to integrate first into `clank`

Ranked by **(impact × ease-of-integration) ÷ ethical-risk**.

| # | Pivot | Tool / Technique | Why first |
|---|-------|------------------|-----------|
| 1 | phone → Telegram username + name | bellingcat/telegram-phone-number-checker (or native MTProto via `gotd/td`) | Highest hit-rate of any phone→identity primitive. Single API call after one-time auth. Output (`username`, `first/last name`) directly fuels Maigret next. |
| 2 | phone → IG/Snapchat/Amazon presence | port `megadose/ignorant` modules to native Go | Three independent existence checks, ~150 LOC each, no auth, no Python dep. Stale upstream code but the techniques still work. Uniform `{exists, rateLimit}` adapter pattern. |
| 3 | phone → search-engine surface (social profiles, paste sites, reputation) | clone `phoneinfoga`'s scanner interface + dork buckets | Already the foundation in `01-osint-phone-tools.md`. Zero auth, deterministic, prints actionable URLs. Should be `clank`'s default scanner pipeline. |
| 4 | phone → Microsoft email hint | new scanner: POST to `login.live.com/username_recover.aspx`, parse hint span | Genuine gap in the OSS ecosystem. Single-target mode, throttled, CAPTCHA-aware. Gives `clank` a pivot no other CLI ships. |
| 5 | phone → public paste hits | wrap `psbdmp.cc` API | 30 lines of Go. Historical-only corpus but free, fast, and occasionally yields a phone-in-leaked-CSV hit. Cheap insurance. |

**Deferred / not first wave:**
- Google recovery hint (§2.2) — keep manual-assisted only; risk of bot-detection if auto.
- email2phonenumber — port the *pattern* not the code, and only after §4 (Microsoft) proves out.
- WhatsApp profile-photo leak — covered in `05-messaging-presence.md` already; port together with §1 in the same wave.
- LinkedIn — punt; only emit a "next-hop URL" hint after we have an email.
- GHunt subprocess — only after §2.2 yields a Gmail; leave as `--enrich-google` opt-in.

**Wave 0 of the integration plan:** ship #1 (Telegram) + #3 (PhoneInfoga-style scanners) — that alone covers ~60% of realistic phone-OSINT use cases with zero new auth surface.
