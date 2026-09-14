<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="Alchemist — video hosting for Bangladesh"
	description="Put your videos online so they start playing straight away in Bangladesh and use as little of your viewers' mobile data as possible. Upload a file or send a link; get back a link that plays."
/>

<section class="shell open">
	<h1>
		<T
			as="span"
			en="Video that plays straight away in Bangladesh, without using up your viewers' data."
			bn="যে ভিডিও বাংলাদেশে সঙ্গে সঙ্গে চলতে শুরু করে, আর দর্শকের ডেটা শেষ করে না।"
		/>
	</h1>
	<div class="flow lead">
		<p class="lede quiet">
			<T
				as="span"
				en="Alchemist keeps your videos and sends them to the people watching. You upload a file, or give us a link and we fetch it; you get back a link to put in your app or your website, and it plays."
				bn="Alchemist আপনার ভিডিও রেখে দেয় আর দর্শকের কাছে পৌঁছে দেয়। আপনি ফাইল আপলোড করেন, বা একটা লিংক দেন আর আমরা সেখান থেকে নিয়ে নিই; বদলে পান এমন একটা লিংক, যেটা আপনার অ্যাপ বা ওয়েবসাইটে বসিয়ে দিলেই ভিডিও চলে।"
			/>
		</p>
		<p class="quiet">
			<T
				as="span"
				en="It is built for Bangladesh in particular. The videos sit on servers inside the country, so they start without the wait, and they are made small enough that watching on mobile data does not cost your viewers a fortune."
				bn="এটি বিশেষভাবে বাংলাদেশের জন্য বানানো। ভিডিওগুলো থাকে দেশের ভেতরের সার্ভারে, তাই অপেক্ষা ছাড়াই চালু হয়; আর আকারে এত ছোট রাখা হয় যে মোবাইল ডেটায় দেখতে গিয়ে দর্শকের অনেক টাকা খরচ হয় না।"
			/>
		</p>
		<div class="row">
			<a class="btn btn--solid" href="/pricing/"><T en="What it costs" bn="খরচ কত" /></a>
			<a class="btn" href="/docs/"><T en="For developers" bn="ডেভেলপারদের জন্য" /></a>
		</div>
	</div>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="It starts, instead of buffering" bn="বাফার নয়, সঙ্গে সঙ্গে চালু" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Your videos are kept on servers inside BDIX, the exchange Bangladeshi internet providers use to pass traffic to each other. From there a video is a few thousandths of a second away from the person watching, instead of the fifth of a second it takes to reach Europe or the United States. That gap is the difference between tapping play and waiting for it."
					bn="আপনার ভিডিও রাখা হয় BDIX-এর ভেতরের সার্ভারে — বাংলাদেশের ইন্টারনেট সরবরাহকারীরা নিজেদের মধ্যে ট্রাফিক আদান-প্রদান করতে যে এক্সচেঞ্জ ব্যবহার করেন। সেখান থেকে ভিডিও দর্শকের কাছ থেকে কয়েক হাজার ভাগের এক সেকেন্ড দূরে, ইউরোপ বা আমেরিকা পর্যন্ত যেতে যেখানে লাগে সেকেন্ডের পাঁচ ভাগের এক ভাগ। এই ফারাকটাই ঠিক করে দেয় — প্লে চাপলে চলবে, নাকি অপেক্ষা করতে হবে।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="On most home broadband packages in Bangladesh, anything coming from inside BDIX also does not count against the data allowance — what providers call FUP. Vimeo, Bunny, Mux and Cloudflare all keep their video outside the country, so none of them can offer that here."
					bn="বাংলাদেশের বেশিরভাগ বাসাবাড়ির ব্রডব্যান্ড প্যাকেজে আবার BDIX-এর ভেতর থেকে আসা কিছু ডেটা লিমিটের হিসাবেই ধরা হয় না — সরবরাহকারীরা যেটাকে FUP বলেন। Vimeo, Bunny, Mux আর Cloudflare সবাই ভিডিও রাখে দেশের বাইরে, তাই এখানে এই সুবিধাটা তাদের কেউ দিতে পারে না।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="Bangladesh's link to the rest of the world also runs through a handful of undersea cables, and when one is cut the whole country slows down. Video held inside the country keeps playing through it."
					bn="দেশের বাইরের দুনিয়ার সঙ্গে বাংলাদেশের সংযোগ আবার চলে হাতেগোনা কয়েকটি সমুদ্রতলের ক্যাবল দিয়ে, আর একটা কাটা পড়লেই সারা দেশ ধীর হয়ে যায়। দেশের ভেতরে রাখা ভিডিও তখনও চলতে থাকে।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="It uses less of your viewers' data" bn="দর্শকের ডেটা কম খরচ হয়" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Nearly nine in ten internet users in Bangladesh are on mobile data, paying by the megabyte — 120.84 million of 135.94 million, by BTRC's count in June 2026. Every megabyte you do not send them is money they keep."
					bn="বাংলাদেশের প্রায় দশজনে নয়জন ইন্টারনেট ব্যবহারকারী মোবাইল ডেটায়, মেগাবাইট হিসেবে টাকা দিয়ে — বিটিআরসির ২০২৬ সালের জুনের হিসাবে ১৩.৫৯ কোটির মধ্যে ১২.০৮ কোটি। যত মেগাবাইট তাঁদের কাছে পাঠানো হলো না, তত টাকা তাঁদের হাতেই থাকল।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="So we send video at a size a phone can actually show, rather than the largest we could. An hour of watching on the default settings uses about 270 MB. The same hour at the settings we would use for an audience abroad uses roughly four times that."
					bn="তাই আমরা ভিডিও পাঠাই এমন আকারে, যেটা ফোনের পর্দা সত্যিই দেখাতে পারে — যতটা বড় পাঠানো সম্ভব ততটা নয়। ডিফল্ট সেটিংসে এক ঘণ্টা দেখতে লাগে প্রায় ২৭০ মেগাবাইট। বিদেশি দর্শকের জন্য যে সেটিংস ব্যবহার করতাম, তাতে একই এক ঘণ্টায় লাগত প্রায় চার গুণ।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="We also leave out the newest video formats. They would save a little more data, but the inexpensive phones most people here own cannot play them smoothly — they stutter and drain the battery, which is worse for the viewer than the data it saves."
					bn="সবচেয়ে নতুন ভিডিও ফরম্যাটগুলোও আমরা রাখিনি। ওতে ডেটা আরেকটু বাঁচত, কিন্তু এখানকার বেশিরভাগ মানুষের হাতে থাকা কমদামি ফোন সেগুলো ঠিকভাবে চালাতে পারে না — আটকে আটকে চলে আর ব্যাটারি দ্রুত ফুরায়, যেটা বাঁচানো ডেটার চেয়েও দর্শকের জন্য খারাপ।"
				/>
			</p>
		</div>

		<div class="measure">
			<table>
				<caption><T en="An hour of watching, at each quality" bn="প্রতিটি মানে এক ঘণ্টা দেখলে" /></caption>
				<thead>
					<tr>
						<th scope="col"><T en="Quality" bn="মান" /></th>
						<th scope="col" class="ta-r"><T en="Data used" bn="ডেটা খরচ" /></th>
					</tr>
				</thead>
				<tbody>
					<tr>
						<th scope="row"><T as="span" en="Lowest" bn="সবচেয়ে কম" /></th>
						<td class="figure">68 MB</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="Low" bn="কম" /></th>
						<td class="figure">135 MB</td>
					</tr>
					<tr class="hi">
						<th scope="row"><T as="span" en="Default" bn="ডিফল্ট" /></th>
						<td class="figure">270 MB</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="Higher" bn="বেশি" /></th>
						<td class="figure">450 MB</td>
					</tr>
					<tr>
						<th scope="row"><T as="span" en="Highest" bn="সবচেয়ে বেশি" /></th>
						<td class="figure">810 MB</td>
					</tr>
				</tbody>
			</table>
			<p class="fine">
				<T
					as="span"
					en="Worked out from the size we send at each quality. A real video varies a little either way. The player picks the quality to suit the connection, and it has a control your viewer can use to hold it at the default and save data."
					bn="প্রতিটি মানে আমরা যে আকারে পাঠাই, তা থেকে কষে বের করা। আসল ভিডিওতে এদিক-ওদিক একটু হয়। সংযোগ বুঝে প্লেয়ার নিজেই মান বেছে নেয়, আর দর্শক চাইলে একটি বোতাম চেপে ডিফল্টেই আটকে রেখে ডেটা বাঁচাতে পারেন।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="Your videos cannot simply be passed around" bn="আপনার ভিডিও হাতে হাতে ছড়িয়ে পড়ে না" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Every video is scrambled, and the key to unscramble it is handed out for one viewing at a time rather than once per course. A link somebody copies and posts in a Telegram group stops working within hours."
					bn="প্রতিটি ভিডিও এলোমেলো করে রাখা হয়, আর সেটা খোলার চাবি দেওয়া হয় একবার দেখার জন্য একবার — পুরো কোর্সের জন্য একবার নয়। কেউ লিংক কপি করে টেলিগ্রাম গ্রুপে দিলে সেটা কয়েক ঘণ্টার মধ্যেই অকেজো হয়ে যায়।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="What it does not stop is somebody recording their own screen. Nothing short of the copy protection the film studios use does, and we have not built that. We would rather say so than use a word like unbreakable."
					bn="তবে কেউ নিজের স্ক্রিন রেকর্ড করলে এটা ঠেকাতে পারে না। ফিল্ম স্টুডিওগুলো যে ধরনের কপি-সুরক্ষা ব্যবহার করে, তা ছাড়া আর কিছুই পারে না — আর সেটা আমরা বানাইনি। ‘ভাঙা অসম্ভব’ জাতীয় কথা বলার চেয়ে এটুকু সোজাসুজি বলাই ভালো।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="wide"><T as="span" en="What using it looks like" bn="ব্যবহার করলে যা যা করতে হয়" /></h2>
	<ol class="steps">
		<li>
			<h3><T as="span" en="Send us the video" bn="ভিডিওটা পাঠান" /></h3>
			<p class="small">
				<T
					as="span"
					en="Upload the file, or give us a link and we fetch it ourselves. Large files go straight to storage, so nothing in your own system has to hold them."
					bn="ফাইলটা আপলোড করুন, অথবা একটা লিংক দিন — আমরা নিজেরাই নিয়ে নেব। বড় ফাইল সরাসরি স্টোরেজে যায়, তাই আপনার নিজের সিস্টেমকে সেটা ধরে রাখতে হয় না।"
				/>
			</p>
		</li>
		<li>
			<h3><T as="span" en="It becomes watchable quickly" bn="খুব দ্রুতই দেখার উপযোগী হয়" /></h3>
			<p class="small">
				<T
					as="span"
					en="The smaller sizes are prepared first, so you can publish before the rest is finished. The larger sizes are made the first time somebody asks for one."
					bn="ছোট আকারগুলো আগে তৈরি হয়, তাই বাকিটা শেষ হওয়ার আগেই আপনি প্রকাশ করতে পারেন। বড় আকারগুলো তৈরি হয় প্রথমবার কেউ সেটা চাইলে।"
				/>
			</p>
		</li>
		<li>
			<h3><T as="span" en="Put the link in your app" bn="লিংকটা আপনার অ্যাপে বসান" /></h3>
			<p class="small">
				<T
					as="span"
					en="It plays in an ordinary video player, on a phone or a browser. The link expires after four hours, which is what stops it being shared indefinitely."
					bn="সাধারণ ভিডিও প্লেয়ারেই চলে, ফোনে বা ব্রাউজারে। লিংকের মেয়াদ চার ঘণ্টা — এটাই ঠেকায় যাতে সেটা অনির্দিষ্টকাল ছড়িয়ে না পড়ে।"
				/>
			</p>
		</li>
		<li>
			<h3><T as="span" en="We tell your system when it is ready" bn="তৈরি হলে আপনার সিস্টেমকে জানিয়ে দিই" /></h3>
			<p class="small">
				<T
					as="span"
					en="And if something fails, you get a reason you can act on rather than a technical error, in Bangla if you ask for it."
					bn="আর কিছু ভুল হলে পাবেন এমন কারণ যা নিয়ে কিছু করা যায় — কোনো কারিগরি বার্তা নয়; চাইলে বাংলাতেই।"
				/>
			</p>
		</li>
	</ol>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="Where this actually is" bn="এখন আসলে কতদূর" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Alchemist is not finished, and the part that is missing is the one this page leans on hardest. The servers inside BDIX are the next piece of work and are not running. Today your viewers would be served from outside the country, like everyone else's."
					bn="Alchemist শেষ হয়নি, আর যেটুকু বাকি সেটাই এই পাতার সবচেয়ে বড় ভরসা। BDIX-এর ভেতরের সার্ভারগুলো পরের ধাপের কাজ, এখনও চালু হয়নি। আজ আপনার দর্শকদের কাছে ভিডিও যেত দেশের বাইরে থেকেই — বাকি সবার মতোই।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="What does work: you can send a video, we prepare it, scramble it, keep it, and give you a link that plays. Your usage is counted from the first request, and one customer's videos can never be read by another."
					bn="যা কাজ করে: আপনি ভিডিও পাঠাতে পারেন, আমরা সেটা তৈরি করি, এলোমেলো করে রাখি, রেখে দিই, আর চলে এমন একটা লিংক দিই। প্রথম অনুরোধ থেকেই আপনার ব্যবহারের হিসাব রাখা হয়, আর এক গ্রাহকের ভিডিও অন্য গ্রাহক কোনোভাবেই দেখতে পান না।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="Live streaming is not built. Studio-grade copy protection is not built. Nor is the automatic per-video size tuning we describe on the course platforms page as something we expect rather than something we have measured."
					bn="লাইভ স্ট্রিমিং তৈরি হয়নি। স্টুডিও মানের কপি-সুরক্ষাও তৈরি হয়নি। প্রতিটি ভিডিও আলাদা করে মেপে আকার ঠিক করার ব্যবস্থাটিও নয় — কোর্স প্ল্যাটফর্মের পাতায় ওটাকে আমরা প্রত্যাশা হিসেবেই লিখেছি, মেপে পাওয়া ফল হিসেবে নয়।"
				/>
			</p>
			<p class="quiet">
				<T
					as="span"
					en="There is no uptime figure anywhere on this site, because nothing is running in production to measure. When there is one it will be a measurement."
					bn="এই সাইটের কোথাও আপটাইমের কোনো হিসাব নেই, কারণ মাপার মতো কিছুই প্রোডাকশনে চলছে না। যখন থাকবে, সেটা মাপা সংখ্যাই হবে।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay closing">
	<div class="flow">
		<h2><T as="span" en="If you put video online in Bangladesh, talk to us early." bn="বাংলাদেশে ভিডিও প্রকাশ করলে আগেভাগেই কথা বলুন।" /></h2>
		<p class="quiet">
			<T
				as="span"
				en="There is no signup form, no waiting list and no trial clock, because none of them would be honest yet. What is useful now is a conversation about how much video you have, how fast you need it published, and what your viewers' data costs them."
				bn="এখানে সাইনআপ ফর্ম নেই, অপেক্ষমাণ তালিকা নেই, ট্রায়ালের ঘড়িও নেই — কারণ এখন এর কোনোটাই সৎ হতো না। এখন কাজের জিনিস একটা আলাপ: আপনার কত ভিডিও আছে, কত দ্রুত প্রকাশ করা দরকার, আর আপনার দর্শকের ডেটায় কত খরচ হয়।"
			/>
		</p>
		<div class="row">
			<a class="btn btn--solid" href="/edtech/"><T en="If you run a course platform" bn="কোর্স প্ল্যাটফর্ম চালালে" /></a>
			<a class="btn" href="/media/"><T en="If you run a news site" bn="নিউজ সাইট চালালে" /></a>
		</div>
	</div>
