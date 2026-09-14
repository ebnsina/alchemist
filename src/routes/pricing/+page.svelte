<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import Money from '$lib/Money.svelte';
</script>

<Seo
	title="Alchemist pricing"
	description="You pay for the video you send us, the video we hold, and the video people watch. No per-person fee, no minimum, no charge to leave, cancel any time. We publish what it costs us to run alongside the rates."
/>

<section class="shell hero">
	<div class="flow">
		<h1>Three things you pay for, and nothing else.</h1>
		<p class="lede quiet">
			You pay for the minutes of video you send us, the gigabytes we hold, and the gigabytes people watch. Everything else is simply how the thing works, and is not charged for separately.
		</p>
	</div>
	<div class="notice">
		<p class="ntag">How the rates work</p>
		<p class="small">
			Rates are per unit, with no plans to pick between and nothing to negotiate. Pricing is not final until launch, so check back before you sign anything off.
		</p>
	</div>
</section>

<section class="shell bay">
	<div class="meters">
		<article class="flow">
			<h2>Preparing your video</h2>
			<p class="price"><Money amount={2} /><span class="note">
					for each minute of video you send · about $0.016
				</span>
			</p>
			<p class="small">
				Charged once, on the length of the video you send us. We make several sizes of it so it plays on any connection, and you are not charged for each one — a one-hour lecture is one hour.
			</p>
		</article>

		<article class="flow">
			<h2>Keeping it</h2>
			<p class="price"><Money amount={1.2} /><span class="note">
					per GB per month · about $0.010
				</span>
			</p>
			<p class="small">
				Counted on what we actually hold, which is less than you would guess. We do not keep your original file unless you ask us to, and the larger sizes only exist for videos somebody has watched.
			</p>
		</article>

		<article class="flow">
			<h2>Sending it to viewers</h2>
			<p class="price"><Money amount={0.35} /><span class="note">
					for each GB watched inside Bangladesh · about $0.003
				</span>
				</p>
				<p class="price"><Money amount={1.2} /><span class="note">
					for each GB watched anywhere else · about $0.010
				</span>
			</p>
			<p class="small">
				Two rates because they cost us two different amounts. Reaching someone in Bangladesh is far cheaper than reaching someone abroad, and charging one flat rate would mean your Bangladeshi viewers quietly paying for the difference.
			</p>
		</article>
	</div>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2>
				What a mid-sized course platform would pay.
			</h2>
			<p class="quiet">
				Twenty hours of new lecture a month, five hundred gigabytes of videos held, and two hundred and fifty gigabytes watched by students in Bangladesh. That is the three rates above multiplied out, and nothing else — there is no monthly platform fee underneath.
			</p>
			<p class="fine">
				Worked out at the rates above. Your own figure depends on how much of your library actually gets watched.
			</p>
		</div>

		<div class="bill">
			<table>
				<caption class="vh">Worked monthly example</caption>
				<tbody>
					<tr>
						<th scope="row">Preparing 20 hours of video</th>
						<td class="figure"><Money amount={2400} decimals={0} /></td>
					</tr>
					<tr>
						<th scope="row">Keeping 500 GB</th>
						<td class="figure"><Money amount={600} decimals={0} /></td>
					</tr>
					<tr>
						<th scope="row">Sending 250 GB inside Bangladesh</th>
						<td class="figure"><Money amount={88} decimals={0} /></td>
					</tr>
				</tbody>
				<tfoot>
					<tr>
						<th scope="row">A month</th>
						<td class="figure total"><Money amount={3088} decimals={0} /></td>
					</tr>
				</tfoot>
			</table>
			<p class="fine conv">
				About $25. Conversions on this page are approximate.
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="wide">Things you will never see on an invoice.</h2>
	<ul class="never">
		<li>Seats. Add as many teammates as you like.</li>
		<li>The several sizes we make of each video so it plays everywhere.</li>
		<li>Asking us questions — how many videos you have, how much you have used.</li>
		<li>Us telling your system when a video is ready.</li>
		<li>Scrambling, expiring links and preview images. They are the product, not an upgrade.</li>
		<li>Leaving, and taking your files with you.</li>
	</ul>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2>What this costs us to run.</h2>
			<p class="quiet">
				Showing you what it costs us is more use than hiding behind a price list. These figures are our own planning numbers, in euros, worked out from what the machines cost to rent and how much video one gets through. They are what the machines cost, not what we charge.
			</p>
			<p class="quiet">
				The gap between that and the rates above is margin, and it stays visible. A number you can check is worth more than one you have to take on faith.
			</p>
		</div>
		<div class="measure">
			<div class="scroll-x">
				<table>
					<caption>What our servers cost us, per hour of video prepared</caption>
					<thead>
						<tr>
							<th scope="col">Volume</th>
							<th scope="col">Servers</th>
							<th scope="col">Per hour</th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<th scope="row">2,000 hours of video a month</th>
							<td>3 machines, €350–500 a month</td>
							<td class="figure">€0.20</td>
						</tr>
						<tr>
							<th scope="row">100,000 hours of video a month</th>
							<td>About €13,630 a month</td>
							<td class="figure">€0.14</td>
						</tr>
					</tbody>
				</table>
			</div>
			<p class="fine">
				At roughly ৳135 to the euro — worth checking the rate yourself — €0.14 is about ৳19 of server cost per hour of video. The time it takes to build and run the thing is not in these numbers, and it is the larger cost of the two.
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<div class="flow narrow">
		<h2>In plain words.</h2>
		<dl class="plainly">
			<dt>Taka for Bangladeshi customers</dt>
			<dd>
				Collected through bKash, Nagad or card — not a foreign-currency charge with a bank's markup stacked on top. USD for everyone else.
			</dd>
			<dt>You can check the numbers yourself</dt>
			<dd>
				Your usage is counted from the first day, and you can read the same numbers the bill is built from at any time. If our bill and your own figures disagree, you should be able to prove it.
			</dd>
			<dt>No minimum, no setup fee, cancel any time</dt>
			<dd>
				There is no asterisk after any of those words, no notice period, and no exit fee.
			</dd>
			<dt>Price changes reach you before they take effect</dt>
			<dd>
				Not after. Keeping a customer by making the way out confusing is not something we are willing to do.
			</dd>
		</dl>
	</div>
