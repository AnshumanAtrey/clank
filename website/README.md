# clank/website — clank.atrey.dev

Marketing landing page for [clank](../README.md) — the phone-number OSINT toolkit.

Lives in the same repo as the CLI (mono-repo with `website/go.mod` stub that excludes this folder from `go install`). See `../path-1/07-repo-architecture.md` for the architecture decision.

## Stack

- [Astro 5](https://astro.build) — static site generator
- Vanilla TypeScript for interactivity (typewriter, install tabs)
- Plain CSS with terminal palette
- No framework islands, no Tailwind, no asciinema-player (v1 — added later)

## Local development

```bash
cd website
npm install
npm run dev
# → http://localhost:4321
```

## Build

```bash
npm run build
# → dist/
```

## Deploy

Cloudflare Pages, configured via dashboard:

- **Root directory**: `website`
- **Build command**: `npm install && npm run build`
- **Build output directory**: `dist`
- **Custom domain**: `clank.atrey.dev`

See `../path-1/06-website-tech.md` for full setup steps.

## Source of truth for copywriting

Canonical copy lives at [`../path-1/05-website-copy.md`](../path-1/05-website-copy.md). When updating copy on the site, update that file too.

## License

MIT — © Anshuman Atrey · <https://atrey.dev>
