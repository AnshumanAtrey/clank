# 02 — Platform Research

> Every channel ranked by ROI for clank specifically. Not "which platforms exist" — which ones convert OSINT-curious developers into clank users. Backed by 2026 launch-tactics research.

## Ranking summary (best → worst for clank)

| Rank | Platform | Audience fit | Effort | Expected stars | Note |
|------|----------|--------------|--------|----------------|------|
| 1 | **Hacker News (Show HN)** | ★★★★★ | Low | 200-2000 if frontpage | All-or-nothing. One shot. |
| 2 | **Reddit r/OSINT** | ★★★★★ | Low | 50-300 | Highest practitioner density |
| 3 | **LinkedIn (Anshuman's profile)** | ★★★★ | Low | 30-150 | Warm audience, good for PR |
| 4 | **Product Hunt** | ★★★ | High | 100-500 + backlinks | One shot, needs prep |
| 5 | **Peerlist Launchpad** | ★★★★ | Low | 20-100 | Indian dev community, weekly |
| 6 | **Reddit r/cybersecurity** | ★★★ | Low | 50-200 | Bigger but less aligned |
| 7 | **Twitter/X infosec** | ★★★ | Low | 20-100 | Network effect with right RT |
| 8 | **DEV.to writeup** | ★★★ | Med | 30-100 | SEO long-tail, evergreen |
| 9 | **Reddit r/golang** | ★★★ | Low | 30-150 | "Built in Go" is the angle |
| 10 | **Reddit r/sideproject** | ★★ | Low | 20-80 | Supportive but soft signal |
| 11 | **Mastodon infosec.exchange** | ★★★ | Low | 10-50 | Small but high-quality |
| 12 | **awesome-* PRs** | ★★★ | Low | 5-30/wk drip | Permanent SEO backlinks |
| 13 | **Indie Hackers** | ★★ | Low | 10-50 | Less technical audience |
| 14 | **OSINT Curious newsletter** | ★★★★★ | Very low (you don't control timing) | Could be massive | Pitch, don't post |
| 15 | **Defcon Demo Labs** | ★★★★ | Very high | Could be massive | Annual deadline, slow ROI |

---

## Tier 1 — Must-launch channels

### Hacker News (Show HN)

**URL**: https://news.ycombinator.com/showhn.html

**Audience fit**: Developers who skew technical, FOSS-positive, privacy-conscious. Dead-center for clank.

**Submission rules** (per HN's official Show HN guidelines):
- Submit only finished work, not demos
- "Pure information products" (blog posts, videos) are not Show HN — submit as regular post instead
- Title must be specific: "Show HN: clank – a phone-number OSINT toolkit (Go)"
- No superlatives ("fastest", "best", "first")
- Link directly to GitHub repo, not to a marketing site
- One submission per project. If it dies, you can't repost the same URL

**Optimal timing**: Tuesday-Thursday, 8-11am Eastern (US developer morning).

**Why HN works for CLI/dev tools**: HN community over-indexes on open-source, privacy-first products. Linking to GitHub signals "real working tool, run it yourself." Repos with issue templates, recent commits, and thoughtful READMEs convert to stars.

**What kills HN posts**:
- Marketing-speak ("revolutionize", "game-changer", "AI-powered")
- Linking to a landing page instead of the repo
- Asking for upvotes (instant flag)
- Disappearing from the comment thread for >30 min

**What to do in the comment thread**: Respond to the FIRST 3 comments within 15 minutes. Be honest about what doesn't work yet (Snapchat broken, ban risk in scan). HN respects honesty more than polish.

**Source**: [Show HN guidelines](https://news.ycombinator.com/showhn.html), [How to launch a dev tool on HN — markepear.dev](https://www.markepear.dev/blog/dev-tool-hacker-news-launch)

### Reddit r/OSINT

**URL**: https://reddit.com/r/OSINT (~45k subscribers)

**Audience fit**: Practitioners — investigators, journalists, security researchers, hobbyist sleuths. The single highest density of "people who would actually use clank tonight."

**Submission rules**:
- Subreddit allows tool releases but emphasizes practitioner posts
- Lead with technique/use case, not "check out my tool"
- Self-promotion ratio: comment 5x more than you post

**Optimal timing**: Tuesday-Wednesday morning (US time), avoid weekends.

**Best post format**: "I built a tool because [specific OSINT scenario]. Here's how it handles [specific edge case from the research/]." End with `[Tool] github.com/AnshumanAtrey/clank`.

### LinkedIn (Anshuman's profile)

**Why it's #3 not #1**: Lower technical conversion (LinkedIn audience installs less software than HN), but warmest audience for *you* personally — friends, ex-colleagues, fellow hackathon-people.

**LinkedIn 2026 algorithm key points** (per research):
- **Golden Hour**: First 60-90 minutes determine 70% of reach. Be online + replying.
- **Comments (5+ words) count 2x more than likes**.
- **Document carousels = 6.6% engagement rate** (highest of any format).
- Hooks should "sound like thoughts, not headlines" — authentic voice beats corporate polish.
- Algorithm tracks dwell time. Carousels generate 2-3x more engagement than single images because each swipe = signal.

**Implication for clank post**: Either a 5-image carousel (one per major subcommand) OR a single-image with strong narrative hook. Don't post a link with no preview image.

**Optimal timing**: Tuesday-Thursday 9-11am IST (matches your network's morning scroll).

**Source**: [LinkedIn Algorithm 2026 — Linkboost](https://blog.linkboost.co/how-to-go-viral-on-linkedin-2026/), [LinkedIn Hooks That Actually Work 2026](https://medium.com/@viralboris/linkedin-hooks-that-actually-work-in-2026-50-examples-bbe7976cde67)

---

## Tier 2 — High-leverage but more effort

### Product Hunt

**URL**: https://www.producthunt.com/launch

**Audience fit**: Maker-and-PM heavy. Less technical than HN, more "who else can use this" mindset. Good for backlinks + permanent profile page.

**2026 launch playbook** (per multiple researched guides):
- Launch Tuesday-Thursday, 2nd or 3rd week of month
- 00:01 Pacific Time start (you fight for 24 hours)
- Need a Hunter ideally with high follower count to submit on your behalf, OR submit yourself if you have a strong PH account
- Pre-launch: 30-day prep window standard
- Day-of: 3 people DM-ing on LinkedIn = 200-300 quality upvotes (don't fake it)
- Maker comment must be written 48 hours in advance: tell your story, not your feature list. Why you built it. What feedback you want.
- "Show the product as much as possible" — Product Hunt is surprisingly visual

**For clank specifically**:
- Hook: "I rebuilt an abandoned 110-line phone-OSINT script into a 10-subcommand toolkit"
- Asset: animated GIF of the scan command in action (use vhs.charm.sh or asciinema)
- Tagline: "Phone-number OSINT toolkit. Single static binary. 10 subcommands. MIT."
- Topics: Developer Tools, Open Source, Security, CLI

**Don't launch PH cold** — wait until clank has 200+ GH stars from HN/LinkedIn to use as social proof.

**Sources**: [PH Launch Playbook (30x #1 winner)](https://dev.to/iris1031/product-hunt-launch-playbook-the-definitive-guide-30x-1-winner-1pbh), [Successfully launch dev tools on PH — Ronak Ganatra](https://ronakganatra.com/posts/successfully-launch-dev-tools-on-producthunt), [Smol Launch PH guide](https://smollaunch.com/guides/launching-on-product-hunt)

### Peerlist Launchpad

**URL**: https://peerlist.io/launchpad

**Audience fit**: Indian developer community + global indie devs. Strong tier-2 platform especially aligned with Anshuman's profile (ISU'28, India-based).

**Submission rules**:
- Launches happen Mondays only (weekly cycle)
- Submit via peerlist.io/projects → "Launch Project" from weekly Spotlight banner
- Project must already be linkable from your Peerlist profile

**Why Peerlist is high-ROI for you specifically**: Your profile is already aligned with this audience. Indian devs are over-represented in OSINT/cybersec adjacency. Launching here builds your personal brand more than the tool's user count, which compounds over time.

**Source**: [Peerlist Launch Help docs](https://peerlist.neetokb.com/articles/how-to-launch-your-project-on-peerlist-project-spotlight)

---

## Tier 3 — Drip campaigns (low effort, additive)

### Other Reddit subs (post sequentially, not simultaneously)

| Subreddit | Subs | Angle | Notes |
|-----------|------|-------|-------|
| r/cybersecurity | 800k | "OSINT toolkit for [scenario]" | Bigger but spammier |
| r/golang | 200k | "Built a phone-OSINT tool in Go, learned [lesson about whatsmeow / gotd]" | Lead with lesson |
| r/sideproject | 200k | "I rebuilt an abandoned 2014 script into a real tool" | Personal narrative |
| r/programming | 5M | Risky — large + harsh, only post a writeup not a tool | Unlikely fit |
| r/AskNetsec | 250k | Don't self-promote. Comment helpfully, link clank when relevant | Build credibility first |

**Anti-pattern**: posting to all 4 in the same hour. Reddit spam-detects this. Stagger over 3-4 days.

### Twitter/X

**Audience**: infosec Twitter is alive in 2026 despite the platform's drama.

**Format**: 5-7 tweet thread. Tweet 1 = hook. Tweets 2-6 = one subcommand each with screenshot. Tweet 7 = repo link + "RT if useful."

**Amplification path**: tag relevant accounts (carefully): @bellingcat, @osintessentials, @dutch_osintguy, @CyberSecMeg. ONE tag, not five.

### Mastodon — infosec.exchange

A meaningful number of OSINT folks migrated post-X drama. Smaller audience, higher quality. Cross-post the Twitter thread here.

### awesome-* PRs (permanent backlinks)

Submit PRs to:
- [jivoi/awesome-osint](https://github.com/jivoi/awesome-osint) — the canonical OSINT list
- [avelino/awesome-go](https://github.com/avelino/awesome-go) — under "Phone Numbers"
- [agarrharr/awesome-cli-apps](https://github.com/agarrharr/awesome-cli-apps)
- [Bellingcat's online toolkit](https://bellingcat.gitbook.io/toolkit) — submit via their contribution channel
- [OSINT-BIBLE](https://github.com/frangelbarrera/OSINT-BIBLE) — 2026 comprehensive guide

These PRs take 5 min each and provide evergreen SEO. Worth doing in week +2 once GH stars demonstrate "real tool."

### Indie Hackers

Post the journey, not the tool. Frame: "rebuilt an abandoned project — here's what I learned about open-source maintenance."

### DEV.to long-form

Title: "I rebuilt an abandoned 110-line phone-OSINT script into a 10-subcommand toolkit. Here's the architecture."

Long-form (3000+ words) ranks for SEO indefinitely. Drives a slow trickle for years.

---

## Tier 4 — High-effort, slow ROI (do later, not for v0.1.0)

### OSINT Curious newsletter

**URL**: https://osintcur.io/

The most influential OSINT newsletter. You don't post — you pitch. Email or DM with: "I built clank, here's what's novel: [Telegram + WhatsApp + EDGAR in one binary]. Worth a mention?"

If they include you, expect 200-500 stars overnight.

**Don't pitch until you have 500 stars + a polished website.** Without those, you waste your one shot.

### Defcon Demo Labs / Black Hat Arsenal

Annual deadlines (Defcon ~April, Black Hat ~April-May for August). Submit clank to demo at the conference. If accepted, this is the single biggest amplifier in OSINT/security — but the accept rate is low and prep is significant.

**Worth doing for v0.5.0 in 2027** when clank has Truecaller + 1000+ stars.

### Bellingcat investigation collaboration

Long game. If clank gets used in a real Bellingcat-style investigation and credited, that's the gold standard.

---

## What NOT to do

- **Don't pay for upvotes** anywhere. Gets you banned, kills credibility.
- **Don't cross-post identical copy** across Reddit subs (filter-detected, removed).
- **Don't tag 10 influencers** in one tweet. Reads as desperate, blocks happen.
- **Don't launch on a Friday** anywhere except DEV.to. Friday-launched posts die over the weekend.
- **Don't run paid ads** for v0.1.0. Too early for the unit economics to make sense.
- **Don't email cold journalists** without a real "this was used in [specific investigation]" angle.

---

## Channel-effort vs expected-impact matrix

```
HIGH IMPACT
   |
   |    [HN]                    [PH] [OSINT Curious pitch]
   |  
   |  [r/OSINT] [LinkedIn]
   |
   |    [Peerlist] [Twitter]
   |              [DEV.to]
   |  [awesome-PR drip]    [Defcon Demo Labs]
   |
   |  [Mastodon] [r/sideproject]
   |
LOW |--------------------------- HIGH
EFFORT                          EFFORT
```

Top-left = best ROI. Hit those first (HN, r/OSINT, LinkedIn, Peerlist) before investing in PH or Defcon.

---

## Sources

- [Show HN Guidelines (HN official)](https://news.ycombinator.com/showhn.html)
- [How to launch a dev tool on Hacker News — markepear.dev](https://www.markepear.dev/blog/dev-tool-hacker-news-launch)
- [PH Launch Playbook 30x #1 Winner — DEV](https://dev.to/iris1031/product-hunt-launch-playbook-the-definitive-guide-30x-1-winner-1pbh)
- [Successfully Launch Dev Tools on PH — Ronak Ganatra](https://ronakganatra.com/posts/successfully-launch-dev-tools-on-producthunt)
- [Smol Launch — How to Launch on PH 2026](https://smollaunch.com/guides/launching-on-product-hunt)
- [Peerlist Launchpad Help](https://peerlist.neetokb.com/articles/how-to-launch-your-project-on-peerlist-project-spotlight)
- [LinkedIn Algorithm 2026 — Linkboost](https://blog.linkboost.co/how-to-go-viral-on-linkedin-2026/)
- [LinkedIn Hooks That Actually Work — Viral Boris](https://medium.com/@viralboris/linkedin-hooks-that-actually-work-in-2026-50-examples-bbe7976cde67)
- [LinkedIn Algorithm Feb 2026 — Dataslayer](https://www.dataslayer.ai/blog/linkedin-algorithm-february-2026-whats-working-now)
- [Bellingcat Online Toolkit](https://bellingcat.gitbook.io/toolkit)
- [awesome-osint](https://github.com/jivoi/awesome-osint)
- [March 2026 OSINT Newsletter](https://www.osintnewsletter.osint-jobs.com/p/march-2026-the-osint-developments)
