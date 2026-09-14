<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
	import Icon from '$lib/Icon.svelte';
</script>

<Seo
	title="Alchemist pricing — what gets billed, and what it costs us"
	description="Three meters: ingest minutes, storage, delivery. No per-seat fee, no minimum, no overage penalty. The numbers on this page are placeholders and are marked as such — the cost basis behind them is not."
/>

<section class="top">
	<div class="wrap">
		<p class="eyebrow"><T en="Pricing" bn="মূল্য" /></p>
		<h1><T as="span" en="Three meters, and nothing else." bn="তিনটি মিটার, আর কিছু নয়।" /></h1>
		<div class="banner" role="note">
			<Icon name="coins" size={22} />
			<p>
				<T
					as="span"
					en="Every price on this page is a placeholder. Alchemist has not launched and no price has been set. The structure below is real and the cost basis further down is real; the numbers are not, and they are marked wherever they appear so they cannot be mistaken for a quote."
					bn="এই পাতার প্রতিটি দাম অনুমান। Alchemist এখনও চালু হয়নি এবং কোনো দাম নির্ধারণ করা হয়নি। নিচের কাঠামোটা সত্যি, আর তারও নিচের খরচের হিসাবটাও সত্যি; শুধু সংখ্যাগুলো নয় — আর যেখানেই সেগুলো এসেছে, আলাদা করে চিহ্নিত করা আছে যাতে কেউ একে দর বলে ভুল না করে।"
				/>
			</p>
		</div>
	</div>
</section>

