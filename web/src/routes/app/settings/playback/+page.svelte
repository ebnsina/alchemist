<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import Check from '$lib/components/Check.svelte';
	import { readOnly } from '$lib/me.svelte';
	import { getDelivery, setDelivery, ApiError, type DeliverySettings } from '$lib/api';

	let settings = $state<DeliverySettings | null>(null);
	let loading = $state(true);
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	// The encryption toggle was reachable only through the API. A customer whose
	// iPhone viewers got a refusal had nowhere in the product to turn it off.
	let encrypt = $state(false);
	let devices = $state(0);

	async function load() {
		error = '';
		try {
			settings = await getDelivery();
			encrypt = settings.encrypt_playback;
			devices = settings.max_viewer_devices;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		load();
	});

	const changed = $derived(
		!!settings &&
			(encrypt !== settings.encrypt_playback || devices !== settings.max_viewer_devices)
	);

	async function save() {
		error = '';
		saving = true;
		saved = false;
		try {
			// Clamped here rather than left to the API, which answers a generic "check
			// the form" that says nothing about what the range is.
			devices = Math.min(20, Math.max(0, Math.round(Number(devices) || 0)));
			// Both fields, always: the API writes both, so sending one would set the
			// other to its zero value and quietly remove the device cap.
			settings = await setDelivery({ encrypt_playback: encrypt, max_viewer_devices: devices });
			encrypt = settings.encrypt_playback;
			devices = settings.max_viewer_devices;
			saved = true;
		} catch (e) {
			error = e instanceof ApiError ? e.message : 'Something went wrong.';
		} finally {
			saving = false;
		}
	}
</script>

<Seo title="Playback — Alchemist" description="How your videos are protected when somebody watches them." />

<p class="sub max-w-xl">
	How a video is delivered, rather than how it is made. Both settings apply to what happens
	when somebody presses play; neither changes a video you have already sent us.
</p>

{#if error}
	<p class="mt-4 text-sm text-red" role="alert">{error}</p>
{/if}

{#if loading}
	<div class="mt-6 grid gap-3">
		{#each [0, 1] as i (i)}<div class="sk h-36"></div>{/each}
	</div>
{:else}
	<div class="card mt-6">
		<p class="title">Scramble the stored files</p>
		<p class="sub mt-2 max-w-xl">
			Your videos are stored scrambled, so a copy of our storage taken without the key is
			worth nothing. It does not stop somebody who is allowed to watch from recording their
			screen — that is what expiring links and the device limit below are for.
		</p>

		<div class="mt-5">
			<Check
				bind:checked={encrypt}
				label="Scramble stored video"
				hint="On for new accounts."
				disabled={readOnly()}
			/>
		</div>

		<!-- The trade, where the decision is made. Finding out from a customer's
		     support ticket is how this went unnoticed while the toggle had no screen. -->
		<div class="mt-5 rounded-xl border border-sunk p-4">
			<p class="text-sm font-medium">Turning this on costs you iPhone and Safari viewers</p>
			<p class="sub mt-2">
				Apple devices cannot unscramble it. They get a clear message telling them to use
				Chrome, Firefox or Edge rather than a black screen, but they cannot watch. If any
				part of your audience is on an iPhone, iPad or a Mac using Safari, leave this off.
			</p>
			<p class="sub mt-2">
				Changing it applies to videos you send from now on. Everything already here keeps
				what it was made with, because re-making a whole library on a settings change would
				cost you real money for a change nobody watching would notice.
			</p>
			<p class="sub mt-2">Live broadcasts are never scrambled, so they play on anything.</p>
		</div>
	</div>

	<div class="card mt-4">
		<p class="title">How many devices one viewer may use</p>
		<p class="sub mt-2 max-w-xl">
			This is the one that answers a login being shared around a class. It counts only
			links you mint for a named viewer; an ordinary link is not affected.
		</p>

		<label class="mt-5 block max-w-32">
			<span class="label">Devices at once</span>
			<input
				bind:value={devices}
				class="field mt-1.5"
				type="number"
				min="0"
				max="20"
				disabled={readOnly()}
			/>
		</label>
		<p class="sub mt-2">Zero means no limit. Twenty is as high as it goes.</p>
	</div>

	{#if !readOnly()}
		<div class="mt-6 flex flex-wrap items-center gap-3">
			<button type="button" class="btn-solid" disabled={!changed || saving} onclick={save}>
				{saving ? 'Saving…' : 'Save'}
			</button>
			{#if saved && !changed}
				<p class="sub">Saved. It applies to the next video you send.</p>
			{/if}
		</div>
	{/if}
{/if}
