export const prerender = true;
export const trailingSlash = 'always';

// No client runtime. The only interactive things on this site are the language and
// theme toggles, and both are a dozen lines of DOM in app.html. Shipping ~40 KB of
// framework to a page that argues for saving the viewer's data would be self-refuting.
export const csr = false;
