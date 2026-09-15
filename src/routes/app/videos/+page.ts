import { redirect } from '@sveltejs/kit';

// The video list lives at /app/. This path is guessable, it is what the breadcrumb
// calls "Videos", and deleting the id from a video URL lands here -- so send people
// on rather than showing them a 404.
export const load = () => redirect(308, '/app/');
