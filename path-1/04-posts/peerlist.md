# Peerlist Launchpad — clank submission

> Submit Monday (the only launch day on Peerlist Launchpad). Per [Peerlist Help](https://peerlist.neetokb.com/articles/how-to-launch-your-project-on-peerlist-project-spotlight): go to peerlist.io/projects → "Launch Project" from the weekly Spotlight banner.

## Why Peerlist matters for you specifically

Peerlist's audience skews toward Indian + global indie devs. Your profile (ISU'28, India-based, 5x hackathon winner per memory) is dead-center for this audience. A Peerlist launch builds personal brand more than user count, and personal brand compounds.

It's also low-effort relative to PH or HN. Submit Monday, get 1 day of front-page Spotlight, accumulate Peerlist reactions/comments.

## Pre-submission checklist

- [ ] Peerlist profile is filled out (bio, projects, GH connected)
- [ ] clank is added as a project on your Peerlist profile (peerlist.io/projects → Add)
- [ ] GitHub repo connected to the Peerlist project entry
- [ ] At least one screenshot uploaded
- [ ] README has a hero image at the top (Peerlist may pull this for the card)

## The Launchpad submission form

Peerlist asks for:
1. **Project title**
2. **Tagline** (~80 chars)
3. **Description** (longer)
4. **Cover image** (1280x720 recommended)
5. **Gallery images** (3-6)
6. **Demo URL**
7. **GitHub URL**
8. **Tags**
9. **Maker note** (the personal context)

## Field-by-field copy

### Project title
```
clank — phone-number OSINT toolkit
```

### Tagline (80 chars)
```
Phone OSINT in your terminal. One binary. Ten subcommands. MIT. No API gate.
```

### Description

```
clank is an open-source CLI for phone-number OSINT. Single 48 MB Go binary,
10 subcommands, MIT licensed, no signup, no API gate.

It's the rewrite I wished existed when I tried investigating a phone number
3 months ago and found that every open-source tool either had broken modules
or hadn't shipped a release since 2023.

Subcommands:
• scan — bulk pattern → enrich every survivor with rate-limit-aware sleeps
• deep — single number, all 6 sources concurrently
• dorks — ~160 Google search URLs across 5 buckets
• imei — Luhn check + 254,993-device TAC database (March 2026 snapshot)
• edgar — SEC EDGAR full-text filings search
• telegram — MTProto phone-to-user via gotd/td
• whatsapp — WhatsApp via whatsmeow (QR-pair session)
• ignorant — Instagram / Snapchat / Amazon presence checks
• history — local audit log
• --version

What works out of the box: libphonenumber lookup, IMEI, dorks, EDGAR.
What needs setup: Telegram (TG_APP_ID + login), WhatsApp (QR pair).
What's deliberately broken: Snapchat (their CSRF moved JS-side in 2024,
documented inline). What's risky at scale: messenger bans above ~50/min.

Built in Go for single-binary distribution across Linux/macOS/Windows
(amd64 + arm64). No Python deps, no Node deps, no Docker required.
Just download and run.

Open-source forever. PRs welcome.
```

### Cover image
1280x720 — terminal screenshot of `clank deep +14155552671` output, with the tagline overlaid:
```
clank
phone OSINT toolkit · 10 subcommands · MIT
```

### Gallery (3-6)
1. Hero terminal output
2. Subcommands grid (visual)
3. `clank scan` in action (asciinema still or GIF)
4. `clank dorks` output (160 URLs)
5. README "what's broken" section

### Demo URL
```
https://clank.atrey.dev
```

### GitHub URL
```
https://github.com/AnshumanAtrey/clank
```

### Tags (Peerlist categories)

Pick 3-5:
- Open Source
- Developer Tools
- CLI
- Security
- OSINT (if available)
- Go

### Maker note (the personal context)

This is where Peerlist's audience really engages — they want to hear from real makers.

```
Hey Peerlist 👋

I'm Anshuman — software engineer at Unitpay, CSE @ ISU'28, and a builder
who can't stop side-projecting. clank is my latest: a phone-number OSINT
toolkit built because every existing open-source tool in this space was
either broken or abandoned.

The story:
3 months ago I tried investigating a phone number for a friend who was
being scammed. I downloaded the top 5 phone-OSINT tools on GitHub. Half
the modules were broken (Snapchat method dead since 2024, pastebin search
service offline). The ones that worked were Python apps with 30+ deps,
took an hour to set up, and returned partial results.

I gave up that night. Started rewriting on the next weekend.

3 months later: clank v0.1.0 — a single Go binary that ships everywhere,
10 subcommands, embedded 254k-device IMEI database, MIT licensed, zero
gating.

What I'm proudest of:
→ It's honest. The README has a "what's broken" section. Snapchat
  doesn't work — I tell you up-front instead of pretending.
→ It's fast. Single static binary, no dep hell, runs offline for
  most subcommands.
→ It's serious about the OSINT use case. Disclaimer makes legit use
  vs misuse explicit.

What I'm asking for from Peerlist:
1. Try it (`go install github.com/AnshumanAtrey/clank@latest`) and tell
   me what surprised you.
2. If you do OSINT/security work, tell me which source clank is missing
   that I should add for v0.2.0 (Truecaller leads the list).
3. If you've launched on Peerlist before, what worked?

Always happy to chat — atrey.dev · build@atrey.dev · 
linkedin.com/in/anshumanatrey/

Thanks for taking a look 🙏
```

---

## Engagement strategy on Peerlist

- Reply to every comment within 1-4 hours (Peerlist's pace is slower than HN)
- Engage substantively — Peerlist's signal-to-noise is high
- Browse Launchpad and react to others' projects (reciprocity)
- Share the Peerlist link on Twitter once: "if you're on Peerlist, clank is up on the Launchpad this week"

## What NOT to do

- ❌ Don't cross-post identical maker note to PH/HN — Peerlist users notice
- ❌ Don't ask for upvotes via DM — Peerlist's culture is more organic than PH
- ❌ Don't submit if your Peerlist profile is empty (judges look)
- ❌ Don't use AI-generated cover images (Peerlist audience pattern-matches and downvotes)

## After the Spotlight week

The Peerlist project page lives forever. Becomes:
- A permanent project showcase backlink
- Part of your Peerlist profile (visible to recruiters, fellow devs)
- A signal in the Peerlist algorithm for future launches

## Realistic outcomes

- 20-100 reactions on the launch
- 5-15 comments
- 2-5 DMs from interested users
- Gradual increase in Peerlist profile views over the following month

Peerlist is more about the cumulative effect than the single-launch spike. Each launch builds your maker reputation here.

## Sources

- [Peerlist Launchpad](https://peerlist.io/launchpad)
- [How to launch on Peerlist Project Spotlight](https://peerlist.neetokb.com/articles/how-to-launch-your-project-on-peerlist-project-spotlight)
- [Peerlist Free Submission Guide 2026](https://launchdirectories.com/directory/peerlist)
