import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
export default {
	preprocess: vitePreprocess(),
	kit: {
		// A fallback page, because the dashboard has a route per asset and the id is
		// not known at build time. Static hosting serves this for anything it has no
		// file for, and the router takes over in the browser.
		adapter: adapter({ fallback: '200.html' }),
		// Absolute asset URLs, not relative. 404.html is served at whatever path the
		// visitor asked for, and relative hrefs would resolve against that and 404 too.
		paths: { relative: false },
		prerender: { entries: ['*'] }
	}
};
