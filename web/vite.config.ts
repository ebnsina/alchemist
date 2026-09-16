import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	// The table ships an uncompiled .svelte file, which Node cannot resolve on its
	// own: any server-rendered route importing it 500s. Bundling it is the fix.
	ssr: { noExternal: ['@tanstack/svelte-table'] }
});
