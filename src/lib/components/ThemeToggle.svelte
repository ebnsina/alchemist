<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Sun03Icon, Moon02Icon } from '@hugeicons/core-free-icons';

	// The system preference decides until somebody says otherwise; the choice is
	// then remembered per browser. Nothing here reaches the server: a theme is a
	// property of the screen you are looking at, not of the account.
	let theme = $state<'light' | 'dark' | null>(null);

	$effect(() => {
		try {
			const saved = localStorage.getItem('theme');
			theme = saved === 'light' || saved === 'dark' ? saved : null;
		} catch {
			theme = null;
		}
	});

	const systemDark = () =>
		typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches;

	const isDark = $derived(theme ? theme === 'dark' : systemDark());

	function toggle() {
		theme = isDark ? 'light' : 'dark';
		document.documentElement.setAttribute('data-theme', theme);
		try {
			localStorage.setItem('theme', theme);
		} catch {
			// A blocked storage is not worth failing over; the choice lasts the visit.
		}
	}
</script>

<svelte:head>
	<!-- Applied before first paint, or the page flashes the wrong theme on every
	     load for anyone who has chosen one. -->
	{@html `<script>try{var t=localStorage.getItem('theme');if(t==='dark'||t==='light')document.documentElement.setAttribute('data-theme',t)}catch(e){}</script>`}
</svelte:head>

<button
	type="button"
	class="icon-btn"
	onclick={toggle}
	aria-label={isDark ? 'Switch to the light theme' : 'Switch to the dark theme'}
	title={isDark ? 'Light' : 'Dark'}
>
	<HugeiconsIcon icon={isDark ? Sun03Icon : Moon02Icon} size={17} strokeWidth={1.7} />
</button>
