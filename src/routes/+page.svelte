<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
	import Icon from '$lib/Icon.svelte';
	import DataPanel from '$lib/DataPanel.svelte';

	const features = [
		{
			icon: 'bolt',
			ten: 'It starts, instead of buffering',
			tbn: 'বাফার নয়, সঙ্গে সঙ্গে চালু',
			ben: 'Your videos sit on servers inside BDIX, the exchange Bangladeshi internet providers use to pass traffic to each other. From there a video is a few thousandths of a second from the viewer, instead of the fifth of a second it takes to reach Europe.',
			bbn: 'আপনার ভিডিও থাকে BDIX-এর ভেতরের সার্ভারে — বাংলাদেশের ইন্টারনেট সরবরাহকারীরা নিজেদের মধ্যে ট্রাফিক আদান-প্রদান করতে যে এক্সচেঞ্জ ব্যবহার করেন। সেখান থেকে ভিডিও দর্শকের কাছ থেকে কয়েক হাজার ভাগের এক সেকেন্ড দূরে, ইউরোপ পর্যন্ত যেতে যেখানে লাগে সেকেন্ডের পাঁচ ভাগের এক ভাগ।'
		},
		{
			icon: 'data',
			ten: "It does not eat the viewer's allowance",
			tbn: 'দর্শকের ডেটা লিমিট শেষ করে না',
			ben: 'On most home broadband packages here, anything coming from inside BDIX does not count against the data allowance — what providers call FUP. Vimeo, Bunny, Mux and Cloudflare all keep their video outside the country, so none of them can offer that.',
			bbn: 'এখানকার বেশিরভাগ বাসাবাড়ির ব্রডব্যান্ড প্যাকেজে BDIX-এর ভেতর থেকে আসা কিছু ডেটা লিমিটের হিসাবেই ধরা হয় না — সরবরাহকারীরা যেটাকে FUP বলেন। Vimeo, Bunny, Mux আর Cloudflare সবাই ভিডিও রাখে দেশের বাইরে, তাই এই সুবিধাটা তাদের কেউ দিতে পারে না।'
		},
		{
			icon: 'lock',
			ten: 'Your videos are not easy to pass around',
			tbn: 'আপনার ভিডিও সহজে ছড়িয়ে পড়ে না',
			ben: 'Every video is scrambled, and the key to unscramble it is handed out for one viewing at a time rather than once per course. A link somebody copies into a Telegram group stops working within hours.',
			bbn: 'প্রতিটি ভিডিও এলোমেলো করে রাখা হয়, আর সেটি খোলার চাবি দেওয়া হয় একবার দেখার জন্য একবার — পুরো কোর্সের জন্য একবার নয়। কেউ লিংক কপি করে টেলিগ্রাম গ্রুপে দিলে তা কয়েক ঘণ্টার মধ্যেই অকেজো হয়ে যায়।'
		},
		{
			icon: 'cable',
			ten: 'It keeps working when the cables cut',
			tbn: 'ক্যাবল কাটলেও চলতে থাকে',
			ben: 'Bangladesh reaches the rest of the world through a handful of undersea cables, and a cut has slowed the whole country more than once. Video held inside the country does not travel those cables.',
			bbn: 'বাংলাদেশ বাকি দুনিয়ার সঙ্গে যুক্ত হাতেগোনা কয়েকটি সমুদ্রতলের ক্যাবল দিয়ে, আর একটা কাটা পড়ে একাধিকবার সারা দেশ ধীর হয়ে গেছে। দেশের ভেতরে রাখা ভিডিও ওই ক্যাবল দিয়ে যায়ই না।'
		},
		{
			icon: 'clock',
			ten: 'Publishable before it is finished',
			tbn: 'পুরো তৈরি হওয়ার আগেই প্রকাশযোগ্য',
			ben: 'The smaller sizes are prepared first, so a video can go out before the rest is done. The larger sizes are made the first time somebody asks for one, once, however many arrive together.',
			bbn: 'ছোট আকারগুলো আগে তৈরি হয়, তাই বাকিটা শেষ হওয়ার আগেই ভিডিও ছেড়ে দেওয়া যায়। বড় আকারগুলো তৈরি হয় প্রথমবার কেউ চাইলে — একবারই, একসঙ্গে যত জনই আসুন।'
		},
		{
			icon: 'script',
			ten: 'Bangla that displays properly',
			tbn: 'বাংলা যেন ঠিকভাবে দেখায়',
			ben: 'Bangla titles, file names and subtitles survive from your upload to the viewer’s screen. Conjunct letters keep their shape instead of turning into empty boxes, which is a real and common failure.',
			bbn: 'বাংলা শিরোনাম, ফাইলের নাম আর সাবটাইটেল আপনার আপলোড থেকে দর্শকের পর্দা পর্যন্ত অক্ষত থাকে। যুক্তাক্ষর ভেঙে ফাঁকা বাক্স হয়ে যায় না — এটা সত্যিই ঘটে, আর প্রায়ই ঘটে।'
		}
	];

	const examples = [
		{
			ten: 'A small course platform',
			tbn: 'ছোট একটি কোর্স প্ল্যাটফর্ম',
			rows: [
				['5 hours of new video a month', 'মাসে ৫ ঘণ্টা নতুন ভিডিও', '৳600'],
				['120 GB held', '১২০ জিবি রাখা', '৳144'],
				['60 GB watched in Bangladesh', 'দেশে দেখা ৬০ জিবি', '৳21']
			],
			total: '৳765',
			usd: '$6'
		},
		{
			ten: 'A mid-sized course platform',
			tbn: 'মাঝারি একটি কোর্স প্ল্যাটফর্ম',
			featured: true,
			rows: [
				['20 hours of new video a month', 'মাসে ২০ ঘণ্টা নতুন ভিডিও', '৳2,400'],
				['500 GB held', '৫০০ জিবি রাখা', '৳600'],
				['250 GB watched in Bangladesh', 'দেশে দেখা ২৫০ জিবি', '৳88']
			],
			total: '৳3,088',
			usd: '$25'
		},
		{
			ten: 'A news site',
			tbn: 'একটি নিউজ সাইট',
			rows: [
				['60 hours of new video a month', 'মাসে ৬০ ঘণ্টা নতুন ভিডিও', '৳7,200'],
				['900 GB held', '৯০০ জিবি রাখা', '৳1,080'],
				['1,500 GB watched in Bangladesh', 'দেশে দেখা ১,৫০০ জিবি', '৳525']
			],
			total: '৳8,805',
			usd: '$70'
		}
	];
