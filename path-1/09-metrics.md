# 09 — Metrics & Feedback

> What to track post-launch. Where to listen. How to decide what to build next based on real signal vs vanity numbers.

## The hierarchy of signals (highest → lowest value)

```
1. Strangers opening detailed bug reports / feature requests   ← ⭐ THE signal
2. Strangers forking and adding a feature                       ← strong signal
3. Inbound DMs asking specific questions                        ← strong signal
4. Real-name Twitter/LinkedIn endorsements ("I use this daily")  ← strong signal
5. GitHub issues opened by anyone who is not you                 ← good signal
6. Release downloads (binaries pulled from GH Releases)         ← decent signal
7. GitHub stars                                                  ← weak signal
8. HN/Reddit upvotes                                            ← weak signal
9. LinkedIn reactions                                           ← noise
10. Twitter likes                                               ← pure noise
```

The temptation is to obsess over 7-10 (visible, high-frequency, easy to refresh). The leverage is in 1-3 (rare, slow, one of these is worth 10,000 stars).

## Metrics to track — instrument these

### GitHub Insights (already free, just look)

GitHub repo → Insights tab:
- **Traffic** → Visitors, page views, clones (last 14 days)
- **Stars** → check the trend graph daily
- **Forks** → low count is fine; *what's in the forks* matters more than count
- **Issues opened** vs closed (label them: bug/feature/question)

### Release downloads

```bash
gh release view v0.1.0 --repo AnshumanAtrey/clank --json assets \
  --jq '.assets[] | "\(.name): \(.downloadCount)"'
```

Run weekly. Track in a simple spreadsheet:
| Date | linux_amd64 | linux_arm64 | darwin_arm64 | darwin_x86_64 | windows |
|------|-------------|-------------|--------------|---------------|---------|
| 2026-05-07 | ? | ? | ? | ? | ? |

The platform mix tells you who your users are (Linux server users? macOS dev laptops? Windows analysts?).

### Issue keyword frequency

Every 2 weeks, scan open issues for repeated phrases:
- "Truecaller" mentioned 5x → build it
- "rate limit" mentioned 3x → connection pooling
- "Snapchat" mentioned 8x → fix it
- "Twilio" mentioned 0x → defer

Cheap version of customer interview research.

### Self-instrumentation (if you want)

Optional: ship a privacy-respecting `clank --report-stats` opt-in that posts anonymized usage to a Cloudflare Worker:
- Subcommand counts (no phone numbers, no IPs, no identifiers)
- OS/arch
- Version
- Crash counts

**Strong recommendation**: don't ship this in v0.1.0. The OSINT/privacy crowd will notice and react badly. Ship it in v0.5.0 if at all, opt-in only, with the source code visible.

## Where to listen for signal

### Active monitoring (poll daily)

