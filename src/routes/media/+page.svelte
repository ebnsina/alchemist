<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
	import Icon from '$lib/Icon.svelte';
</script>

<Seo
	title="Alchemist for news and media — live in under a minute, local when the cables cut"
	description="A news clip that is live in 40 seconds at 480p beats one that is perfect in six minutes. Alchemist encodes the low rungs first, publishes immediately, and serves from inside BDIX."
/>

<section class="top">
	<div class="wrap">
		<p class="eyebrow"><T en="For news and media" bn="সংবাদ ও মিডিয়ার জন্য" /></p>
		<h1>
			<T
				as="span"
				en="A clip live in forty seconds beats a clip that is perfect in six minutes."
				bn="চল্লিশ সেকেন্ডে লাইভ হওয়া ক্লিপ ছয় মিনিটে নিখুঁত হওয়া ক্লিপকে হারিয়ে দেয়।"
			/>
		</h1>
		<p class="lede">
			<T
				as="span"
				en="For a newsroom the encoder is not a quality tool, it is a clock. Alchemist is built so the clock stops early: the low rungs are encoded on ingest and the asset becomes playable then, rather than when the whole ladder finishes."
				bn="নিউজরুমে এনকোডার আসলে কোয়ালিটির যন্ত্র নয়, ঘড়ি। Alchemist এমনভাবে তৈরি যাতে ঘড়ি আগেই থামে: ইনজেস্টেই নিচের ধাপগুলো এনকোড হয় আর তখনই অ্যাসেট চালানোর উপযোগী হয়ে যায় — পুরো ল্যাডার শেষ হওয়ার অপেক্ষা করতে হয় না।"
			/>
		</p>
	</div>
</section>

<section class="alt">
	<div class="wrap">
		<div class="grid">
			<article class="card">
				<div class="ic"><Icon name="bolt" size={24} /></div>
				<h2><T as="span" en="Publish before the ladder finishes" bn="ল্যাডার শেষ হওয়ার আগেই প্রকাশ" /></h2>
				<p>
					<T
						as="span"
						en="Ingest encodes 144p–360p and the asset moves to partially_ready, which is a playable state, not a waiting room. The 480p and 720p rungs are generated on the first playback request — one encode no matter how many people arrive at once."
						bn="ইনজেস্টে ১৪৪পি–৩৬০পি এনকোড হয় আর অ্যাসেট চলে যায় partially_ready অবস্থায় — যেটা অপেক্ষার ঘর নয়, চালানো যায় এমন অবস্থা। ৪৮০পি আর ৭২০পি ধাপ তৈরি হয় প্রথম প্লেব্যাক রিকোয়েস্টে — একসঙ্গে যত দর্শকই আসুক, এনকোড হয় একবারই।"
					/>
				</p>
			</article>
			<article class="card">
				<div class="ic"><Icon name="cable" size={24} /></div>
				<h2><T as="span" en="When the cables cut, local keeps working" bn="ক্যাবল কাটলে লোকালটা চলতেই থাকে" /></h2>
				<p>
					<T
						as="span"
						en="Bangladesh's international connectivity depends on a small number of submarine cables and has degraded nationally more than once. A breaking story is exactly when that happens and exactly when it matters. Content served inside BDIX never leaves the country."
						bn="বাংলাদেশের আন্তর্জাতিক সংযোগ নির্ভর করে হাতেগোনা কয়েকটি সাবমেরিন ক্যাবলের উপর, আর একাধিকবার সারা দেশে তা ধীর হয়ে গেছে। বড় খবরের সময়েই এমনটা ঘটে, আর তখনই এটা সবচেয়ে বেশি গুরুত্বপূর্ণ। BDIX-এর ভেতর থেকে দেওয়া কনটেন্ট দেশের বাইরেই যায় না।"
					/>
				</p>
			</article>
			<article class="card">
				<div class="ic"><Icon name="layers" size={24} /></div>
				<h2><T as="span" en="Import what you already have" bn="যা আছে তা-ই নিয়ে আসুন" /></h2>
				<p>
					<T
						as="span"
						en="POST an asset with a URL and it pulls the file itself, so an existing library or an existing publishing workflow does not have to be rewritten to try this. Bucket sources are supported for bulk."
						bn="একটি লিংক দিয়ে অ্যাসেট POST করলে ফাইলটা নিজেই টেনে নেয় — তাই পুরোনো লাইব্রেরি বা চালু পাবলিশিং পদ্ধতি নতুন করে লেখার দরকার পড়ে না। অনেক ফাইল একসঙ্গে আনার জন্য বাকেট সোর্স আছে।"
					/>
				</p>
			</article>
			<article class="card">
				<div class="ic"><Icon name="phone" size={24} /></div>
				<h2><T as="span" en="Your audience is on a phone, on a budget" bn="আপনার দর্শক ফোনে, আর হিসেব করে চলেন" /></h2>
				<p>
					<T
						as="span"
						en="89% of BD internet users are on mobile (BTRC, June 2026). The news profile runs 240p–720p and optimises for time-to-ready rather than peak quality, because a reader on 4G in a crowd wants the clip to start, not to be beautiful."
						bn="বাংলাদেশের ৮৯% ইন্টারনেট ব্যবহারকারী মোবাইলে (বিটিআরসি, জুন ২০২৬)। নিউজ প্রোফাইল চলে ২৪০পি–৭২০পি-তে আর অগ্রাধিকার দেয় দ্রুত প্রস্তুত হওয়াকে, সর্বোচ্চ মানকে নয় — কারণ ভিড়ের মধ্যে 4G-তে থাকা পাঠক চান ক্লিপটা চালু হোক, সুন্দর হোক তা নয়।"
					/>
				</p>
			</article>
		</div>
	</div>
