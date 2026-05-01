# Reddit r/golang — submission

> r/golang (~200k subscribers). Lead with the Go technical angle, NOT the OSINT angle. r/golang readers care about Go patterns, not phone enrichment. Submit Thursday morning (~10am ET = different audience from OSINT post).

## Title

```
Shipped a phone-OSINT CLI in Go — single static binary, 10 subcommands, embedded 254k-record SQLite. Lessons learned about gotd/td, whatsmeow, and goreleaser
```

**Why this title**:
- Names specific Go libraries (`gotd/td`, `whatsmeow`, `goreleaser`) — Go folks search for these
- "Lessons learned" frames it as content, not promotion
- Specific scope ("10 subcommands", "254k-record") = real
- No "introducing", no "announcing"

## Body

```
Just shipped clank v0.1.0 — a phone-OSINT CLI tool, single Go binary,
10 subcommands. Sharing the technical lessons in case useful.

Repo: https://github.com/AnshumanAtrey/clank

The interesting Go decisions:

**Why Go for this in particular**: 
The previous-generation tools in this space are Python (PhoneInfoga,
megadose/ignorant). Their dep trees are painful — pip install across
3.10/3.11/3.12 + half-deprecated cryptography libs. Go's single static
binary distribution + cross-compile to 5 platforms via goreleaser was
worth the trade of fewer rich-ecosystem libs.

**Library wins**:
- `nyaruka/phonenumbers` — Go port of Google's libphonenumber. Active,
  matches upstream within a release. ~3 MB embedded data.
- `gotd/td` — Telegram MTProto client. Same lib Bellingcat's
  telegram-phone-number-checker uses. Excellent maintenance.
- `go.mau.fi/whatsmeow` — WhatsApp Web multi-device protocol. Powers
  Beeper's mautrix-whatsapp bridge. Production-grade.
- `modernc.org/sqlite` — pure-Go SQLite driver, no CGo. Required for
  whatsmeow's session DB. Saved my cross-compile setup.
- `mdp/qrterminal/v3` — terminal QR rendering for WhatsApp pair flow.
- `PuerkitoBio/goquery` — for the Amazon HTML scraping in ignorant
  module.

**The annoying parts**:
1. **whatsmeow #1086**: WhatsApp silently drops invalid numbers from
   IsOnWhatsApp responses. The result map has fewer keys than the
   request. I had to reconcile by Query field rather than response
   index, and surface "no response — likely invalid" instead of
   false-negative. Worth a writeup on its own.

2. **gotd/td session persistence**: gotd's `bin.Buffer`-based session
   file format isn't the most ergonomic — went with a simple custom
   wrapper at ~/.clank/telegram.session. Their `auth.Flow` is great
   though, handles 2FA cleanly.

3. **goreleaser brew tap setup**: brew formula publish requires a
   token + an existing tap repo. Without both, the entire release
   workflow fails — even if binaries built fine. Solved with a
   `disable:` template in `.goreleaser.yaml` that skips the brew
   pipe when token is empty:
   ```yaml
   disable: '{{ if eq .Env.HOMEBREW_TAP_GITHUB_TOKEN "" }}true{{ end }}'
   ```
   This was the single biggest gotcha during the v0.1.0 release.

4. **go:embed for the IMEI database**: 254,993 device records as a
   3.5 MB CSV embedded into the binary. Total binary: 48 MB.
   Considered SQLite-with-go:embed but plain CSV + a sorted-slice
   binary search was faster and simpler.

**Concurrency patterns**:
The `deep` subcommand fans out to 6 sources concurrently using
sync.WaitGroup + per-source timeouts via context.WithTimeout.
Each source returns a Block that's exactly one of {Result, Skipped,
Error}, so the aggregator never has to deal with nil-vs-zero
ambiguity. Worked well — easy to add a 7th source.

**Cross-platform binary distribution**:
Goreleaser config produces 5 archives (linux/darwin/windows × amd64/arm64)
on every `git tag v*` push. CGo disabled (modernc.org/sqlite makes this
work). Total CI time: 4-5 min per release.

Anything you'd do differently? Particularly interested in:
- Better patterns for "exactly-one-of-{result, skipped, error}"
  block types — I went with three pointer fields, considered tagged
  union via interface but it was awkward
- Whether anyone's done streaming-MTProto for Telegram bulk lookups
  (the `gotd/td` API seems batch-oriented)

Repo with all the code: https://github.com/AnshumanAtrey/clank
```