- GitHub issues page (set up email notifications via Watch → All Activity)
- GitHub Discussions (enable in repo settings; OSS folks like to chat there before opening issues)
- Twitter/X search: `clank phone-osint OR "clank scan" OR "clank deep"` — saved search
- Reddit search: `r/OSINT clank` — saved search
- HN search via [hnsearch.dev](https://hnsearch.dev) for "clank"
- Mastodon search across infosec.exchange

### Passive monitoring (weekly)

- Google Alert for "clank phone OSINT" → email when articles mention it
- GitHub trending Go: `https://github.com/trending/go?since=weekly` — scroll for mentions
- awesome-osint repo activity — watch for clank PRs being merged

### What signals are worth dropping work for

- A real journalist or Bellingcat researcher tweets "I used clank to find X" → engage immediately, ask if they want to do an interview/case study
- A security researcher files a CVE-worthy issue → patch within 24h, publish coordinated disclosure
- A 1k+ follower OSS maintainer stars or RTs → reply with a substantive followup that opens conversation

## Vanity-metric guardrails

Set these limits to protect your time:

- **Stars**: check max once/day. Trends matter, single-day spikes don't.
- **HN upvotes**: check 3x on launch day, 1x/day for the week, 0x after.
- **Twitter likes**: don't track at all.
- **Mentions feed**: schedule 15 min/day, not constant.

## Build-vs-don't-build decision framework

When week +2 inbound feature requests start arriving, sort them by this matrix:

| Request | Frequency | Effort | Strategic | Decision |
|---------|-----------|--------|-----------|----------|
| Truecaller | High (predicted #1 ask) | High (paired-phone import) | Core | **Build for v0.2.0** |
| Snapchat fix | Medium (known broken) | Medium (RE the JS flow) | Polish | **Fix for v0.1.x** |
| Twilio Lookup paid | Low (only paying users) | Low (provider plug) | Optional | **Build if 5+ requests** |
| Web UI | Medium (non-tech askers) | Very high (rewrite) | Off-thesis | **Defer indefinitely** |
| Random one-off (e.g. "support [Z]") | One person | Any | None | **Tag `wontfix-needs-discussion`, ask for use case** |

**Default to no.** A clear "this is out of scope, here's why" is better than a half-built feature that drifts the project.

## Monthly review template

Once a month, write to yourself (just a file, doesn't need publishing):

```
clank — month N review (date)

By the numbers:
  - Stars: X (delta vs last month)
  - Downloads: Y (delta)
  - Open issues: Z (label breakdown)
  - PRs merged from contributors: N

Top 3 user requests this month:
  1.
  2.
  3.

Top 3 things I learned:
  1.
  2.
  3.

What I built:
  -

What I should have built:
  -

What I should NOT have built:
  -

Next month's #1 priority:
  -
```

This becomes the engine of v0.2.0+ planning. 12 of these = a year of compounded learning.

## Definition of "this thing has legs"

clank has legs if, by 3 months post-launch, ANY of these are true:
- 1,000+ GitHub stars (any source)
- 5+ contributors with merged PRs not from you
- Used in 1+ public investigation (journalist, researcher, conference talk)
- Receives 50+ release downloads/week sustained for 4 weeks
- Mentioned in OSINT Curious newsletter or Bellingcat blog
- Forks-with-real-changes count > 3

If NONE of these are true at month 3, clank is a fine portfolio piece but probably won't grow into a community. That's also a valid outcome — most personal projects are this. Decide whether to keep iterating or move on.

## Definition of "this thing is going viral"

If, in any 7-day window, you see:
- 500+ new stars
- A 5-figure-follower account RT/posting about clank without you asking
- A spike of release downloads above 5x baseline
- 10+ issues opened in 24h

Drop everything. Reply to every issue within 6h. Ship a v0.X.Y patch within 48h that addresses the most common new feedback. The goal: convert virality into sustained users by being unusually responsive. Most viral tools die in week 4 because the maker stops shipping.

## What success at month 6 looks like

Realistic stretch case:
- 2,000-5,000 GitHub stars
- 50+ open issues, 30+ closed
- 5-10 regular contributors
- 1-2 public mentions in major OSINT/security press
- A `clank-tutorials` repo started by a fan
- An invite to speak at a small conference
- Inbound from a paid-tier OSINT vendor wanting to "talk partnership"

Realistic floor case:
- 200-500 stars
- 10-20 closed issues
- 0-1 outside contributors
- 0 press mentions
- Steady ~10 releases/week downloaded
- A handful of personal "thank you" DMs

Both are fine. The first is impressive. The second is the median outcome for a single-maintainer side project. Don't measure against vaporware launches you saw on Twitter.

## Tooling

- **GH stars over time**: [star-history.com](https://star-history.com/#AnshumanAtrey/clank)
- **GH dependents**: enable Insights → Dependency graph → Dependents
- **Release downloads over time**: manual via `gh release view` (no built-in time series)
- **Twitter mentions**: use the Twitter advanced search saved-link, or [Tweet Hunter](https://tweethunter.io) free tier
- **Newsletter mention scanner**: [Refind.com](https://refind.com) — track infosec/OSINT newsletters that link clank

## The one metric that matters most

If you only track one thing, track this:

**Number of unique strangers who opened a GitHub issue that wasn't a "thank you" or a duplicate.**

Each one represents a real human who:
1. Found clank
2. Installed it
3. Used it for something
4. Hit something they cared about enough to write up

That's the entire conversion funnel in one number. Optimize the launch + the tool + your replies for moving this number, and everything else takes care of itself.
