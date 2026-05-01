# LinkedIn Post — clank v0.1.0 launch

> Anshuman's voice (per memory: raw, casual, anti-corporate, hacker tone). Single image OR 5-card carousel. Goal: 250 reactions, 30 comments. Post Tue 09:00 IST.

---

## VARIANT A — Pratfall + curiosity (recommended)

```
The most popular phone-OSINT tool on GitHub hasn't shipped a release in 2 years.

The Snapchat method is broken. Half the messengers don't work. The TAC database is from 2014.

Last month I needed to investigate a phone number for a project. Nothing worked end-to-end.

So I rebuilt it.

clank — phone-number OSINT in your terminal:
↳ 10 subcommands
↳ single 48 MB binary
↳ MIT, no signup, no API gate
↳ 254,993-device IMEI database (March 2026 snapshot)
↳ Telegram + WhatsApp + Instagram lookup
↳ SEC EDGAR full-text search

What's deliberately broken: Snapchat (their CSRF moved JS-side in 2024 — documented).
What's risky at scale: WhatsApp/Telegram bans above ~50/min — also documented.
What works out of the box: everything else.

Open-source. Pull the binary or `go install`.

→ github.com/AnshumanAtrey/clank
→ clank.atrey.dev

If you've ever tried to figure out who a number belongs to and given up, this is for you.

#OSINT #cybersecurity #opensource #golang #buildinpublic
```

**Hashtags rationale**: 5 max, mix general (OSINT, cybersecurity) with specific (golang, buildinpublic, opensource). LinkedIn caps reach for >5 hashtags.

**Image**: Single hero — terminal screenshot of `clank deep +14155552671` output (the same one in 05-website-copy.md hero). Format 1080x1350 (LinkedIn portrait, biggest in feed).

---

## VARIANT B — Contrarian (backup if A flops)

```
Hot take: most "OSINT toolkits" are graveyards.

I checked the top 10 phone-OSINT tools on GitHub last month. 7 of them haven't merged a PR in 18 months. 3 of them have broken core features that nobody's fixed.

The OSINT community is huge. The maintainers are exhausted. The tools rot.

So I built clank. Phone OSINT. One binary. 10 subcommands. Open-source.

The pact I'm making with myself:
↳ Ship a release every 2 weeks
↳ Reply to every issue within 48h
↳ Be honest about what's broken (currently: Snapchat lookup)
↳ Never gate features behind API keys I control

If you've ever opened a phone-OSINT tool's GitHub and seen "last commit: 2 years ago" — this one's the opposite.

→ github.com/AnshumanAtrey/clank

What other OSINT tools deserve a 2026 rewrite? Drop them in the comments — I might pick one for v0.3.0.
```

**Why B works as backup**: stronger contrarian hook, ends with engagement-prompt question (LinkedIn algo loves comments).

---

## VARIANT C — Personal narrative (longest, story-driven)

```
3 months ago I had a phone number. +91-something. Suspected scam. 4 hours to figure out who owned it.

I tried every open-source phone-OSINT tool I could find:
↳ PhoneInfoga: half the modules deprecated, Snapchat broken, last release 2023
↳ ignorant: Python deps from hell, 2/3 sites broken
↳ phone-checker: actually worked but only 1 source

I gave up after 3 hours. Spent the 4th hour reading research papers on phone-OSINT instead.

That weekend I started writing clank.

3 months later it's:
- 10 subcommands
- 48 MB single binary
- libphonenumber + IMEI (254k device DB) + EDGAR + Telegram + WhatsApp + Instagram + Google dorks
- MIT, no signup, no API gate
- v0.1.0 shipped today

The number from month 3? Still unknown. But now I have the tool I wished I'd had then.

If you're an OSINT investigator, journalist, security researcher, or just curious about your own digital footprint — give it a spin.

→ github.com/AnshumanAtrey/clank
→ clank.atrey.dev

Honest about what's broken (Snapchat), risky (messenger bans at scale), and what just works (everything else).
```

**Why C is variant**: longer, more personal, slightly higher emotional pull. Use if your network responds to story-driven posts.

---

## Carousel option (5 slides — for highest LinkedIn engagement per 2026 algo research)

