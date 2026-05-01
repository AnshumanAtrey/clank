# Reddit r/cybersecurity — submission

> r/cybersecurity (~800k). Bigger but less aligned than r/OSINT. Frame as a tool useful for defensive use cases (SOC analysts investigating suspicious numbers, security researchers). Submit Thursday afternoon (different day from r/OSINT to avoid spam detect).

## Title

```
Free open-source phone-OSINT CLI for SOC/IR investigators — single binary, 10 subcommands, no signup
```

**Why this title**:
- "SOC/IR investigators" — frames the audience for r/cybersecurity (broader than OSINT, more enterprise-y)
- "Free open-source" up front — heads off "is this a paid tool?" objections
- "No signup" — paywall-resistant audience appreciates this
- Specific scope ("10 subcommands") proves real

## Body

```
Hi r/cybersecurity,

Releasing clank — a phone-number OSINT CLI built specifically because
the existing open-source options weren't reliable for incident response
work. Sharing in case useful for SOC analysts, fraud investigators, or
security researchers dealing with phone-number pivots.

**The use case I built it for**:
You get an alert. Number tied to a phishing campaign. You need to know:
- Is the number valid? Mobile or VOIP?
- Carrier (and is the carrier known for spoof-friendliness)?
- Is it associated with any registered messenger account (Telegram,
  WhatsApp)?
- Does it appear in any SEC filings (insider/social-engineering pivot)?
- Is it in any spam blacklist?

The existing tools either required paid API subscriptions, had broken
modules, or didn't ship a release since 2023.

**clank v0.1.0**:
- `clank deep <phone>` — fans out to 6 sources concurrently in ~5-10s
  - libphonenumber + carrier + spam blacklist (offline)
  - 3 free-tier API providers (NumVerify, Veriphone, IPQS — bring your
    own keys, all have generous free tiers)
  - Telegram + WhatsApp presence (requires session pair)
  - Instagram + Snapchat + Amazon presence
  - SEC EDGAR full-text search
- `clank scan <pattern>` — bulk lookup over a number pattern with
  rate-limit-aware sleeps (so messenger lookups don't trigger bans)
- `clank imei <15-digit>` — Luhn check + 254k-device TAC database
  for handset identification (March 2026 snapshot)
- `clank dorks <phone>` — generates ~160 Google search URLs across
  social/disposable/reputation/individuals/general categories
- `clank edgar <query>` — SEC full-text filings search
- All output is JSON (--json flag) for SIEM/SOAR integration
- Local audit log at ~/.clank/history.jsonl (or disable via env var)

**Privacy**: zero telemetry. Zero outbound calls except to the OSINT
sources you explicitly invoke. Audit log is local. Set `CLANK_NO_AUDIT=1`
to disable even that.

**Honest limitations** (in README):
- Snapchat lookup broken since their CSRF moved JS-side in 2024
- Carrier portability returns originally-allocated, not current carrier
- Messenger ban risk at scale (~50 lookups/min) — use a burner account
- No HLR/SS7 lookup (paid services required for that)

**Single static Go binary** (48 MB, no Python/Node deps):
```
go install github.com/AnshumanAtrey/clank@latest
# or
gh release download v0.1.0 -R AnshumanAtrey/clank
```

**Repo**: https://github.com/AnshumanAtrey/clank

**Use cases I built it for**:
1. SOC investigating suspicious caller IDs from a phishing alert
2. Fraud team verifying that a customer's claimed phone is real
3. Security researchers mapping social-engineering campaigns
4. Incident response pivoting on numbers extracted from breach data

Happy to answer technical questions in comments.

[MIT license, no telemetry, no signup. Disclaimer in README addresses
the "this could be misused" concern explicitly.]
```

## Why this body works for r/cybersecurity

