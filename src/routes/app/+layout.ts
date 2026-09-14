// The dashboard is a shell the browser fills from the API. Prerendering the shell
// is fine; prerendering its data is not, and there is none to prerender.
export const prerender = true;
export const ssr = false;
