import { redirect } from '@sveltejs/kit';

// Both moved under Settings; a bookmark from before still lands somewhere real.
export const load = () => redirect(307, '/app/settings/profile/');
