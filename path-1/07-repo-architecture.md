# 07 — Repo Architecture: Where Does the Website Live?

> User question: "If we are making a website, where should we keep it? Same repo or different? I'd host it on Cloudflare connected to clank.atrey.dev. If users install via npm/brew/go install, we don't want to ship our website right?"

## TL;DR — Recommendation

**Separate repo: `AnshumanAtrey/clank-web`** (or `clank-site`).

The clank CLI repo stays binary-focused. The website lives separately, deploys to Cloudflare Pages, and serves at `clank.atrey.dev`.

Copywriting authoritative source remains in `clank/path-1/05-website-copy.md` — the website repo pulls from it (manual copy or scripted sync).

## The user's intuition was right

> "If user is using npm to install our tool or some other package manager, we do not want to ship our website right?"

Yes. Let's verify exactly what each install path actually downloads:

### What `go install github.com/AnshumanAtrey/clank@latest` downloads

`go install` clones the **entire module** (everything in the same git repo, defined by `go.mod`) into the user's module cache temporarily, builds the binary, and discards the source. So:

- Website files would be downloaded to `~/go/pkg/mod/github.com/anshumanatrey/clank@v0.1.0/`
- They wouldn't end up in the compiled binary (Go's compiler only includes what's imported)
- BUT the user pays the bandwidth + disk for the source download
- For a typical Astro site (pre-build): ~100-500 files, ~5-15 MB
- For a Next.js site with `node_modules` git-ignored: still ~50-200 source files, ~1-5 MB

**Impact**: small but non-zero. More importantly, *aesthetic*. The user's `go install` log shows "downloading 423 files" instead of "downloading 87 files." Smells weird for a CLI tool.

### What `brew install AnshumanAtrey/tap/clank` downloads

Goreleaser archives only what `.goreleaser.yaml` says. Currently:
```yaml
files:
  - README.md
  - LICENSE*
```

