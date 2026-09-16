<script lang="ts" module>
	// Lime for work that is done or going out, red for work that is lost, plain for
	// work still happening. The icon carries the color; the word stays readable ink,
	// so the two hues never have to do the job of the label.
	export type Tone = 'good' | 'bad' | 'busy' | 'idle';
</script>

<script lang="ts">
	import { HugeiconsIcon } from '@hugeicons/svelte';
	import {
		CheckmarkCircle02Icon,
		AlertCircleIcon,
		Loading03Icon,
		CircleIcon
	} from '@hugeicons/core-free-icons';

	let { label, tone = 'idle' }: { label: string; tone?: Tone } = $props();

	const ICON = {
		good: CheckmarkCircle02Icon,
		bad: AlertCircleIcon,
		busy: Loading03Icon,
		idle: CircleIcon
	};
</script>

<span class="badge badge--{tone}">
	<HugeiconsIcon icon={ICON[tone]} size={14} strokeWidth={2} class="badge__icon" />
	{label}
</span>

<style>
	.badge {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		padding: 4px 10px 4px 8px;
		border-radius: var(--radius-sm);
		border: 1px solid var(--color-sunk);
		border-left-width: 3px;
		background: var(--color-sunk);
		font-size: 12px;
		white-space: nowrap;
		color: var(--color-ink);
	}
	.badge--good {
		border-left-color: var(--color-brand);
	}
	.badge--good :global(.badge__icon) {
		color: var(--color-accent);
	}
	.badge--bad {
		border-left-color: var(--color-red);
	}
	.badge--bad :global(.badge__icon) {
		color: var(--color-red);
	}
	.badge--busy {
		border-left-color: var(--color-dim);
	}
	.badge--busy :global(.badge__icon) {
		color: var(--color-dim);
	}
	.badge--idle {
		border-left-color: var(--color-muted, var(--color-sunk));
	}
	.badge--idle :global(.badge__icon) {
		color: var(--color-faint);
	}
</style>
