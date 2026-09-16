import { ROUTES, SITE } from '$lib/site';

export const prerender = true;

export function GET() {
	const urls = ROUTES.map((r) => {
		const loc = SITE.origin + (r === '/' ? '/' : r + '/');
		return `\t<url>\n\t\t<loc>${loc}</loc>\n\t\t<xhtml:link rel="alternate" hreflang="en" href="${loc}"/>\n\t\t<xhtml:link rel="alternate" hreflang="bn" href="${loc}"/>\n\t</url>`;
	}).join('\n');

	return new Response(
		`<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">\n${urls}\n</urlset>\n`,
		{ headers: { 'content-type': 'application/xml' } }
	);
}
