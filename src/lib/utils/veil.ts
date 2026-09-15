import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { ScrollToPlugin } from 'gsap/ScrollToPlugin';

gsap.registerPlugin(ScrollTrigger, ScrollToPlugin);

/**
 * The page never scrolls. It stays on one screen, and each section wipes in from the
 * right over the one before it — the way a Barba transition brings a page in.
 *
 * The stack is pinned for as long as there are sections left, so the wheel spends
 * its distance on the reveal rather than on moving the page. Scrub ties the veil to
 * the scroll position — dragging back closes it again — and snap settles on whole
 * sections when the gesture stops, so the page is never left half-veiled.
 *
 * The mechanism is clip-path rather than transform: the incoming section does not
 * move, it is uncovered by an edge travelling right to left. Sliding it would drag
 * type across type, which is what makes stacked-section pages feel cheap.
 *
 * Off below 1024px and under reduced motion. There the sections are ordinary blocks
 * in ordinary flow, which is what the markup already is — this only adds behaviour,
 * it never holds the content.
 */
export function veil(stack: HTMLElement) {
	const media = gsap.matchMedia();

	media.add(
		{ wide: '(min-width: 1024px)', motion: '(prefers-reduced-motion: no-preference)' },
		(context) => {
			const { wide, motion } = context.conditions as { wide: boolean; motion: boolean };
			if (!wide || !motion) return;

			const sections = Array.from(stack.querySelectorAll<HTMLElement>('section.screen'));
			if (sections.length < 2) return;
			const steps = sections.length - 1;

			// The browser's own smooth scrolling fights a scrubbed trigger and an
			// anchor jump alike: it lags behind every write and lands short.
			const root = document.documentElement;
			const inherited = root.style.scrollBehavior;
			root.style.scrollBehavior = 'auto';

			stack.classList.add('stack--on');
			sections.forEach((s, i) => {
				gsap.set(s, {
					position: 'absolute',
					inset: 0,
					zIndex: i,
					clipPath: i === 0 ? 'inset(0% 0% 0% 0%)' : 'inset(0% 0% 0% 100%)'
				});
			});

			const tl = gsap.timeline();
			sections.slice(1).forEach((s, i) => {
				tl.to(s, { clipPath: 'inset(0% 0% 0% 0%)', ease: 'none', duration: 1 }, i);
			});

			const trigger = ScrollTrigger.create({
				trigger: stack,
				start: 'top top',
				end: () => `+=${steps * window.innerHeight}`,
				pin: true,
				scrub: 0.6,
				animation: tl,
				snap: {
					snapTo: 1 / steps,
					duration: { min: 0.2, max: 0.5 },
					ease: 'power2.inOut'
				},
				invalidateOnRefresh: true
			});

			// An anchor cannot scroll to a section that never moves: it has to scroll to
			// the point in the pin where that section is uncovered.
			const onClick = (e: MouseEvent) => {
				const link = (e.target as HTMLElement | null)?.closest?.('a[href*="#"]');
				if (!link) return;
				const href = link.getAttribute('href') ?? '';
				const hash = href.slice(href.indexOf('#'));
				if (hash.length < 2) return;
				const target = document.querySelector<HTMLElement>(hash);
				const index = target ? sections.findIndex((s) => s === target || s.contains(target)) : -1;
				if (index < 0) return;
				e.preventDefault();
				history.replaceState(null, '', hash);
				const span = trigger.end - trigger.start;
				gsap.to(window, {
					scrollTo: { y: trigger.start + (index / steps) * span, autoKill: false },
					duration: 0.6,
					ease: 'power2.inOut'
				});
			};
			document.addEventListener('click', onClick);

			return () => {
				document.removeEventListener('click', onClick);
				trigger.kill();
				tl.kill();
				gsap.set(sections, { clearProps: 'position,inset,zIndex,clipPath' });
				stack.classList.remove('stack--on');
				root.style.scrollBehavior = inherited;
			};
		}
	);

	return () => media.revert();
}
