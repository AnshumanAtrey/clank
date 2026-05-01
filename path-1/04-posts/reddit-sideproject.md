# Reddit r/sideproject — submission

> r/sideproject (~200k). Supportive audience, soft signal. Best for the personal narrative + journey angle. Submit Tue 14:30 IST (after r/OSINT, different angle, no spam-detect).

## Title

```
Rebuilt an abandoned 2014 phone-OSINT script into a 10-subcommand CLI — 3 months of weekends, finally shipped v0.1.0
```

**Why this title**:
- "3 months of weekends" — relatable to side-project audience
- "Finally shipped" — celebrates the milestone (this sub loves shipping stories)
- Specific scope ("10-subcommand CLI") = real, not vaporware
- "Rebuilt an abandoned 2014 script" = origin story hook

## Body

```
Title pretty much sums it up. Sharing the journey in case it's useful
to anyone else dealing with project paralysis.

**The starting point**:
3 months ago I needed to investigate a phone number for a friend. Tried
every open-source tool — most either had broken modules or hadn't shipped
a release since 2023. The OSINT space is full of abandonware.

I found one specific tool — a 110-line phone-number combinator from 2014
by a guy named jwoertink. It worked, barely. No license file, no commits
in 8+ years. I forked it on a Saturday afternoon thinking I'd add one
feature.

**3 months later**:
- 10 subcommands
- Single 48 MB Go binary (cross-platform, brew-installable)
- Embedded 254k-device IMEI database
- 8 working integrations (libphonenumber, IMEI, EDGAR, Telegram, WhatsApp,
  Instagram, Google dorks, free-tier APIs)
- 11 markdown files of OSINT research
- v0.1.0 GitHub Release with downloadable archives

**What I learned about side projects** (in case it's useful):

1. **Scope creep is fine if you check in with yourself weekly.**
   I wrote a "what's done / what's next" markdown every Sunday. Made
   sure I was still excited. Twice I almost gave up — but the markdown
   showed real progress, which made it easier to keep going.

2. **"Rewrite from scratch" works for forgotten projects.**
   The original 2014 script had no test coverage, no CI, dead deps.
   Rewriting from scratch was faster than refactoring. Counterintuitive.

3. **Ship a release before adding more features.**
   I spent week 11 just on the v0.1.0 release pipeline (goreleaser, GH
   Actions, brew tap setup attempts). Felt like wasted week. But now
   every future feature has a "ship it" path. v0.2.0 release will take
   1 hour instead of 11 days.

4. **Honest documentation beats polished documentation.**
   The README has a "What's broken" section listing things that don't
   work (Snapchat lookup since 2024). Counterintuitively, this seems
   to build trust faster than pretending everything works.

5. **The launch is week +1, not week +0.**
   I shipped the binary today, but I'm not "launching" until next
   Tuesday after I have a website + post drafts ready. Forcing myself
   to wait — it's better than scattered launching across 4 hours and
   getting nothing done.

**Repo**: https://github.com/AnshumanAtrey/clank
**Site**: clank.atrey.dev (in progress, going live next week)

**For others working on side projects**: what's keeping you stuck?
What's the next 1-week milestone for your project? Share — sometimes
saying it out loud helps.

(Not selling anything. Free, MIT, open-source. This is just a celebration
post + lessons learned. Hopefully useful.)
```

## Why this body works for r/sideproject

1. **Personal narrative leads** — sideproject sub loves journey posts
2. **Concrete numbers** ("110 lines → 10 subcommands", "3 months", "48 MB")
3. **Lessons learned section** — generalizable advice for other side-projecters
4. **Asks a question at end** — invites others to share their projects (reciprocity)
5. **Disclaimer at end** — heads off "is this an ad?" replies

## After posting

- Reply to people who share THEIR side projects in comments (this is the unwritten rule of r/sideproject)
- Upvote good replies generously
- Don't push the GitHub link in every reply — let curious folks find it
- Engagement here builds long-tail reputation, not viral spike

## Reply templates for r/sideproject

### "How did you stay motivated for 3 months?"
```
Honestly: the weekly markdown checkin. Every Sunday I wrote down
what was done that week + what was next. Made progress visible.
Twice I considered quitting — but the doc showed too much momentum
to drop. Recommend the practice for any 8+ week project.
```

### "How did you find the time?"
```
Two evenings a week + Saturday afternoons. ~10 hours/week, average.
3 months × 10h = 120 hours total. Probably 30 of those were
unproductive (yak-shaving, false starts). 90 productive hours = a
shippable v0.1.0.

Time-budgeting helped: I told myself "1 subcommand per weekend max."
Forced focus.
```

### "Did you write tests?"
```
Sparse but yes. Each subcommand has 5-15 unit tests for the trickiest
parts. Total coverage is probably ~30%, which is shameful for production
software but fine for a side project that's still figuring out its
shape. I'll push coverage higher in v0.2.0 once features stabilize.
```

### Someone shares their own project
```
[Read their post first, ALWAYS]
Cool — [specific technical observation about their project]. The
[specific feature] reminds me of [thing], have you considered
[adjacent feature]?
```

(Substantive engagement with others' posts builds your sub-credibility, which feeds back to your own post's reach.)

### "Would you do it again?"
```
Yes, but I'd ship sooner. I held back the v0.1.0 release for 3 weeks
trying to add "one more thing." Should have shipped at week 8 with 7
subcommands and added 3 more in v0.2.0. The shipping pipeline IS the
project — building it once teaches you what you don't know.
```

## What NOT to do in r/sideproject

- ❌ Don't post a wall-of-text without paragraph breaks
- ❌ Don't make it sound like a press release (this isn't TechCrunch)
- ❌ Don't use marketing words ("revolutionize", "game-changer")
- ❌ Don't drop the link 3 times in different comments
- ❌ Don't ignore commenters who share their own projects
- ❌ Don't crosspost the same copy from r/OSINT — Reddit detects this

## Post timing

- Tuesday-Wednesday morning EST best
- Avoid Sunday (sub is dead Sundays)
- Avoid Friday after 12 (everyone's checking out for the weekend)

## Realistic outcomes

- 50-200 upvotes
- 20-50 comments
- 3-10 GitHub stars from this single post
- A few DMs from other side-project makers asking about your process

r/sideproject is more about reciprocity and community than spike conversion. Plant the seed, water it via engagement, the value comes back over months.
