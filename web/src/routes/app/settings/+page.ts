import { redirect } from '@sveltejs/kit';

// The tabs are the page; /app/settings/ itself opens the first of them.
export const load = () => redirect(307, '/app/settings/profile/');
