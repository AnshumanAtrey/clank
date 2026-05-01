# Reddit r/OSINT — submission

> r/OSINT (~45k subscribers). Highest-density audience for clank. Lead with technique, not product. Submit Tue 14:00 IST (~04:30 ET = US morning).

## Title (paste exactly)

```
Built a phone-OSINT toolkit in Go after every existing one I tried had broken modules — sharing in case it's useful
```

**Why this title works for r/OSINT**:
- Doesn't start with "I built" or "Show off" (which mods sometimes flag)
- Explains the WHY (every existing tool had broken modules) — practitioners relate
- "in case it's useful" reads humble, not sales-y
- No emoji, no exclamation marks (r/OSINT is technical, not r/sideproject)

**Alternative titles (test if first doesn't perform)**:
- `Phone-OSINT toolkit with Telegram + WhatsApp + EDGAR + IMEI in one binary (Go)`
- `clank — open-source phone-OSINT CLI, single binary, honest about what's broken`

## Body

```
Hey r/OSINT,

Long-time lurker, occasional commenter. Built something I think this
sub will appreciate, sharing in case it's useful for your work.

**Background**: 3 months ago I tried investigating a number for a friend
who was being scammed. Pulled down the top open-source phone-OSINT tools
on GitHub — most had broken modules:
- PhoneInfoga: Snapchat method dead, last release 2023
- ignorant: Python 3.11+ broke half the imports
- phone-checker: only 1 source

I gave up and started rewriting that weekend. 3 months later it's a
single Go binary with 10 subcommands.

**What it does** (clank v0.1.0):
- `clank scan <pattern>` — bulk pattern (e.g. `+918115605xxx`) → libphonenumber
  filter → enrich each candidate with rate-limit-aware sleeps + JSONL checkpoint
- `clank deep <number>` — fans out to 6 sources concurrently, unified report
- `clank dorks <number>` — generates ~160 Google search URLs across 5 buckets
  (PhoneInfoga's killer feature, ported)
- `clank imei <15-digit>` — Luhn check + 254,993-device TAC lookup (Mar 2026
  snapshot from MoazEb/tac-database)
- `clank edgar <query>` — SEC EDGAR full-text search (no API key)
- `clank telegram lookup` — MTProto via gotd/td (the same lib Bellingcat uses)
- `clank whatsapp lookup` — whatsmeow (powers the mautrix-whatsapp bridge)
- `clank ignorant <number>` — Instagram/Snapchat/Amazon presence (port of
  megadose/ignorant + native Go HTTP)

**What's broken** (transparent in README):
- Snapchat: their CSRF moved JS-side in 2024, method broken — documented
- Messenger ban risk at scale: WhatsApp/Telegram >50/min triggers bans
- psbdmp pastebin search: service went dark 2024
- Carrier portability: libphonenumber returns originally-allocated, not
  current — same limitation as all phone-OSINT tools

**Why Go**: single static binary on every OS, no Python dep hell, the
messenger libraries (gotd/td, whatsmeow) are actually maintained.

**Install**:
```
go install github.com/AnshumanAtrey/clank@latest
# or
gh release download v0.1.0 -R AnshumanAtrey/clank
```

**Repo**: https://github.com/AnshumanAtrey/clank

**Genuine asks**:
1. If you do phone-OSINT in your workflow: what source is clank missing?
   Truecaller is at the top of v0.2.0. Other top contenders?
2. Anyone here have a working Snapchat phone-validate technique post-2024?
   Their JS flow has a bearer token I haven't reverse-engineered yet.
3. Pastebin/leak search — is there a working free-tier replacement for
   psbdmp.cc? Considering IntelligenceX free API.

Happy to nerd out on the implementation in comments. The research/ folder
in the repo has 11 markdown files mapping the OSINT-phone-tool ecosystem
in case you want the deeper context.

[Not a paid product. MIT. Side project. No tracking, no signup.]
```

## Why this body works

1. **Establishes credibility as a community member** ("long-time lurker, occasional commenter")
2. **Concrete origin story** with specific tools you tried — practitioners pattern-match
3. **Bullet list of subcommands** = scannable, proves it's not vaporware
4. **"What's broken" section** = unusual for promo posts, signals honesty
5. **Why-Go technical justification** — r/OSINT respects implementation choices
6. **Genuine asks** at the end — invites comment thread engagement
7. **Last line caveat** — disclaims commercial intent (mods see this and let it stand)

## After posting

### First 30 min — critical window

- Refresh Reddit once at +5 min — confirm post is live and visible (not shadowbanned)
- Reply to every comment within 10 min for the first hour
- If a mod removes the post, DM them politely:
  ```
  Hi mods — my post was removed. Could you let me know which rule
  it violated so I can fix and resubmit? I tried to lead with
  technique and avoid sales-y language. Happy to adjust.
  ```

### First 4 hours

- Respond to every question substantively (1-3 sentences per reply)
- If asked "how does this compare to PhoneInfoga?" — be specific and fair
- Share the post link on Twitter ONCE, no more
- Don't share it on LinkedIn the same day (cross-channel spam detection)

## Comment-reply templates for r/OSINT specifically

### "How does it differ from PhoneInfoga?"
```
PhoneInfoga has the web UI and more provider integrations — strengths.
clank's faster (Go binary, no Python), has Telegram + WhatsApp + EDGAR
+ IMEI that PhoneInfoga doesn't, and is honest about what's broken
(Snapchat). They're complementary tools, not replacements.

The dorks subcommand is directly inspired by PhoneInfoga's killer
feature — credit where due.
```

### "Is the messenger lookup safe? Will it ban my account?"
```
Honest answer: at low volume (<50 lookups/min), no real ban risk in
my testing. At higher volume, yes — WhatsApp's `events.TemporaryBan`
fires hard, Telegram's `FLOOD_WAIT_X` slows you down. The README has
the full risk note. For bulk scanning, I recommend a burner account
just for clank — that's what I do.
```

### "How recent is the IMEI database?"
```
March 2026 snapshot from MoazEb/tac-database (MIT). 254,993 entries.
Covers everything from old Alcatels through iPhone 15 Pro Max and
Galaxy S24+. New device launches won't be in there for 1-2 months
typically — `clank imei --update` is on the v0.2.0 roadmap to refresh
without binary update.
```

### "What about Truecaller?"
```
Top of v0.2.0. The path I'm planning: paired-Android-device import
flow, since Truecaller's API isn't open. Documented in research file
07. ETA: ~3 weeks.

If anyone has tested approaches that work (paid API, intermediate
proxy, etc.), I'd love to hear.
```

### "Looks great, will star/contribute"
```
Thanks! If you do try it, I'd love to hear which subcommand you used
first — always curious what people pivot to. Issues open at
github.com/AnshumanAtrey/clank/issues — top wishlist is in the path-1/
folder of the repo.
```

### "Why not just use [paid OSINT vendor]?"
```
[Vendor] is great for pros who need SLAs and support contracts.
clank is for the journalists, students, and hobbyists who can't
justify $99-499/mo for occasional investigations. Different market.
```

### Critical comment about a specific technique
```
Fair criticism. The reason I went with [decision] was [specific],
but you're right that [tradeoff]. Opened issue #X to track —
appreciate the eye on it.
```

## What to avoid in r/OSINT

- ❌ Don't link to the website (clank.atrey.dev) — link to GitHub
- ❌ Don't use emoji in the body
- ❌ Don't bold every other word (looks like LinkedIn-speak)
- ❌ Don't end with "Let me know what you think!" (lazy CTA)
- ❌ Don't post then disappear — r/OSINT mods notice
- ❌ Don't post on weekends (low traffic, posts die)

## If the post does NOT get traction (under 5 upvotes in 2 hours)

1. Don't delete it
2. Comment on a few popular r/OSINT posts substantively (build sub-karma)
3. Wait 4-6 weeks before posting clank again here
4. The next clank post should be a writeup ("How clank handles WhatsApp's #1086 silently-drops bug") — long-form value, not promotional
