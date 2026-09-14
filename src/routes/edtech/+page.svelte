<script lang="ts">
	import Seo from '$lib/Seo.svelte';
</script>

<Seo
	title="Alchemist for course platforms"
	description="Lectures that play on a cheap phone without stalling, that do not drain a student's data, and that cannot simply be downloaded and passed around."
/>

<section class="shell open">
	<h1>
		Your two real problems are copying and your students' data bill.
	</h1>
	<p class="lede quiet lead">
		Everything else about putting a lecture online has been solved for years. These two have not, and both are harder here than in the countries the big video companies designed for.
	</p>
</section>

<section class="shell bay">
	<div class="split split--even">
		<div class="flow">
			<h2>Courses getting copied</h2>
			<p class="quiet">
				Every video is scrambled, and the key to unscramble it is issued for one viewing rather than once per course. A link a student copies into a Telegram group stops working within hours instead of never.
			</p>
			<p class="cav">
				It does not stop somebody recording their own screen. Only the copy protection the film studios use does, and we have not built that. We would rather tell you than sell you the word unbreakable.
			</p>
		</div>
		<div class="flow">
			<h2>The data bill</h2>
			<p class="quiet">
				A lecture is a person talking or a screen being shared. It squeezes down much further than film footage does, and most video services are tuned for film. An hour of lecture at our default size costs a student about 270 MB.
			</p>
			<p class="cav">
				We expect lectures and screen recordings to come down another forty to sixty per cent once we finish sizing each video individually. That is what we expect from this kind of material, not something we have measured on your videos — that work is not finished, and we will not quote you a saving we have not measured.
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="wide">Things that matter in a classroom and nowhere else</h2>
	<dl class="pts">
		<div>
			<dt>Audio on its own</dt>
			<dd>
				Most of a lecture is the talking. A student revising can listen instead of watch, for roughly a tenth of the data. On a metered connection that is the difference between finishing the course and rationing it.
			</dd>
		</div>
		<div>
			<dt>Publishable before it is finished</dt>
			<dd>
				A video becomes watchable as soon as the smaller sizes are ready. A teacher uploading at eleven the night before class does not sit and wait for the rest.
			</dd>
		</div>
		<div>
			<dt>Bangla that displays properly</dt>
			<dd>
				Bangla titles, file names and subtitles survive from your upload to the student's screen. Conjunct letters keep their shape instead of turning into empty boxes, which is a real and common failure.
			</dd>
		</div>
		<div>
			<dt>A data saver the student can see</dt>
			<dd>
				The player has a visible button that holds the video at the smaller size and shows what an hour will cost. Hidden in a settings menu it is a feature; in plain sight it is a reason to trust you.
			</dd>
		</div>
	</dl>
	<div class="tail flow">
		<div class="row">
			<a class="btn btn--solid" href="/pricing/">What it costs</a>
			<a class="btn" href="/docs/">For developers</a>
		</div>
	</div>
</section>

<style>
	.open {
		padding-block: clamp(7.5rem, 5.5rem + 7vw, 10rem) 0;
	}
	.open h1 {
		max-width: 17ch;
		margin-bottom: 1.4rem;
	}
	.lead {
		max-width: 56ch;
	}
	.cav {
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		font-size: var(--step-0);
		color: var(--graphite);
	}
	.wide {
		max-width: 22ch;
		margin-bottom: 2rem;
	}
	.pts {
		margin: 0;
		display: grid;
		gap: 0 clamp(1.5rem, 4vw, 3.5rem);
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
	}
	.pts dt {
		font-weight: 600;
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		margin-bottom: 0.35rem;
	}
	.pts dd {
		margin: 0 0 1.6rem;
		color: var(--graphite);
		font-size: var(--step-0);
	}
	.tail {
		margin-top: 1.2rem;
		padding-top: 1.4rem;
		border-top: 1px solid var(--rule-firm);
	}
</style>
