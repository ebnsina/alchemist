<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="About Alchemist — what this project believes"
	description="Transparent pricing, no manipulative growth tactics, no dark patterns, and no claims we have not measured. A short statement of how this project intends to behave."
/>

<section class="top">
	<div class="wrap narrow">
		<p class="eyebrow"><T en="About" bn="পরিচিতি" /></p>
		<h1><T as="span" en="Software that behaves itself." bn="যে সফটওয়্যার ভদ্রভাবে চলে।" /></h1>
		<p class="lede">
			<T
				as="span"
				en="Alchemy is al-kimiya — you put in something base and take out something refined. That is what a transcoder does, and it is a better name than another word ending in -ly."
				bn="Alchemy শব্দটি এসেছে আরবি আল-কিমিয়া থেকে — সাধারণ কিছু দিয়ে পরিশোধিত কিছু বের করে আনা। ট্রান্সকোডার ঠিক এই কাজটাই করে, আর -ly দিয়ে শেষ হওয়া আরেকটা নামের চেয়ে এটা ভালো।"
			/>
		</p>
	</div>
</section>

<section>
	<div class="wrap narrow">
		<h2><T as="span" en="What we hold ourselves to" bn="আমরা নিজেদের যেসব নিয়মে বাঁধি" /></h2>
		<dl class="beliefs">
			<div>
				<dt><T as="span" en="Say the price. All of it." bn="দামটা বলুন। পুরোটাই।" /></dt>
				<dd>
					<T
						as="span"
						en="Every meter is published, the cost model behind it is published, and there is no line item you discover on an invoice. Where we have not set a price yet, the page says so in plain words rather than saying 'contact sales'."
						bn="প্রতিটি মিটার প্রকাশ্য, তার পেছনের খরচের হিসাবও প্রকাশ্য, আর বিলে এমন কোনো খাত থাকবে না যা আপনি পরে আবিষ্কার করবেন। যেখানে এখনও দাম ঠিক হয়নি, সেখানে ‘সেলসে যোগাযোগ করুন’ না লিখে সোজাসুজি সেটাই লেখা থাকে।"
					/>
				</dd>
			</div>
			<div>
				<dt><T as="span" en="No manufactured urgency." bn="বানানো তাড়াহুড়ো নয়।" /></dt>
				<dd>
					<T
						as="span"
						en="No countdown timers, no 'only three spots left', no pre-ticked boxes, no trial that quietly becomes a subscription. If a decision is worth making it is worth making slowly, and a product that needs a timer to close a deal is telling you something."
						bn="কোনো কাউন্টডাউন ঘড়ি নেই, ‘আর মাত্র তিনটি জায়গা বাকি’ নেই, আগে থেকে টিক দেওয়া ঘর নেই, আর এমন ট্রায়াল নেই যা চুপচাপ সাবস্ক্রিপশন হয়ে যায়। যে সিদ্ধান্ত নেওয়ার মতো, সেটা ধীরেসুস্থে নেওয়ারই মতো — আর যে পণ্যের চুক্তি করাতে ঘড়ি লাগে, সেটা নিজেই একটা কথা বলে দেয়।"
					/>
				</dd>
			</div>
			<div>
				<dt><T as="span" en="No borrowed credibility." bn="ধার করা বিশ্বাসযোগ্যতা নয়।" /></dt>
				<dd>
					<T
						as="span"
						en="You will not find customer logos, testimonials, case studies, uptime percentages or compliance badges on this site, because we do not have any of them yet. When we do, they will name real people who agreed to be named."
						bn="এই সাইটে গ্রাহকের লোগো, প্রশংসাবাক্য, কেস স্টাডি, আপটাইমের শতাংশ বা কমপ্লায়েন্স ব্যাজ পাবেন না, কারণ এগুলোর কোনোটাই আমাদের এখনও নেই। যখন থাকবে, তখন সেখানে সত্যিকারের মানুষের নাম থাকবে — যাঁরা নাম দিতে রাজি হয়েছেন।"
					/>
				</dd>
			</div>
			<div>
				<dt><T as="span" en="Measured, or marked." bn="হয় মাপা, নয়তো চিহ্নিত।" /></dt>
				<dd>
					<T
						as="span"
						en="A number on this site is either something we measured, something published by someone who measured it and is cited, or something labelled as a projection or a placeholder. There is no fourth category."
						bn="এই সাইটের কোনো সংখ্যা হয় আমাদের মাপা, নয়তো এমন কারও প্রকাশ করা যিনি মেপেছেন এবং যাঁর নাম উল্লেখ করা আছে, নয়তো পরিষ্কারভাবে অনুমান বা প্লেসহোল্ডার বলে চিহ্নিত। চতুর্থ কোনো ধরন নেই।"
					/>
				</dd>
			</div>
			<div>
				<dt><T as="span" en="The door out stays open." bn="বেরোনোর দরজা খোলা থাকে।" /></dt>
				<dd>
					<T
						as="span"
						en="Your files are yours, the API is the whole product with no privileged internal surface we keep for ourselves, and leaving costs nothing. Lock-in built by making export difficult is a tax on customers who stopped being happy, and we would rather lose them honestly."
						bn="আপনার ফাইল আপনারই, API-ই পুরো পণ্য — নিজেদের জন্য আলাদা করে রাখা গোপন কোনো সুবিধা নেই — আর চলে যেতে কোনো খরচ নেই। রপ্তানি কঠিন করে আটকে রাখা আসলে অসন্তুষ্ট গ্রাহকের উপর বসানো কর; তার চেয়ে সৎভাবে তাঁদের হারানোই ভালো।"
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
	</div>
