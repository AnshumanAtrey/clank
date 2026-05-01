# 05 — Website Copy (clank.atrey.dev)

> Canonical copywriting for the landing page. Written to be dropped into the Astro site at `clank/website/` (see `06-website-tech.md` for stack + `07-repo-architecture.md` for the same-repo decision). Tight, direct, terminal-aesthetic. No marketing fluff.

## Page architecture (single page, scroll-driven)

```
[Hero — animated terminal demo + tagline]
       ↓
[What it does — 4 lines, 4 commands]
       ↓
[Live terminal — real asciinema cast looping]
       ↓
[10 subcommands grid]
       ↓
[Install — 3 options, copy-button each]
       ↓
[Honesty section — what's broken / risky]
       ↓
[Author / FAQ / Footer]
```

---

## SECTION 1 — Hero

### Tagline (h1)
```
phone-number OSINT. one binary. ten subcommands.
```

### Subtagline (h2)
```
Open-source. MIT. No signup. No API gate.
clank — what PhoneInfoga should have been in 2026.
```

### Animated terminal in the hero (autoplay loop)
Show this asciinema cast on loop, 30 seconds:
```
$ clank deep +14155552671

Deep lookup — +14155552671

1. Local — libphonenumber + spam check
  +14155552671  FIXED_OR_MOBILE  US  San Francisco, CA

2. Free APIs
  veriphone    valid · MOBILE · AT&T · United States
  ipqs         valid · MOBILE · AT&T · fraud=12

3. Telegram
  registered · @jane_d · Jane Doe

4. WhatsApp
  registered · jid=14155552671@s.whatsapp.net

5. ignorant (Instagram / Snapchat / Amazon)
  instagram  registered

6. SEC EDGAR full-text
  0 hits

Took 6.2s
```

### Hero CTAs (buttons)
```
[ Install (1 line) ]   [ View on GitHub →  ]
```

---

## SECTION 2 — What it does (one-line each)

```
SCAN     —  bulk pattern → enrich every survivor
DEEP     —  one number, all six sources concurrently
DORKS    —  ~160 Google search URLs across 5 buckets
IMEI     —  Luhn check + 254,993-device TAC database
```

---

## SECTION 3 — Live terminal demo (real asciinema, looping)

Headline above the embed:
```
Watch it run.
```

Subline:
```
Real terminal session. Real commands. Real output. No mocks.
```

Asciinema embed showing:
- `clank scan +918115605xxx --max 5 --quick --no-edgar --region IN`
- The progress lines streaming
- The summary table at the end
- Loop after 90s with a 3s pause

---

## SECTION 4 — All 10 subcommands grid

Headline:
```
Every subcommand. No paywalls.
```

Grid (10 cards, 2 cols on mobile, 5 cols on desktop):

| Subcommand | One-liner | Status |
|------------|-----------|--------|
| `scan` | Bulk lookup over a pattern with rate-limit-aware sleeps | ✅ |
| `deep` | Single number, all 6 sources concurrently | ✅ |
| `dorks` | Generate Google-dork URLs across 5 buckets | ✅ |
| `imei` | Luhn check + 254,993-device TAC lookup (Mar 2026) | ✅ |
| `edgar` | SEC EDGAR full-text filings search | ✅ |
| `telegram` | MTProto phone-to-user via gotd/td | ✅ requires session |
| `whatsapp` | WhatsApp via whatsmeow (QR-pair) | ✅ requires session |
| `ignorant` | IG / Snapchat / Amazon presence check | ⚠️ Snapchat broken |
| `history` | Audit log of every clank invocation | ✅ |
| `--version` | Print version + author | ✅ |

---

## SECTION 5 — Install

Headline:
```
Three ways to install. Pick yours.
```

Three tabbed panels (each with copy-button):

### Tab 1 — Homebrew (recommended)
```
brew install AnshumanAtrey/tap/clank
```
*[note if brew tap not yet live: "Coming soon — see issue #3"]*

### Tab 2 — Pre-built binary
```
# macOS / Linux / Windows — grab from GitHub Releases
curl -L https://github.com/AnshumanAtrey/clank/releases/latest/download/clank_$(uname -s)_$(uname -m).tar.gz | tar xz
```

### Tab 3 — Go install
```
go install github.com/AnshumanAtrey/clank@latest
```

After install, verify:
```
clank --version
```

---

## SECTION 6 — Honest section (the differentiator)

Headline:
```
What's broken.
```

Subline:
```
Most "OSINT toolkit" landing pages oversell. This one doesn't. Here's what doesn't work yet.
```

