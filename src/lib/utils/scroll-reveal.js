/**
 * Svelte action: fades an element up the first time it enters the viewport.
 * Does nothing at all when the reader has asked for reduced motion — the element
 * is simply left in its final state, which is how `.scroll-fade` is authored.
 * @param {HTMLElement} node
 * @param {{ delay?: number }} [options]
 */
export function scrollReveal(node, options = {}) {
	const still = window.matchMedia('(prefers-reduced-motion: reduce)');
	if (still.matches || !('IntersectionObserver' in window)) {
		node.classList.add('is-visible');
		return {};
	}

	node.classList.add('scroll-fade');
	if (options.delay) node.style.transitionDelay = `${options.delay}ms`;

	const io = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				if (!entry.isIntersecting) continue;
				entry.target.classList.add('is-visible');
				io.unobserve(entry.target);
			}
		},
		{ rootMargin: '0px 0px -12% 0px', threshold: 0.08 }
	);
	io.observe(node);

	return {
		destroy() {
			io.disconnect();
		}
	};
}
