# path-1 — Get Users, Let Demand Drive Priorities

> **Path 1 of 3** from the post-v0.1.0 strategy session. The thesis: a great tool with no users dies. Before building v0.2.0 features, get clank in front of OSINT investigators, security devs, and the indie-hacker community. Their feedback shapes what to build next.

## What's in this folder

This is the entire launch + website + community playbook for clank v0.1.0. **18 files, ~5000 lines.** Read in order, or jump to what you need.

### Strategy & decisions

| File | What it covers |
|------|----------------|
| [`01-launch-strategy.md`](./01-launch-strategy.md) | 14-day phased rollout. Why LinkedIn → Reddit → HN → PH order. Failure-mode planning. Success metrics. |
| [`02-platforms.md`](./02-platforms.md) | All 14 launch platforms ranked by ROI. Submission rules, optimal timing, audience fit per platform. Backed by 2026 research. |
| [`03-persuasion.md`](./03-persuasion.md) | Cialdini's 7 principles applied to clank. Hook patterns. Dark-pattern catalog with reasons each is a trap. Reply scripts. |
| [`07-repo-architecture.md`](./07-repo-architecture.md) | Should the website live in clank's repo or a separate one? Decision matrix + recommendation: **separate repo `clank-web`**. |

### Ready-to-paste post drafts

The `04-posts/` subfolder has platform-specific copy. Each file is a drop-in launch.

| File | Platform | Notes |
|------|----------|-------|
| [`04-posts/linkedin.md`](./04-posts/linkedin.md) | LinkedIn | 3 variants (Pratfall, Contrarian, Narrative) + 5-card carousel option + reply scripts |
| [`04-posts/hn-show.md`](./04-posts/hn-show.md) | Hacker News (Show HN) | Title, body, comment-thread reply templates, account-hygiene tips |
| [`04-posts/twitter-thread.md`](./04-posts/twitter-thread.md) | Twitter/X | 7-tweet thread + Mastodon cross-post variant |
| [`04-posts/producthunt.md`](./04-posts/producthunt.md) | Product Hunt | T-30 / T-7 / T-1 / launch-day playbook + maker comment |
| [`04-posts/peerlist.md`](./04-posts/peerlist.md) | Peerlist Launchpad | Submission form fields + maker note |
| [`04-posts/reddit-osint.md`](./04-posts/reddit-osint.md) | r/OSINT | Lead with technique. The most aligned audience. |
| [`04-posts/reddit-go.md`](./04-posts/reddit-go.md) | r/golang | Lead with Go libraries (gotd/td, whatsmeow, goreleaser) |
| [`04-posts/reddit-sideproject.md`](./04-posts/reddit-sideproject.md) | r/sideproject | Personal narrative + lessons learned |
| [`04-posts/reddit-cybersecurity.md`](./04-posts/reddit-cybersecurity.md) | r/cybersecurity | Frame for SOC/IR defenders |

**Important**: every Reddit draft is *different copy*. Don't cross-post identical text — Reddit's spam filter detects this and shadow-bans.

### Website (clank.atrey.dev)

| File | What it covers |
|------|----------------|
| [`05-website-copy.md`](./05-website-copy.md) | All landing page text. Hero, features, install, FAQ, footer. 4 hero-variant tests. Copy that goes into the Astro site. |
| [`06-website-tech.md`](./06-website-tech.md) | Stack: **Astro + asciinema-player + Cloudflare Pages**. Animation strategy. Custom domain setup. Performance budget. |

### Operations

| File | What it covers |
|------|----------------|
| [`08-launch-day.md`](./08-launch-day.md) | Hour-by-hour Tuesday playbook. Pin this in a tab. Crisis-mode protocols. Energy management. |
| [`09-metrics.md`](./09-metrics.md) | What to track post-launch. Where to listen. Build-vs-don't-build decision framework. The one metric that matters. |

## Suggested reading order

If you have ~30 minutes and just want the highlights:
1. [`01-launch-strategy.md`](./01-launch-strategy.md) — the master plan (10 min)
2. [`07-repo-architecture.md`](./07-repo-architecture.md) — answer to your "where does the website live?" question (5 min)
3. [`08-launch-day.md`](./08-launch-day.md) — the operational playbook for launch day (10 min)
4. Skim [`02-platforms.md`](./02-platforms.md) (5 min)

If you have a full evening (~2 hours):
- Read everything in order: 01 → 02 → 03 → 07 → 05 → 06 → 08 → 09 → 04-posts/*

## Decisions already made (so you don't have to re-litigate)

1. **Website lives in a separate repo** — `AnshumanAtrey/clank-web` (see `07-repo-architecture.md`)
2. **Stack is Astro + Cloudflare Pages** (see `06-website-tech.md`)
3. **Domain is `clank.atrey.dev`** via Cloudflare Pages custom domain
4. **Launch order is LinkedIn → Reddit → HN → DEV → Peerlist → ProductHunt** (over 14 days)
5. **Tone is honest about brokenness** — Snapchat broken, scan ban risk, etc. Pratfall Effect.
6. **No dark patterns** — fake scarcity, fake testimonials, bought upvotes are off-limits.
7. **Author identity travels everywhere** — atrey.dev / linkedin.com/in/anshumanatrey/ / build@atrey.dev in every artifact (already shipped in commit `e6b67ad`)

## What this folder is NOT

- ❌ Not a marketing strategy for v0.2.0+ — those launches will need fresh research (TAM grows, channels change)
- ❌ Not a comprehensive PR playbook — no journalist outreach yet (premature without v0.2.0 + 1k stars)
- ❌ Not a paid-acquisition playbook — running ads on a free CLI tool doesn't make sense yet
- ❌ Not engineering/feature roadmap — that lives in GH issues and the main README

## How to use this when you actually launch

1. **2 weeks before launch**: read everything once. Schedule the launch day. Build the website (`clank-web` repo).
2. **1 week before**: soft-share to ~10 contacts. Final QA on website + GH repo. Charge laptop.
3. **Launch day Tuesday**: follow `08-launch-day.md` hour by hour. Don't deviate.
4. **Week +1**: amplification (Peerlist Mon, PH Tue, DEV.to Fri).
5. **Week +2 onwards**: triage GH issues using `09-metrics.md` framework.

## When this folder becomes outdated

These files are calibrated to v0.1.0 launch. They'll go stale when:
- v0.2.0 ships with major new features (Truecaller, etc.) — rewrite hooks
- Platform algorithms change significantly (LinkedIn algo updates regularly)
- A platform shuts down or pivots (e.g. ProductHunt launches a new format)

Re-research and rewrite this folder before v0.5.0. The principles (Cialdini, Pratfall, golden hour) are timeless. The specifics (which subreddit, which timing) decay.

## What's NOT in this folder

Two future docs that should exist but don't yet:
- `10-content-calendar.md` — week +2 onwards content (technical writeups, tutorials, case studies). Build after seeing what the launch reveals.
- `11-community-handbook.md` — how to manage GH issues, set up Discord, run a contributor-friendly project. Build at month +1 if there's a real community to manage.

## Author

Anshuman Atrey · [atrey.dev](https://atrey.dev) · [linkedin.com/in/anshumanatrey/](https://linkedin.com/in/anshumanatrey/) · [build@atrey.dev](mailto:build@atrey.dev)

## License

Same as parent repo: MIT.
