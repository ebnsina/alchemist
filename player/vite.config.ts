import { defineConfig, type Plugin } from 'vite';

// The embed bundle is pinned by major version: an iframe deployed against /v1/
// keeps getting a /v1/ player forever. Breaking changes ship as /v2/.
const BASE = '/v1/';

/** Dev-only stand-in for the production rewrite of /e/{asset} to the embed page. */
const embedRoute = (): Plugin => ({
  name: 'alchemist-embed-route',
  configureServer(server) {
    server.middlewares.use((req, _res, next) => {
      if (req.url && /^\/e\//.test(req.url)) {
        const q = req.url.indexOf('?');
        req.url = `${BASE}embed.html${q < 0 ? '' : req.url.slice(q)}`;
      }
      next();
    });
  },
});

export default defineConfig({
  base: BASE,
  plugins: [embedRoute()],
  build: {
    outDir: 'dist/v1',
    emptyOutDir: true,
    target: ['es2019', 'chrome70', 'safari12'],
    cssMinify: true,
    rollupOptions: {
      input: { embed: 'embed.html', index: 'index.html' },
    },
  },
  server: { port: 5180 },
  preview: { port: 5180 },
});
