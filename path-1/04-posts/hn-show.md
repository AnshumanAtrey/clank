# Hacker News — Show HN submission

> One shot. Title is everything. Submit Tue 19:30 IST (≈09:00 EST). Stay in the comment thread for 90+ min after submission.

## Title (paste exactly)

```
Show HN: Clank – phone-number OSINT toolkit (Go)
```

**Why this title**:
- "Show HN:" prefix is mandatory for the Show HN section
- "Clank" capitalized (HN convention for project names in titles)
- Em-dash (–, not hyphen -) is HN's house style
- "phone-number OSINT toolkit" is concrete — tells you what it does in 4 words
- "(Go)" parens signal language for filter readers; HN community values language disclosure

**What NOT to put in the title**:
- ❌ "Show HN: Clank — the BEST phone-OSINT tool" (no superlatives, per HN guidelines)
- ❌ "Show HN: I built clank" (no first person in titles)
- ❌ "Show HN: Clank | A revolutionary OSINT platform" (no "platform", no "revolutionary")
- ❌ "Show HN: Clank, a Free Phone OSINT Tool That Works in 2026" (too long, too marketing)

## URL field

```
https://github.com/AnshumanAtrey/clank
```

**Critical**: link to the GitHub repo, NOT clank.atrey.dev. HN's data shows GitHub-linked Show HN posts convert ~2x better than landing-page-linked ones for dev tools. The repo proves "real working code" instantly. The README does the marketing.

## Body (paste in the text field)

```
I rebuilt an abandoned 2014 phone-number combinator into a 10-subcommand
OSINT toolkit. Single static Go binary, 48 MB, MIT.

Subcommands:
- scan: bulk pattern → libphonenumber filter → enrich each survivor
- deep: single number, all 6 sources concurrently
- dorks: ~160 Google search URLs across 5 buckets
- imei: Luhn check + 254,993-device TAC lookup (Mar 2026 snapshot)
- edgar: SEC EDGAR full-text filings search (no auth)
- telegram: MTProto phone-to-user via gotd/td
- whatsapp: phone-to-user via whatsmeow (QR-pair session)
- ignorant: Instagram + Snapchat + Amazon presence checks
- history: ~/.clank/history.jsonl audit log
- --version

What works out of the box: libphonenumber lookup, IMEI, dorks, EDGAR.
What needs setup: Telegram (TG_APP_ID + login), WhatsApp (QR pair).
What's deliberately broken: Snapchat (their CSRF moved JS-side in 2024,
documented inline). What's risky at scale: messenger bans above ~50/min.

The path here was: I needed to investigate a phone number, tried every
open-source tool, all of them either had broken modules or hadn't shipped
a release since 2023. Started rewriting on a weekend. 3 months later it
became this.

Origin commits + research notes are in the repo. README has the full
subcommand docs and an honest "what doesn't work" section. Roadmap
includes Truecaller integration, Snapchat CSRF rediscovery, and
connection pooling for the scan command.

Happy to dig into any of:
- whatsmeow's bug #1086 (silently-drops-invalid-numbers) and how
  clank reconciles
- gotd/td's session persistence + FLOOD_WAIT handling
- Why I bundled the 254k-device TAC database (MoazEb's) instead of
  fetching at runtime
- The trade-off between "ship as Go single-binary" vs "Python with
  deps for richer ecosystem"
```

**Length**: ~280 words. HN body norms: 100-400 words. Below 100 = too thin. Above 500 = TLDR, people skip.

