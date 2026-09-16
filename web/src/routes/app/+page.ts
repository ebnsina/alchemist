import { redirect } from '@sveltejs/kit';

// The list lives at /app/videos/ now. Keeping this redirect means old links, the
// logo in the header and anyone's bookmark still land somewhere real.
export const load = () => redirect(308, '/app/videos/');