<section>
	<div class="wrap">
		<div class="grid">
			<article class="card meter">
				<p class="mlabel"><T en="Meter 1" bn="মিটার ১" /></p>
				<h2><T as="span" en="Ingest" bn="ইনজেস্ট" /></h2>
				<p class="price">
					<span class="ph"><span class="num">0.00</span></span>
					<span class="unit"
						><T as="span" en="per minute of source video" bn="সোর্স ভিডিওর প্রতি মিনিটে" /></span
					>
				</p>
				<p>
					<T
						as="span"
						en="Charged once, on the duration of the file you send us — not on the number of renditions produced. Encoding a one-hour lecture into five rungs is one hour of ingest, not five."
						bn="একবারই নেওয়া হয়, আপনি যে ফাইল পাঠাচ্ছেন তার দৈর্ঘ্যের উপর — কতগুলো রেন্ডিশন তৈরি হলো তার উপর নয়। এক ঘণ্টার লেকচার পাঁচটি ধাপে এনকোড করা মানে এক ঘণ্টার ইনজেস্ট, পাঁচ ঘণ্টার নয়।"
					/>
				</p>
			</article>

			<article class="card meter">
				<p class="mlabel"><T en="Meter 2" bn="মিটার ২" /></p>
				<h2><T as="span" en="Storage" bn="স্টোরেজ" /></h2>
				<p class="price">
					<span class="ph"><span class="num">0.00</span></span>
					<span class="unit"><T as="span" en="per GB per month" bn="প্রতি জিবি প্রতি মাসে" /></span>
				</p>
				<p>
					<T
						as="span"
						en="Measured on what is actually stored, which is less than you would expect: the original is dropped after the mezzanine is made unless you ask us to keep it, and the higher rungs only exist for content somebody has watched."
						bn="যা সত্যিই জমা আছে তার উপর হিসাব — আর সেটা আপনার ধারণার চেয়ে কম: মেজানিন তৈরি হয়ে গেলে মূল ফাইল ফেলে দেওয়া হয়, যদি না আপনি রাখতে বলেন; আর উপরের ধাপগুলো শুধু সেই কনটেন্টের জন্যই থাকে যা কেউ দেখেছে।"
					/>
				</p>
			</article>

			<article class="card meter">
				<p class="mlabel"><T en="Meter 3" bn="মিটার ৩" /></p>
				<h2><T as="span" en="Delivery" bn="ডেলিভারি" /></h2>
				<p class="price">
					<span class="ph"><span class="num">0.00</span></span>
					<span class="unit"
						><T as="span" en="per GB delivered inside BDIX" bn="BDIX-এর ভেতরে প্রতি জিবিতে" /></span
					>
				</p>
				<p class="price second">
					<span class="ph"><span class="num">0.00</span></span>
					<span class="unit"
						><T as="span" en="per GB delivered internationally" bn="দেশের বাইরে প্রতি জিবিতে" /></span
					>
				</p>
				<p>
					<T
						as="span"
						en="Two rates because they are two different costs, and pretending otherwise would mean charging BD viewers for transit nobody paid for. BDIX delivery is the cheaper one and we intend to price it that way."
						bn="দুটি আলাদা হার, কারণ এ দুটোর খরচই আলাদা — এক দেখানো মানে দেশের দর্শকদের এমন ট্রানজিটের জন্য টাকা নেওয়া যার খরচ কেউ দেয়নি। BDIX ডেলিভারির খরচ কম, আর আমরা দামও সেভাবেই রাখতে চাই।"
					/>
				</p>
			</article>
		</div>

		<div class="nots">
			<h2><T as="span" en="What you will never be billed for" bn="যেসবের বিল কখনও আসবে না" /></h2>
			<ul>
				<li><Icon name="check" size={17} /><T as="span" en="Seats. Add as many teammates as you like." bn="সিট। যত খুশি সহকর্মী যোগ করুন।" /></li>
				<li><Icon name="check" size={17} /><T as="span" en="Renditions. One ingest, however many rungs your profile has." bn="রেন্ডিশন। একবার ইনজেস্ট, আপনার প্রোফাইলে যত ধাপই থাকুক।" /></li>
				<li><Icon name="check" size={17} /><T as="span" en="API requests to the control plane." bn="কন্ট্রোল প্লেনে পাঠানো API রিকোয়েস্ট।" /></li>
				<li><Icon name="check" size={17} /><T as="span" en="Webhook deliveries and retries." bn="ওয়েবহুক ডেলিভারি ও পুনঃচেষ্টা।" /></li>
				<li><Icon name="check" size={17} /><T as="span" en="Encryption, signed URLs, or thumbnails. They are how the product works, not an upsell." bn="এনক্রিপশন, সাইন করা লিংক বা থাম্বনেইল। এগুলো প্রোডাক্টের কাজ করার ধরন, আলাদা করে বেচার জিনিস নয়।" /></li>
			</ul>
		</div>
	</div>
</section>