</section>

<style>
	.hero {
		padding-block: clamp(7.5rem, 5.5rem + 7vw, 10rem) 0;
	}
	.hero h1 {
		max-width: 14ch;
	}
	.notice {
		margin-top: 2.5rem;
		padding: 1.1rem 1.3rem;
		background: var(--wash);
		border-top: 2px solid var(--brass);
		border-radius: 0 0 4px 4px;
	}
	.ntag {
		font-weight: 600;
		color: var(--brass);
		margin-bottom: 0.25rem;
	}
	.notice .small {
		max-width: 78ch;
	}

	.meters {
		display: grid;
		gap: clamp(1.75rem, 4vw, 3rem);
		grid-template-columns: repeat(auto-fit, minmax(245px, 1fr));
	}
	.meters h2 {
		font-size: var(--step-3);
	}

	.measure {
		border-top: 1px solid var(--rule-firm);
		padding-top: 1.2rem;
	}
	.measure .fine {
		margin-top: 1rem;
	}
	.bill {
		border: 1px solid var(--rule);
		border-radius: var(--r);
		background: var(--leaf);
		box-shadow: var(--lift);
		padding: 1rem 1.2rem 1.1rem;
	}
	.bill table {
		margin-top: 0.5rem;
		font-size: var(--step-0);
	}
	.bill th {
		font-weight: 400;
		color: var(--graphite);
	}
	.bill td {
		text-align: right;
		padding-right: 0;
	}
	.bill tfoot th,
	.bill tfoot td {
		border-bottom: 0;
		padding-top: 0.8rem;
		font-weight: 600;
		color: var(--ink);
	}
	.total {
		font-size: var(--step-2);
	}
	.conv {
		margin-top: 0.6rem;
	}

	.wide {
		max-width: 20ch;
		margin-bottom: 1.8rem;
	}
	.never {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 0;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		column-gap: clamp(1.5rem, 4vw, 3rem);
	}
	.never li {
		padding: 0.8rem 0;
		border-top: 1px solid var(--rule);
		color: var(--graphite);
		font-size: var(--step-0);
	}

	.narrow {
		max-width: 760px;
	}
	.plainly {
		margin: 0;
	}
	.plainly dt {
		font-weight: 600;
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		margin-bottom: 0.3rem;
	}
	.plainly dt:first-of-type {
		border-top-color: var(--rule-firm);
	}
	.plainly dd {
		margin: 0 0 1.4rem;
		color: var(--graphite);
		font-size: var(--step-0);
	}
</style>
