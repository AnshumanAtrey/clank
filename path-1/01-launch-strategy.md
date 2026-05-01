# 01 — Launch Strategy

> Phased rollout for clank v0.1.0. The goal isn't a one-day spike — it's a 14-day rolling launch that compounds across channels and ends with a self-sustaining trickle of inbound users.

## The thesis

A "Show HN + ProductHunt same day" launch is high variance. You either crack the front page or the post dies in 4 hours. Better: spread the launch across **3 weeks across 8 channels**, where each channel feeds the next:

1. **LinkedIn first** — your warmest audience. Generates the early stars + screenshots that the next channels need as social proof.
2. **Reddit second** — niche subreddits validate technical fit. Pull quotes for HN.
3. **HN third** — the make-or-break for technical credibility. Use the Reddit + LinkedIn momentum as ammo in the comments.
4. **Product Hunt + Peerlist later** — these reward polish, not freshness. Land them after the GH README has been hardened by the first 3 channels' feedback.

## 14-day timeline

```
Week -1 (prep):
  Mon-Wed: Demo recordings (asciinema, 4 casts: scan, deep, dorks, imei)
  Thu:     Website MVP live at clank.atrey.dev (Astro on Cloudflare Pages)
  Fri:     Soft-share to ~10 personal contacts in OSINT/security for QA

Week 0 (launch):
  Mon: Final polish + 1-shot proofread of all post drafts
  Tue 09:00 IST: LinkedIn post goes live (golden-hour engagement window)
  Tue 14:00 IST: Reddit r/OSINT + r/sideproject (afternoon EU + morning US)
  Tue 19:30 IST: HN Show submission (~10am EST, peak HN traffic)
  Wed: Twitter thread (recap with metrics from day 1)
  Thu: Reddit r/golang + r/cybersecurity (different audiences, different copy)
  Fri: DEV.to long-form technical writeup ("rebuilding an abandoned 110-line script")

Week +1 (amplification):
  Mon: Peerlist Launchpad submission (Mondays only)
  Tue: Product Hunt launch (00:01 PT — coordinate with hunters in advance)
  Wed: Post mortem on GH discussions
  Thu-Fri: Submit PR to awesome-osint, awesome-go, awesome-cli-apps

Week +2 (sustained):
  - Weekly LinkedIn: "what I learned from N installs"
  - Address GH issues publicly (visible activity = credibility)
  - Pitch to OSINT Curious newsletter for inclusion
```

## Why this order

**LinkedIn before HN**: HN is a popularity contest in the first hour. If your post starts with 0 upvotes from cold, it dies. LinkedIn warms up your immediate network so when HN goes live, ~5-10 people who saw the LinkedIn post are primed to upvote within minutes. **Don't ask for upvotes** (HN bans for this). But sharing an HN link organically on LinkedIn after submitting is fair game.

**Reddit before HN**: The r/OSINT crowd will give you sharp technical questions. Field 20-30 of these and you'll know exactly what HN commenters will ask. The HN comment thread becomes the most important part of the launch — you need to be ready.

**Peerlist + ProductHunt after**: Both reward polished GitHub READMEs and a working website. You'll only have those after the first wave of feedback. Launching cold on PH wastes the launch slot (you only get 1).

## Pre-launch readiness checklist

- [ ] `clank --version` prints v0.1.0 (currently `dev`)
- [ ] GitHub Release v0.1.0 has download counts visible (already live)
- [ ] README has 1 GIF/asciinema embed at top — first impression matters
- [ ] `clank.atrey.dev` resolves and shows landing page
- [ ] LICENSE exists (✅ done in commit 08a3117)
- [ ] At least one other person has installed clank end-to-end (canary user)
- [ ] All API providers documented with sign-up links in README (✅ done)
- [ ] Issue templates set up in `.github/ISSUE_TEMPLATE/` for bug + feature
- [ ] `gh release view v0.1.0` shows all 5 archives (✅ done)

## Failure-mode planning

**If HN post dies in 2 hours (<10 upvotes)**:
- Don't repost. HN flags reposts.
- Pivot energy to DEV.to long-form. Writeup-style content has 7-day reach.
- Try Show HN again in 6 weeks with a v0.2.0 + a major new feature (Truecaller, etc.).

**If LinkedIn underperforms (<50 reactions in 4 hours)**:
- Adjust the hook (the second-version lives in `04-posts/linkedin.md`).
- Reshare with a different angle 5-7 days later. LinkedIn doesn't penalize this.

**If Reddit moderators remove the post**:
- r/OSINT tolerates self-promo if the post leads with technique, not "check out my tool". The drafted version follows that rule.
- If removed: DM mods politely, ask why, edit, resubmit once.

**If Product Hunt launches at #4-#10 (not #1)**:
- That's still a win. PH page becomes a permanent backlink + Stripe-Atlas-style social proof.
- Don't sweat the rank — sweat the comments. Comments convert.

## What success looks like in 14 days

| Metric | Floor | Target | Stretch |
|--------|-------|--------|---------|
| GitHub stars | 100 | 500 | 2,000 |
| Release downloads | 200 | 1,000 | 5,000 |
| GH issues opened | 5 | 20 | 50 |
| LinkedIn reactions | 50 | 250 | 1,000 |
| HN points | 30 | 150 | 500+ |
| Reddit upvotes (combined) | 100 | 400 | 1,500 |
| Inbound DMs from real users | 3 | 15 | 50 |

**The single most important metric**: number of strangers who open a GitHub issue with a real bug report or feature request. That's the only signal that says "this tool entered someone's actual workflow." Everything else is vanity.

## What to do AFTER launch (week +2 onward)

The launch isn't the goal. The goal is to find the **3-5 power users** whose feedback will shape v0.2.0. Watch:
- GH issues with detailed reproduction steps
- DMs that ask a specific feature request
- Forks that add a feature (then offer to upstream it)

Reply to every one of these within 24 hours for the first month. That converts strangers into champions.

---

> **Remember**: a great tool with bad marketing dies. A mediocre tool with great marketing dies slower. Only a great tool with great marketing AND a tight feedback loop with early users compounds. v0.1.0 is great; the marketing is in this folder; the feedback loop is what you build in week +2.
