<script lang="ts">
	import Seo from '$lib/Seo.svelte';
</script>

<Seo
	title="Alchemist for news sites"
	description="A clip online in under a minute beats a clip that is perfect in six. Videos become publishable before they are fully prepared, and they keep playing when the country's international links go down."
/>

<section class="shell open">
	<h1>
		A clip online in forty seconds beats a clip that is perfect in six minutes.
	</h1>
	<p class="lede quiet lead">
		In a newsroom the thing that prepares your video is not a quality tool, it is a clock. Ours is built so the clock stops early: the smaller sizes are ready first, and the video can go out then.
	</p>
</section>

<section class="shell bay">
	<dl class="pts">
		<div>
			<dt>Publish before it is fully prepared</dt>
			<dd>
				A video becomes watchable as soon as the smaller sizes exist, not when everything is finished. The larger sizes are made the first time a reader asks for one — once, however many arrive at the same moment.
			</dd>
		</div>
		<div>
			<dt>It keeps working when the cables cut</dt>
			<dd>
				Bangladesh reaches the rest of the world through a handful of undersea cables, and a cut has slowed the whole country more than once. That tends to happen exactly when a story is breaking. Video kept inside the country does not travel those cables.
			</dd>
		</div>
		<div>
			<dt>Bring what you already have</dt>
			<dd>
				Give us a link and we fetch the file ourselves, so you can try this without rebuilding how you already publish. A whole existing library can be brought over at once.
			</dd>
		</div>
		<div>
			<dt>Your readers are on a phone and on a budget</dt>
			<dd>
				Nearly nine in ten internet users here are on mobile data, by BTRC's June 2026 count. A reader standing in a crowd on patchy 4G wants the clip to start, not to be beautiful, so we favour speed over size.
			</dd>
		</div>
	</dl>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2>What is not here yet</h2>
			<p class="quiet">
				Live streaming is not built. If you need it — and in Bangladesh that usually means sport — say so early, because it changes the order of the work rather than slotting into it.
			</p>
			<div class="row">
				<a class="btn btn--solid" href="/pricing/">What it costs</a>
				<a class="btn" href="/docs/">For developers</a>
			</div>
		</div>
	</div>
</section>

<style>
	.open {
		padding-block: clamp(7.5rem, 5.5rem + 7vw, 10rem) 0;
	}
	.open h1 {
		max-width: 19ch;
		margin-bottom: 1.4rem;
	}
	.lead {
		max-width: 56ch;
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
	.row {
		margin-top: 1.6rem;
	}
</style>
