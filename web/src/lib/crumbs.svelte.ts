// The trail a page sits in. The layout draws it, pages set it — a page knows the
// video it belongs to and the layout does not, so deriving it from the URL alone
// would mean a crumb that says "21110c8f" instead of the name of the thing.
export type Crumb = { label: string; href?: string };

const state = $state<{ trail: Crumb[] }>({ trail: [] });

export function setCrumbs(trail: Crumb[]) {
	state.trail = trail;
}

export function crumbs(): Crumb[] {
	return state.trail;
}