1. **Frames the use case as defensive** (SOC, IR, fraud) — this sub is enterprise-leaning
2. **"Bring your own keys"** language signals respect for users' existing tooling stack
3. **JSON output for SIEM/SOAR** — buzzwords this audience cares about
4. **Privacy section** addresses the "is this tool snooping?" objection
5. **Honest limitations** — transparency works in security audiences
6. **Specific use cases at end** — helps undecided readers map clank to their job

## After posting

- Reply to comments quickly (first 30-60 min)
- Be ready for "is this legal" / "ethics" debates — they're more common in r/cybersecurity
- Don't engage with bad-faith "OSINT = stalking" comments — link to README disclaimer once, move on

## Comment-reply templates

### "How is this different from MaltGOAT / PhoneInfoga / X paid tool?"
```
Different scope: clank is one CLI binary you `go install`, no Python deps,
no signup. PhoneInfoga has the web UI + more provider integrations.
Maltego is enterprise-tier link analysis on top of multiple sources.
Each tool has its niche.

For SOC use specifically, clank's --json output is designed to feed
into your existing SIEM/SOAR pipeline. Same for the audit log.
```

### "Is this OSINT or stalker-ware?"
```
README has an explicit Disclaimer section addressing this:
- Built for legitimate use: SOC, fraud investigation, journalism,
  security research, knowing your own footprint
- Misuse (stalking, harassment) is illegal regardless of tool
- All data sources queried are public — same as a Google search

If your concern is "this lowers the bar for misuse," I take that
seriously. Counter-argument: the data is already accessible to bad
actors. clank levels the playing field for defenders.
```

### "What about HLR / SS7 lookup?"
```
Out of scope for v0.1.0 — those require paid services (HLR-Lookups,
Twilio Lookup v2 paid tier, Telnyx). On the v0.2.0 roadmap as
optional integrations. For free-tier OSINT, libphonenumber + IPQS
gets you 80% of what HLR provides.
```

### "Audit log — what does it contain?"
```
Just the timestamp + subcommand name + first phone-shaped argument.
NEVER the full args (so API keys aren't logged), NEVER the lookup
results. It's purely a "what did I run when" tracker for personal
investigation reproducibility. Disable entirely with
CLANK_NO_AUDIT=1 in your env.
```

### "Will it work in an enterprise environment behind a proxy?"
```
Currently no — clank uses Go's default HTTP client. Adding HTTPS_PROXY
support is on the v0.2.0 roadmap (and trivial — Go's net/http respects
the env var natively, just need to verify all sub-clients honor it).
File an issue if you need this prioritized.
```

### "Open-source license? Can my employer use it?"
```
MIT. Use commercially, modify, redistribute, no attribution required
(though appreciated). Same license as Express, jQuery, Vim, etc. —
unrestricted for enterprise use.
```

## What NOT to do in r/cybersecurity

- ❌ Don't oversell ("this revolutionizes incident response")
- ❌ Don't make it sound like a sales pitch
- ❌ Don't ignore the inevitable "this enables stalkers" comment
- ❌ Don't argue with bad-faith ethics critics — answer once, move on
- ❌ Don't crosspost identical copy from r/OSINT (sub-mod will spam-flag)

## Subreddit rules to check

r/cybersecurity has stricter rules than r/sideproject:
- Self-promo allowed but with limits (1:5 ratio of own content : others')
- Tool releases generally OK if framed as "useful for the community"
- AMA-style posts encouraged

If your account doesn't have prior r/cybersecurity engagement, comment substantively on 3-5 popular posts in the week before launching.

## Realistic outcomes

- 50-200 upvotes
- 15-40 comments (more critical than r/OSINT, more "this could be misused" debates)
- 5-20 GitHub stars from this single post
- 1-3 DMs from enterprise users asking about specific integrations

This is a high-volume sub but lower conversion. The volume gives you long-tail reach (post stays in search results); conversion is lower because most r/cybersecurity readers are SOC analysts who want polished commercial tools, not CLI side projects.

Worth the post — just calibrate expectations.
