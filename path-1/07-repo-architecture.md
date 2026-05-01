# 07 — Repo Architecture: Where Does the Website Live?

> **DECISION REVISED.** Original recommendation was "separate repo `clank-web`." After deeper analysis with the user, the cleaner answer is: **same repo, with a one-line trick that excludes the website from every package manager's install path.**

## The user's question

> "What if we keep the website in the same repo? And while installing it using Go, npm, or brew — the package we serve, we do not include the website in the hosted package?"

**Yes. This works cleanly. Here's how.**

---

## TL;DR — Recommendation

**Keep the website in `clank/website/` inside the same repo.** Add a stub `website/go.mod` file (3 lines) and the website is automatically excluded from every Go install path. Brew and any future npm distribution are unaffected because they only archive what `.goreleaser.yaml` lists. Cloudflare Pages natively supports building from a subdirectory.

The original "separate repo" recommendation was over-defensive. The user's intuition was right.

---

## How each install path works (and why mono is safe)

### `go install github.com/AnshumanAtrey/clank@latest`

**What happens**:
1. Go resolves `github.com/AnshumanAtrey/clank` to a module
2. Go module proxy (proxy.golang.org) fetches the module ZIP for the version
3. Go builds `main.go`, installs the binary
4. Source files go to `~/go/pkg/mod/...` temporarily