</script>

<Seo
	title="Alchemist — video hosting for Bangladesh"
	description="Put your videos online so they start playing straight away in Bangladesh and use as little of your viewers' mobile data as possible. Send a file or a link; get back a link that plays."
/>

<section class="hero">
	<div class="mesh" aria-hidden="true"></div>
	<div class="shell hero__in">
		<p class="pill">
			<b></b>
			<T
				as="span"
				en="Hosted inside BDIX · sized for mobile data"
				bn="BDIX-এর ভেতরে হোস্ট করা · মোবাইল ডেটার উপযোগী আকারে"
			/>
		</p>
		<h1>
			<T as="span" en="Video that plays straight away in Bangladesh," bn="যে ভিডিও বাংলাদেশে সঙ্গে সঙ্গে চলে," />
			<span class="accent">
				<T as="span" en="without using up your viewers' data." bn="আর দর্শকের ডেটা শেষ করে না।" />
			</span>
		</h1>
		<p class="lede quiet">
			<T
				as="span"
				en="Alchemist keeps your videos and sends them to the people watching. You upload a file, or give us a link and we fetch it; you get back a link to put in your app or your website, and it plays."
				bn="Alchemist আপনার ভিডিও রেখে দেয় আর দর্শকের কাছে পৌঁছে দেয়। আপনি ফাইল আপলোড করেন, বা একটা লিংক দেন আর আমরা সেখান থেকে নিয়ে নিই; বদলে পান এমন একটা লিংক, যেটা আপনার অ্যাপ বা ওয়েবসাইটে বসিয়ে দিলেই ভিডিও চলে।"
			/>
		</p>
		<div class="row hero__cta">
			<a class="btn btn--solid" href="/pricing/"><T en="What it costs" bn="খরচ কত" /></a>
			<a class="btn" href="/docs/"><T en="For developers" bn="ডেভেলপারদের জন্য" /></a>
		</div>
		<ul class="assure">
			<li><Icon name="check" size={15} /><T as="span" en="About 270 MB an hour of watching" bn="ঘণ্টায় প্রায় ২৭০ মেগাবাইট" /></li>
			<li><Icon name="check" size={15} /><T as="span" en="A fresh link for every viewing" bn="প্রতিবার দেখার জন্য নতুন লিংক" /></li>
			<li><Icon name="check" size={15} /><T as="span" en="Bangla and English throughout" bn="সবখানে বাংলা ও ইংরেজি" /></li>
		</ul>
	</div>
