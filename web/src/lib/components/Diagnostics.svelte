<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import { ArrowDown01Icon, ArrowRight01Icon } from '@hugeicons/core-free-icons';
	import type { DiagnosticJob } from '$lib/api';

	let {
		jobs,
		facts,
		loading = false,
		error = ''
	}: {
		jobs: DiagnosticJob[];
		facts: [string, string | null | undefined][];
		loading?: boolean;
		error?: string;
	} = $props();

	let open = $state(false);

	const when = (iso?: string | null) =>
		iso
			? new Intl.DateTimeFormat('en', { dateStyle: 'medium', timeStyle: 'medium' }).format(
					new Date(iso)
				)
			: '—';

	// How long it actually took, which is the first question about a slow encode.
	const took = (j: DiagnosticJob) => {
		if (!j.started_at || !j.finished_at) return '';
		const ms = new Date(j.finished_at).getTime() - new Date(j.started_at).getTime();
		return new Intl.NumberFormat('en', { maximumFractionDigits: 1 }).format(ms / 1000) + 's';
	};
</script>

<section class="card mt-4">
	<button
		type="button"
		class="flex w-full items-center justify-between gap-3 text-left"
		aria-expanded={open}
		onclick={() => (open = !open)}
	>
		<span class="min-w-0">
			<span class="title block">Technical detail</span>
			<span class="sub mt-1 block">
				What the machines said. Only Alchemist staff see this — a customer gets the plain
				reason instead.
			</span>
		</span>
		<HugeiconsIcon
			icon={open ? ArrowDown01Icon : ArrowRight01Icon}
			size={17}
			strokeWidth={2}
			class="flex-none text-dim"
		/>
	</button>

	{#if open}
		{#if error}
			<p class="mt-4 text-sm text-red" role="alert">{error}</p>
		{:else if loading}
			<div class="mt-4 grid gap-2">
				{#each [0, 1, 2] as i (i)}<div class="sk h-4 w-full"></div>{/each}
			</div>
		{:else}
			<dl class="mt-5 grid gap-3 sm:grid-cols-2">
				{#each facts as [k, v] (k)}
					<div class="min-w-0">
						<dt class="label">{k}</dt>
						<dd class="mono mt-1 truncate">{v ?? '—'}</dd>
					</div>
				{/each}
			</dl>

			<p class="label mt-6">Jobs</p>
			{#if jobs.length === 0}
				<p class="sub mt-2">No job has run for this yet.</p>
			{:else}
				<ul class="mt-3 grid gap-3">
					{#each jobs as j (j.id)}
						<li class="rounded-md border border-sunk p-4">
							<div class="flex flex-wrap items-baseline justify-between gap-3">
								<p class="text-sm font-semibold">{j.kind}</p>
								<p class="mono">
									{j.state} · attempt {j.attempt} of {j.max_attempts}
									{#if took(j)} · {took(j)}{/if}
								</p>
							</div>
							<p class="mono mt-1.5">
								queued {when(j.queued_at)} · started {when(j.started_at)} · finished {when(
									j.finished_at
								)}
							</p>
							{#each j.errors as e, i (i)}
								<!-- Verbatim, including the trace: a paraphrased stack trace is no
								     use to whoever has to read it. -->
								<pre
									class="mt-3 overflow-x-auto rounded-md border border-sunk bg-bg p-3 font-mono text-xs text-red">attempt {e.attempt ?? '?'} · {when(
										e.at
									)}
{e.error ?? ''}{e.trace ? '\n\n' + e.trace : ''}</pre>
							{/each}
						</li>
					{/each}
				</ul>
			{/if}
		{/if}
	{/if}
</section>