</section>

<section class="alt">
	<div class="wrap narrow">
		<h2><T as="span" en="Where it stands" bn="এখন কোথায় দাঁড়িয়ে" /></h2>
		<p>
			<T
				as="span"
				en="Alchemist is self-hosted on bare metal — Go, PostgreSQL, SeaweedFS, ffmpeg and shaka-packager — with no managed transcoding service underneath it. That is a deliberate choice about cost and control, and it is also what makes hosting inside BDIX possible at all."
				bn="Alchemist চলে নিজেদের হার্ডওয়্যারে — Go, PostgreSQL, SeaweedFS, ffmpeg আর shaka-packager দিয়ে — নিচে ভাড়া করা কোনো ট্রান্সকোডিং সার্ভিস নেই। এটা খরচ আর নিয়ন্ত্রণ নিয়ে ভেবেচিন্তে নেওয়া সিদ্ধান্ত, আর BDIX-এর ভেতরে হোস্ট করা সম্ভব হচ্ছে এর কারণেই।"
			/>
		</p>
		<p>
			<T
				as="span"
				en="It is not generally available. The encode, storage, encryption and playback path works; the BDIX edge that the whole argument rests on is the next phase of work and is not running yet. Saying that clearly on a marketing site is uncomfortable, which is roughly the point."
				bn="এটি এখনও সবার জন্য উন্মুক্ত নয়। এনকোড, স্টোরেজ, এনক্রিপশন আর প্লেব্যাকের পথটা কাজ করে; কিন্তু পুরো যুক্তিটা যার উপর দাঁড়িয়ে, সেই BDIX এজ পরের ধাপের কাজ — এখনও চালু হয়নি। একটা মার্কেটিং সাইটে এ কথা স্পষ্ট করে লেখা অস্বস্তিকর, আর মোটামুটি সেটাই এখানে বলার বিষয়।"
			/>
		</p>
	</div>
</section>

<style>
	.top {
		padding-block: clamp(2.5rem, 7vw, 4.5rem) clamp(1rem, 3vw, 2rem);
		background: radial-gradient(70ch 34ch at 10% -20%, var(--brass-soft), transparent 70%);
	}
	.narrow {
		max-width: 720px;
	}
	.alt {
		background: var(--raise);
		border-block: 1px solid var(--line);
	}
	.alt p {
		color: var(--muted);
	}
	.beliefs {
		margin: 1.5rem 0 0;
		display: grid;
		gap: 1.6rem;
	}
	dt {
		font-weight: 600;
		font-size: 1.08rem;
		margin-bottom: 0.3rem;
	}
	dd {
		margin: 0;
		color: var(--muted);
		font-size: 0.95rem;
	}
</style>
