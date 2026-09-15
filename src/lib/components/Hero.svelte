<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { Shield01Icon } from '@hugeicons/core-free-icons';
	import ConversionFlow from './ConversionFlow.svelte';
</script>

<section class="mx-auto max-w-[1200px] px-4 pt-[104px] pb-16 sm:px-6">
	<div class="grid items-center gap-10 lg:grid-cols-2 lg:gap-[64px]">
		<div>
			<p class="label flex items-center gap-2 text-dim">
				<HugeiconsIcon icon={Shield01Icon} size={14} strokeWidth={2} class="flex-none" />
				One API for upload, encode and delivery
			</p>

			<h1 class="big mt-4">
				Video infrastructure for your product.
			</h1>

			<p class="measure mt-4 text-dim">
				Your users upload video. We encode it, hold it, and stream it to their viewers behind
				links only they can hand out. You get an API and a key, not a project to run.
			</p>

			<div class="mt-6 flex flex-wrap gap-2">
				<a href="/signup/" class="btn-solid">Start free</a>
				<a href="/#how" class="btn">See how it works</a>
			</div>

			<ul class="sub mt-6 flex flex-wrap gap-x-6 gap-y-1 text-dim">
				<li>Free while you build</li>
				<li>Usage pricing, no seats</li>
				<li>Every account isolated</li>
			</ul>
		</div>

		<ConversionFlow />
	</div>
</section>
