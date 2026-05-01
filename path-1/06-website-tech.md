# 06 — Website Tech Spec

> The website is animated, terminal-themed, and serves at clank.atrey.dev via Cloudflare Pages. This document specs the stack, animation strategy, file structure, and deploy pipeline.

## Stack decision — Astro

**Recommended**: [Astro](https://astro.build) static + island components

| Option | Bundle | Cloudflare fit | Animation friendliness | Verdict |
|--------|--------|----------------|------------------------|---------|
| **Astro** | ~5-15 KB JS | Native ([docs](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/)) | Islands hydrate only what needs JS | ✅ Pick this |
| Next.js | 80-200 KB JS | Works but heavier (App Router on Workers) | Full React, more capable | Overkill |
| SvelteKit | 30-80 KB JS | Works | Best animations | Solid alt |
| Plain HTML + Alpine | 5 KB | Native | Manual JS | Underbuild |
| Hugo / Jekyll | 0 KB | Native | Static only, minimal JS | Underbuild for animations |

**Why Astro wins for clank**:
- Static-by-default (zero client JS unless you opt in)
- Cloudflare Pages has first-class Astro adapter
- Per-component hydration ("islands") — perfect for an asciinema embed in an otherwise-static page
- Markdown content support natively (we can render `path-1/05-website-copy.md` directly via Content Collections if we mono-source later)
- Build time: ~3-8s for a single-page site
- Bundle: <20KB total client JS for our use case

## Project structure — `clank/website/` (mono with parent repo)

The website lives inside the `clank` repo at `website/`, NOT in a separate repo. A stub `website/go.mod` excludes it from the parent Go module so `go install`, brew, and any future npm path don't pick it up. See [`07-repo-architecture.md`](./07-repo-architecture.md) for the full reasoning.

```
clank/website/
├── go.mod                               ← STUB → excludes from parent Go module
├── src/
│   ├── pages/
│   │   └── index.astro                  ← single landing page
│   ├── components/
│   │   ├── Hero.astro                   ← static hero + tagline
│   │   ├── HeroTerminal.astro           ← animated typewriter (Astro island)
│   │   ├── AsciinemaCast.astro          ← <script type="text/asciinema-cast"> embed
│   │   ├── SubcommandsGrid.astro        ← static grid
│   │   ├── InstallTabs.astro            ← copy-button tabs (small island)
│   │   ├── HonestyBlock.astro           ← static
│   │   └── Footer.astro                 ← static
│   ├── styles/
│   │   ├── global.css                   ← terminal palette, font setup
│   │   └── terminal.css                 ← scanline effect, cursor blink
│   ├── layouts/
│   │   └── Base.astro                   ← html shell, meta tags
│   └── lib/
│       └── copy.ts                      ← optional read from ../path-1/05-website-copy.md
├── public/
│   ├── casts/
│   │   ├── deep.cast                    ← clank deep lookup recording
│   │   ├── scan.cast                    ← clank scan recording
│   │   └── dorks.cast                   ← clank dorks recording
│   ├── favicon.svg                      ← terminal cursor icon
│   ├── og.png                           ← 1200x630 OG image (terminal screenshot)
│   └── robots.txt
├── astro.config.mjs
├── package.json
├── tsconfig.json
├── README.md                            ← "Marketing site — see ../README.md for the CLI"
└── .gitignore                           ← node_modules/, dist/, .astro/
```

The `website/go.mod` stub (the magic 3 lines):
```
// Stub go.mod — excludes this dir from parent Go module.
module github.com/AnshumanAtrey/clank/website

go 1.25
```

Bonus: copywriting source-of-truth at `../path-1/05-website-copy.md` is now a **relative-path read away** instead of a cross-repo sync. A `lib/copy.ts` at build time can `import` and use the MD directly via Astro's `getCollection`.

## Animation strategy — three layers

### Layer 1: Hero typewriter (custom JS, ~2 KB)

A pseudo-terminal in the hero that types out the tagline character-by-character, blinks the cursor, then "executes" `clank --version`. Custom Vanilla JS, no library. Re-types every 30s.

Pseudocode:
```js
const lines = [
  { prompt: '$ ', text: 'clank --version' },
  { prompt: '', text: 'clank v0.1.0' },
  { prompt: '', text: 'by Anshuman Atrey · https://atrey.dev · build@atrey.dev' },
  { prompt: '$ ', text: '' }, // cursor blinks here
];
typeLines(lines, { charDelay: 50, lineDelay: 800 });
```

**Why custom**: typewriter libraries pull 30+ KB. Hand-rolled is 50 lines.

### Layer 2: Real asciinema cast (mid-page demo)

Use [asciinema-player](https://github.com/asciinema/asciinema-player) (the official one) as a self-hosted JS bundle. ~150 KB but only loads when the section scrolls into view (Astro `client:visible` directive).

**Recording the casts**:
```bash
brew install asciinema
asciinema rec deep.cast
clank deep +14155552671
exit
asciinema upload deep.cast  # optional, or self-host the file
```

For self-hosted (recommended — no external dependency):
```html
<script src="/asciinema-player.min.js"></script>
<link rel="stylesheet" href="/asciinema-player.css" />
<div id="demo"></div>
<script>
  AsciinemaPlayer.create('/casts/deep.cast', document.getElementById('demo'), {
    autoPlay: true,
    loop: true,
    speed: 1.5,
    theme: 'monokai',
    poster: 'npt:0:00',
  });
</script>
```

**Source**: [asciinema-player docs](https://docs.asciinema.org/manual/player/), [asciinema embedding docs](https://docs.asciinema.org/manual/server/embedding/)

### Layer 3: Subtle micro-animations (CSS only)

- Cursor blink on hero terminal (CSS keyframes, 1s cycle)
- Subcommand-grid hover (slight scale + green outline)
- "Copy" button → "Copied ✓" transition (200ms ease-out)
- Scanlines overlay on terminal blocks (optional, retro CRT vibe — CSS gradient)

No Lottie, no GSAP, no heavy animation libs. CSS + tiny JS only.

## Color palette (terminal aesthetic)

```css
:root {
  --bg: #0a0e14;            /* near-black */
  --bg-soft: #11151c;       /* card bg */
  --fg: #b3b1ad;            /* primary text */
  --fg-bright: #e6e1cf;     /* headings */
  --accent: #95e6cb;        /* mint — for "registered/yes" */
  --accent-warn: #ffb454;   /* amber — warnings, "broken" */
  --accent-danger: #f07178; /* coral — errors, "not paired" */
  --accent-link: #59c2ff;   /* blue — links */
  --muted: #4d5566;         /* dim text, separators */
  --selection: #1f2430;
}
```

Inspired by the Ayu Mirage / Ayu Dark editor theme — terminal-dev-friendly, not Apple-glassy.

## Typography

- **Mono everywhere**: `JetBrains Mono` (Google Fonts, weight 400/500/700)
- **Headlines**: same font, larger size, lower line-height
- **Fallback stack**: `JetBrains Mono, "SF Mono", Menlo, Consolas, monospace`

Self-host the font for performance + privacy (no Google Fonts external requests).

## Build + deploy pipeline

### Local dev
```bash
git clone github.com/AnshumanAtrey/clank
cd clank/website
npm install
npm run dev    # → http://localhost:4321
```

(If you're already in the clank repo, just `cd website && npm install && npm run dev`.)

### astro.config.mjs
```js
import { defineConfig } from 'astro/config';

export default defineConfig({
  site: 'https://clank.atrey.dev',
  output: 'static',
  build: {
    inlineStylesheets: 'auto',
  },
  compressHTML: true,
});
```

### Cloudflare Pages setup (one-time, monorepo mode)

1. Cloudflare dashboard → Workers & Pages → Create application → Pages → Connect to Git
2. Select `AnshumanAtrey/clank` (the same repo as the CLI)
3. Build settings:
   - Framework preset: **Astro**
   - Build command: `npm install && npm run build`
   - Build output: `dist`
   - **Root directory: `website`** ← critical — tells Cloudflare to treat `website/` as the project root
   - Node version: 20+ (set via `NODE_VERSION` env var if needed)
4. Save → first deploy runs (~30s)
5. Project URL: `clank.pages.dev` (or whatever Cloudflare assigns)

Cloudflare diffs paths on push — only website changes trigger a rebuild. CLI-only commits don't waste a Pages build.

### Custom domain — clank.atrey.dev

In Cloudflare Pages project → Custom domains → Set up a domain:
- Enter `clank.atrey.dev`
- Cloudflare will detect the parent zone (`atrey.dev`)
  - **If `atrey.dev` is on Cloudflare DNS**: it auto-creates the CNAME for you
  - **If `atrey.dev` is on another DNS**: it shows you the CNAME record to add (`clank` → `<your-project-name>.pages.dev`, e.g. `clank.pages.dev`)
- SSL provisions in 1-15 min
- Done

**Source**: [Cloudflare Pages custom domains](https://developers.cloudflare.com/pages/configuration/custom-domains/), [Astro on Cloudflare Pages docs](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/), [Beginner's guide to Astro + Cloudflare](https://dev.to/warish/astro-cloudflare-pages-a-beginners-guide-to-fast-and-easy-deployment-558e)

### CI/CD via GitHub
Cloudflare Pages auto-builds on every push to `main` that touches `website/**`. Preview deployments for PRs. Zero new workflow file needed.

You should also add path-filtering to the existing Go CI workflow so website-only changes don't trigger Go tests:

```yaml
# .github/workflows/ci.yml
on:
  push:
    branches: [main]
    paths:
      - '**.go'
      - 'go.mod'
      - 'go.sum'
      - '.github/workflows/ci.yml'
```

If you want preview-deploy protection (private previews behind Cloudflare Access), see [Aaron Czichon's guide](https://aaronczichon.de/blog/28-cloudflare-pages-astro-github/).

## OG image (social preview)

Critical for LinkedIn/Twitter/HN previews:
- 1200x630 PNG
- Terminal screenshot of `clank deep +14155552671` output
- Bold tagline overlay: "phone OSINT toolkit · single binary · MIT"
- Author tag: `@anshumanatrey · clank.atrey.dev`

Tools to make it:
- Carbon (carbon.now.sh) for the terminal block
- Then composite via Figma / Photopea
- Or use [og-image-as-a-service](https://github.com/vercel/og-image)

## Performance budget

- **Lighthouse score target**: 100/100/100/100 (Perf/A11y/BP/SEO)
- **First Contentful Paint**: <500ms
- **Total page weight**: <500 KB including casts
- **Total JS**: <30 KB (excluding asciinema-player which loads on scroll)
- **Total CSS**: <10 KB (inlined in `<head>`)

How to hit this:
- No web fonts other than self-hosted JetBrains Mono (subset to needed glyphs)
- No analytics (no Google Analytics, no Plausible needed for v1)
- No images larger than 50 KB (use SVG where possible)
- Asciinema casts are text-based — typically 30-60 KB
- Astro's static output ships near-zero JS by default

## SEO

- `<title>`: `clank — phone-number OSINT toolkit`
- `<meta description>`: `Open-source CLI for phone-number OSINT. Single binary, ten subcommands, no API gate. MIT licensed.`
- `<meta og:image>`: see above
- `<link rel="canonical" href="https://clank.atrey.dev">`
- `robots.txt`: allow all
- `sitemap.xml`: optional for a single-page site
- Schema.org `SoftwareApplication` JSON-LD with `applicationCategory: SecurityApplication`, `operatingSystem`, `downloadUrl`, `softwareVersion`

## Future expansion (don't build for v1)

- `/docs` → MD-based docs, generated from clank's research/ folder
- `/blog` → long-form posts (e.g., "How clank handles WhatsApp's #1086 silently-drops bug")
- `/changelog` → autopulled from GH Releases via build-time API call
- `/install` → expanded install guide for Windows, NixOS, etc.
- WASM playground — actually run clank in browser (clank-wasm build) for "try without installing"

For v1 launch: just the single page. Don't multiply scope.

## Cost

- Cloudflare Pages free tier: unlimited requests, 500 builds/month, 100 custom domains. **$0**.
- Domain `atrey.dev`: already owned.
- Total ongoing cost: $0.

## Alternative if you want to skip Astro

A purely static one-page HTML+CSS+vanilla-JS site is also fine. ~1 day of work. Lighter than Astro, no build step at all. Trade-off: no Markdown content collections, manual file management.

For v1, Astro adds enough convenience (component reuse, copy-button as a component, MD support for FAQ) to justify its 5MB of build dependencies.

## Sources

- [Astro on Cloudflare Pages docs](https://developers.cloudflare.com/pages/framework-guides/deploy-an-astro-site/)
- [Astro deploy guide — Cloudflare](https://docs.astro.build/en/guides/deploy/cloudflare/)
- [Cloudflare Pages custom domains](https://developers.cloudflare.com/pages/configuration/custom-domains/)
- [Astro + Cloudflare beginner's guide — DEV](https://dev.to/warish/astro-cloudflare-pages-a-beginners-guide-to-fast-and-easy-deployment-558e)
- [Asciinema player docs](https://docs.asciinema.org/manual/player/)
- [Asciinema embedding docs](https://docs.asciinema.org/manual/server/embedding/)
- [Enhance README with asciinema — Cesar Soto Valero](https://www.cesarsotovalero.net/blog/enhance-your-readme-with-asciinema.html)