**What this body does**:
1. Opens with the concrete scope (10 subcommands, single binary)
2. Lists every subcommand briefly — proves it's not vaporware
3. Honest status (Pratfall — what's broken)
4. Personal origin story (1 paragraph — humanizes)
5. Ends with 4 specific technical-rabbit-hole topics — primes commenters to ask deep questions

The last paragraph is the most important. It tells HN commenters *which* directions are worth probing, which generates the substantive comment thread that drives upvotes.

## Pre-submit checklist

- [ ] You've eaten in the last 2 hours
- [ ] You have 90 min of clear schedule ahead (no calls, no meetings)
- [ ] You've reread the README and can speak to any line
- [ ] You've run `clank --version` and `clank deep +<your-own-number>` in the last hour (so you can describe live behavior)
- [ ] All your post-launch demo URLs are live (clank.atrey.dev resolves)
- [ ] HN account is at least 2 weeks old with some prior comments (HN penalizes fresh accounts on Show HN)

## After submitting

### First 15 minutes
- Refresh `news.ycombinator.com/show` once after 5 min — confirm post appears
- Don't refresh obsessively after that

### First 60 minutes
- Reply to every comment within 8 min
- Each reply: 1-3 sentences, technical, honest
- If criticism: acknowledge directly. Never argue.
- Pin a comment? No — HN doesn't allow pinning, but the first comment you write yourself will sit visibly near the top.

### First 90 minutes — peak window
- This is when post either climbs to /show (Show HN page) or dies
- 30+ points + 10+ comments at this mark = climbing
- Under 5 points + 0 comments = dying; pivot energy elsewhere
- Front page (/news) requires ~50+ points in 1-2 hours typically

### Reply templates for HN

**Critical comment about your code/design**:
```
Fair criticism. The reason for [decision] was [specific reason], but
you're right that [tradeoff]. Opened issue #X to track refactoring
this — appreciate the eye on it.
```

**"Why didn't you use [alternative]?"**:
```
Considered [alternative]. Went with [chose] because [specific tradeoff].
Open to revisiting if there's a real-world case where [chose] hits a wall —
which scenario specifically were you thinking?
```

**"How does it compare to PhoneInfoga / [competitor]?"**:
```
PhoneInfoga has the web UI, more API providers, and is more battle-tested.
clank is faster (Go binary, no Python deps), has Telegram/WhatsApp/EDGAR
that PhoneInfoga lacks, and is honest about what's broken (Snapchat).
Use both — they're complementary. The README's roadmap section has the
full comparison.
```

**"Is this legal? Can I get in trouble?"**:
```
clank queries public data via documented endpoints — same as a web search.
The Disclaimer in the README spells out legitimate use: fraud investigation,
journalism, security research, knowing your own footprint. Misuse is illegal
regardless of tool. I'm not a lawyer; check your jurisdiction.
```

**"Why bundle the IMEI database instead of fetching live?"**:
```
Three reasons: (1) offline reliability — works without network, (2) the
dataset is 254k records / 3.5 MB compressed, fits in a Go binary fine,
(3) MoazEb's source updates monthly, so I plan a `clank imei --update`
subcommand to refresh without binary update. Deferred to v0.2.0 — issue #X.
```

**"Show HN posts shouldn't be this polished — feels promotional"**:
```
Fair concern. Promise it's not corporate — single maintainer, weekend
project. The README is detailed because the OSINT space is full of broken
abandoned tools and I wanted the docs to outlast me if I get hit by a
bus. Nothing's behind a paywall, no signup, MIT.
```

## What to do if it hits front page (/news)

If HN's algo upranks past Show HN to the main page:
1. **Cancel everything** for the next 4 hours. Inbox-zero on the comment thread.
2. Reply to comments faster (sub-5-min)
3. **Don't** push code changes (something will break)
4. **Don't** tweet "we're on HN" — gauche
5. **Do** screenshot the moment for personal records
6. Post a thank-you on LinkedIn ~6h later when the spike is over

## What to do if it dies

If 2 hours in you have <10 points and 0-1 comments:
1. Don't delete the post — HN tracks deletions
2. Don't repost — HN flags reposts
3. Move on to Reddit/Twitter
4. Try Show HN again in 6 weeks with a v0.2.0 + a major new feature
5. The fact that the post died doesn't mean clank is bad — HN is high variance

## A note on HN account hygiene

If your HN account is brand new or has 0 prior comments, your Show HN gets penalized hard.

**Pre-launch (week before)**:
- Comment substantively on 5-10 posts in `/news` — technical replies, not "great post!"
- Submit 1-2 non-promotional links (interesting articles you actually read)
- Build account karma to ~10+ before launch

This is unmanipulable signal that shows you're a real HN-er, not just here to promote.

## Sources

- [Show HN Guidelines (HN official)](https://news.ycombinator.com/showhn.html)
- [How to launch a dev tool on Hacker News — markepear.dev](https://www.markepear.dev/blog/dev-tool-hacker-news-launch)
