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
	let origins = $state<string[]>([]);
	let newOrigin = $state('');
	let ttl = $state(14400);

	async function load() {
		error = '';
		try {
			settings = await getDelivery();
			encrypt = settings.encrypt_playback;
			devices = settings.max_viewer_devices;
			origins = [...settings.playback_origins];
			ttl = settings.playback_ttl_seconds;
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
			(encrypt !== settings.encrypt_playback ||
				devices !== settings.max_viewer_devices ||
				ttl !== settings.playback_ttl_seconds ||
				origins.join('\n') !== settings.playback_origins.join('\n'))
	);

	// Added to the list rather than saved immediately, so one Save covers the whole
	// page and a half-typed address never becomes a rule.
	function addOrigin() {
		const v = newOrigin.trim();
		if (!v || origins.includes(v)) return;
		origins = [...origins, v];
		newOrigin = '';
	}

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
			ttl = Math.min(86400, Math.max(60, Math.round(Number(ttl) || 14400)));
			settings = await setDelivery({
				encrypt_playback: encrypt,
				playback_origins: origins,
				playback_ttl_seconds: ttl,
				max_viewer_devices: devices
			});
			encrypt = settings.encrypt_playback;
			devices = settings.max_viewer_devices;
			ttl = settings.playback_ttl_seconds;
			// Echoed back normalised — a trailing slash or capitals are tidied by the
			// API, and the list on screen has to be the list that is enforced.
			origins = [...settings.playback_origins];
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
				hint="Off for new accounts."
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
		<p class="title">Where your videos are allowed to play</p>
		<p class="sub mt-2 max-w-xl">
			Name the sites your videos are embedded on and a playback link only works inside a
			page on one of them. A link copied out and pasted somewhere else does nothing, for
			as long as it would otherwise have lasted.
		</p>

		{#if origins.length}
			<ul class="mt-5 divide-y divide-sunk border-y border-sunk">
				{#each origins as o (o)}
					<li class="flex items-center gap-3 py-2.5">
						<code class="min-w-0 flex-1 truncate font-mono text-xs">{o}</code>
						{#if !readOnly()}
							<button
								type="button"
								class="flex-none text-xs text-dim hover:text-red"
								onclick={() => (origins = origins.filter((x) => x !== o))}
							>
								Remove
							</button>
						{/if}
					</li>
				{/each}
			</ul>
		{:else}
			<p class="sub mt-5">
				Nothing listed, so your videos play wherever your link is used.
			</p>
		{/if}

		{#if !readOnly()}
			<div class="mt-4 flex flex-wrap items-end gap-3">
				<label class="min-w-56 flex-1">
					<span class="label">Add a site</span>
					<input
						bind:value={newOrigin}
						class="field mt-1.5"
						type="url"
						placeholder="https://app.yourschool.com"
						onkeydown={(e) => {
							if (e.key === 'Enter') {
								e.preventDefault();
								addOrigin();
							}
						}}
					/>
				</label>
				<button type="button" class="btn flex-none" onclick={addOrigin} disabled={!newOrigin.trim()}>
					Add
				</button>
			</div>
		{/if}

		<!-- The limit, said where the decision is, because finding it out means a
		     customer's app has already stopped playing. -->
		<div class="mt-5 rounded-xl border border-sunk p-4">
			<p class="text-sm font-medium">Only browsers can be checked this way</p>
			<p class="sub mt-2">
				A web page tells us which site it is; a phone app or a TV app does not, so a link
				locked to a site will not play in one. If your videos are watched in an app of
				your own, leave this list empty — expiring links still protect them.
			</p>
		</div>
	</div>

	<div class="card mt-4">
		<p class="title">How long a playback link lasts</p>
		<p class="sub mt-2 max-w-xl">
			A link you hand out stops working after this. Your app asks for the video again
			whenever somebody presses play, so this is not how long a video lives — only how
			long one link to it keeps working.
		</p>

		<label class="mt-5 block max-w-40">
			<span class="label">Minutes</span>
			<input
				value={Math.round(ttl / 60)}
				oninput={(e) => (ttl = Math.round(Number(e.currentTarget.value) * 60))}
				class="field mt-1.5"
				type="number"
				min="1"
				max="1440"
				disabled={readOnly()}
			/>
		</label>

		<!-- The floor, where it is chosen. Set this to fifteen minutes and a
		     forty-minute lecture stops playing two thirds of the way through. -->
		<div class="mt-5 rounded-xl border border-sunk p-4">
			<p class="text-sm font-medium">Set this from your longest video, not your shortest</p>
			<p class="sub mt-2">
				The same link carries the whole video, so if it runs out while somebody is
				watching, the video stops there. A four hour link covers a three hour lecture; a
				fifteen minute one does not.
			</p>
			<p class="sub mt-2">
				If what you are worried about is a link being passed around rather than kept too
				long, the list of allowed sites above is the setting for that — it does not care
				how long the link lasts.
			</p>
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
