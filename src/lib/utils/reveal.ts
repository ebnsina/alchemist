import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

gsap.registerPlugin(ScrollTrigger);

/**
 * Svelte action: the elements inside a section arrive as that section does.
 *
 * One ScrollTrigger per section rather than one per element — a landing page has
 * dozens of children, and a trigger each is dozens of scroll listeners doing the
 * same arithmetic.
 *
 * Reduced motion is not a lighter animation, it is none: the resting state is the
 * final state, so doing nothing leaves the section correct.
 */
export function reveal(
	node: HTMLElement,
	options: { selector?: string; stagger?: number; y?: number } = {}
) {
	const still = window.matchMedia('(prefers-reduced-motion: reduce)');
	if (still.matches) return {};

	const { selector = ':scope > *', stagger = 0.06, y = 14 } = options;
	const targets = Array.from(node.querySelectorAll<HTMLElement>(selector));
	if (targets.length === 0) return {};

	const ctx = gsap.context(() => {
		gsap.set(targets, { opacity: 0, y });
		gsap.to(targets, {
			opacity: 1,
			y: 0,
			duration: 0.5,
			ease: 'power2.out',
			stagger,
			scrollTrigger: {
				trigger: node,
				// The section is snapped to the top of the viewport, so waiting for the
				// usual 80% mark would fire only after it had already arrived.
				start: 'top 85%',
				once: true
			}
		});
	}, node);

	return {
		destroy() {
			ctx.revert();
		}
	};
}