<section class="alt">
	<div class="wrap narrow">
		<p class="eyebrow"><T en="The cost basis" bn="খরচের ভিত্তি" /></p>
		<h2><T as="span" en="What this actually costs to run" bn="এটা চালাতে আসলে কত খরচ" /></h2>
		<p>
			<T
				as="span"
				en="We would rather show the cost model than hide behind a pricing page. These figures come from our own capacity planning, in euros, and they are modelled from hardware prices and measured encoder throughput — not billed revenue, because there is none yet."
				bn="মূল্যতালিকার আড়ালে লুকানোর চেয়ে খরচের হিসাবটাই দেখানো ভালো মনে করি। এই সংখ্যাগুলো আমাদের নিজেদের ক্যাপাসিটি পরিকল্পনা থেকে নেওয়া, ইউরোতে, আর হিসাব করা হয়েছে হার্ডওয়্যারের দাম ও মেপে দেখা এনকোডার গতির উপর ভিত্তি করে — বিল করা আয়ের ভিত্তিতে নয়, কারণ সেটা এখনও নেই।"
			/>
		</p>
		<div class="scroll-x">
			<table>
				<thead>
					<tr>
						<th scope="col"><T en="Scenario" bn="পরিস্থিতি" /></th>
						<th scope="col"><T en="Infrastructure" bn="অবকাঠামো" /></th>
						<th scope="col"><T en="Cost per source hour" bn="প্রতি সোর্স ঘণ্টায় খরচ" /></th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<th scope="row"><T as="span" en="2,000 hrs/month (start)" bn="২,০০০ ঘণ্টা/মাস (শুরু)" /></th>
						<td><T as="span" en="3 machines · €350–500/mo" bn="৩টি মেশিন · €৩৫০–৫০০/মাস" /></td>
						<td class="num">≈ €0.20</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="100,000 hrs/month (scaled)" bn="১,০০,০০০ ঘণ্টা/মাস (বড় পরিসরে)" /></th>
						<td><T as="span" en="≈ €13,630/mo" bn="≈ €১৩,৬৩০/মাস" /></td>
						<td class="num">≈ €0.14</td>
					</tr>
				</tbody>
			</table>
		</div>
		<p class="small">
			<T
				as="span"
				en="At roughly 135 BDT to the euro — verify the current rate — €0.14 is about 19 BDT per source hour ingested. Those are our costs, not your price. The gap between them is margin and it will be visible rather than hidden; we would rather set a number you can sanity-check than one you have to take on faith."
				bn="প্রতি ইউরো মোটামুটি ১৩৫ টাকা ধরে — বর্তমান রেট যাচাই করে নেবেন — €০.১৪ মানে প্রতি সোর্স ঘণ্টায় প্রায় ১৯ টাকা। এটা আমাদের খরচ, আপনার দাম নয়। দুইয়ের মাঝের ফারাকটাই মুনাফা, আর সেটা লুকানো থাকবে না; বিশ্বাস করে মেনে নিতে হয় এমন সংখ্যার চেয়ে যাচাই করা যায় এমন সংখ্যাই আমরা দিতে চাই।"
			/>
		</p>
		<div class="note">
			<p>
				<T
					as="span"
					en="For comparison, from published list prices recorded in our cost model and worth re-checking before you rely on them: Mux is around $0.05 per minute of encoding, roughly $3.00 per hour, plus delivery. Bunny Stream is around $0.005 per minute plus delivery. Neither of them can deliver inside BDIX."
					bn="তুলনার জন্য, আমাদের খরচের হিসাবে লিখে রাখা প্রকাশ্য তালিকামূল্য — নির্ভর করার আগে আবার যাচাই করে নেওয়া ভালো: Mux-এ এনকোডিং প্রতি মিনিটে প্রায় $০.০৫, অর্থাৎ ঘণ্টায় প্রায় $৩.০০, তার উপর ডেলিভারির খরচ। Bunny Stream-এ প্রতি মিনিটে প্রায় $০.০০৫, সঙ্গে ডেলিভারি। তবে এদের কেউই BDIX-এর ভেতর থেকে দিতে পারে না।"
				/>
			</p>
		</div>
	</div>
</section>

