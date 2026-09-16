// trailingSlash 'never' makes the static adapter write build/404.html, which is the
// file a static host serves for an unknown path. The adapter's SPA `fallback` cannot
// be used here: with csr off there is no JavaScript to fill the shell in.
export const trailingSlash = 'never';
