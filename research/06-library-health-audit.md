# Library Health Audit — clank

Audit date: 2026-04-29. All stats fetched live from GitHub API and pkg.go.dev.

---

## 1. `github.com/nyaruka/phonenumbers` v1.7.2

- **Repo:** https://github.com/nyaruka/phonenumbers
- **Stars:** 1,549 - **Forks:** 173 - **Open issues:** 2
- **Last push:** 2026-04-27 - **Last release:** v1.7.2 on 2026-04-27 (95 total releases)
- **License:** MIT - **Archived:** No
- **pkg.go.dev imported by:** 377 known importers

**Maintainer:** Nyaruka (org with 86 public repos), creators of RapidPro and TextIt (textit.com). Multi-contributor: 103 closed PRs, with PRs merged as recently as April 1, 2026 (#225 and #224). Releases cluster around metadata syncs (e.g., 4 releases in late March/early April 2026). README explicitly states "We use this library daily in production."

**Production users:** Nyaruka uses it internally to power RapidPro/TextIt (the messaging platform behind UNICEF, IFRC/Red Cross, and other NGO deployments). 377 known Go importers per pkg.go.dev — wide adoption across the Go ecosystem. It is the de facto Go libphonenumber port (more active than ttacon's original).

**Risk: GREEN** — company-backed, used in production by the maintainer, regular releases tracking upstream Google libphonenumber metadata, very strong dependent count.

**Alternatives:**
- `github.com/ttacon/libphonenumber` — 633 stars, last push 2024-04 (stale, not archived but mostly dormant). The original; nyaruka is a fork of this.
- `github.com/printesoi/libphonenumber` — minor fork of ttacon with go.mod cleanup; low activity, not a serious replacement.

---

## 2. `github.com/fatih/color` v1.18.0

- **Repo:** https://github.com/fatih/color
- **Stars:** 7,948 - **Forks:** 637 - **Open issues:** 25
- **Last push:** 2026-04-28 - **Last release:** v1.19.0 on 2026-03-20
- **License:** MIT - **Archived:** No
- **pkg.go.dev imported by:** 27,995 packages

**Maintainer:** Fatih Arslan (long-time Go community contributor, ex-DigitalOcean, ex-HashiCorp). Windows colorable layer maintained by mattn. Single-maintainer at the top, but with substantial community contribution and fresh activity (push within last 24 hours of audit).

**Production users:** Used by tens of thousands of Go projects including HashiCorp Terraform, Vault, Consul; GitHub CLI (`gh`); kubectl plugins; nearly every notable Go CLI. 28k importers makes it one of the most-imported terminal libraries in Go.

**Risk: GREEN** — single-maintainer-at-the-top but the dependency graph is so huge that an abandonment would trigger immediate community fork; active commits as of audit date.

**Alternatives (if ever needed):**
- `github.com/charmbracelet/lipgloss` — richer styling, used by Charm tooling. Heavier API.
- `github.com/gookit/color` — feature-equivalent fallback.
- `github.com/muesli/termenv` — lower-level.

---

## 3. `github.com/pbakondy/mcc-mnc-list` (JSON data)

- **Repo:** https://github.com/pbakondy/mcc-mnc-list
- **Stars:** 128 - **Forks:** 60 - **Open issues:** 14
- **Last push:** 2023-04-04 (3 years stale at audit time)
- **License:** MIT - **Archived:** No (but effectively dormant)
- **No tagged releases.**

**Maintainer:** pbakondy (single author, individual GitHub user). Commit cadence over the last 5 years averaged once every ~6 months and stopped April 2023. 14 open issues, no recent triage.

**Production users:** Several npm wrappers (`ktalebian/mcc-mnc`, others) ship this JSON. It's the most popular Wikipedia-derived MCC/MNC dataset on GitHub, but its Wikipedia-source-of-truth nature means staleness is tolerable for legacy carriers; new MVNOs since 2023 are missing.

**Risk: YELLOW** — single-author, no commits in 3 years, but data half-life for MCC/MNC is long (carriers rarely change codes). Not catastrophic, but we should not depend on it for fresh MVNO mappings.

**Alternatives (better options exist):**
- `github.com/musalbas/mcc-mnc-table` — **archived November 2024** despite "Updated daily" tagline; 363 stars. Not viable.
- `github.com/P1sec/MCC_MNC` — actively maintained, combines Wikipedia + ITU-T bulletins + txtNation. **Best technical replacement.**
- `github.com/CursedHardware/mno-list` — weekly updates from mcc-mnc.com/.net + Google carrier data. **Best freshness.**
- ITU-T Operational Bulletin (authoritative but PDF-only, hard to parse).

---

## 4. `github.com/Oros42/phone-blacklist` (CSV data)

- **Repo:** https://github.com/Oros42/phone-blacklist
- **Stars:** 3 - **Forks:** 3 - **Open issues:** 0
- **Last push:** 2015-04-07 (11 years stale)
- **License:** Unlicense (public domain) - **Archived:** No (but effectively abandoned)
- **Total commits:** 4, all on 2015-04-07.

**Maintainer:** Oros42 (single author hobby project). README itself frames it as "phone numbers I blacklisted" — never claimed to be a community-curated list.

**Production users:** None visible. No mirrors, no downstream packages.

**Risk: RED** — abandoned for over a decade, hobbyist scope, almost zero stars/forks. The numbers are 11 years old; many are reassigned by carriers. Shipping this as "blacklist data" inside clank borders on misleading.

**Alternatives:**
- **FCC Robocall data** — `https://reportrobo.fcc.gov` complaint feed (USA, public).
- **FTC Do Not Call complaints** — public download.
- **community-spam-list** style repos: `chucknorris-io/swd` (varied), `phlite/spam-numbers`, plus telecomvulnerabilities feeds.
- **NumVerify / IPQS / Twilio Lookup** if API access is OK.
- **Hiya / Truecaller / YouMail** datasets — commercial, better signal.

---

## 5. `github.com/jwoertink/blocked-numbers` (CSV data)

- **Repo:** https://github.com/jwoertink/blocked-numbers
- **Stars:** 2 - **Forks:** 1 - **Open issues:** 0
- **Last push:** 2017-10-16 (8.5 years stale)
- **License:** None specified (legally ambiguous to redistribute) - **Archived:** No
- **Total commits:** 3, all on 2017-10-16.

**Maintainer:** jwoertink (Crystal lang community member, individual). One-day project, never returned to.

**Production users:** None visible.

**Risk: RED** — abandoned + no license = we have no clear right to redistribute the CSV inside our binary. Data is also 8.5 years old. Worse posture than #4.

**Alternatives:** Same as #4. Plus:
- **YouMail Robocall Index** (commercial API, monthly free tier).
- **Nomorobo** (reseller licensing).
- **OpenCNAM** spam flagging.

---

## Risk summary table

| # | Library / Data | Rating | Why |
|---|---|---|---|
| 1 | nyaruka/phonenumbers v1.7.2 | GREEN | Company-backed, 377 importers, releases this week |
| 2 | fatih/color v1.18.0 | GREEN | 28k importers, push within 24h of audit |
| 3 | pbakondy/mcc-mnc-list | YELLOW | Single author, 3yr stale, but slow-changing data |
| 4 | Oros42/phone-blacklist | RED | 11yr stale hobby project, 3 stars |
| 5 | jwoertink/blocked-numbers | RED | 8.5yr stale, no license, 2 stars |

---

## Recommended actions

**Do now:**

1. **Drop `Oros42/phone-blacklist` and `jwoertink/blocked-numbers`.** Both are RED. The jwoertink CSV has no license — shipping it in our binary is technically copyright-ambiguous. Replace with one of:
   - FTC Do-Not-Call public data (cleaner provenance), or
   - A live API hit to YouMail/Hiya/Truecaller/IPQS at lookup time, or
   - At minimum, mark the spam-list output in clank as "stale community data, last updated 2015/2017" so users don't trust it.

2. **Switch MCC/MNC source to `CursedHardware/mno-list` or `P1sec/MCC_MNC`.** `pbakondy/mcc-mnc-list` is a single-author repo with no commits since April 2023 and `musalbas/mcc-mnc-table` (the obvious next choice) was archived in Nov 2024. CursedHardware/mno-list updates weekly — that's the live one. Add a CI job that re-pulls the JSON quarterly.

3. **Pin nyaruka/phonenumbers to a tagged minor (^1.7).** Already pinned to v1.7.2 in `go.mod`; just keep using `go mod tidy` quarterly to track upstream Google libphonenumber metadata releases (they ship every 1-2 weeks).

4. **Pin fatih/color** loosely (^1.18); no concerns.

**Add to CI:**

- Monthly check that none of our direct deps are GitHub-archived (gh api `repos/{owner}/{name}` → `.archived`).
- Monthly check on push date freshness; alert if a critical dep has no push in 6+ months.

**Single most important action right now:** **Remove or replace the two RED data sources (Oros42, jwoertink).** They are both abandoned, one has no license, and their staleness actively misrepresents clank's signal quality. This is the only finding worth acting on today.