**Module ZIP contents** are governed by [Go's module reference](https://go.dev/ref/mod#zip-files):

> A module zip must contain only files in the module... certain files and directories are excluded:
> - vcs files like .git, .hg
> - **Directories with their own go.mod file (nested modules)**

**The trick**: putting a `go.mod` in `website/` makes Go treat that directory as a separate module. The Go module proxy excludes it from the parent module's ZIP. **`go install` will not download website files.**

### `brew install AnshumanAtrey/tap/clank`

**What happens**:
1. Brew downloads the pre-built binary archive from the GitHub Release
2. Archive contents are determined by `.goreleaser.yaml`:
   ```yaml
   archives:
     - files:
         - README.md
         - LICENSE*
   ```
3. The website folder is NOT in this list, so it's never in the archive

**Conclusion**: brew is unaffected by mono-vs-split. The website folder simply isn't archived.

### `gh release download` (raw binary)

Same as brew. Goreleaser's archive contents are defined explicitly. Website not included.

### Future `npm install -g clank` (if you add it later)

Some Go CLIs (esbuild, swc, biome) ship via npm using a tiny wrapper package that downloads the platform binary at install time. The `package.json` for such a wrapper would live separately (e.g. `clank/npm-wrapper/`) with its own `package.json` controlling exactly what npm publishes. **The website doesn't ship via npm any more than the CLI does.**

If you add this later, structure:
```
clank/
├── npm-wrapper/
│   ├── package.json    ← defines `files:` field, lists ONLY install.js + README
│   ├── install.js      ← downloads platform binary from GH Release
│   └── README.md
├── website/
└── ...
```

Inside `npm-wrapper/package.json`:
```json
{
  "name": "clank",
  "files": ["install.js", "README.md"],
  ...
}
```

The `files` field in package.json is a whitelist — npm publishes only what's listed. Website is not included.

### `git clone github.com/AnshumanAtrey/clank.git`

This DOES download everything (cloning is not installing). Contributors who want to fix a CLI typo will pull website source too. Trade-off:
- **Cost**: ~1-5 MB of additional source files for cloners
- **Benefit**: contributors can update website + CLI in a single PR
- Acceptable for a project this size

---

## The full structure

```
clank/                                  ← github.com/AnshumanAtrey/clank
├── go.mod                              ← MAIN Go module
├── go.sum
├── main.go                             ← CLI entry
├── internal/                           ← CLI internals (Go)
│   ├── api/, scan/, deep/, etc.
├── website/                            ← Astro site
│   ├── go.mod                          ← STUB → excludes from parent module
│   ├── package.json                    ← real Astro setup
│   ├── astro.config.mjs
│   ├── tsconfig.json
│   ├── src/
│   │   ├── pages/index.astro
│   │   ├── components/
│   │   └── styles/
│   ├── public/
│   │   └── casts/                       ← asciinema recordings
│   └── README.md                        ← "Marketing site for clank — see ../README.md"
├── path-1/                              ← launch playbook (this folder)
├── research/                            ← OSINT research notes
├── .github/workflows/
│   ├── ci.yml                           ← Go CI (path-filtered to .go files)
│   └── release.yml                      ← goreleaser on tag push
├── .goreleaser.yaml
├── README.md                            ← root README (Go CLI focus)
├── LICENSE                              ← MIT — covers everything
└── .gitignore                           ← website/dist, website/node_modules, etc.
```

### The `website/go.mod` stub (the magic 3-line file)

```
// Stub go.mod — exists solely to exclude this directory from the parent
// Go module (github.com/AnshumanAtrey/clank). Without this, `go install`
// would download website source files unnecessarily. The website is an
// Astro project, not Go code. See path-1/07-repo-architecture.md.
module github.com/AnshumanAtrey/clank/website

go 1.25
```

That's it. No Go code in the website folder. The file just declares: "this directory is its own module, exclude from parent."

---

## Verification — let's prove this works

After creating `website/go.mod`, you can verify the exclusion locally:

```bash
# Show what files would be included in the module ZIP
cd /Users/atrey/Desktop/code/side-claude-income/clank
go mod download -x github.com/AnshumanAtrey/clank@latest

# OR more directly, use 'go list' to see module boundaries:
go list -m github.com/AnshumanAtrey/clank
go list -m github.com/AnshumanAtrey/clank/website
# Both will resolve, confirming they're separate modules

# Most direct: package the module and inspect
go mod download -x . 2>&1 | grep -i website
# Should return nothing — website not in download
```

Or: build a tagged release with goreleaser (or just `go install` from your local clone) and inspect the cached module:

```bash
ls ~/go/pkg/mod/github.com/anshumanatrey/clank@<version>/
# Should NOT show a website/ directory
```

---

## Cloudflare Pages — monorepo build setup

Cloudflare Pages supports specifying a subdirectory as the build root. Setup:

1. **Create Cloudflare Pages project**:
   - Workers & Pages → Create application → Pages → Connect to Git
   - Select `AnshumanAtrey/clank` (the same repo)

2. **Build settings**:
   ```
   Framework preset: Astro
   Build command: npm install && npm run build
   Build output directory: website/dist
   Root directory: website
   ```
   The `Root directory: website` setting is the key — Cloudflare runs the build inside `website/`, treating it as the project root.

3. **Custom domain**: `clank.atrey.dev` — same as before. Pages → Custom domains → Set up a domain → enter `clank.atrey.dev`.

4. **Auto-deploy**: Cloudflare Pages auto-builds on every push to `main` that touches `website/`. A push that only touches `internal/` Go files won't trigger a website rebuild (Cloudflare diffs paths).

**Source**: [Cloudflare Pages Astro docs](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/), [Cloudflare Pages monorepo support](https://developers.cloudflare.com/pages/configuration/monorepos/)

---

## CI workflow path-filtering

To avoid running the Go test suite on website-only changes, update `.github/workflows/ci.yml`:

```yaml
on:
  push:
    branches: [main]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/ci.yml'
  pull_request:
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
```

Now Go CI only runs on Go-relevant changes. Website-only commits skip the Go workflow entirely.

Cloudflare Pages does the inverse — only runs on `website/**` changes.

---

## Updated decision matrix (mono vs split)

| Concern | Mono (with go.mod stub) | Split-repo | Winner |
|---------|--------------------------|------------|--------|
| **`go install` payload** | Clean (excluded via stub) | Clean | **Tie** |
| **`brew install` payload** | Clean (goreleaser archive) | Clean | **Tie** |
| **`npm install` payload** (if added) | Clean (`files:` whitelist) | Clean | **Tie** |
| **`git clone` payload** | Includes website (~5 MB extra) | Clean | Split |
| **Repo browse-ability** | One repo, clear subfolders | Concern split | **Tie** |
| **CI complexity** | Path filters needed | Two simpler workflows | **Tie** |
| **Cloudflare Pages binding** | Native monorepo support | Native | **Tie** |
| **Single source of truth** | Same git history | Two repos to sync | **Mono** |
| **Single PR for cross-cutting** | One PR | Two PRs | **Mono** |
| **Issue tracking** | One issue tracker | Two trackers | **Mono** |
| **Star consolidation** | One repo, one star count | Stars split | **Mono** |
| **Language detection** | Go + TS/Astro mixed | Each repo single-lang | Split |
| **License clarity** | Single LICENSE | Need duplicate | **Mono** |

**New tally**: Mono wins 5-1-7-tied. The user's instinct was correct — once you know the `go.mod` stub trick, mono is strictly better for this size of project.

---

## Why the original recommendation was wrong

The first version of this doc recommended split-repo. Reasons cited:
1. "`go install` would download website" — **WRONG**, the go.mod stub fixes this
2. "Repo browse-ability" — actually fine when website lives in a clearly-named subfolder
3. "CI complexity" — path filters solve this in 5 lines
4. "Cloudflare Pages preference" — Cloudflare supports monorepo natively, my prior assumption was wrong

The original analysis missed the Go module's nested-submodule behavior. The fix is genuinely 3 lines (the stub go.mod). Once you know it exists, the trade-off flips entirely.

This is a useful debugging note for future architecture decisions: when something seems to require splitting, check first whether the language/tool has a built-in scoping mechanism (Go submodules, Python `__init__.py` controls, npm `files:` whitelists, etc.). Almost always there's a way to keep things together.

---

## What this means for the rest of path-1

References to "separate repo `clank-web`" in other path-1 docs are now stale. Affected:
- `path-1/README.md` — the "Decisions already made" section (item #1)
- `path-1/06-website-tech.md` — the project structure + repo references
- `path-1/05-website-copy.md` — minor reference updates

These are being updated in the same commit that lands this revised doc.

---

## Migration plan if you ever WANT split repos later

If down the road clank's website grows to be its own thing (docs site + blog + changelog + tutorials = 100+ files), the migration is:

1. Use `git subtree split` to extract `website/` history into a new repo
2. Push to `AnshumanAtrey/clank-web`
3. Delete `website/` from clank repo
4. Update Cloudflare Pages project to point at the new repo
5. Add a redirect file in clank/README.md pointing to the new repo

Reversible decision. Going mono first → split later is easy.

---

## Final action items

1. **In this commit**: stub `website/go.mod` is NOT yet created (no `website/` folder exists). Create it when you scaffold the Astro site. Until then, the only thing in `website/` should be the stub go.mod (so even if someone runs `go install` between now and the website launch, nothing weird happens).

2. **When you scaffold**: 
   ```bash
   cd /Users/atrey/Desktop/code/side-claude-income/clank
   mkdir website
   cat > website/go.mod <<'EOF'
   // Stub go.mod — excludes this dir from parent Go module.
   module github.com/AnshumanAtrey/clank/website
   
   go 1.25
   EOF
   cd website
   npm create astro@latest .  # initialize Astro
   ```

3. **Before next clank Go release** (v0.1.1+): verify the stub works by running:
   ```bash
   go mod download -x . 2>&1 | grep -i website  # should be empty
   ```

4. **Cloudflare Pages**: connect to the same `AnshumanAtrey/clank` repo, set Root directory to `website`.

---

## Sources

- [Go modules reference — module zip files](https://go.dev/ref/mod#zip-files) — official spec on what's included/excluded in module ZIPs
- [Cloudflare Pages Astro framework guide](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/)
- [Cloudflare Pages monorepo support](https://developers.cloudflare.com/pages/configuration/monorepos/)
- [Astro deployment to Cloudflare Pages](https://docs.astro.build/en/guides/deploy/cloudflare/)
- [GitHub Actions path filters](https://docs.github.com/en/actions/using-workflows/triggering-a-workflow#using-filters)
- [npm package.json `files` field](https://docs.npmjs.com/cli/v10/configuring-npm/package-json#files)
