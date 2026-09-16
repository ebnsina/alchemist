<script lang="ts" module>
	// Called tag, not badge: app.css already has a badge utility and it is a 24px
	// circle, which quietly turned this into a crescent with the label outside it.
	// Lime for done or going out, red for lost, plain for still happening. The icon
	// and the word both take the tone, so a state reads at a glance from across a
	// column without anybody parsing the label.
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

<span class="tag tag--{tone}">
	<HugeiconsIcon icon={ICON[tone]} size={14} strokeWidth={2} />
	{label}
</span>

<style>
	.tag {
		display: inline-flex;
		align-items: center;
		gap: 7px;
		padding: 4px 10px 4px 8px;
		border-radius: var(--radius-sm);
		background: var(--color-sunk);
		font-size: 12px;
		white-space: nowrap;
		color: var(--color-ink);
	}
	.tag--good {
		color: var(--color-accent);
	}
	.tag--bad {
		color: var(--color-red);
	}
	.tag--busy {
		color: var(--color-dim);
	}
	.tag--idle {
		color: var(--color-faint);
	}
</style>