</section>

<section class="shell diagram">
	<DataPanel />
	<p class="built-for fine">
		<T
			as="span"
			en="Built for course platforms, news sites and streaming services in Bangladesh."
			bn="বাংলাদেশের কোর্স প্ল্যাটফর্ম, নিউজ সাইট আর স্ট্রিমিং সার্ভিসের কথা ভেবে বানানো।"
		/>
	</p>
</section>

<section class="shell bay">
	<h2 class="lead-h"><T as="span" en="What it does for you" bn="এটি আপনার জন্য যা করে" /></h2>
	<div class="cards">
		{#each features as f (f.icon)}
			<article class="card">
				<span class="card__icon"><Icon name={f.icon} size={17} /></span>
				<h3><T as="span" en={f.ten} bn={f.tbn} /></h3>
				<p><T as="span" en={f.ben} bn={f.bbn} /></p>
			</article>
		{/each}
	</div>
</section>

<section class="shell bay">
	<h2 class="lead-h"><T as="span" en="How it works" bn="কীভাবে কাজ করে" /></h2>
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
			<h3><T as="span" en="We prepare it and tell you" bn="আমরা তৈরি করি, আর জানিয়ে দিই" /></h3>
			<p class="small">
				<T
					as="span"
					en="The smaller sizes are prepared first and your system is told as soon as the video is watchable. If something fails you get a reason you can act on, in Bangla if you ask for it."
					bn="ছোট আকারগুলো আগে তৈরি হয়, আর ভিডিও দেখার উপযোগী হওয়ামাত্র আপনার সিস্টেমকে জানিয়ে দেওয়া হয়। কিছু ভুল হলে পাবেন এমন কারণ যা নিয়ে কিছু করা যায় — চাইলে বাংলাতেই।"
				/>
			</p>
		</li>
		<li>
			<h3><T as="span" en="Put the link in your app" bn="লিংকটা আপনার অ্যাপে বসান" /></h3>
			<p class="small">
				<T
					as="span"
					en="It plays in an ordinary video player, on a phone or in a browser. The link expires after four hours, which is what stops it being shared indefinitely."
					bn="সাধারণ ভিডিও প্লেয়ারেই চলে, ফোনে বা ব্রাউজারে। লিংকের মেয়াদ চার ঘণ্টা — এটাই ঠেকায় যাতে তা অনির্দিষ্টকাল ছড়িয়ে না পড়ে।"
				/>
			</p>
		</li>
	</ol>

	<div class="frame worked">
		<div class="frame__bar">
			<span class="frame__title">
				<T as="span" en="A 45-minute lecture, from the student's side" bn="৪৫ মিনিটের একটি লেকচার, শিক্ষার্থীর দিক থেকে" />
			</span>
			<span class="frame__note"><T as="span" en="At the default size" bn="ডিফল্ট আকারে" /></span>
		</div>
		<div class="frame__body worked__body">
			<div>
				<p class="wlabel"><T as="span" en="Watched at the default size" bn="ডিফল্ট আকারে দেখলে" /></p>
				<p class="wval">203<span> MB</span></p>
			</div>
			<div>
				<p class="wlabel"><T as="span" en="Listened to as audio only" bn="শুধু অডিও হিসেবে শুনলে" /></p>
				<p class="wval">20<span> MB</span></p>
			</div>
			<p class="small wnote">
				<T
					as="span"
					en="Most of a lecture is the talking, so a student revising can listen instead of watch for about a tenth of the data. On a metered connection that is the difference between finishing a course and rationing it."
					bn="লেকচারের বেশিরভাগটাই আসলে কথা, তাই রিভিশনের সময় শিক্ষার্থী দেখার বদলে শুনতে পারেন — ডেটা লাগে প্রায় দশ ভাগের এক ভাগ। মিটারড কানেকশনে এটাই ঠিক করে দেয়, কোর্সটা শেষ হবে নাকি মেপে মেপে দেখতে হবে।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="lead-h"><T as="span" en="What it would cost" bn="খরচ কেমন পড়ত" /></h2>
	<p class="quiet sub">
		<T
			as="span"
			en="There are no plans to choose between. You pay for the video you send us, the video we hold, and the video people watch — so the bill follows what you actually do. Three worked examples at the same three rates:"
			bn="বেছে নেওয়ার মতো কোনো প্যাকেজ নেই। আপনি টাকা দেন আমাদের পাঠানো ভিডিওর জন্য, আমরা যা রাখি তার জন্য, আর মানুষ যা দেখে তার জন্য — তাই বিল চলে আপনি আসলে যা করেন তার পেছন পেছন। একই তিনটি হারে কষা তিনটি উদাহরণ:"
		/>
	</p>
	<div class="tiers">
		{#each examples as ex (ex.ten)}
			<article class="tier" class:tier--featured={ex.featured}>
				<h3><T as="span" en={ex.ten} bn={ex.tbn} /></h3>
				<p class="tier__total">{ex.total}<span> / <T as="span" en="month" bn="মাস" /></span></p>
				<p class="tier__usd"><T as="span" en="About {ex.usd}" bn="প্রায় {ex.usd}" /></p>
				<ul>
					{#each ex.rows as [en, bn, amount] (en)}
						<li><span><T as="span" {en} {bn} /></span><span class="figure">{amount}</span></li>
					{/each}
				</ul>
			</article>
		{/each}
	</div>
	<p class="fine tiers__note">
		<T
			as="span"
			en="Worked out at the rates on the pricing page, where we also publish what it costs us to run. Pricing is not final until launch."
			bn="মূল্যের পাতার হারে কষা — সেখানে আমাদের চালাতে কত খরচ হয় তাও দেওয়া আছে। চালু হওয়ার আগপর্যন্ত দাম চূড়ান্ত নয়।"
		/>
	</p>
	<div class="row"><a class="btn" href="/pricing/"><T en="See the rates" bn="হারগুলো দেখুন" /></a></div>
</section>

<section class="shell bay" id="faq">
	<h2 class="lead-h"><T as="span" en="Questions" bn="প্রশ্ন" /></h2>
	<div class="faq">
		<details>
			<summary><T as="span" en="How do I get my videos in?" bn="ভিডিওগুলো কীভাবে দেব?" /></summary>
			<p>
				<T
					as="span"
					en="Upload them, or hand us a link and we fetch the file ourselves. A whole existing library can come across in one go, and nothing about how you already publish has to change to try it."
					bn="আপলোড করে দিন, অথবা একটা লিংক দিন — ফাইলটা আমরা নিজেরাই নিয়ে নেব। পুরোনো পুরো লাইব্রেরিটাই একবারে আনা যায়, আর যাচাই করে দেখতে আপনি এখন যেভাবে প্রকাশ করেন তার কিছুই বদলাতে হয় না।"
				/>
			</p>
		</details>
		<details>
			<summary><T as="span" en="How does the billing work?" bn="বিল কীভাবে হয়?" /></summary>
			<p>
				<T
					as="span"
					en="Three things are counted: the minutes of video you send us, the gigabytes we hold, and the gigabytes people watch. No per-person fee, no minimum, and no charge to leave and take your files. You can cancel any time, with no notice period and no asterisk after the sentence."
					bn="তিনটি জিনিসের হিসাব রাখা হয়: আপনার পাঠানো ভিডিওর মিনিট, আমরা যত গিগাবাইট রাখি, আর মানুষ যত গিগাবাইট দেখে। মাথাপিছু কোনো ফি নেই, ন্যূনতম নেই, আর ছেড়ে গিয়ে নিজের ফাইল নিয়ে যেতেও খরচ নেই। যেকোনো সময় বন্ধ করতে পারেন — আগে জানানোর বাধ্যবাধকতা নেই, বাক্যের পাশে তারকাচিহ্নও নেই।"
				/>
			</p>
		</details>
		<details>
			<summary><T as="span" en="Can students still download my course?" bn="শিক্ষার্থীরা কি তবু কোর্স নামিয়ে নিতে পারবে?" /></summary>
			<p>
				<T
					as="span"
					en="Copying a link and sharing it stops working, because the key is issued for one viewing and the link expires. Somebody recording their own screen is not stopped by any of this. Only the copy protection the film studios use does that, and we have not built it."
					bn="লিংক কপি করে শেয়ার করা কাজ করবে না, কারণ চাবি দেওয়া হয় একবার দেখার জন্য আর লিংকের মেয়াদ শেষ হয়ে যায়। তবে কেউ নিজের স্ক্রিন রেকর্ড করলে এর কোনোটাই তা ঠেকাতে পারে না। কেবল ফিল্ম স্টুডিওগুলোর কপি-সুরক্ষাই সেটা পারে, আর তা আমরা বানাইনি।"
				/>
			</p>
		</details>
		<details>
			<summary><T as="span" en="Do I have to change how I publish?" bn="আমি এখন যেভাবে প্রকাশ করি, তা কি বদলাতে হবে?" /></summary>
			<p>
				<T
					as="span"
					en="No. Give us a link and we fetch the file ourselves, so an existing library or workflow can be brought over without rebuilding it. The video plays in an ordinary player on a phone or in a browser."
					bn="না। একটা লিংক দিন, ফাইলটা আমরা নিজেরাই নিয়ে নেব — তাই পুরোনো লাইব্রেরি বা কাজের ধারা নতুন করে না বানিয়েও আনা যায়। ভিডিও চলে ফোনে বা ব্রাউজারে সাধারণ প্লেয়ারেই।"
				/>
			</p>
		</details>
		<details>
			<summary><T as="span" en="What happens when an undersea cable is cut?" bn="সমুদ্রতলের ক্যাবল কাটা পড়লে কী হয়?" /></summary>
			<p>
				<T
					as="span"
					en="Video held inside Bangladesh does not travel those cables, so it keeps playing while the rest of the country's internet slows down. For a newsroom mid-story or a course platform mid-exam-season, that is the whole argument for keeping it local."
					bn="বাংলাদেশের ভেতরে রাখা ভিডিও ওই ক্যাবল দিয়ে যায় না, তাই দেশের বাকি ইন্টারনেট ধীর হয়ে গেলেও সেটা চলতে থাকে। বড় খবরের মাঝপথে থাকা নিউজরুম বা পরীক্ষার মৌসুমে থাকা কোর্স প্ল্যাটফর্মের জন্য কনটেন্ট দেশে রাখার পুরো যুক্তিটাই এটা।"
				/>
			</p>
		</details>
		<details>
			<summary><T as="span" en="Do you do live streaming?" bn="লাইভ স্ট্রিমিং কি হয়?" /></summary>
			<p>
				<T
					as="span"
					en="No, and it is not close. If you need live — in Bangladesh that usually means sport — say so early, because it changes the order of the work rather than slotting into it."
					bn="না, আর কাছাকাছিও নয়। লাইভ দরকার হলে — বাংলাদেশে যার মানে সাধারণত খেলা — আগেভাগে বলবেন, কারণ এটা কাজের ক্রমই বদলে দেয়, চুপচাপ কোনো ফাঁকে ঢোকে না।"
				/>
			</p>
		</details>
	</div>
</section>

<section class="shell bay">
	<div class="closing">
		<h2><T as="span" en="If you put video online in Bangladesh, talk to us early." bn="বাংলাদেশে ভিডিও প্রকাশ করলে আগেভাগেই কথা বলুন।" /></h2>
		<p class="quiet">
			<T
				as="span"
				en="Every large video service was built for somewhere else, and it shows here: video that crosses an ocean to reach Dhaka, sized for people who never look at their data bill. Alchemist exists because the network in this country works differently, and because a viewer paying by the megabyte deserves a service that knows it."
				bn="বড় ভিডিও সার্ভিসগুলোর সবকটিই তৈরি হয়েছে অন্য কোথাও বসে, আর এখানে এলে সেটা ধরা পড়ে যায়: ঢাকায় পৌঁছাতে ভিডিওকে পেরোতে হয় সাগর, আর আকার ঠিক করা হয় এমন মানুষের কথা ভেবে যাঁরা কখনও ডেটার বিল দেখেন না। Alchemist আছে কারণ এই দেশের নেটওয়ার্ক অন্যরকম, আর কারণ মেগাবাইট হিসেবে টাকা দেওয়া দর্শক এমন একটা সার্ভিস পাওয়ার যোগ্য যেটা এ কথা জানে।"
			/>
		</p>
		<div class="row">
			<a class="btn btn--solid" href="/edtech/"><T en="If you run a course platform" bn="কোর্স প্ল্যাটফর্ম চালালে" /></a>
			<a class="btn" href="/media/"><T en="If you run a news site" bn="নিউজ সাইট চালালে" /></a>
		</div>
	</div>
</section>

<style>
	/* One soft light source, built from four overlapping radial gradients rather than
	   a canvas or an image: a shader costs battery and bytes on the cheap phones this
	   product exists for. The drift is transform-only on a single layer. */
	.hero {
		position: relative;
		isolation: isolate;
		overflow: hidden;
	}
	.mesh {
		position: absolute;
		inset: -40% -20% auto -20%;
		height: 150%;
		z-index: -1;
		pointer-events: none;
		background:
			radial-gradient(38rem 22rem at 22% 30%, rgba(232, 163, 61, 0.16), transparent 62%),
			radial-gradient(30rem 20rem at 62% 12%, rgba(232, 163, 61, 0.09), transparent 60%),
			radial-gradient(34rem 24rem at 82% 48%, rgba(120, 96, 200, 0.09), transparent 64%),
			radial-gradient(26rem 18rem at 44% 68%, rgba(70, 150, 170, 0.07), transparent 62%);
		animation: drift 120s var(--ease) infinite alternate;
	}
	:global(html[data-theme='light']) .mesh {
		background:
			radial-gradient(38rem 22rem at 22% 30%, rgba(138, 90, 11, 0.11), transparent 62%),
			radial-gradient(30rem 20rem at 62% 12%, rgba(138, 90, 11, 0.07), transparent 60%),
			radial-gradient(34rem 24rem at 82% 48%, rgba(90, 70, 160, 0.07), transparent 64%),
			radial-gradient(26rem 18rem at 44% 68%, rgba(40, 120, 140, 0.06), transparent 62%);
	}
	@keyframes drift {
		from {
			transform: translate3d(-2%, -1%, 0) scale(1.04);
		}
		to {
			transform: translate3d(2%, 1%, 0) scale(1.1);
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.mesh {
			animation: none;
		}
	}

	.hero__in {
		padding-block: clamp(2.5rem, 1.8rem + 3.5vw, 5rem) clamp(2rem, 1.5rem + 2vw, 3.5rem);
	}
	.hero h1 {
		max-width: 21ch;
		margin: 1.1rem 0 1.1rem;
	}
	.accent {
		display: block;
		color: var(--brass);
	}
	.hero .lede {
		max-width: 56ch;
	}
	.hero__cta {
		margin: 1.5rem 0 1.5rem;
	}

	.diagram {
		padding-bottom: var(--bay);
	}
	.built-for {
		margin-top: 1.1rem;
		max-width: 70ch;
	}

	.lead-h {
		margin-bottom: 1.6rem;
	}
	.sub {
		max-width: 62ch;
		margin: -0.8rem 0 1.6rem;
	}

	.steps {
		list-style: none;
		counter-reset: s;
		padding: 0;
		margin: 0 0 2rem;
		display: grid;
		gap: 1.5rem clamp(1.25rem, 3vw, 2.5rem);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
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
		color: var(--brass);
		margin-bottom: 0.5rem;
	}
	.steps h3 + p {
		margin-top: 0.35rem;
	}

	.worked__body {
		display: grid;
		gap: 1rem clamp(1.5rem, 4vw, 3rem);
		grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
		align-items: start;
	}
	.wlabel {
		font-size: var(--step--1);
		color: var(--ink-3);
		margin-bottom: 0.3rem;
	}
	:global(html[lang='bn']) .wlabel {
		font-size: var(--step-0);
	}
	.wval {
		font-size: var(--step-5);
		font-weight: 600;
		letter-spacing: -0.035em;
		line-height: 1;
		color: var(--brass);
	}
	.wval span {
		font-size: var(--step-2);
		font-weight: 400;
		color: var(--ink-3);
	}
	.wnote {
		grid-column: 1 / -1;
		margin: 0;
		padding-top: 0.9rem;
		border-top: 1px solid var(--rule);
	}

	.tiers {
		display: grid;
		gap: var(--gap);
		grid-template-columns: repeat(auto-fit, minmax(245px, 1fr));
		align-items: start;
	}
	.tier {
		background: var(--leaf);
		border: 1px solid var(--rule);
		border-radius: var(--r);
		box-shadow: var(--lift);
		padding: 1.2rem;
	}
	.tier--featured {
		border-color: var(--brass);
		background: var(--sink);
	}
	.tier__total {
		margin-top: 0.8rem;
		font-size: var(--step-5);
		font-weight: 600;
		letter-spacing: -0.035em;
		line-height: 1.1;
	}
	.tier__total span {
		font-size: var(--step-1);
		font-weight: 400;
		letter-spacing: -0.01em;
		color: var(--ink-3);
	}
	.tier__usd {
		font-size: var(--step-0);
		color: var(--ink-3);
		margin-bottom: 1rem;
	}
	.tier ul {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 0;
	}
	.tier li {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		gap: 1rem;
		padding: 0.5rem 0;
		border-top: 1px solid var(--rule);
		font-size: var(--step-0);
		color: var(--graphite);
	}
	:global(html[lang='bn']) .tier li {
		font-size: var(--step-1);
	}
	.tiers__note {
		margin: 1.2rem 0 1.2rem;
		max-width: 74ch;
	}

	.closing {
		max-width: 56ch;
	}
	.closing h2 {
		margin-bottom: 0.9rem;
	}
	.closing .row {
		margin-top: 1.4rem;
	}
</style>
