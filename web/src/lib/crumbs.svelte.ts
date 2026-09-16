import { page } from '$app/state';

// The trail a page sits in. The layout draws it, pages set it — a page knows the
// video it belongs to and the layout does not, so deriving it from the URL alone
// would mean a crumb that says "21110c8f" instead of the name of the thing.
export type Crumb = { label: string; href?: string };

// Stamped with the path it was set for. The layout used to clear the trail on every
// navigation, which raced the page that was setting one: whichever effect ran last
// won, and a detail page could render with only its section showing.
const state = $state<{ trail: Crumb[]; path: string }>({ trail: [], path: '' });

export function setCrumbs(trail: Crumb[]) {
	state.trail = trail;
	state.path = page.url.pathname;
}

export function crumbs(): Crumb[] {
	return state.path === page.url.pathname ? state.trail : [];
}
