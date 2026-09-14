<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="Alchemist pricing"
	description="You pay for the video you send us, the video we hold, and the video people watch. No per-person fee, no minimum, no charge to leave. The figures shown are indicative examples — no price has been set — and what it costs us to run is published beside them."
/>

<section class="shell hero">
	<div class="flow">
		<h1><T as="span" en="Three things you pay for, and nothing else." bn="আপনি টাকা দেন তিনটি জিনিসের জন্য, আর কিছুর নয়।" /></h1>
		<p class="lede quiet">
			<T
				as="span"
				en="You pay for the minutes of video you send us, the gigabytes we hold, and the gigabytes people watch. Everything else is simply how the thing works, and is not charged for separately."
				bn="আপনি টাকা দেন আমাদের কাছে পাঠানো ভিডিওর মিনিটের জন্য, আমরা যত গিগাবাইট রাখি তার জন্য, আর মানুষ যত গিগাবাইট দেখে তার জন্য। বাকি সবকিছু নিছক এর কাজ করার ধরন, আলাদা করে তার দাম নেই।"
			/>
		</p>
	</div>
	<div class="notice">
		<p class="ntag"><T as="span" en="Indicative figures" bn="সম্ভাব্য হিসাব" /></p>
		<p class="small">
			<T
				as="span"
				en="Alchemist has not launched and no price has been set. The numbers here are worked examples, sized against what it costs us to run — shown further down — so you can see the real shape rather than blanks. Treat them as an illustration, not a quote."
				bn="Alchemist এখনও চালু হয়নি, কোনো দামও নির্ধারিত হয়নি। এখানকার সংখ্যাগুলো উদাহরণ হিসেবে কষা — আমাদের চালাতে কত খরচ হয় তার সঙ্গে মিলিয়ে বসানো, যা নিচে দেখানো আছে — যাতে ফাঁকা ঘরের বদলে আসল আকারটা বুঝতে পারেন। এগুলোকে দৃষ্টান্ত ধরুন, দর নয়।"
			/>
		</p>
	</div>
</section>