<section>
	<div class="wrap narrow">
		<h2><T as="span" en="How we intend to charge, in plain words" bn="সহজ কথায়, কীভাবে টাকা নেওয়ার ইচ্ছা" /></h2>
		<ul class="plain">
			<li>
				<T
					as="span"
					en="Prices in BDT for Bangladeshi customers, collected through bKash, Nagad or card — not a foreign currency charge with a bank's markup on top. USD for everyone else."
					bn="বাংলাদেশি গ্রাহকের জন্য দাম টাকায়, আদায় হবে বিকাশ, নগদ বা কার্ডে — ব্যাংকের বাড়তি চার্জসহ বিদেশি মুদ্রার বিল নয়। বাকি সবার জন্য ডলারে।"
				/>
			</li>
			<li>
				<T
					as="span"
					en="Usage is metered from the first request and the same numbers you are billed on are readable from the API at any time. If our invoice and your meter disagree, you should be able to prove it."
					bn="প্রথম রিকোয়েস্ট থেকেই ব্যবহার মাপা হয়, আর যে সংখ্যার উপর বিল হবে সেই একই সংখ্যা যেকোনো সময় API থেকে দেখা যায়। আমাদের বিল আর আপনার হিসাব না মিললে আপনি সেটা প্রমাণ করতে পারবেন।"
				/>
			</li>
			<li>
				<T
					as="span"
					en="No minimum commitment, no setup fee, no charge to leave and take your files. There will be no asterisk after any of those words."
					bn="ন্যূনতম চুক্তি নেই, সেটআপ ফি নেই, ছেড়ে চলে যেতে বা নিজের ফাইল নিয়ে যেতে কোনো খরচ নেই। এর কোনোটির পাশেই তারকাচিহ্ন থাকবে না।"
				/>
			</li>
			<li>
				<T
					as="span"
					en="When a price changes, existing customers will be told before it takes effect, not after. Retaining a customer by making the exit confusing is not something we are willing to do."
					bn="দাম বদলালে চালু গ্রাহকদের তা কার্যকর হওয়ার আগেই জানানো হবে, পরে নয়। বেরোনোর পথ জটিল করে গ্রাহক ধরে রাখা আমরা করব না।"
				/>
			</li>
		</ul>
		<p class="small">
			<T
				as="span"
				en="Contact address, also a placeholder until the domain is registered:"
				bn="যোগাযোগের ঠিকানা — ডোমেইন নিবন্ধন না হওয়া পর্যন্ত এটিও অনুমান:"
			/>
			<span class="ph"><span class="mono">hello@alchemist.example</span></span>
		</p>
	</div>
</section>

<style>
	.top {
		padding-block: clamp(2.5rem, 7vw, 4rem) 0;
	}
	.banner {
		display: flex;
		gap: 0.9rem;
		align-items: flex-start;
		border: 1px dashed var(--brass);
		background: var(--brass-soft);
		border-radius: var(--radius);
		padding: 1.1rem 1.2rem;
		margin-top: 1.5rem;
	}
	.banner :global(svg) {
		color: var(--brass);
		flex: 0 0 auto;
		margin-top: 0.2rem;
	}
	.banner p {
		margin: 0;
		font-size: 0.94rem;
	}
	.mlabel {
		font-family: var(--mono);
		font-size: 0.7rem;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	.meter h2 {
		font-size: 1.25rem;
		margin-bottom: 0.8rem;
	}
	.price {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 0.35rem;
		margin-bottom: 0.4rem;
		font-size: 1.3rem;
	}
	.price.second {
		margin-top: 0.5rem;
	}
	.price .unit {
		font-size: 0.85rem;
		color: var(--muted);
	}
	.meter > p:last-child {
		color: var(--muted);
		font-size: 0.92rem;
		margin: 0.9rem 0 0;
	}
	.nots {
		margin-top: 2.5rem;
	}
	.nots h2 {
		font-size: 1.25rem;
	}
	.nots ul {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 0.6rem;
	}
	.nots li {
		display: flex;
		gap: 0.6rem;
		align-items: flex-start;
		color: var(--muted);
		font-size: 0.94rem;
	}
	.nots :global(svg) {
		color: var(--save);
		flex: 0 0 auto;
		margin-top: 0.3rem;
	}
	.alt {
		background: var(--raise);
		border-block: 1px solid var(--line);
	}
	.narrow {
		max-width: 760px;
	}
	.narrow > p {
		color: var(--muted);
	}
	.small {
		font-size: 0.88rem;
		color: var(--muted);
	}
	.plain {
		list-style: none;
		padding: 0;
		margin: 1.2rem 0 1.5rem;
		display: grid;
		gap: 1rem;
	}
	.plain li {
		padding-inline-start: 1rem;
		border-inline-start: 2px solid var(--line);
		color: var(--muted);
		font-size: 0.95rem;
	}
</style>
