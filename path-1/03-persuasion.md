# 03 — Persuasion Playbook

> The user asked for "manipulative" tactics. This document treats that ask seriously: it covers the full spectrum from ethical persuasion (Cialdini's 7 principles) to dark patterns, marks the line, and explains why crossing it in 2026 is an actively bad bet.

## The 2026 ethical floor

Per researched guidance ([Yu-kai Chou on Cialdini 2026](https://yukaichou.com/gamification-analysis/cialdini-6-principles-persuasion-octalysis-gamification-framework/)):

> Regulators in the UK, EU, and US are moving harder against urgency countdowns and other dark-pattern tactics, so in 2026, every scarcity claim on a site should be defensible in a deposition.

**Practical translation for clank**: don't fake scarcity, don't fake urgency, don't fake testimonials. The legal cost-of-getting-caught now exceeds the conversion benefit. Beyond legality, OSINT/security developers are unusually pattern-matching for manipulation. They'll detect and broadcast it. One tweet calling out a fake "X people viewing right now" badge would erase 6 months of credibility.

**This document recommends ethical persuasion techniques only.** "Manipulative" is reframed as "high-conversion via consent" — copy that pulls people in because it's *true and well-framed*, not because it tricks them.

---

## Cialdini's 7 principles applied to clank

Per [Robert Cialdini, Influence Institute](https://www.influenceatwork.com/7-principles-of-persuasion/), the 7 principles are: **Reciprocity, Commitment & Consistency, Social Proof, Authority, Liking, Scarcity, Unity**.

### 1. Reciprocity — give first

clank is free, MIT, no signup. That's the default reciprocity. To deepen it:
- **Free research files**: the 11 markdown files in `research/` are *useful* even to people who never run clank. Surface them in the README. People who read research files feel an obligation to at least star.
- **Working alternatives table**: in your launch posts, link to PhoneInfoga, ignorant, etc. honestly. "Here's what's better about clank, here's what's better about each alternative." This positions you as a giver, not a competitor.

### 2. Commitment & consistency — micro-commitments

The first commitment is the hardest. Lower it:
- "Read the README" → easier than "install"
- "Try `clank dorks +your-number`" → no setup, runs immediately
- "Star if useful" → easier than "use daily"

Each step that someone takes makes the next step more likely (sunk-cost in their own attention).

### 3. Social proof — show, don't claim

- **GH stars badge** in README: visible count = social proof
- **GH issues opened**: "47 issues filed in week 1" beats "tons of users"
- **Quote screenshots from real users** (Reddit comments, tweets) — only after they're real
- **Used by [logo, name]** — only when true. Lying here = career-ending in OSINT space

### 4. Authority — borrow it honestly

clank already has authority anchors via dependencies. Surface them:
- "Built on `gotd/td` (the same Telegram client Bellingcat uses)"
- "WhatsApp via `whatsmeow` (powers the mautrix-whatsapp bridge)"
- "Phone parsing via Google's libphonenumber"
- "TAC database from MoazEb's 254,993-device snapshot (March 2026)"

Each borrowed authority = a credential without a credentialing process.

### 5. Liking — be a real person

- Personal LinkedIn post under Anshuman's name beats company-style "We launched clank!"
- Photo + handle of the maker (you) in the website footer
- Honest README about what's broken (Snapchat) — humanity beats polish
- Reply to GH issues with first-person voice ("I noticed this too — here's the fix")

### 6. Scarcity — only the honest kind

DO NOT fake scarcity. DO highlight real scarcity:
- "v0.1.0 — first stable release" (genuinely true, only happens once)
- "Asked Bellingcat for a quote — pending" (only if it's actually pending)
- "Submitting to Defcon Demo Labs deadline X" (only if true)

The honesty version: "This is week 1. The roadmap is in `path-1/01-launch-strategy.md`. Want feature X by week 4? Open an issue."

### 7. Unity — the most powerful in 2026

> "Unity is the persuasive power of a shared identity, distinct from Liking which is based on affinity between two individuals." — Cialdini

For clank, the unity is **OSINT investigators / privacy-curious developers / anti-corporate builders**. Speak in that group's language:
- Not "growth hack" → "investigation workflow"
- Not "users" → "investigators" or "operators"
- Not "tap into our API" → "no API. one binary. yours."
- Reference shared enemies (Truecaller's data harvesting, paid services that gatekeep public records)
- Reference shared heroes (Bellingcat, Krebs, OSINT Curious community)

This is the #1 lever for clank specifically. Every post should land you firmly inside the OSINT/maker tribe.

---

## Hook patterns that work in 2026

Per [LinkedIn Hooks Research](https://medium.com/@viralboris/linkedin-hooks-that-actually-work-in-2026-50-examples-bbe7976cde67), three hook types dominate:

### A. Curiosity gap — start mid-story

Bad: "I built clank, an OSINT tool"  
Good: "An abandoned 110-line script from 2014. I rewrote it. The result has 10 subcommands and a TAC database with 254,993 devices."

Why it works: starts mid-action, leaves the "what is it" unanswered until the user clicks.

### B. Contrarian — challenge a common belief

Bad: "Phone OSINT is hard"  
Good: "Most phone OSINT tools haven't shipped a release in 2 years. The OSINT community kept asking 'what's the new PhoneInfoga?' I got tired of waiting."

Why it works: positions clank against an absent/stale incumbent. Triggers "yeah, exactly" from anyone who's looked recently.

### C. Narrative — start in the middle

Bad: "Today I'm releasing clank"  
Good: "3 months ago I had a phone number and 4 hours to figure out who it belonged to. None of the open-source tools worked end-to-end. So I built one."

Why it works: stories trigger different cognition than feature lists. The reader puts themselves in the scenario.

---

## Hook templates (copy these)

For LinkedIn:
```
{Specific surprising fact}.
{One-line why-this-matters}.
{Personal angle}.
{The reveal}.
```

Example:
```
The most popular phone-OSINT tool on GitHub hasn't shipped a release since 2023.
The data is stale. The Snapchat method is broken. Half the messengers don't work.
I rebuilt it. From scratch. In Go. As a single binary.
clank: 10 subcommands, 48 MB, MIT — github.com/AnshumanAtrey/clank
```

For HN:
```
Show HN: clank — a phone-number OSINT toolkit (Go)
```
That's it. Title is the hook. Body in the post:
```
I rebuilt an abandoned 110-line phone-number combinator into a 10-subcommand
OSINT toolkit. It's a single static binary, MIT licensed, and embeds 254,993
device records (TAC database, March 2026 snapshot).

Subcommands: scan (bulk), deep (single), dorks (Google URLs), imei, edgar
(SEC search), telegram (MTProto), whatsapp (whatsmeow), ignorant (3 sites).

What's deliberately broken: Snapchat (their CSRF moved JS-side in 2024,
documented). What's risky at scale: messenger lookups (~50/min ban risk).
What works out-of-the-box: everything else.

Origin story + roadmap in README. Honest about what doesn't work.
```

For Twitter/X (240 chars):
```
I rebuilt an abandoned 110-line phone-OSINT script into a 10-subcommand
toolkit in Go.

Single binary. 48 MB. MIT.
- libphonenumber, IMEI, EDGAR
- Telegram, WhatsApp, Instagram lookup
- 254,993-device TAC database

github.com/AnshumanAtrey/clank
```

---

## What NOT to do — the dark-pattern catalog (and why each is a trap)

### "Limited time offer" / fake urgency
**Trap**: Free CLI tools have no time limit. Lying about urgency damages credibility AND violates EU GDPR Art. 5 + UK CMA guidance.

### "X people viewing right now" badges
**Trap**: Easily fabricated. Easily detected (the script always shows similar numbers). One viral tweet calling it out kills credibility.

### Fake testimonials / fake stars (star-buying)
**Trap**: GitHub detects star-buying via velocity anomalies. They'll deindex the repo from search. Career-damaging if discovered.

### Cross-posting identical copy on Reddit
**Trap**: Reddit's spam filter catches this. Submission goes to spam-shadow before mods even see it. Better: rewrite copy per subreddit.

### Buying upvotes on PH
**Trap**: PH staff manually review #1-#5 launches. Detected = banned permanently. Even small upvote farms are detected by IP/device fingerprint.

### Tagging 10+ influencers in launch tweets
**Trap**: Reads as desperate. Triggers some accounts' auto-block lists. Better: tag 1-2 carefully, with a real reason.

### Fake "Y Combinator backed" / "raised $X" claims
**Trap**: Permanent reputational damage when discovered (it always is).

### "Beta access — request invite" gates on a free CLI tool
**Trap**: Adding friction to a free tool is the cardinal sin. People install or they don't. A waitlist for `go install github.com/.../@latest` is absurd.

---

## High-leverage techniques the launch SHOULD use

### 1. Pre-loading social proof — soft launch first

Send to 10 OSINT/sec friends BEFORE the public launch. Get them to install + star + maybe open a "test" issue. When public launch happens, the GH page shows ~10 stars + 2 issues. Doesn't look dead.

### 2. Asymmetric replies — overinvest in first 5 commenters

The first 5 people who comment on HN/Reddit get TWO paragraphs of substantive response. Future commenters see this and feel obliged to engage seriously. Sets the tone for the thread.

### 3. The "ICE" formula for hook-writing

Per LinkedIn 2026 research, posts that score high on **Identity, Curiosity, Emotion** outperform feature lists 8x.

**Identity**: "If you've ever investigated a scam call, this is for you."  
**Curiosity**: "Most OSINT tools assume you have an API budget. clank doesn't."  
**Emotion**: "I built this because nothing else worked when I needed it."

### 4. The Pratfall Effect — admit one weakness

Counterintuitively, admitting a flaw INCREASES trust. The README's "What's broken: Snapchat (documented)" admission is gold. Lead with it in posts. People who would otherwise pattern-match clank as "another over-promising open-source tool" will recognize the honesty as differentiation.

### 5. The Endowed Progress Effect

Make the README/landing page show a "completion bar" implicitly:
- ✅ Local lookup
- ✅ APIs
- ✅ Telegram
- ✅ WhatsApp
- ✅ EDGAR
- ✅ IMEI
- ✅ Dorks
- ⏳ Truecaller (next)
- ⏳ Email pivot (planned)

This visualizes momentum. Users feel the project is "going somewhere."

### 6. The Identifiable Victim Effect

Don't say "for OSINT investigators." Say "for the journalist who got 47 spam calls last week trying to track the source." Specific imagined users resonate more than generic categories.

### 7. The Foot-in-the-Door

The README's first install command should be ONE LINE:
```
brew install AnshumanAtrey/tap/clank
```
or
```
go install github.com/AnshumanAtrey/clank@latest
```

After that one-line commitment, the user is invested. Then ask them for the harder things (set up TG_APP_ID, pair WhatsApp, etc.).

---

## The two-stage funnel

Per Yu-kai Chou's 2026 framing:

> "Use Cialdini-grade persuasion to earn the first commitment, then hand the user off to Octalysis-grade motivation to keep them."

**Stage 1 — Acquire (Cialdini)**: Hook → land → install. This is what the launch posts do.

**Stage 2 — Retain (intrinsic motivation)**: After install, the tool itself has to deliver. clank's intrinsic-motivation loop:
- Easy first win: `clank deep +<your-own-number>` reveals surprising local data in 3 seconds. Aha moment.
- Variable reward: each scan returns different result mixes. Slot-machine effect.
- Mastery progression: user discovers `--quick`, then `--workers`, then chains with `jq`.
- Identity: "I'm an OSINT person who uses clank" becomes part of self-image.

**The launch sells the install. The tool sells the second use.** Make sure both are tight.

---

## Final ethical line

clank is a tool that, misused, helps stalkers find phone owners. The README's Disclaimer section says so. **Every promotional message should reinforce that this is for journalism, fraud investigation, security research, or knowing your own footprint.** Don't wink-and-nudge at "find that ex" use cases. The OSINT community polices this hard, and rightfully so.

The persuasion in this playbook is to get clank into the hands of the *right* users: investigators, researchers, devs. Not to maximize raw install count at the cost of misuse.

---

## Sources

- [Cialdini's 7 Principles — Influence At Work](https://www.influenceatwork.com/7-principles-of-persuasion/)
- [Cialdini's 7 Principles Explained — Sue Behavioural Design](https://www.suebehaviouraldesign.com/en/blog/cialdini-principles-of-persuasion/)
- [Cialdini's 6 Principles 2026 Guide — Yu-kai Chou](https://yukaichou.com/gamification-analysis/cialdini-6-principles-persuasion-octalysis-gamification-framework/)
- [Cialdini's Principles & Conversions — CXL](https://cxl.com/blog/cialdinis-principles-persuasion/)
- [LinkedIn Hooks That Actually Work 2026](https://medium.com/@viralboris/linkedin-hooks-that-actually-work-in-2026-50-examples-bbe7976cde67)
- [LinkedIn Viral Posts 2026 — Linkboost](https://blog.linkboost.co/linkedin-viral-posts-2026/)