<section class="shell bay">
	<div class="meters">
		<article class="flow">
			<h2><T as="span" en="Preparing your video" bn="ভিডিও তৈরি করা" /></h2>
			<div class="indic">
				<span class="indic__tag"><T as="span" en="Indicative" bn="সম্ভাব্য" /></span>
				<span class="indic__value">৳2.00</span>
				<span class="indic__unit">
					<T as="span" en="for each minute of video you send · about $0.016" bn="আপনার পাঠানো ভিডিওর প্রতি মিনিটে · প্রায় $০.০১৬" />
				</span>
			</div>
			<p class="small">
				<T
					as="span"
					en="Charged once, on the length of the video you send us. We make several sizes of it so it plays on any connection, and you are not charged for each one — a one-hour lecture is one hour."
					bn="একবারই নেওয়া হয়, আপনার পাঠানো ভিডিওর দৈর্ঘ্যের উপর। যেকোনো সংযোগে যাতে চলে, সেজন্য আমরা কয়েকটি আকার বানাই — প্রতিটির জন্য আলাদা টাকা নেওয়া হয় না। এক ঘণ্টার লেকচার মানে এক ঘণ্টাই।"
				/>
			</p>
		</article>

		<article class="flow">
			<h2><T as="span" en="Keeping it" bn="রেখে দেওয়া" /></h2>
			<div class="indic">
				<span class="indic__tag"><T as="span" en="Indicative" bn="সম্ভাব্য" /></span>
				<span class="indic__value">৳1.20</span>
				<span class="indic__unit">
					<T as="span" en="per GB per month · about $0.010" bn="প্রতি জিবি প্রতি মাসে · প্রায় $০.০১০" />
				</span>
			</div>
			<p class="small">
				<T
					as="span"
					en="Counted on what we actually hold, which is less than you would guess. We do not keep your original file unless you ask us to, and the larger sizes only exist for videos somebody has watched."
					bn="আমরা সত্যিই যতটা রাখি তার উপর হিসাব — আর সেটা আপনার ধারণার চেয়ে কম। আপনি না বললে আপনার মূল ফাইলটা আমরা রাখি না, আর বড় আকারগুলো কেবল সেই ভিডিওর জন্যই থাকে যা কেউ দেখেছেন।"
				/>
			</p>
		</article>

		<article class="flow">
			<h2><T as="span" en="Sending it to viewers" bn="দর্শকের কাছে পাঠানো" /></h2>
			<div class="indic">
				<span class="indic__tag"><T as="span" en="Indicative" bn="সম্ভাব্য" /></span>
				<span class="indic__value">৳0.35</span>
				<span class="indic__unit">
					<T as="span" en="for each GB watched inside Bangladesh · about $0.003" bn="বাংলাদেশের ভেতরে দেখা প্রতি জিবিতে · প্রায় $০.০০৩" />
				</span>
				<span class="indic__value second">৳1.20</span>
				<span class="indic__unit">
					<T as="span" en="for each GB watched anywhere else · about $0.010" bn="অন্য যেকোনো জায়গায় দেখা প্রতি জিবিতে · প্রায় $০.০১০" />
				</span>
			</div>
			<p class="small">
				<T
					as="span"
					en="Two rates because they cost us two different amounts. Reaching someone in Bangladesh is far cheaper than reaching someone abroad, and charging one flat rate would mean your Bangladeshi viewers quietly paying for the difference."
					bn="দুটি আলাদা হার, কারণ আমাদের খরচও দুই রকম। বাংলাদেশের কারও কাছে পৌঁছানো বিদেশের কারও কাছে পৌঁছানোর চেয়ে অনেক সস্তা — একটাই হার রাখলে সেই ফারাকের টাকাটা চুপচাপ আপনার দেশি দর্শকরাই দিতেন।"
				/>
			</p>
		</article>
	</div>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2>
				<T
					as="span"
					en="What a mid-sized course platform would pay."
					bn="মাঝারি আকারের একটি কোর্স প্ল্যাটফর্মের কত পড়ত।"
				/>
			</h2>
			<p class="quiet">
				<T
					as="span"
					en="Twenty hours of new lecture a month, five hundred gigabytes of videos held, and two hundred and fifty gigabytes watched by students in Bangladesh. That is the three rates above multiplied out, and nothing else — there is no monthly platform fee underneath."
					bn="মাসে বিশ ঘণ্টা নতুন লেকচার, পাঁচশ গিগাবাইট ভিডিও রাখা, আর বাংলাদেশে শিক্ষার্থীদের দেখা আড়াইশ গিগাবাইট। এটা উপরের তিনটি হার গুণ করেই পাওয়া, আর কিছু নয় — নিচে কোনো মাসিক প্ল্যাটফর্ম ফি নেই।"
				/>
			</p>
			<p class="fine">
				<T
					as="span"
					en="An illustration at the indicative rates above, not a quotation. Your own numbers depend on how much your library is actually watched."
					bn="উপরের সম্ভাব্য হারে কষা একটি দৃষ্টান্ত, কোনো দর নয়। আপনার নিজের হিসাব নির্ভর করবে আপনার লাইব্রেরি আসলে কতটা দেখা হয় তার উপর।"
				/>
			</p>
		</div>

		<div class="indic bill">
			<span class="indic__tag"><T as="span" en="Indicative" bn="সম্ভাব্য" /></span>
			<table>
				<caption class="vh"><T en="Worked monthly example" bn="মাসিক উদাহরণের হিসাব" /></caption>
				<tbody>
					<tr>
						<th scope="row"><T as="span" en="Preparing 20 hours of video" bn="২০ ঘণ্টা ভিডিও তৈরি" /></th>
						<td class="figure">৳2,400</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="Keeping 500 GB" bn="৫০০ জিবি রেখে দেওয়া" /></th>
						<td class="figure">৳600</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="Sending 250 GB inside Bangladesh" bn="দেশের ভেতরে ২৫০ জিবি পাঠানো" /></th>
						<td class="figure">৳88</td>
					</tr>
				</tbody>
				<tfoot>
					<tr>
						<th scope="row"><T as="span" en="A month" bn="মাসে" /></th>
						<td class="figure total">৳3,088</td>
					</tr>
				</tfoot>
			</table>
			<p class="fine conv">
				<T as="span" en="About $25. Conversions on this page are approximate." bn="প্রায় $২৫। এই পাতার মুদ্রা রূপান্তরগুলো আনুমানিক।" />
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="wide"><T as="span" en="Things you will never see on an invoice." bn="যেগুলো কখনও বিলে দেখবেন না।" /></h2>
	<ul class="never">
		<li><T as="span" en="Seats. Add as many teammates as you like." bn="সিট। যত খুশি সহকর্মী যোগ করুন।" /></li>
		<li><T as="span" en="The several sizes we make of each video so it plays everywhere." bn="সব জায়গায় যাতে চলে, সেজন্য প্রতিটি ভিডিওর যে কয়েকটি আকার বানানো হয়।" /></li>
		<li><T as="span" en="Asking us questions — how many videos you have, how much you have used." bn="আমাদের কাছে জানতে চাওয়া — কতগুলো ভিডিও আছে, কতটা ব্যবহার হয়েছে।" /></li>
		<li><T as="span" en="Us telling your system when a video is ready." bn="ভিডিও তৈরি হলে আপনার সিস্টেমকে জানানো।" /></li>
		<li><T as="span" en="Scrambling, expiring links and preview images. They are the product, not an upgrade." bn="এলোমেলো করে রাখা, মেয়াদওয়ালা লিংক আর প্রিভিউ ছবি। এগুলোই পণ্য, আলাদা কোনো আপগ্রেড নয়।" /></li>
		<li><T as="span" en="Leaving, and taking your files with you." bn="চলে যাওয়া, আর নিজের ফাইল সঙ্গে নিয়ে যাওয়া।" /></li>
	</ul>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="What this costs us to run." bn="এটা চালাতে আমাদের কত খরচ।" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Showing you what it costs us is more use than hiding behind a price list. These figures are our own planning numbers, in euros, worked out from what the machines cost to rent and how much video one gets through. They are not revenue, because there is none yet."
					bn="দামের তালিকার আড়ালে লুকানোর চেয়ে আমাদের কত খরচ হয় সেটা দেখানোই আপনার বেশি কাজে লাগবে। এই সংখ্যাগুলো আমাদের নিজেদের পরিকল্পনার হিসাব, ইউরোতে — মেশিন ভাড়ায় কত পড়ে আর একটা মেশিন কতটা ভিডিও সামলাতে পারে, তা থেকে কষা। এগুলো আয় নয়, কারণ আয় বলে এখনও কিছু নেই।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="The gap between that and the rates above is margin, and it stays visible. A number you can check is worth more than one you have to take on faith."
					bn="ওটার সঙ্গে উপরের হারগুলোর যে ফারাক, সেটাই মুনাফা — আর সেটা চোখের সামনেই থাকবে। বিশ্বাস করে মেনে নিতে হয় এমন সংখ্যার চেয়ে যাচাই করা যায় এমন সংখ্যার দাম বেশি।"
				/>
			</p>
		</div>
		<div class="measure">
			<div class="scroll-x">
				<table>
					<caption><T en="What our servers cost us, per hour of video prepared" bn="প্রতি ঘণ্টা ভিডিও তৈরিতে আমাদের সার্ভারের খরচ" /></caption>
					<thead>
						<tr>
							<th scope="col"><T en="Volume" bn="পরিমাণ" /></th>
							<th scope="col"><T en="Servers" bn="সার্ভার" /></th>
							<th scope="col"><T en="Per hour" bn="ঘণ্টাপ্রতি" /></th>
						</tr>
					</thead>
					<tbody>
						<tr>
							<th scope="row"><T as="span" en="2,000 hours of video a month" bn="মাসে ২,০০০ ঘণ্টা ভিডিও" /></th>
							<td><T as="span" en="3 machines, €350–500 a month" bn="৩টি মেশিন, মাসে €৩৫০–৫০০" /></td>
							<td class="figure">€0.20</td>
						</tr>
						<tr>
							<th scope="row"><T as="span" en="100,000 hours of video a month" bn="মাসে ১,০০,০০০ ঘণ্টা ভিডিও" /></th>
							<td><T as="span" en="About €13,630 a month" bn="মাসে প্রায় €১৩,৬৩০" /></td>
							<td class="figure">€0.14</td>
						</tr>
					</tbody>
				</table>
			</div>
			<p class="fine">
				<T
					as="span"
					en="At roughly ৳135 to the euro — worth checking the rate yourself — €0.14 is about ৳19 of server cost per hour of video. The time it takes to build and run the thing is not in these numbers, and it is the larger cost of the two."
					bn="প্রতি ইউরো মোটামুটি ১৩৫ টাকা ধরে — রেটটা নিজে যাচাই করে নেওয়া ভালো — €০.১৪ মানে প্রতি ঘণ্টা ভিডিওতে প্রায় ১৯ টাকার সার্ভার খরচ। এটা বানাতে আর চালাতে যত সময় যায় সেটা এই হিসাবে নেই, অথচ দুটোর মধ্যে বড় খরচ সেটাই।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<div class="flow narrow">
		<h2><T as="span" en="In plain words." bn="সহজ কথায়।" /></h2>
		<dl class="plainly">
			<dt><T as="span" en="Taka for Bangladeshi customers" bn="বাংলাদেশি গ্রাহকের জন্য টাকায়" /></dt>
			<dd>
				<T
					as="span"
					en="Collected through bKash, Nagad or card — not a foreign-currency charge with a bank's markup stacked on top. USD for everyone else."
					bn="আদায় হবে বিকাশ, নগদ বা কার্ডে — ব্যাংকের বাড়তি চার্জ চাপানো বিদেশি মুদ্রার বিল নয়। বাকি সবার জন্য ডলারে।"
				/>
			</dd>
			<dt><T as="span" en="You can check the numbers yourself" bn="সংখ্যাগুলো আপনি নিজেই যাচাই করতে পারবেন" /></dt>
			<dd>
				<T
					as="span"
					en="Your usage is counted from the first day, and you can read the same numbers the bill is built from at any time. If our bill and your own figures disagree, you should be able to prove it."
					bn="প্রথম দিন থেকেই আপনার ব্যবহারের হিসাব রাখা হয়, আর বিল যে সংখ্যাগুলোর উপর তৈরি সেই একই সংখ্যা আপনি যেকোনো সময় দেখতে পারেন। আমাদের বিল আর আপনার হিসাব না মিললে আপনি তা প্রমাণ করতে পারবেন।"
				/>
			</dd>
			<dt><T as="span" en="No minimum, no setup fee, no exit fee" bn="ন্যূনতম নেই, সেটআপ ফি নেই, বেরোনোর ফি নেই" /></dt>
			<dd>
				<T
					as="span"
					en="There will be no asterisk after any of those words."
					bn="এর কোনোটির পাশেই তারকাচিহ্ন থাকবে না।"
				/>
			</dd>
			<dt><T as="span" en="Price changes arrive before they take effect" bn="দামের বদল কার্যকর হওয়ার আগেই জানানো হবে" /></dt>
			<dd>
				<T
					as="span"
					en="Not after. Keeping a customer by making the way out confusing is not something we are willing to do."
					bn="পরে নয়। বেরোনোর পথটা জটিল করে গ্রাহক ধরে রাখা আমরা করব না।"
				/>
			</dd>
		</dl>
	</div>
</section>

<style>
	.hero {
		padding-block: clamp(2.5rem, 1.5rem + 4vw, 4.5rem) 0;
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
	.indic__value.second {
		display: block;
		margin-top: 0.7rem;
	}

	.measure {
		border-top: 1px solid var(--rule-firm);
		padding-top: 1.2rem;
	}
	.measure .fine {
		margin-top: 1rem;
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