## Why this body works for r/golang

1. **Opens with a single sentence** scope statement — Go folks like brevity
2. **Lists specific libraries with attribution** — Go community over-indexes on this
3. **Honest about pain points** with code snippets — proves it's real engineering
4. **Names a specific bug (#1086)** with the workaround — bookmark-worthy
5. **Asks for technical feedback** at the end — opens substantive comment thread
6. **Doesn't oversell the OSINT angle** — r/golang doesn't care about phone OSINT, they care about Go patterns

## What r/golang readers WILL ask about

### "Why CGo disabled?"
```
modernc.org/sqlite is pure-Go (transpiled from C, hosted in Go). Means
the entire binary cross-compiles without a C toolchain. Worth the slight
performance hit (sqlite-CGo is ~30% faster on heavy reads) for a tool
that only uses SQLite for whatsmeow's session DB.
```

### "Why 3 pointer fields instead of an interface?"
```
Considered:
type Block interface { isBlock() }
type Result struct{...}
type Skipped struct{...}
type Error struct{...}

Trade-off was JSON marshaling. With pointer fields you get a clean
{result: {...}} or {error: "..."} top-level field in the JSON output.
Interface-based needed custom MarshalJSON. Pointer fields won for
readability. Open to alternatives.
```

### "Why not gRPC for the deep aggregator?"
```
Six in-process source calls, no network boundary inside clank. gRPC would
add complexity for zero benefit. Each source is a Go package directly
invoked — no IPC.
```

### "Have you tried [newer phone-parse lib]?"
```
nyaruka/phonenumbers is a Go port of Google's official libphonenumber.
For OSINT work, you want exactly Google's regional rules — diverging
would cause subtle valid-vs-invalid disagreements with other tools.
That's why I went with it.
```

### "Why not use a config file instead of env vars for API keys?"
```
Two reasons: (1) twelve-factor — env vars are ergonomic for CI/CD and
container deployment, (2) most users only need 1-2 providers, so the
config file would have a lot of empty fields. Env vars + --key override
covers all cases.
```

### "Did you consider building a TUI with bubbletea?"
```
For v1, no — wanted to focus on subcommand correctness first. v0.5.0+
might have a `clank tui` mode using bubbletea, especially for the scan
subcommand which currently just streams progress lines. Would also
make the WhatsApp QR pair flow nicer.
```

### "How did goreleaser work for you?"
```
Mostly great. Trips:
- v2 brews → homebrew_casks deprecation, will need to migrate
- The `disable:` template above for skipping brew when token unset
- Build time scales with number of platforms (4-5 min for 5 archives)

Overall I'd recommend it. The .goreleaser.yaml in the repo is a working
example.
```

## Posting strategy

- Post Thursday morning ET (~9am-11am)
- Reply to all comments within 30 min for first hour
- Engage technical questions with specifics + code snippets
- DON'T cross-link to the OSINT post
- DON'T cross-link to clank.atrey.dev (r/golang prefers GitHub)

## Subreddit rules to verify

r/golang has fairly relaxed rules but:
- Self-promo limited to "show your project" posts (this qualifies)
- Pure marketing posts get downvoted hard
- Long-form technical writeups always do better than "look at my project"

## What success looks like

- 50-200 upvotes
- 10-30 substantive technical comments
- 5-15 GitHub stars from this single post
- Some comments suggesting library improvements (gold)

## Followup post (week +2 if this performs)

```
Title: Lessons from shipping a Go CLI: goreleaser tap conditional, whatsmeow #1086, exactly-one-of block pattern

Body: Long-form follow-up to my r/golang post 2 weeks ago. Here's what
I learned shipping clank v0.1.0 in detail.
[full writeup, 1500 words]
```

This second post (after engagement on the first) builds your r/golang reputation as a technical contributor, not just a self-promoter.