</section>

<style>
	/* One soft light source behind the opening, as a gradient rather than an image —
	   two lines of CSS, no bytes on the wire beyond them. */
	.open {
		position: relative;
		padding-block: clamp(3rem, 2rem + 4vw, 6rem) clamp(2.5rem, 2rem + 2vw, 4rem);
	}
	.open::before {
		content: '';
		position: absolute;
		inset: -30% 0 auto -10%;
		height: 130%;
		background: radial-gradient(46rem 26rem at 18% 12%, var(--glow), transparent 68%);
		pointer-events: none;
		z-index: -1;
	}
	.open h1 {
		max-width: 19ch;
		margin-bottom: 1.3rem;
	}
	.lead {
		max-width: 58ch;
	}
	.lead .row {
		margin-top: 1.6rem;
	}

	.measure {
		border-top: 1px solid var(--rule-firm);
		padding-top: 1rem;
	}
	.ta-r,
	.measure td {
		text-align: right;
		padding-right: 0;
	}
	.hi th,
	.hi td {
		color: var(--ink);
		font-weight: 500;
	}
	.hi td {
		color: var(--brass);
	}
	.measure .fine {
		margin-top: 0.9rem;
	}

	.wide {
		max-width: 20ch;
		margin-bottom: 1.8rem;
	}
	.steps {
		list-style: none;
		counter-reset: s;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 1.5rem clamp(1.25rem, 3vw, 2.5rem);
		grid-template-columns: repeat(auto-fit, minmax(205px, 1fr));
	}
	.steps li {
		counter-increment: s;
		padding-top: 0.75rem;
		border-top: 1px solid var(--rule-firm);
	}
	.steps li::before {
		content: counter(s);
		display: block;
		font-family: var(--mono);
		font-size: var(--step--1);
		color: var(--ink-3);
		margin-bottom: 0.5rem;
	}
	.steps h3 + p {
		margin-top: 0.35rem;
	}
	.closing h2 {
		max-width: 22ch;
	}
	.closing .row {
		margin-top: 1.4rem;
	}
</style>