If you want max engagement (carousels = 6.6% engagement vs 2-3% for single image):

**Slide 1**: Hook (text only on dark bg)
```
The most popular phone-OSINT tool
hasn't shipped in 2 years.

So I built a new one.
[swipe →]
```

**Slide 2**: What it does
```
clank — phone OSINT in your terminal

→ 10 subcommands
→ single 48 MB binary
→ MIT, open source
→ no signup, no API gate
[swipe →]
```

**Slide 3**: Demo screenshot (terminal output of `clank deep`)

**Slide 4**: What's broken (the Pratfall slide)
```
What I won't pretend works:

× Snapchat lookup (broken since 2024)
× WhatsApp/Telegram at scale (~50/min ban risk)
× Pastebin search (psbdmp.cc dead)

Documented inline. v0.1.0 doesn't oversell.
[swipe →]
```

**Slide 5**: CTA
```
Try it now:

go install github.com/AnshumanAtrey/clank@latest

Or download the binary:
github.com/AnshumanAtrey/clank/releases

Anshuman Atrey · clank.atrey.dev
```

Carousel takes longer to make but converts ~2x better. Use Canva or Figma to make 5 portrait images (1080x1350), upload as PDF document to LinkedIn.

---

## Reply scripts (prepare these in advance)

When comments come in, you have 3-5 minutes to respond before the algo loses interest. Prepare these:

**"How does this compare to PhoneInfoga?"**
```
Honestly: PhoneInfoga has the web UI and more provider integrations.
clank is faster (Go binary, no Python deps), has Telegram + WhatsApp +
EDGAR (PhoneInfoga doesn't), and is honest about what's broken.

Use both. They're complementary.
```

**"Is this legal?"**
```
Same as any web search — clank queries public data via documented
endpoints. Misuse (stalking, harassment) is illegal regardless of tool.
README has a Disclaimer section spelling out the legitimate use cases:
fraud investigation, journalism, security research, knowing your own
footprint.
```

**"Will you add Truecaller?"**
```
Top of v0.2.0 roadmap. Manual paired-phone import flow per the research
files. ETA: ~3 weeks.
```

**"Built in Go — why?"**
```
Single static binary, no Python dep hell, 48 MB ships everywhere
(Linux/macOS/Windows × amd64/arm64), and the Telegram/WhatsApp libraries
in Go (gotd/td and whatsmeow) are actually maintained. The only language
where I could ship all 10 features in one binary.
```

**"Do you log my queries?"**
```
Zero telemetry. Zero analytics. Zero outbound calls except to the OSINT
sources you explicitly invoke. Local audit log at ~/.clank/history.jsonl
(your machine, never sent anywhere). Set CLANK_NO_AUDIT=1 to disable
even that.
```

**"How can I contribute?"**
```
PRs welcome. The roadmap is in the path-1/ folder + GH issues.
Top wishlist items: Truecaller, Snapchat fix, connection pooling
in scan, email-pivot via Hunter.io. Pick one + open a PR, I'll
review same day.
```

**Generic positive comment** ("nice work!", "love this", etc.):
```
Thanks [name]! If you try it, I'd love to hear what you scan first
— always curious what people pivot to.
```

(The follow-up question keeps the convo open and signals to the algo that the post is generating active discussion.)

---

## Post-launch followup (day 3 if first post lands well)

```
clank update — 48 hours after launch:

→ 247 GitHub stars (thank you)
→ 19 issues opened (3 are real bugs, fixing this weekend)
→ 1 person used it on a real journalism investigation (their words)
→ 14 inbound DMs asking about Truecaller — moved to top of v0.2.0

The most surprising thing I learned in 48 hours: people install OSINT tools faster than I expected, and bug-report them slower. Half the issues are "this would be useful if it ALSO did X." That's a great problem to have.

Next: Truecaller integration + connection pooling for scan.

Roadmap: github.com/AnshumanAtrey/clank/issues
```

This week-2 followup keeps the audience engaged and converts curious-stargazers into actual users.

---

## Don't post if

- It's a Friday (LinkedIn dies for the weekend)
- You can't be online for the next 90 minutes (Golden Hour wasted)
- You haven't slept (replies will read flat)
- A major news story is dominating LinkedIn (your post will be invisible)
