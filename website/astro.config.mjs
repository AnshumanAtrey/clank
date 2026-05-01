import { defineConfig } from 'astro/config';

// Static-only landing page deployed to Cloudflare Pages at clank.atrey.dev.
// No adapter needed — pure static build outputs to dist/.
export default defineConfig({
  site: 'https://clank.atrey.dev',
  output: 'static',
  build: {
    inlineStylesheets: 'auto',
  },
  compressHTML: true,
});
