<script lang="ts">
	import Seo from '$lib/Seo.svelte';
</script>

<Seo
	title="About Alchemist — what this project believes"
	description="Transparent pricing, no manipulative growth tactics, no dark patterns, and no borrowed credibility. A short statement of how this project behaves."
/>

<section class="shell hero">
	<div class="flow">
		<h1>Software that behaves itself.</h1>
		<p class="lede quiet">
			Alchemy is al-kimiya — you put in something base and take out something refined. That is roughly what this does to a video file, and it is a better name than another word ending in -ly.
		</p>
	</div>
</section>

<section class="shell bay">
	<dl class="beliefs">
		<div>
			<dt>Say the price. All of it.</dt>
			<dd>
				Every rate is published, what it costs us to run is published beside it, and there is no line item you discover on an invoice. No page on this site says contact sales.
			</dd>
		</div>
		<div>
			<dt>No manufactured urgency.</dt>
			<dd>
				No countdown timers, no last three places, no pre-ticked boxes, no trial that quietly becomes a subscription. A product that needs a clock to close a deal is telling you something.
			</dd>
		</div>
		<div>
			<dt>No borrowed credibility.</dt>
			<dd>
				You will not find customer logos, testimonials, case studies, uptime percentages or compliance badges here, because we do not have any of them yet. When we do, they will name real people who agreed to be named.
			</dd>
		</div>
		<div>
			<dt>Measured, or marked.</dt>
			<dd>
				A number here is either something we measured, or something published by whoever measured it and cited to them. Where a figure is still moving — pricing, before launch — the page says so once and does not pretend otherwise.
			</dd>
		</div>
		<div>
			<dt>The door out stays open.</dt>
			<dd>
				Your files are yours, the API is the whole product with no privileged internal surface kept back for ourselves, and leaving costs nothing. Lock-in built by making export difficult is a tax on customers who stopped being happy.
			</dd>
		</div>
		<div>
			<dt>Bangla is not a translation layer.</dt>
			<dd>
				Error messages, player chrome and subtitles are written in Bangla, not machine-translated into it, and the text is checked for correct conjunct shaping rather than assumed to work. This page is bilingual for the same reason.
			</dd>
		</div>
	</dl>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2>Where it stands</h2>
			<p class="quiet">
				Alchemist runs on machines we rent and look after ourselves, with no other company’s video service underneath it. That is a deliberate choice about cost and control, and it is also the only reason keeping your videos inside Bangladesh is possible at all.
			</p>
			<p class="quiet">
				We are onboarding a first set of customers by hand rather than opening sign-ups, and the BDIX edge is rolling out node by node. If you are weighing us against Mux or Bunny for something that starts next term, talk to us and we will tell you exactly where we are that week.
			</p>
		</div>
	</div>
</section>

<style>
	.hero {
		padding-block: clamp(7.5rem, 5.5rem + 7vw, 10rem) 0;
	}
	.hero h1 {
		max-width: 13ch;
	}
	.beliefs {
		margin: 0;
		display: grid;
		gap: 0 clamp(1.5rem, 4vw, 3.5rem);
		grid-template-columns: repeat(auto-fit, minmax(290px, 1fr));
	}
	.beliefs dt {
		font-weight: 600;
		font-size: var(--step-2);
		letter-spacing: -0.012em;
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		margin-bottom: 0.4rem;
	}
	.beliefs dd {
		margin: 0 0 1.8rem;
		color: var(--graphite);
		font-size: var(--step-0);
	}
</style>
