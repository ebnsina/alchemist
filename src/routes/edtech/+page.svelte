<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
	import Icon from '$lib/Icon.svelte';
</script>

<Seo
	title="Alchemist for course platforms — stop the leak, cut the data bill"
	description="Per-viewer encrypted playback, a mobile-first ladder tuned for lectures and screen recordings, and delivery from inside BDIX so students' data quota survives the semester."
/>

<section class="top">
	<div class="wrap">
		<p class="eyebrow"><T en="For course platforms" bn="কোর্স প্ল্যাটফর্মের জন্য" /></p>
		<h1>
			<T
				as="span"
				en="Your two real problems are piracy and your students' data bill."
				bn="আপনার আসল সমস্যা দুটি — কনটেন্ট চুরি আর শিক্ষার্থীর ডেটার বিল।"
			/>
		</h1>
		<p class="lede">
			<T
				as="span"
				en="Everything else about hosting a lecture is solved. These two are not, and both of them are worse in Bangladesh than in the markets the big video platforms were designed for."
				bn="লেকচার হোস্ট করার বাকি সব সমস্যার সমাধান হয়ে গেছে। এই দুটোর হয়নি — আর বড় ভিডিও প্ল্যাটফর্মগুলো যেসব বাজারের কথা ভেবে তৈরি, বাংলাদেশে এই দুটো সমস্যাই সেখানকার চেয়ে বেশি তীব্র।"
			/>
		</p>
	</div>
</section>

<section class="alt">
	<div class="wrap">
		<div class="grid">
			<article class="card">
				<div class="ic"><Icon name="lock" size={24} /></div>
				<h2><T as="span" en="Course piracy" bn="কোর্স চুরি" /></h2>
				<p>
					<T
						as="span"
						en="Media is packaged with SAMPLE-AES. The content key is served from an endpoint that requires the same short-lived signature as the manifest, and that signature is issued per playback rather than per course. A link shared in a Telegram group stops working within hours, not never."
						bn="মিডিয়া প্যাকেজ করা হয় SAMPLE-AES দিয়ে। কনটেন্ট কি আসে এমন একটি এন্ডপয়েন্ট থেকে, যেটির জন্য ম্যানিফেস্টের মতো একই স্বল্পমেয়াদি সিগনেচার লাগে — আর সেই সিগনেচার দেওয়া হয় প্রতিটি প্লেব্যাকের জন্য আলাদা করে, পুরো কোর্সের জন্য একবার নয়। টেলিগ্রাম গ্রুপে শেয়ার করা লিংক কয়েক ঘণ্টার মধ্যেই অকেজো হয়ে যায় — চিরকাল চলে না।"
					/>
				</p>
				<p class="caveat">
					<T
						as="span"
						en="What it does not do: stop someone re-recording their own screen. Nothing short of studio DRM does, and studio DRM is not built here. We would rather say that than sell you a word like 'unbreakable'."
						bn="যা এটি পারে না: কেউ নিজের স্ক্রিন রেকর্ড করলে তা ঠেকাতে। স্টুডিও DRM ছাড়া সেটা কেউই পারে না, আর স্টুডিও DRM এখানে তৈরি হয়নি। ‘ভাঙা অসম্ভব’ জাতীয় কথা বলার চেয়ে এটুকু সোজাসুজি বলাই ভালো মনে করি।"
					/>
				</p>
			</article>
			<article class="card">
				<div class="ic save"><Icon name="coins" size={24} /></div>
				<h2><T as="span" en="The data bill" bn="ডেটার বিল" /></h2>
				<p>
					<T
						as="span"
						en="A lecture is a slow-moving talking head or a screen recording. It compresses far better than the film footage most default encoding ladders are tuned against, and the default ladder here targets 360p at 0.6 Mbps rather than 1080p at 3.5."
						bn="একটি লেকচার মানে ধীর গতির কথা বলা মুখ বা স্ক্রিন রেকর্ডিং। বেশিরভাগ ডিফল্ট এনকোডিং ল্যাডার যে সিনেমার ফুটেজ ধরে তৈরি, তার চেয়ে এসব কনটেন্ট অনেক ভালো কমপ্রেস হয়। এখানকার ডিফল্ট ল্যাডারের লক্ষ্য ৩৬০পি-তে ০.৬ Mbps, ১০৮০পি-তে ৩.৫ নয়।"
					/>
				</p>
				<p class="caveat">
					<T
						as="span"
						en="Our encoding notes expect lecture and screen content to land 40–60% below that table at equal quality once per-title analysis is tuned. That is a projection from the content type, not a measurement of your library — per-title analysis is not shipped yet, and we will not quote you a saving we have not measured on your files."
						bn="আমাদের এনকোডিং নোট বলছে, পার-টাইটেল বিশ্লেষণ ঠিকমতো দাঁড়ালে একই মানে লেকচার ও স্ক্রিন কনটেন্ট ওই তালিকার চেয়ে ৪০–৬০% কম বিটরেটে নামবে। এটি কনটেন্টের ধরন দেখে করা হিসাব, আপনার লাইব্রেরিতে মেপে পাওয়া ফল নয় — পার-টাইটেল বিশ্লেষণ এখনও চালু হয়নি, আর আপনার ফাইলে না মেপে আমরা কোনো সাশ্রয়ের সংখ্যা বলব না।"
					/>
				</p>
			</article>
		</div>
	</div>
