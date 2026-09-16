import type { Session } from '$lib/api';

// One copy of who is signed in. The layout loads it; pages read it to know whether
// they are looking at their own account or somebody else's, which decides whether a
// destructive control should be on screen at all.
const state = $state<{ me: Session | null }>({ me: null });

export function setMe(s: Session | null) {
	state.me = s;
}

export function me(): Session | null {
	return state.me;
}

// While staff are viewing another account the API refuses every write, so a button
// that writes is a button that can only fail.
export function readOnly(): boolean {
	return !!state.me?.impersonating;
}
