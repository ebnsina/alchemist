<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="About Alchemist — what this project believes"
	description="Transparent pricing, no manipulative growth tactics, no dark patterns, and no claims we have not measured. A short statement of how this project intends to behave."
/>

<section class="shell hero">
	<div class="flow">
		<h1><T as="span" en="Software that behaves itself." bn="যে সফটওয়্যার ভদ্রভাবে চলে।" /></h1>
		<p class="lede quiet">
			<T
				as="span"
				en="Alchemy is al-kimiya — you put in something base and take out something refined. That is roughly what this does to a video file, and it is a better name than another word ending in -ly."
				bn="Alchemy শব্দটি এসেছে আরবি আল-কিমিয়া থেকে — সাধারণ কিছু দিয়ে পরিশোধিত কিছু বের করে আনা। একটা ভিডিও ফাইলের সঙ্গে এখানে মোটামুটি সেটাই করা হয়, আর -ly দিয়ে শেষ হওয়া আরেকটা নামের চেয়ে এটা ভালো।"
			/>
		</p>
	</div>
</section>

<section class="shell bay">
	<dl class="beliefs">
		<div>
			<dt><T as="span" en="Say the price. All of it." bn="দামটা বলুন। পুরোটাই।" /></dt>
			<dd>
				<T
					as="span"
					en="Every meter is published, the cost model behind it is published, and there is no line item you discover on an invoice. Where a price has not been set, the page says so plainly instead of saying contact sales."
					bn="প্রতিটি মিটার প্রকাশ্য, তার পেছনের খরচের হিসাবও প্রকাশ্য, আর বিলে এমন কোনো খাত থাকবে না যা আপনি পরে আবিষ্কার করবেন। যেখানে দাম ঠিক হয়নি, সেখানে ‘সেলসে যোগাযোগ করুন’ না লিখে সোজাসুজি সেটাই লেখা থাকে।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="No manufactured urgency." bn="বানানো তাড়াহুড়ো নয়।" /></dt>
			<dd>
				<T
					as="span"
					en="No countdown timers, no last three places, no pre-ticked boxes, no trial that quietly becomes a subscription. A product that needs a clock to close a deal is telling you something."
					bn="কোনো কাউন্টডাউন ঘড়ি নেই, ‘আর মাত্র তিনটি জায়গা বাকি’ নেই, আগে থেকে টিক দেওয়া ঘর নেই, আর এমন ট্রায়াল নেই যা চুপচাপ সাবস্ক্রিপশন হয়ে যায়। যে পণ্যের চুক্তি করাতে ঘড়ি লাগে, সেটা নিজেই একটা কথা বলে দেয়।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="No borrowed credibility." bn="ধার করা বিশ্বাসযোগ্যতা নয়।" /></dt>
			<dd>
				<T
					as="span"
					en="You will not find customer logos, testimonials, case studies, uptime percentages or compliance badges here, because we do not have any of them yet. When we do, they will name real people who agreed to be named."
					bn="এখানে গ্রাহকের লোগো, প্রশংসাবাক্য, কেস স্টাডি, আপটাইমের শতাংশ বা কমপ্লায়েন্স ব্যাজ পাবেন না, কারণ এগুলোর কোনোটাই আমাদের এখনও নেই। যখন থাকবে, তখন সেখানে সত্যিকারের মানুষের নাম থাকবে — যাঁরা নাম দিতে রাজি হয়েছেন।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Measured, or marked." bn="হয় মাপা, নয়তো চিহ্নিত।" /></dt>
			<dd>
				<T
					as="span"
					en="A number here is either something we measured, something published by whoever measured it and cited to them, or something labelled as a projection or an indicative example. There is no fourth category."
					bn="এখানকার কোনো সংখ্যা হয় আমাদের মাপা, নয়তো এমন কারও প্রকাশ করা যিনি মেপেছেন এবং যাঁর নাম উল্লেখ করা আছে, নয়তো পরিষ্কারভাবে অনুমান বা সম্ভাব্য উদাহরণ বলে চিহ্নিত। চতুর্থ কোনো ধরন নেই।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="The door out stays open." bn="বেরোনোর দরজা খোলা থাকে।" /></dt>
			<dd>
				<T
					as="span"
					en="Your files are yours, the API is the whole product with no privileged internal surface kept back for ourselves, and leaving costs nothing. Lock-in built by making export difficult is a tax on customers who stopped being happy."
					bn="আপনার ফাইল আপনারই, API-ই পুরো পণ্য — নিজেদের জন্য আলাদা করে রাখা গোপন কোনো সুবিধা নেই — আর চলে যেতে কোনো খরচ নেই। রপ্তানি কঠিন করে আটকে রাখা আসলে অসন্তুষ্ট গ্রাহকের উপর বসানো কর।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Bangla is not a translation layer." bn="বাংলা কোনো অনুবাদের আস্তরণ নয়।" /></dt>
			<dd>
				<T
					as="span"
					en="Error messages, player chrome and subtitles are written in Bangla, not machine-translated into it, and the text is checked for correct conjunct shaping rather than assumed to work. This page is bilingual for the same reason."
					bn="এরর বার্তা, প্লেয়ারের লেখা আর সাবটাইটেল বাংলায় লেখা হয় — যন্ত্রে অনুবাদ করে বসিয়ে দেওয়া হয় না — আর যুক্তাক্ষর ঠিকভাবে বসছে কিনা তা ধরে না নিয়ে দেখে নেওয়া হয়। এই পাতাটি দুই ভাষায়, সেই একই কারণে।"
				/>
			</dd>
		</div>
	</dl>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="Honest about the unfinished parts." bn="যেটুকু শেষ হয়নি, তা নিয়ে সোজাসুজি।" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Alchemist runs on machines we rent and look after ourselves, with no other company’s video service underneath it. That is a deliberate choice about cost and control, and it is also the only reason keeping your videos inside Bangladesh is possible at all."
					bn="Alchemist চলে এমন মেশিনে, যেগুলো আমরা নিজেরা ভাড়া নিই আর নিজেরাই দেখাশোনা করি — নিচে অন্য কোনো কোম্পানির ভিডিও সার্ভিস নেই। এটা খরচ আর নিয়ন্ত্রণ নিয়ে ভেবেচিন্তে নেওয়া সিদ্ধান্ত, আর আপনার ভিডিও বাংলাদেশের ভেতরে রাখা সম্ভব হচ্ছে কেবল এই কারণেই।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="It is not open to everyone yet. Preparing, keeping, scrambling and playing a video all work; the servers inside Bangladesh that the whole argument rests on are the next piece of work and are not running. Saying that plainly on a page meant to sell something is uncomfortable, which is roughly the point."
					bn="এটি এখনও সবার জন্য খোলা নয়। ভিডিও তৈরি করা, রেখে দেওয়া, এলোমেলো করে রাখা আর চালানো — সবই কাজ করে; কিন্তু পুরো যুক্তিটা যার উপর দাঁড়িয়ে, দেশের ভেতরের সেই সার্ভারগুলো পরের ধাপের কাজ, এখনও চালু হয়নি। কিছু বিক্রি করার পাতায় এ কথা সোজাসুজি লেখা অস্বস্তিকর, আর মোটামুটি সেটাই এখানে বলার বিষয়।"
				/>
			</p>
		</div>
	</div>
</section>

<style>
	.hero {
		padding-block: clamp(2.5rem, 1.5rem + 4vw, 4.5rem) 0;
	}
	.hero h1 {
		max-width: 13ch;
	}
	.beliefs {
		margin: 0;
		display: grid;
		gap: 0 clamp(1.5rem, 4vw, 3.5rem);
		grid-template-columns: repeat(auto-fit, minmax(290px, 1fr));
	}
	.beliefs dt {
		font-weight: 600;
		font-size: var(--step-2);
		letter-spacing: -0.012em;
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		margin-bottom: 0.4rem;
	}
	.beliefs dd {
		margin: 0 0 1.8rem;
		color: var(--graphite);
		font-size: var(--step-0);
	}
</style>