</section>

<section>
	<div class="wrap narrow">
		<h2><T as="span" en="Things that matter more in a classroom than a cinema" bn="সিনেমার চেয়ে ক্লাসরুমে যেগুলো বেশি জরুরি" /></h2>
		<ul class="pts">
			<li>
				<h3><T as="span" en="An audio-only rendition" bn="শুধু অডিও রেন্ডিশন" /></h3>
				<p>
					<T
						as="span"
						en="Most of a lecture is the audio. An audio-only track costs roughly a tenth of the data of video, and for revision it is often all a student wants. On a metered connection that is the difference between finishing the course and rationing it."
						bn="লেকচারের বেশিরভাগটাই আসলে কথা। শুধু অডিও ট্র্যাকে ভিডিওর প্রায় দশ ভাগের এক ভাগ ডেটা লাগে, আর রিভিশনের সময় শিক্ষার্থীর প্রায়ই এটুকুই দরকার। মিটারড কানেকশনে এটাই ঠিক করে দেয় — কোর্সটা শেষ হবে, নাকি মেপে মেপে দেখতে হবে।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="Playable before it is finished" bn="পুরো শেষ হওয়ার আগেই চালানো যায়" /></h3>
				<p>
					<T
						as="span"
						en="An asset reaches partially_ready as soon as the low rungs exist, and that state is playable. A teacher who uploads at 11pm the night before class does not wait for the whole ladder."
						bn="নিচের ধাপগুলো তৈরি হলেই অ্যাসেট partially_ready অবস্থায় পৌঁছে যায়, আর সেই অবস্থাতেই চালানো যায়। ক্লাসের আগের রাত ১১টায় আপলোড করা শিক্ষককে পুরো ল্যাডারের জন্য বসে থাকতে হয় না।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="Bangla that renders correctly" bn="বাংলা যেন ঠিকভাবে দেখায়" /></h3>
				<p>
					<T
						as="span"
						en="Bangla titles, filenames and subtitles survive the whole chain — database, object keys, manifest metadata, player. Conjuncts shape properly instead of breaking into boxes, which is a real failure mode with a careless font stack."
						bn="বাংলা শিরোনাম, ফাইলের নাম আর সাবটাইটেল পুরো পথ ধরে অক্ষত থাকে — ডেটাবেজ, স্টোরেজ কি, ম্যানিফেস্ট মেটাডেটা, প্লেয়ার। যুক্তাক্ষর ভেঙে বাক্স হয়ে যায় না, ঠিকভাবেই বসে — অসতর্ক ফন্ট বাছাইয়ে এটা সত্যিই ঘটে।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="A Data Saver the student can see" bn="দর্শকের চোখে পড়ার মতো ডেটা সেভার" /></h3>
				<p>
					<T
						as="span"
						en="The player carries a visible Data Saver control that pins playback to 360p and shows the estimated megabytes per hour. Buried in a settings menu it would be a feature; on the bar it is a reason to trust you."
						bn="প্লেয়ারে চোখে পড়ার মতো একটি ডেটা সেভার বোতাম আছে, যেটি প্লেব্যাক ৩৬০পি-তে বেঁধে রাখে আর ঘণ্টায় আনুমানিক কত মেগাবাইট লাগবে তা দেখায়। সেটিংস মেনুতে লুকানো থাকলে এটা নিছক একটা ফিচার হতো; সামনে থাকলে এটা আপনাকে বিশ্বাস করার একটা কারণ।"
					/>
				</p>
			</li>
		</ul>
		<div class="note">
			<p>
				<T
					as="span"
					en="One more thing worth knowing before you plan a term around it: BDIX peering and the Dhaka edge node are the next phase of work and are not live. The latency and FUP advantages described across this site are properties of BDIX delivery, which is what the system is being built to do — not a measurement of Alchemist serving your students today."
					bn="একটা সেমিস্টারের পরিকল্পনা করার আগে আরেকটি কথা জেনে রাখা দরকার: BDIX পিয়ারিং আর ঢাকার এজ নোড পরের ধাপের কাজ, এখনও চালু হয়নি। এই সাইটে বলা লেটেন্সি আর FUP-এর সুবিধাগুলো BDIX ডেলিভারির বৈশিষ্ট্য — সিস্টেমটা সেভাবেই তৈরি হচ্ছে — কিন্তু আজ Alchemist আপনার শিক্ষার্থীদের সেবা দিচ্ছে, এমন কোনো মাপা ফল নয়।"
				/>
			</p>
		</div>
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
	.ic.save {
		color: var(--save);
	}
	.card h2 {
		font-size: 1.2rem;
	}
	.card p {
		color: var(--muted);
		font-size: 0.94rem;
	}
	.caveat {
		border-top: 1px dashed var(--line);
		padding-top: 0.9rem;
		margin-bottom: 0;
		font-size: 0.88rem !important;
	}
	.narrow {
		max-width: 760px;
	}
	.pts {
		list-style: none;
		padding: 0;
		margin: 1.5rem 0 2rem;
		display: grid;
		gap: 1.5rem;
	}
	.pts h3 {
		margin-bottom: 0.3rem;
	}
	.pts p {
		color: var(--muted);
		font-size: 0.94rem;
		margin: 0;
	}
	.cta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.7rem;
		margin-top: 2rem;
	}
</style>