Website files would NOT be in the brew archive (they're not in the `files:` list). So brew is unaffected by mono-vs-split repo.

### What `git clone` downloads

Everything. Mono-repo means contributors clone both CLI source AND website source. For someone who wants to fix a CLI typo, that's wasted bandwidth and a confusing "what is all this Astro stuff?" moment.

### What pre-built GitHub Releases ship

Same as brew — only what's in goreleaser's `files:` list. Unaffected.

---

## Decision matrix

| Concern | Mono-repo | Split-repo | Winner |
|---------|-----------|------------|--------|
| **`go install` payload** | Bigger (5-15 MB extra) | Clean | Split |
| **Repo browse-ability** | "What is this clank?" confusion | Clear concern separation | Split |
| **CI complexity** | Single workflow but mixed (Go + Node) | Two simpler workflows | Split |
| **Cloudflare Pages binding** | Subfolder support exists but adds friction | Native (one repo → one project) | Split |
| **Single source of truth for docs** | One PR can update both | Two PRs needed | **Mono** |
| **Versioning alignment** | CLI v0.1.0 ↔ website v0.1.0 mechanical | Manual to keep in sync | **Mono** |
| **Star/discoverability** | One repo accumulates all stars | Stars split between repos | **Mono** |
| **Contribution scope clarity** | "Where does my PR go?" confusion | Obvious | Split |
| **Deploy-on-push simplicity** | Need path filters in workflows | Trivial | Split |
| **Issue tracking** | Mixed CLI bugs + website tweaks | Naturally separated | Split |
| **License consistency** | Same LICENSE applies everywhere | Need duplicate LICENSE files | Mono |
| **README signal** | Buried under website files | Front and center | Split |

**Score**: Split wins 8-3-1 (1 = "License consistency" is trivially solved by copying the file).

## The hybrid that solves the only real mono advantage

The strongest pro-mono argument is **single source of truth for copywriting**. Solve this without going mono:

1. **Canonical copy lives in `clank/path-1/05-website-copy.md`** (this folder, what we're writing now).
2. **Website repo (`clank-web`) pulls copy via either**:
   - Manual copy on website updates (low frequency, fine for v1)
   - A submodule pointing at clank repo (overkill but works)
   - A build-time script that fetches the raw MD via GitHub API (clean, decouples deploy)

For v0.1.0 launch: **manual copy is fine**. Update website ~weekly, keep MD in path-1 as the canonical source for posts AND landing page.

---

## Recommended repo structure

```
github.com/AnshumanAtrey/clank/         ← existing, CLI only
├── main.go
├── internal/
├── path-1/
│   ├── 05-website-copy.md              ← CANONICAL copy source
│   ├── 06-website-tech.md              ← spec the website builds against
│   └── ...
├── README.md
├── LICENSE
└── ...

github.com/AnshumanAtrey/clank-web/     ← NEW, website only
├── src/
│   ├── pages/
│   │   ├── index.astro                 ← reads copy from path-1/05-website-copy.md
│   │   └── docs/                        ← optional later
│   ├── components/
│   │   ├── Terminal.astro
│   │   ├── AsciinemaCast.astro
│   │   └── Hero.astro
│   ├── styles/
│   └── assets/casts/                   ← .cast files for asciinema
├── public/
├── astro.config.mjs
├── package.json
├── wrangler.toml                        ← Cloudflare Pages config
├── README.md                            ← "Marketing site for github.com/AnshumanAtrey/clank"
└── LICENSE                              ← MIT, points back to clank repo
```

---

## Cloudflare Pages binding

**Domain plan**:
- Apex `atrey.dev` → existing personal site (already hosted somewhere)
- Subdomain `clank.atrey.dev` → new Cloudflare Pages project for `clank-web`

**Setup steps** (one-time):

1. Create new Pages project in Cloudflare dashboard:
   - Workers & Pages → Create application → Pages → Connect to Git
   - Select `AnshumanAtrey/clank-web` repo
   - Build command: `npm run build`
   - Build output: `dist`
   - Framework preset: Astro

2. Project gets a `*.pages.dev` URL automatically (e.g. `clank-web.pages.dev`).

3. Add custom domain (subdomain):
   - Pages project → Custom domains → Set up a domain → `clank.atrey.dev`
   - Cloudflare detects the parent zone (atrey.dev) and either:
     - Auto-creates the CNAME if atrey.dev is on Cloudflare DNS
     - OR shows you the CNAME to add at your DNS provider (`clank` → `clank-web.pages.dev`)
   - SSL cert provisions in 1-15 minutes

**Source**: [Cloudflare Pages custom domains docs](https://developers.cloudflare.com/pages/configuration/custom-domains/), [Astro on Cloudflare Pages](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/)

---

## Migration plan if you ever WANT a mono-repo later

If down the road clank grows into something that benefits from mono (e.g., shared TS types between CLI and website), here's the painless migration:

1. Move `clank-web/*` into `clank/website/` via `git subtree`. Preserves history.
2. Update Cloudflare Pages project to point at the new path (build root `website/`, output `website/dist`).
3. Add `path:` filters to GH Actions workflows so website pushes don't trigger Go CI and vice versa.
4. Archive `clank-web` repo with a note pointing at the new monorepo location.

**Reversible decision.** Going split first → mono later is easy. Going mono first → split later requires git surgery.

---

## What about `pkg.go.dev` and module download size?

`go install` actually only downloads the **module zip**, not the full git history. Go's module proxy strips `.git/` and only ships listed files. But it doesn't strip arbitrary subfolders — so `website/` still ends up in the module zip.

You CAN tell go to ignore directories via `.gitignore` patterns... no wait, that's not how it works. The Go module zip respects `go:embed` and module paths, not `.gitignore`. You'd need to use **goreleaser-style "tags" or sub-modules** to formally exclude the website. Doable but adds complexity.

The cleanest exclusion is **don't put it in the same repo in the first place**.

---

## Naming the new repo

**Recommended**: `clank-web` (matches conventions like `bun-sh/bun.sh`, `docusaurus/website`)

**Alternatives considered**:
- `clank-site` — also fine
- `clank.atrey.dev` — too cute, breaks for SSH cloning
- `clank-marketing` — too corporate
- `clank-landing` — narrower than what it'll grow into (landing → docs → blog)

---

## Final recommendation

**Do this:**
1. Create `clank-web` repo on GitHub (private at first, public when website is live)
2. Build with Astro (see `06-website-tech.md`)
3. Cloudflare Pages → connect `clank-web` repo → custom domain `clank.atrey.dev`
4. Keep canonical copy in `clank/path-1/05-website-copy.md`
5. Cross-link: clank README links to `clank.atrey.dev`; website footer links to GH repo

**Don't do this:**
- ❌ Build the website inside `clank/website/` — it bloats `go install`
- ❌ Use Astro server-side rendering — Cloudflare Pages prefers static (zero cold-start)
- ❌ Use Next.js — heavier than needed for a marketing site
- ❌ Skip the website and just rely on README — landing pages convert 3-5x better than READMEs for non-technical visitors