Bullets:
- **Snapchat phone lookup**: Their CSRF moved JS-side in 2024. Method broken. Documented inline. Will fix in v0.2.0.
- **Messenger ban risk at scale**: WhatsApp/Telegram lookups can ban your session above ~50/min. Use a burner account or stick to small scans.
- **Carrier portability**: libphonenumber returns the originally-allocated carrier, not current. After number portability the answer may be wrong. Honest about it.
- **No pastebin search**: psbdmp.cc went dark in 2024. Hunting for a replacement.
- **Stars-to-features ratio**: still proving real-world value. v0.1.0 is week 1.

Headline (smaller):
```
What works out of the box.
```

Bullets:
- ✅ libphonenumber lookup (instant, no key)
- ✅ MCC/MNC carrier list (embedded)
- ✅ Spam blacklist (embedded)
- ✅ IMEI decoder + 254k-device TAC DB (embedded)
- ✅ Google-dork URL generator
- ✅ SEC EDGAR full-text search (no key)
- ✅ Telegram + WhatsApp + Instagram (with one-time setup)

---

## SECTION 7 — Footer / FAQ / Author

### FAQ (collapsibles)

**Why open-source?**  
Because phone OSINT shouldn't require a $99/mo Truecaller-Pro subscription. The data is public. The tool should be too.

**Is this legal?**  
Depends on your jurisdiction. clank queries public data sources via documented endpoints. Same as any web search. Misuse (stalking, harassment) is illegal regardless of tool. Read the README's Disclaimer.

**Will you add Truecaller?**  
Yes — top of the v0.2.0 roadmap. See [issue tracker](https://github.com/AnshumanAtrey/clank/issues) for ETA.

**How does it compare to PhoneInfoga?**  
clank: faster (Go binary, no Python deps), more sources (Telegram/WhatsApp/EDGAR), honest about brokenness. PhoneInfoga: web UI, more provider integrations, more battle-tested. Use both.

**Can I contribute?**  
Yes. Issues + PRs welcome at [github.com/AnshumanAtrey/clank](https://github.com/AnshumanAtrey/clank). Roadmap is public in `path-1/01-launch-strategy.md`.

**Does it phone home?**  
No. Zero telemetry. Zero analytics. Zero outbound calls except to the OSINT data sources you explicitly invoke. Audit log lives at `~/.clank/history.jsonl` on your machine, never sent anywhere. Set `CLANK_NO_AUDIT=1` to disable even that.

### Author block

```
Built by Anshuman Atrey
Software engineer · OSINT toolmaker · 5x hackathon winner

atrey.dev   ·   linkedin.com/in/anshumanatrey/   ·   build@atrey.dev
```

### Footer

```
clank — phone-number OSINT toolkit
MIT License · © 2026 Anshuman Atrey
github.com/AnshumanAtrey/clank   ·   issue tracker   ·   releases
```

---

## Microcopy

### Copy-button states
- Default: "copy"
- After click (2s): "copied ✓"
- After 2s: revert

### Loading / pending states
- Asciinema buffering: `▌` blinking cursor, no spinner
- Install command failed copy (rare): "couldn't copy — select manually"

### Error states (404, 500)
- 404: `command not found: clank --help` (terminal-themed joke)
- 500: `panic: runtime error — try refreshing or open an issue`

---

## Tone rules (for any future copy)

1. **Lowercase first letters** in headings — terminal aesthetic, anti-corporate
2. **Numbers always specific** — "254,993 devices" not "huge database"
3. **No buzzwords** — never "AI-powered", "revolutionary", "next-gen"
4. **Honest about limits** — Pratfall Effect; people trust admitted weakness
5. **Imperative voice** — "install" not "try installing"
6. **No exclamation marks** — terminal output doesn't shout
7. **Code blocks for any command** — make it copy-able
8. **Em-dashes for asides** — they read more conversational than parens

---

## Alternative hero variants (for A/B if first doesn't perform)

### Variant B — contrarian
```
Most phone-OSINT tools haven't shipped a release in 2 years.
clank ships every week. Open-source. MIT. No signup.
```

### Variant C — narrative
```
3 months ago I had a phone number and 4 hours.
None of the open-source tools worked end-to-end.
So I built clank.
```

### Variant D — utility-first
```
phone-number OSINT in your terminal.
no API gate. no signup. one binary. ten commands.
```

Pick A (default) for launch. Test B vs C vs D in week +2 if the bounce rate is high.

---

## What NOT to put on the landing page

- ❌ Pricing — it's free
- ❌ "Trusted by" logos — until it's true
- ❌ Email signup form — premature; build trust first
- ❌ Testimonials — until they're real
- ❌ "Limited time" anything
- ❌ Cookie banner (no analytics, no cookies = no banner needed)
- ❌ Long manifesto about OSINT philosophy — link to a blog post instead

The whole page should fit on one scroll for desktop. Mobile: 4-5 scrolls max.
