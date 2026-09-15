import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	kit: {
		// A Node server, because the AI routes hold provider keys and a key cannot
		// live in page script. Everything that was static still is: the marketing
		// pages and the dashboard shell prerender at build time and are served as
		// files, so this only adds a server for the handful of routes that need one.
		adapter: adapter(),
		// Absolute asset URLs, not relative. An error page is served at whatever path
		// the visitor asked for, and relative hrefs would resolve against that.
		paths: { relative: false },
		prerender: { entries: ['*'] }
	}
};