</section>

<section>
	<div class="wrap narrow">
		<h2><T as="span" en="What is not here yet" bn="যা এখনও নেই" /></h2>
		<p>
			<T
				as="span"
				en="Live streaming is not built. If you need live — and in Bangladesh that usually means sport — say so early, because it changes the order of the work rather than slotting into it. The packager was written as a library specifically so live is a phase and not a second product, but that is a design decision, not a shipped feature."
				bn="লাইভ স্ট্রিমিং তৈরি হয়নি। লাইভ দরকার হলে — বাংলাদেশে যার মানে সাধারণত খেলা — আগেভাগে বলবেন, কারণ এটা কাজের ক্রমেই বদল আনে, চুপচাপ কোনো ফাঁকে ঢুকে যায় না। প্যাকেজারটা লাইব্রেরি হিসেবে লেখা হয়েছে ঠিক এই কারণেই যে লাইভ যেন আলাদা প্রোডাক্ট না হয়ে একটা ধাপ হয় — তবে সেটা নকশার সিদ্ধান্ত, চালু ফিচার নয়।"
			/>
		</p>
		<p>
			<T
				as="span"
				en="BDIX peering and the Dhaka edge node are also not live yet. The origin was built to sit behind an edge cache — correct byte ranges, ETags, conditional requests and cache headers — but the edge itself is the next phase."
				bn="BDIX পিয়ারিং আর ঢাকার এজ নোডও এখনও চালু হয়নি। অরিজিন তৈরি হয়েছে একটি এজ ক্যাশের পেছনে বসার জন্য — ঠিকঠাক বাইট রেঞ্জ, ETag, কন্ডিশনাল রিকোয়েস্ট আর ক্যাশ হেডারসহ — কিন্তু এজ নোডটাই পরের ধাপ।"
			/>
		</p>
		<div class="cta">
			<a class="btn btn-primary" href="/docs/"><T en="See the API" bn="API দেখুন" /><Icon name="arrow" size={18} /></a>
			<a class="btn" href="/pricing/"><T en="What gets billed" bn="কীসের বিল হয়" /></a>
		</div>
	</div>
</section>

<style>
	.top {
		padding-block: clamp(2.5rem, 7vw, 4.5rem) clamp(1.5rem, 4vw, 2.5rem);
		background: radial-gradient(70ch 34ch at 10% -20%, var(--brass-soft), transparent 70%);
	}
	.top h1 {
		max-width: 20ch;
	}
	.alt {
		background: var(--raise);
		border-block: 1px solid var(--line);
	}
	.ic {
		color: var(--brass);
		margin-bottom: 0.6rem;
	}
	.card h2 {
		font-size: 1.15rem;
	}
	.card p {
		color: var(--muted);
		font-size: 0.94rem;
		margin-bottom: 0;
	}
	.narrow {
		max-width: 760px;
	}
	.narrow p {
		color: var(--muted);
	}
	.cta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.7rem;
		margin-top: 2rem;
	}
</style>
