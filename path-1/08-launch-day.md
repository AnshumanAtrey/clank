# 08 — Launch Day Playbook

> Hour-by-hour timeline for the actual day clank goes public. Pin this in a tab. Reference it through the day. Times in IST (Anshuman's timezone) with US/EU conversions.

## The day before (Monday)

- [ ] Final read of all post drafts in `04-posts/` — proofread, fix typos
- [ ] Verify clank.atrey.dev is live and OG image renders
- [ ] Verify GitHub repo:
  - [ ] README has the demo GIF/asciinema at top
  - [ ] LICENSE is present (✅ done)
  - [ ] Issue templates exist
  - [ ] Latest release v0.1.0 visible with all 5 archives
  - [ ] At least one open issue (to show "real project, real activity")
- [ ] Soft-share to ~10 personal contacts in OSINT/sec — collect early stars
- [ ] Charge laptop, plug in, hydrate
- [ ] Block calendar 6am-10pm Tuesday for launch
- [ ] Set Slack/Discord status to "launching clank, slow to respond"

## Launch day (Tuesday) — hour by hour

### 06:00 IST — Pre-flight (00:30 UTC, evening US Pacific Mon)

- Wake up early or stay up late, your call. The next 18 hours are dense.
- Coffee/chai. Don't skip food — you'll forget by 11am.
- Re-verify: clank.atrey.dev still live? GitHub repo public? `gh release view v0.1.0` works?

### 07:00 IST — Final checks

- [ ] Open these tabs and pin them:
  1. clank.atrey.dev (your own landing page)
  2. github.com/AnshumanAtrey/clank
  3. github.com/AnshumanAtrey/clank/issues
  4. linkedin.com (your own profile)
  5. news.ycombinator.com
  6. reddit.com/r/OSINT
  7. (later: ProductHunt, Peerlist tabs)
- [ ] Open `04-posts/` folder for copy-paste reference
- [ ] Open metrics tab: GitHub Insights → Traffic
- [ ] Have a notepad open for "people who DM'd / commented" — you'll thank everyone individually post-launch

### 09:00 IST — LinkedIn post goes live (03:30 UTC, 23:30 ET Mon)

**Why now**: Your LinkedIn network is South Asian + global devs. 9am IST catches India's morning scroll AND lunch-break Europe.

- [ ] Copy `04-posts/linkedin.md` (Variant A) into LinkedIn composer
- [ ] Upload the 5-image carousel (one per major subcommand) OR single hero image
- [ ] Hit Post
- [ ] **DO NOT close LinkedIn** — the next 60 minutes are the Golden Hour (70% of post reach is determined here)
- [ ] Reply to every comment within 5 minutes for the first hour
- [ ] DM your closest 5-10 contacts asking them to comment (not like — comment with 5+ words)

### 09:30-10:30 IST — Golden Hour engagement

- [ ] Reply to every comment with a substantive sentence (not just "thanks!")
- [ ] If a comment asks a question, reply + tag a related person who'd find it interesting
- [ ] Reshare 2-3 of your own LinkedIn connections' recent posts (LinkedIn rewards reciprocity in 1-day windows)

### 11:00 IST — First metrics check (05:30 UTC)

If LinkedIn has 30+ reactions and 5+ comments, you're on track. If under 10 reactions in 2 hours, the hook isn't landing — abandon this post and try Variant B at 3pm.

### 12:00 IST — Lunch + buffer (06:30 UTC)

Eat. Don't post during lunch. Let LinkedIn breathe.

### 14:00 IST — Reddit r/OSINT (08:30 UTC, 04:30 ET, 09:30 GMT)

**Why now**: Catches end-of-morning EU and start-of-morning US East.

- [ ] Copy `04-posts/reddit-osint.md` to Reddit
- [ ] Title format: lead with the use case, not the tool name
- [ ] Submit
- [ ] **CRITICAL**: monitor for 30 min. If a mod removes it, DM mods politely and ask why — adjust + resubmit *once*
- [ ] When comments come in, reply within 10 min for the first 3-5 comments
- [ ] If asked "how does this compare to PhoneInfoga?", be ready with the honest comparison (drafted in `02-platforms.md`)

### 14:30 IST — Reddit r/sideproject

- [ ] Different copy from r/OSINT (anti-cross-post-detection)
- [ ] Lead with the personal story angle
- [ ] Same engagement rules — first 30 min critical

### 16:00 IST — Reddit metrics check (10:30 UTC)

If r/OSINT post has 20+ upvotes and 3+ comments, the technical audience is biting. If under 5 upvotes in 2 hours, post will die — let it.

### 17:00 IST — Tea break (11:30 UTC)

Breathe. Don't post. Don't refresh LinkedIn obsessively (you've already lost the dwell-time game if you're not active in real conversations).

### 18:00 IST — Twitter thread (12:30 UTC)

**Why now**: Catches US morning. Twitter is the bridge from EU+India audiences to US.

- [ ] Post the 5-7 tweet thread from `04-posts/twitter-thread.md`
- [ ] First tweet has the OG image attached
- [ ] Each subsequent tweet has either a screenshot or a short demo GIF
- [ ] Final tweet: link to GH repo + ask "what tool would you build next on top of this?"
- [ ] Tag 1-2 (NOT MORE) accounts: try @bellingcat or @osintessentials with a *real* question, not just self-promo

### 19:30 IST — HN Show post (14:00 UTC, 09:00 EST, 15:00 CET)

**Why now**: This is HN's peak window for technical posts. You want to land before the US workday's mid-morning coffee scroll.

- [ ] Title (exactly): `Show HN: Clank – phone-number OSINT toolkit (Go)`
  - Note: HN sentence-cases the project name. "Show HN: Clank" not "Show HN: clank"
  - The `(Go)` parens signal language to filter readers
- [ ] URL: link to the GitHub repo, NOT clank.atrey.dev
- [ ] Body: the carefully-drafted post from `04-posts/hn-show.md`
- [ ] Submit
- [ ] **STAY ON HN FOR 90 MINUTES MINIMUM**. Reply to every comment within 10 min for the first hour. The first 5 commenters set the tone.
- [ ] If a comment is critical, reply substantively — never defensive. HN respects "you're right, I'll fix that" 100x more than excuses.
- [ ] If the post hits the front page, your day shifts: focus 100% on the comment thread for the next 4 hours.

### 20:00 IST — Concurrent: monitor everything (14:30 UTC)

Tab cycle every 5 min:
- HN comments
- LinkedIn replies
- Reddit comments
- GH issues (people who tried and hit a bug)
- GH stars trend (Insights → Traffic shows real-time)

### 22:00 IST — Wind down (16:30 UTC)

- [ ] Final comment-reply sweep across all channels
- [ ] Set up a "post-launch follow-up" reminder for tomorrow morning
- [ ] Close all tabs except GH issues
- [ ] Sleep. Don't pull an all-nighter — Wednesday matters too.

## Wednesday morning — recovery + amplification

### 08:00 IST

- [ ] Read every comment from yesterday with fresh eyes — reply to ones you missed
- [ ] Triage GH issues: anything urgent? Anything that suggests a real bug?
- [ ] Tweet a "thank you" + metrics update if numbers are good ("clank just crossed 100 stars overnight — thanks everyone")
- [ ] Reply to inbound DMs

### 14:00 IST — DEV.to long-form

- [ ] Post the writeup version: "I rebuilt an abandoned 110-line script into a 10-subcommand toolkit. Here's the architecture."
- [ ] Tag: #go #osint #cli #opensource #devtools

### Thursday-Friday

- [ ] Reddit r/golang (the "built in Go, learned X" angle)
- [ ] Reddit r/cybersecurity (broader audience, OSINT use case)
- [ ] DM the 5-10 power users who installed and starred — ask what's missing

## Following Monday — Peerlist + ProductHunt prep

- [ ] Submit to Peerlist Launchpad (Mondays only)
- [ ] Schedule ProductHunt launch for Tuesday following

## Crisis-mode protocols

### "HN comment thread is going wrong"

If HN commenters are pile-on critical (rare but possible):
1. Don't argue
2. Don't delete comments
3. Pick the most substantive criticism, reply with "you're right — opened issue #X to track" — public commitment defuses
4. Stop replying to bad-faith threads after 1 substantive reply
5. Move on

### "Someone found a security bug"

If a launch comment reveals a real vulnerability (e.g., hardcoded API key, unsafe parsing):
1. Don't dismiss
2. Reply: "thank you — opening private issue, will patch within 24h"
3. Patch it that day
4. Push v0.1.1
5. Announce the fix publicly — vulnerability response done well = trust gained

### "Subreddit removed my post"

1. Don't argue with mods publicly
2. DM mods politely, ask the rule
3. Adjust copy to fit the rule
4. Resubmit ONCE
5. If removed again, accept and move on

### "Numbers are below floor (HN <30, LinkedIn <50, etc.)"

Don't panic. v0.1.0 is the start, not the end. The launch isn't a referendum on whether clank is good — it's the first ping. Iterate to v0.2.0 with feedback from the comments you DID get, and try a smaller-scale relaunch in 6 weeks.

## What to NOT do on launch day

- ❌ Don't delete bad comments (looks defensive)
- ❌ Don't respond to trolls (drains your energy, energy is your scarcest resource today)
- ❌ Don't push code changes (something will break; only fix CRITICAL bugs)
- ❌ Don't engage in unrelated drama on Twitter (you'll get tagged, ignore it)
- ❌ Don't repost on Reddit if removed (mods will ban)
- ❌ Don't cross-post HN to other HN-style sites (lobsters, etc.) for 24h after HN
- ❌ Don't @ random influencers begging for retweets (career-damaging)
- ❌ Don't open more PRs to your own repo today (looks staged)

## Energy management

Launch day is a full-day endurance event. Plan:
- 8 hours of sleep before
- 3 meals + 2 snacks
- Hydration: water bottle on desk, refill 4x
- Two 20-min walks (10am and 4pm) — refreshes your responses
- Don't doom-scroll metrics. Set 30-min check intervals, not constant refreshing.

## After 48 hours

The launch is over. Whatever happened, happened. Pivot to:
- Building the most-requested feature
- Replying to GH issues with first-class care
- Writing the week-2 follow-up content (`16-followup-content.md` — TBD)

The launch makes a tool *visible*. The week after the launch makes it *real*.
