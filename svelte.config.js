import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	kit: {
		adapter: adapter(),
		// Absolute asset URLs, not relative. 404.html is served at whatever path the
		// visitor asked for, and relative hrefs would resolve against that and 404 too.
		paths: { relative: false },
		prerender: { entries: ['*'] }
	}
};
