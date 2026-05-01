# Product Hunt — clank launch

> Launch in week +1, NOT same day as HN. PH rewards polish + momentum, both of which you'll have after the HN/Reddit waves. Tuesday 00:01 PT start time.

## Pre-launch (T-30 days)

Per [PH Launch Playbook research](https://dev.to/iris1031/product-hunt-launch-playbook-the-definitive-guide-30x-1-winner-1pbh):

- [ ] Create PH "Coming soon" page (4 weeks ahead — builds upvote-day reminders)
- [ ] DM 3-5 Hunters with high follower counts asking if they'd hunt clank
  - Best Hunters for dev tools: search PH for "developer tools" hunters with 3k+ followers
  - Ask 1 week ahead, not day-of
  - Backup: hunt yourself if you have a strong PH account (50+ upvotes given to others, profile filled out)
- [ ] Pre-write the maker comment (48h advance, per playbook research)
- [ ] Prepare assets (see below)
- [ ] Add the PH page to your email signature, LinkedIn bio for "follow this launch"

## Pre-launch (T-7 days)

- [ ] Confirm Hunter is still on board
- [ ] Verify all assets are uploaded
- [ ] Write maker comment final draft
- [ ] Prepare 30-50 personal DMs (LinkedIn/Twitter/email) ready to send launch morning
- [ ] Block calendar 00:00-23:59 PT on launch day (this is a 24h sprint)

## Pre-launch (T-1 day)

- [ ] Final check that github.com/AnshumanAtrey/clank loads, has stars, has a recent commit
- [ ] Confirm clank.atrey.dev loads, OG image renders
- [ ] Get sleep — launch starts midnight Pacific (= 12:30 PM IST next day, manageable for India-based maker)

---

## Asset list

### Logo / icon (240x240 PNG)
Terminal cursor `▌` icon on dark background, accent color from website palette (`--accent: #95e6cb` or similar). Or just lowercase `clank` wordmark.

### Gallery images (4-6 screenshots, 1270x760 PNG/JPG)

1. **Hero**: terminal screenshot of `clank deep +14155552671` — full output visible
2. **Subcommands grid**: visual showing all 10 subcommands as a grid
3. **Honest section**: "what's broken" callout (Pratfall as differentiator)
4. **Install one-liner**: terminal showing `go install github.com/AnshumanAtrey/clank@latest` + version output
5. **Asciinema embed teaser**: still frame from the asciinema demo with play button overlay
6. **The README screenshot**: top of README showing badges (stars, license, downloads)

### Animated GIF (optional but high-impact)
30-second GIF of `clank scan` running. Convert from asciinema cast via `vhs` (charm.sh tool).

### Tagline (60 chars max)
```
Phone-number OSINT toolkit. One binary. Ten subcommands. MIT.
```

### Description (260 chars max)
```
Open-source phone OSINT in your terminal. 10 subcommands: bulk pattern
scans, deep enrichment, IMEI decode, SEC EDGAR search, Telegram/WhatsApp
lookup, Google dorks. Single 48 MB binary. No signup, no API gate.
Honest about what's broken.
```

---

## The maker comment (post 30 min after launch)

This is the most important text on the entire PH page. Per the research:
> "Your maker comment should be written 48 hours in advance, tell your story not your feature list, explain what problem you personally faced, why you built this, and what feedback you want."

```
Hey Product Hunt 👋

3 months ago I had a phone number and 4 hours to figure out who it
belonged to. Suspected scam targeting a friend. I tried every
open-source phone-OSINT tool I could find — most either had broken
modules (Snapchat lookups dead since 2024), or hadn't shipped a
release since 2023, or required a Python dep tree that took longer
to install than the investigation should have.

I gave up at hour 3.

That weekend I started rewriting it. 3 months later it's clank —
a phone-number OSINT toolkit, single Go binary, 10 subcommands,
open-source MIT.

What's in v0.1.0:
→ libphonenumber + carrier + spam check (instant, no API key)
→ Free-tier API providers (NumVerify, Veriphone, IPQS)
→ IMEI decoder + 254,993-device TAC database (March 2026 snapshot)
→ Telegram phone-to-user (MTProto via gotd/td)
→ WhatsApp phone-to-user (whatsmeow, QR-pair)
→ Instagram + Snapchat + Amazon presence (port of megadose/ignorant)
→ SEC EDGAR full-text filings search
→ Google-dork URL generator (5 buckets, ~160 URLs)
→ Bulk pattern scanning with rate-limit-aware sleeps + JSONL checkpoint

Honest about what's broken: Snapchat method (their CSRF moved JS-side
in 2024 — documented), messenger ban risk at scale (~50/min), psbdmp
pastebin search (service went dark). All called out in the README.

I'd love feedback on:
1. What OSINT source should I add for v0.2.0? Truecaller is at the top
   of the list but I'm open to suggestions.
2. The honest-about-brokenness section is unusual — is it useful or off-putting?
3. Anyone here actively use phone-OSINT in their workflow? Would love
   to hear what you'd want.

Free, MIT, no signup, no API gate ever.

→ github.com/AnshumanAtrey/clank
→ clank.atrey.dev

Thanks for taking a look 🙏
— Anshuman (atrey.dev · build@atrey.dev)
```

**Why this comment works**:
- Story-first (per playbook): the "I needed this" hook
- Honest about brokenness (Pratfall — differentiator)
- Three specific feedback questions (asks for engagement, signals you'll listen)
- No marketing-speak
- Personal sign-off

---

## Topic tags (PH categories)

Pick 3 max:
- **Developer Tools** (primary)
- **Open Source**
- **Security**

Avoid: "AI" (false signal), "Productivity" (off-thesis), "SaaS" (literally not a SaaS)

---

## Launch day (Tuesday)

### 00:01 PT — Launch goes live

- Hunter (or you) hits Submit
- Page is live within 1 min
- Verify all assets render correctly

### 00:30 PT — Post the maker comment

Wait 30 min so it's not the very first comment (looks staged). Then post the prepared comment.

### 01:00 PT — Send the personal DMs

Use the prepared list of 30-50 contacts. Personalize each first line:
```
Hey [name] —
clank just launched on PH (the phone-OSINT tool I've been building).
[1 line referencing your last conversation with them].
If you want to upvote: [PH link]
No pressure if not your scene.
```

**DON'T** send identical copy to everyone. PH staff and savvy users notice. Personalize the first line minimum.

**DON'T** send to people you haven't talked to in 6+ months without context. Reads as desperate.

### 02:00-08:00 PT — Sleep (it's midnight in India)

The first 8 hours of PH are about volume. You can't be active here. Trust the prep.

### 08:00 PT (20:30 IST) — Check in

- Where are you ranked? (Goal: top 5)
- How many comments? (Goal: 10+ by hour 8)
- Reply to every comment that's not yours within 15 min

### 09:00-18:00 PT — The grind

- Reply to every new comment within 15 min
- Reshare PH page on Twitter once at hour 9 ("clank is on PH today, here's what people are saying")
- DON'T spam upvote requests on Twitter — looks desperate, PH staff notices
- DM follow-ups to people who said they'd upvote but you don't see their vote yet (gentle: "how's your day going?")

### 23:59 PT — Day over

- Final rank locked in
- Post a thank-you comment regardless of rank
- Take screenshots of the badge for portfolio + future LinkedIn posts

---

## Post-launch (next day)

- [ ] Reply to every comment from launch day you missed
- [ ] DM thank-you notes to people who upvoted publicly
- [ ] Post a LinkedIn followup: "Yesterday's PH launch — here's what I learned"
- [ ] Add the "Featured on Product Hunt" badge to README (if you got top 5)

---

## Realistic outcomes

- **Top 1 — unlikely** (would need 1000+ upvotes, hunter with massive following)
- **Top 5 — achievable** (~300-500 upvotes, requires 50+ active DM list)
- **Top 10 — likely** (~150-300 upvotes, with prep)
- **Top 50 — very likely** even with no prep
- **Permanent PH page** — guaranteed, this is the biggest long-term win

The PH page lives forever. Becomes a permanent backlink + Stripe-Atlas-style social proof. Don't sweat the rank — sweat the quality of the comment thread.

## What to do if numbers are catastrophic (rank 100+)

- Don't delete the launch
- Don't apologize publicly
- Engage substantively with the few who DID comment
- The PH page still exists as a permanent backlink
- Move on. Try again in 6+ months with a v0.5.0 + a major new feature.

## A note on PH Hunters

Finding a high-profile Hunter is hard. Most won't reply to cold DMs.

**Workarounds**:
1. Self-hunt — build PH karma (upvote 30+ products in the prior month, comment substantively, complete profile)
2. Pay for a "Hunter service" — DON'T (PH staff disqualifies launches found to be hunter-bought)
3. Ask in indie-hacker communities (Indie Hackers forum, Twitter) for someone willing to hunt
4. Attend a PH-organized event (online meetups happen monthly) and meet hunters organically

**For v0.1.0**: self-hunt is fine. Save the high-profile-hunter ask for a v0.5.0+ launch where the polish justifies it.

## Sources

- [PH Launch Playbook 30x #1 Winner — DEV](https://dev.to/iris1031/product-hunt-launch-playbook-the-definitive-guide-30x-1-winner-1pbh)
- [Successfully Launch Dev Tools on PH — Ronak Ganatra](https://ronakganatra.com/posts/successfully-launch-dev-tools-on-producthunt)
- [Smol Launch Guide](https://smollaunch.com/guides/launching-on-product-hunt)
- [How to launch a developer tool on PH — Hackmamba](https://hackmamba.io/developer-marketing/how-to-launch-on-product-hunt/)
- [Product Hunt official launch guide](https://www.producthunt.com/launch)
