# Twitter/X — Launch Thread

> 7-tweet thread. Post Tue 18:00 IST (~07:30 EST = US morning scroll). Single image attached to tweet 1, screenshots/GIFs in 2-6, link in 7.

## The thread

### Tweet 1 (the hook + image)

```
i rebuilt an abandoned 110-line phone-OSINT script into a 10-subcommand
toolkit.

single 48 MB binary. MIT. no signup. no API gate.

clank — what PhoneInfoga should have been in 2026.

🧵
```

**Image**: Hero terminal screenshot of `clank deep +14155552671` output (same one from website hero / OG image). 1200x630 or 1080x1080.

**Why "🧵"**: signals thread → primes scrollers to keep reading.

**Why lowercase**: matches your terminal aesthetic + reads as a real person on dev Twitter.

---

### Tweet 2 — the scope

```
10 subcommands shipped in v0.1.0:

scan      — bulk pattern → enrich each survivor
deep      — one number, 6 sources concurrent
dorks     — ~160 Google URLs across 5 buckets
imei      — Luhn + 254,993-device TAC DB (Mar 2026)
edgar     — SEC full-text filings search
telegram  — MTProto via gotd/td
whatsapp  — whatsmeow (QR pair)
ignorant  — IG/Snap/Amazon presence
history   — local audit log
--version
```

No image needed — list is the visual. (Twitter renders fixed-width font for code-like blocks if you use indentation.)

---

### Tweet 3 — what's broken (Pratfall)

```
honest about what doesn't work yet:

× Snapchat: their CSRF moved JS-side in 2024, broken
× messenger bans at scale (~50 lookups/min)
× pastebin search: psbdmp.cc went dark
× carrier portability: shows original carrier, not current

every one of these is documented inline. no oversell.
```

**Why this tweet matters**: dev Twitter is allergic to "this thing solves everything!" launches. Admitting weakness in tweet 3 (after the hook + scope) reads as confidence, not weakness.

---

### Tweet 4 — the "live" subcommand demo (image/GIF)

```
the bulk pattern scan in action.

clank scan --quick --max 5 +918115605xxx --region IN

   ↳ libphonenumber filters invalid candidates
   ↳ enriches each in parallel
   ↳ JSONL checkpoint per result
   ↳ score-ranked summary table
   ↳ 5 candidates in ~3 seconds
```

**Image/GIF**: short asciinema cast (15s) of the scan command. Use vhs.charm.sh to convert .cast → .gif.

---

### Tweet 5 — origin story

```
why this exists:

3 months ago i needed to investigate a number. tried every open-source
phone-OSINT tool. all of them either had broken modules or hadn't shipped
since 2023.

started writing clank that weekend. 3 months later, this.

the OSINT community deserves better than abandonware.
```

**Why this works**: personal stake + community framing (Cialdini's Unity principle).

---

### Tweet 6 — the install (3 ways)

```
three ways to install:

go install github.com/AnshumanAtrey/clank@latest

# or grab the binary
gh release download v0.1.0 -R AnshumanAtrey/clank

# brew tap coming soon (issue #3)
brew install AnshumanAtrey/tap/clank
```

---

### Tweet 7 — links + CTA

```
→ github.com/AnshumanAtrey/clank
→ clank.atrey.dev

honest roadmap, open issues, and a Disclaimer that takes the
"don't use this for stalking" angle seriously.

what would YOU build on top of this? RT if useful, reply if you'd
contribute.
```

**Why the CTA at the end**: asks two things — RT (passive amplification) AND reply (engagement). Letting people pick reduces the ask, more people do at least one.

---

## Variants by tweet

### Tweet 1 — alternative hooks

**Variant B** (contrarian):
```
the most popular phone-OSINT tool on github
hasn't shipped a release in 2 years.

so i built a new one.

clank: 10 subcommands, single binary, MIT.

🧵
```

**Variant C** (data-flex):
```
in 3 months i:
- ported PhoneInfoga to Go (faster, single binary)
- added Telegram, WhatsApp, EDGAR
- bundled a 254,993-device IMEI database
- documented what's broken (Snapchat) so users aren't surprised

shipped today as clank v0.1.0. open-source, MIT.

🧵
```

---

## Tagging strategy

In tweet 7 (or as quote-tweet replies later), tag carefully:

**Worth tagging** (1-2 per thread, never in opening tweet):
- @bellingcat — if you have a relevant question, not just for visibility
- @osintessentials — OSINT community account, sometimes RTs niche tools
- @dutch_osintguy — well-known OSINT personality
- @CyberSecMeg — security community amplifier
- @megadose — author of `ignorant`, the Python lib clank ports — reciprocity
- @whatsmeow — WhatsApp Go lib maintainers

**Do NOT tag**:
- Random influencers with no clank-relevance
- Multiple accounts in one tweet (auto-block triggers)
- Anyone you don't have a real reason to engage

**Tag pattern**: "thanks @megadose for the original Python `ignorant` — clank's a Go port + extension of their work."

---

## Reply templates

### "Cool but how does it compare to [X]?"
```
[X] has [their strengths]. clank's faster (Go binary), has Telegram +
WhatsApp + EDGAR (which [X] doesn't), and is honest about what's broken.
Use both. The README has the full comparison.
```

### "Is this legal?"
```
clank queries public data via documented endpoints — same as a web search.
Misuse (stalking, harassment) is illegal regardless of tool. README has a
Disclaimer covering legitimate use: journalism, fraud investigation,
security research, your own footprint.
```

### "Will you add [feature]?"
```
On the roadmap — open an issue if it's not already there. Top of v0.2.0:
Truecaller integration + connection pooling for scan.
```

### "Why Go?"
```
single static binary on every OS, no python dep hell, the messenger
libraries (gotd/td, whatsmeow) are actually maintained in Go, and i
wanted shipping speed.
```

### "Cool stuff"
```
thanks! if you try it, let me know what you scan first — always curious
what people pivot to.
```

---

## Cross-post to Mastodon (infosec.exchange)

Same thread, posted to https://infosec.exchange — much smaller audience but high-quality OSINT/security devs. Some Twitter-skeptic OSINT folks live there exclusively in 2026.

Adjust tweet 1 slightly for context (some Mastodon users see your cross-post on both platforms):
```
crosspost from twitter — building clank, a phone-OSINT toolkit in Go.

10 subcommands, single binary, MIT. honest about what's broken.

🧵 below
```

---

## Cross-post to LinkedIn (different post, NOT same copy)

Don't paste the Twitter thread into LinkedIn. The LinkedIn post is in `linkedin.md` — different format (carousel or single post), different tone, different timing.

Cross-promote: in the LinkedIn post, say "also threading this on twitter [link]" once. Once is fine. More is spammy.

---

## What NOT to do on Twitter launch

- ❌ Don't post the thread, then immediately quote-tweet it 5 times for "more reach"
- ❌ Don't reply to your own tweets with "btw — [more info]" loops (looks desperate)
- ❌ Don't tag 5+ people
- ❌ Don't include a tracking-link (utm_*) in tweet 7 — looks corporate
- ❌ Don't delete and repost if engagement is low (Twitter penalizes deletion+repost)
- ❌ Don't pin the thread to your profile until it has 50+ likes (early pinning of low-engagement threads looks desperate)
- ❌ Don't add 10+ hashtags to the thread (Twitter caps reach above 2)

## Hashtag strategy

Max 2 hashtags total across the thread. Add them to tweet 1 OR tweet 7, not both.

**Good options**: `#OSINT`, `#opensource`, `#golang`, `#cybersecurity`, `#buildinpublic`

**Pick 1-2** based on which audience you most want to attract.
