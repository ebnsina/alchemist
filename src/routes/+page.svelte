<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
	import Icon from '$lib/Icon.svelte';

	const ladder = [
		{ rung: '144p', mbps: '0.15', en: 'Rural 3G, edge of coverage. It plays.', bn: 'গ্রামের 3G, নেটওয়ার্কের কিনারায়। অন্তত চলে।' },
		{ rung: '240p', mbps: '0.30', en: 'Congested 4G. The realistic floor.', bn: 'ভিড় করা 4G। বাস্তব সর্বনিম্ন ধাপ।' },
		{ rung: '360p', mbps: '0.60', en: 'The default target. Where most BD mobile viewing lands.', bn: 'মূল লক্ষ্য। বাংলাদেশের বেশিরভাগ মোবাইল দর্শক এখানেই থাকেন।' },
		{ rung: '480p', mbps: '1.00', en: 'Good 4G, cheap phones at full screen.', bn: 'ভালো 4G, কমদামি ফোনে ফুল স্ক্রিন।' },
		{ rung: '720p', mbps: '1.80', en: 'Top rung by default. Matches most BD handsets.', bn: 'ডিফল্টে সর্বোচ্চ ধাপ। বেশিরভাগ হ্যান্ডসেটের পর্দার সমান।' }
	];
</script>

<Seo
	title="Alchemist — video hosting inside BDIX, built for metered mobile"
	description="A video API for Bangladesh. Upload or import a video, get an encrypted adaptive-bitrate stream and a signed playback URL — delivered from inside BDIX, encoded for viewers who pay per megabyte."
/>

<section class="hero">
	<div class="wrap">
		<p class="eyebrow"><T en="Video API · Bangladesh first" bn="ভিডিও API · আগে বাংলাদেশ" /></p>
		<h1>
			<T
				as="span"
				en="Video that starts instantly in Bangladesh — and doesn't eat your viewers' data."
				bn="যে ভিডিও বাংলাদেশে সঙ্গে সঙ্গে চালু হয় — আর দর্শকের ডেটা শেষ করে না।"
			/>
		</h1>
		<p class="lede">
			<T
				as="span"
				en="Alchemist is a video transcoding and delivery API. Upload or import a file; get back an encrypted adaptive-bitrate stream and a signed playback URL. It is hosted inside BDIX and encoded for a mobile ladder, because that is what Bangladesh actually runs on."
				bn="Alchemist একটি ভিডিও ট্রান্সকোডিং ও ডেলিভারি API। ফাইল আপলোড করুন বা লিংক থেকে আনুন; ফেরত পাবেন এনক্রিপ্টেড অ্যাডাপটিভ-বিটরেট স্ট্রিম আর সাইন করা প্লেব্যাক লিংক। এটি হোস্ট করা BDIX-এর ভেতরে, আর এনকোড করা মোবাইল ল্যাডারে — কারণ বাংলাদেশ আসলে এর উপরেই চলে।"
			/>
		</p>
		<div class="cta">
			<a class="btn btn-primary" href="/docs/"
				><T en="Read the API docs" bn="API ডকুমেন্টেশন দেখুন" /><Icon name="arrow" size={18} /></a
			>
			<a class="btn" href="/pricing/"><T en="See how billing works" bn="বিলিং কীভাবে হয়" /></a>
		</div>
		<p class="honest">
			<T
				as="span"
				en="Alchemist is in development. Below, the network facts are public and sourced; everything about Alchemist itself is marked as either working today or not built yet."
				bn="Alchemist এখনও তৈরি হচ্ছে। নিচে নেটওয়ার্ক সংক্রান্ত তথ্যগুলো প্রকাশ্য ও সূত্রসহ; আর Alchemist নিজে কী এখন কাজ করে আর কী এখনও তৈরি হয়নি, দুটোই আলাদা করে লেখা আছে।"
			/>
		</p>
	</div>
</section>

<section class="facts">
	<div class="wrap">
		<div class="grid">
			<article class="card fact">
				<div class="ficon"><Icon name="globe" size={24} /></div>
				<p class="fnum num">1–5<span>ms</span></p>
				<h2 class="fh">
					<T as="span" en="Inside BDIX, not across an ocean" bn="সাগর পেরিয়ে নয়, BDIX-এর ভেতর থেকে" />
				</h2>
				<p>
					<T
						as="span"
						en="Content served from inside the Bangladesh Internet Exchange reaches BD viewers in 1–5ms at line rate. The same file from a European or US region is 80–200ms away. Vimeo, Bunny, Mux and Cloudflare Stream are not in BDIX."
						bn="বাংলাদেশ ইন্টারনেট এক্সচেঞ্জের ভেতর থেকে দেওয়া কনটেন্ট দেশের দর্শকের কাছে পৌঁছায় ১–৫ মিলিসেকেন্ডে, পুরো লাইন স্পিডে। একই ফাইল ইউরোপ বা আমেরিকা থেকে এলে লাগে ৮০–২০০ মিলিসেকেন্ড। Vimeo, Bunny, Mux বা Cloudflare Stream কেউই BDIX-এ নেই।"
					/>
				</p>
			</article>

			<article class="card fact save">
				<div class="ficon"><Icon name="coins" size={24} /></div>
				<p class="fnum word"><T as="span" en="No FUP" bn="FUP লাগে না" /></p>
				<h2 class="fh">
					<T as="span" en="Local traffic usually doesn't count" bn="লোকাল ট্রাফিক সাধারণত হিসাবেই আসে না" />
				</h2>
				<p>
					<T
						as="span"
						en="On most BD broadband packages, BDIX traffic is exempt from the subscriber's FUP — it does not consume their data quota. An entire consumer content culture in Bangladesh exists because of this. It is the reason a foreign CDN cannot match a local one here."
						bn="বাংলাদেশের বেশিরভাগ ব্রডব্যান্ড প্যাকেজে BDIX ট্রাফিক FUP-এর বাইরে — গ্রাহকের ডেটা কোটা থেকে কাটে না। দেশের পুরো একটা কনটেন্ট সংস্কৃতি দাঁড়িয়েই আছে এর উপর। এখানে বিদেশি CDN কেন লোকাল সার্ভারের সঙ্গে পারে না, কারণটা এটাই।"
					/>
				</p>
			</article>

			<article class="card fact">
				<div class="ficon"><Icon name="phone" size={24} /></div>
				<p class="fnum num">89<span>%</span></p>
				<h2 class="fh">
					<T as="span" en="are on metered mobile" bn="মানুষ মিটারড মোবাইলে" />
				</h2>
				<p>
					<T
						as="span"
						en="BTRC, June 2026: 120.84 million mobile internet users out of 135.94 million total. For them BDIX fixes latency but not the bill — every megabit you don't send is money left in a viewer's pocket. That is why the default ladder tops out at 720p."
						bn="বিটিআরসি, জুন ২০২৬: মোট ১৩.৫৯ কোটি ইন্টারনেট ব্যবহারকারীর মধ্যে ১২.০৮ কোটিই মোবাইলে। তাঁদের জন্য BDIX লেটেন্সি ঠিক করে, বিল নয় — যত কম মেগাবিট পাঠানো হবে, দর্শকের পকেটে তত টাকা থাকবে। এই কারণেই ডিফল্ট ল্যাডার ৭২০পি-তে থামে।"
					/>
				</p>
			</article>

			<article class="card fact">
				<div class="ficon"><Icon name="cable" size={24} /></div>
				<p class="fnum word"><T as="span" en="Cable cuts" bn="ক্যাবল কাটা" /></p>
				<h2 class="fh">
					<T as="span" en="Local content keeps working" bn="লোকাল কনটেন্ট চলতেই থাকে" />
				</h2>
				<p>
					<T
						as="span"
						en="Bangladesh's international connectivity runs through a small number of submarine cables, and cuts have degraded the country's access more than once. Content served from inside BDIX does not cross them. Mid-exam-season or mid-event, that is not a footnote."
						bn="বাংলাদেশের আন্তর্জাতিক ইন্টারনেট চলে হাতেগোনা কয়েকটি সাবমেরিন ক্যাবলের উপর দিয়ে, আর ক্যাবল কাটা পড়ে একাধিকবার সারা দেশের ইন্টারনেট ধীর হয়ে গেছে। BDIX-এর ভেতর থেকে দেওয়া কনটেন্ট ওই পথ মাড়ায় না। পরীক্ষার মৌসুমে বা বড় কোনো ঘটনার সময় এটা ছোট বিষয় নয়।"
					/>
				</p>
			</article>
		</div>
	</div>
</section>

<section>
	<div class="wrap">
		<p class="eyebrow"><T en="The ladder" bn="ল্যাডার" /></p>
		<h2><T as="span" en="Encoded for the phone people actually hold" bn="মানুষ যে ফোন হাতে নিয়ে দেখে, সেটার জন্যই এনকোড করা" /></h2>
		<p class="lede">
			<T
				as="span"
				en="The default profile is 144p to 720p, H.264 Main, no AV1. 1080p is opt-in per tenant rather than on by default, because most handsets here are 720p-class and shipping 1080p by default spends a viewer's money on pixels their screen cannot show."
				bn="ডিফল্ট প্রোফাইল ১৪৪পি থেকে ৭২০পি, H.264 Main, AV1 নেই। ১০৮০পি প্রতি গ্রাহকের জন্য আলাদা করে চালু করতে হয়, ডিফল্টে নয় — কারণ এখানকার বেশিরভাগ ফোনের পর্দা ৭২০পি মানের, আর ডিফল্টে ১০৮০পি পাঠানো মানে দর্শকের টাকা খরচ করে এমন পিক্সেল পাঠানো যা তাঁর পর্দাই দেখাতে পারে না।"
			/>
		</p>
		<div class="scroll-x ladderwrap">
			<table>
				<caption class="sr-only"
					><T en="Default mobile encoding ladder" bn="ডিফল্ট মোবাইল এনকোডিং ল্যাডার" /></caption
				>
				<thead>
					<tr>
						<th scope="col"><T en="Rung" bn="ধাপ" /></th>
						<th scope="col"><T en="Bitrate" bn="বিটরেট" /></th>
						<th scope="col"><T en="Who it is for" bn="কার জন্য" /></th>
					</tr>
				</thead>
				<tbody>
					{#each ladder as r (r.rung)}
						<tr>
							<th scope="row" class="num">{r.rung}</th>
							<td class="num nowrap">{r.mbps} Mbps</td>
							<td><T as="span" en={r.en} bn={r.bn} /></td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
		<div class="note">
			<p>
				<T
					as="span"
					en="AV1 is excluded, not deferred: the Helio-G and Unisoc chipsets common in the Bangladeshi budget market have no hardware decode for it, so software decoding drops frames and drains battery. It is worse for the viewer than H.264, so we do not ship it here."
					bn="AV1 পিছিয়ে দেওয়া হয়নি, একেবারেই বাদ দেওয়া হয়েছে: বাংলাদেশের সস্তা ফোনে চলা Helio-G ও Unisoc চিপে এর হার্ডওয়্যার ডিকোড নেই, ফলে সফটওয়্যারে চালালে ফ্রেম পড়ে যায় আর ব্যাটারি দ্রুত ফুরায়। দর্শকের জন্য সেটা H.264-এর চেয়েও খারাপ, তাই এখানে আমরা এটা দিই না।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="alt">
	<div class="wrap">
		<p class="eyebrow"><T en="How it works" bn="কীভাবে কাজ করে" /></p>
		<h2><T as="span" en="Four calls, and it plays" bn="চারটি কল, তারপরই চলে" /></h2>
		<ol class="steps">
			<li>
				<h3><T as="span" en="Get an upload target" bn="আপলোডের জায়গা নিন" /></h3>
				<p>
					<T
						as="span"
						en="POST /v1/uploads returns a presigned URL and an asset ID. The bytes go straight to storage — they never pass through the API. Or skip it and POST /v1/assets with a URL to import from."
						bn="POST /v1/uploads দিলে একটি প্রি-সাইন করা লিংক আর একটি অ্যাসেট আইডি পাবেন। ফাইলের বাইটগুলো সরাসরি স্টোরেজে যায়, API-এর ভেতর দিয়ে নয়। অথবা এটা বাদ দিয়ে POST /v1/assets-এ একটি লিংক দিলে সেখান থেকেই নিয়ে নেবে।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="It encodes the low rungs first" bn="আগে নিচের ধাপগুলো এনকোড হয়" /></h3>
				<p>
					<T
						as="span"
						en="Ingest encodes 144p–360p and the asset becomes partially_ready — which means playable. The higher rungs are generated on first play instead of up front, so nobody waits for quality nobody has asked to watch."
						bn="ইনজেস্টে প্রথমে ১৪৪পি–৩৬০পি এনকোড হয় আর অ্যাসেট হয়ে যায় partially_ready — অর্থাৎ চালানো যায়। উপরের ধাপগুলো আগে থেকে নয়, প্রথমবার কেউ চালালে তখন তৈরি হয়; তাই কেউ এমন কোয়ালিটির জন্য অপেক্ষা করে না যা এখনও কেউ দেখতেই চায়নি।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="Take the signed playback URL" bn="সাইন করা প্লেব্যাক লিংক নিন" /></h3>
				<p>
					<T
						as="span"
						en="GET /v1/assets/:id returns playback URLs carrying an expiry and a signature. One signature authorises the manifest, every segment, the content key, the poster and the thumbnails. They expire after four hours; re-fetch the asset for a fresh set."
						bn="GET /v1/assets/:id দিলে মেয়াদ আর সিগনেচারসহ প্লেব্যাক লিংক পাওয়া যায়। একটি সিগনেচারই ম্যানিফেস্ট, প্রতিটি সেগমেন্ট, কনটেন্ট কি, পোস্টার আর থাম্বনেইল — সবকিছুর অনুমতি দেয়। চার ঘণ্টা পর মেয়াদ শেষ; নতুন লিংকের জন্য অ্যাসেটটা আবার নিন।"
					/>
				</p>
			</li>
			<li>
				<h3><T as="span" en="Listen for the webhook" bn="ওয়েবহুক শুনুন" /></h3>
				<p>
					<T
						as="span"
						en="asset.ready, asset.failed and rendition.ready are delivered with an HMAC signature over the raw body. Verify it. Errors everywhere carry a stable machine code and a human message, translated when the request asks for Bangla."
						bn="asset.ready, asset.failed আর rendition.ready ইভেন্টগুলো আসে বডির উপর HMAC সিগনেচারসহ। সেটি যাচাই করে নেবেন। সব জায়গায় এররের সঙ্গে থাকে একটি স্থির কোড আর মানুষের পড়ার মতো বার্তা — রিকোয়েস্টে বাংলা চাইলে বাংলাতেই।"
					/>
				</p>
			</li>
		</ol>
	</div>
</section>

<section>
	<div class="wrap">
		<p class="eyebrow"><T en="Who it is for" bn="কাদের জন্য" /></p>
		<div class="grid who-grid">
			<a class="card who" href="/edtech/">
				<Icon name="book" size={22} />
				<h3><T as="span" en="Course platforms" bn="কোর্স প্ল্যাটফর্ম" /></h3>
				<p>
					<T
						as="span"
						en="Long lectures and screen recordings, students on cheap phones, and piracy as the number one worry."
						bn="লম্বা লেকচার আর স্ক্রিন রেকর্ডিং, শিক্ষার্থীরা কমদামি ফোনে, আর সবচেয়ে বড় দুশ্চিন্তা পাইরেসি।"
					/>
				</p>
				<span class="more"><T en="For edtech" bn="এডটেকের জন্য" /><Icon name="arrow" size={16} /></span>
			</a>
			<a class="card who" href="/media/">
				<Icon name="bolt" size={22} />
				<h3><T as="span" en="News and media" bn="সংবাদ ও মিডিয়া" /></h3>
				<p>
					<T
						as="span"
						en="A clip live in under a minute beats a clip that is perfect in six. Turnaround is the product."
						bn="এক মিনিটের মধ্যে লাইভ হওয়া ক্লিপ ছয় মিনিটে নিখুঁত হওয়া ক্লিপের চেয়ে ভালো। দ্রুততাই এখানে আসল জিনিস।"
					/>
				</p>
				<span class="more"><T en="For media" bn="মিডিয়ার জন্য" /><Icon name="arrow" size={16} /></span>
			</a>
			<div class="card who">
				<Icon name="layers" size={22} />
				<h3><T as="span" en="OTT and entertainment" bn="ওটিটি ও বিনোদন" /></h3>
				<p>
					<T
						as="span"
						en="The one profile where 1080p by default makes sense, on BDIX broadband. Studio DRM and live are not built yet — ask before you plan around them."
						bn="একমাত্র এই প্রোফাইলেই ডিফল্টে ১০৮০পি যুক্তিসঙ্গত, BDIX ব্রডব্যান্ডে। স্টুডিও DRM আর লাইভ এখনও তৈরি হয়নি — এগুলো ধরে পরিকল্পনা করার আগে জিজ্ঞেস করে নিন।"
					/>
				</p>
			</div>
			<div class="card who">
				<Icon name="code" size={22} />
				<h3><T as="span" en="Developers" bn="ডেভেলপার" /></h3>
				<p>
					<T
						as="span"
						en="An HTTP API with an API key. No shared database, no direct object-storage access, no internal endpoints. If something is not in the API, that is a gap in the API."
						bn="API কি দিয়ে চলা একটি HTTP API। কোনো শেয়ার করা ডেটাবেজ নেই, স্টোরেজে সরাসরি ঢোকা নেই, ভেতরের গোপন এন্ডপয়েন্ট নেই। কিছু যদি API-তে না থাকে, সেটা API-এরই ঘাটতি।"
					/>
				</p>
				<a class="more" href="/docs/"><T en="Read the docs" bn="ডকুমেন্টেশন" /><Icon name="arrow" size={16} /></a>
			</div>
		</div>
	</div>
</section>

<section class="alt">
	<div class="wrap two">
		<div>
			<p class="eyebrow"><T en="Encryption" bn="এনক্রিপশন" /></p>
			<h2><T as="span" en="Every stream is encrypted, and the token is per viewer" bn="প্রতিটি স্ট্রিম এনক্রিপ্টেড, আর টোকেন প্রতি দর্শকের আলাদা" /></h2>
			<p>
				<T
					as="span"
					en="Media is packaged with SAMPLE-AES (cbcs). The player fetches the content key from the manifest using the same short-lived signature that authorised the manifest, so a copied URL stops working when it expires. shaka-player and hls.js handle this without custom code."
					bn="মিডিয়া প্যাকেজ করা হয় SAMPLE-AES (cbcs) দিয়ে। প্লেয়ার কনটেন্ট কি নেয় ম্যানিফেস্ট থেকে, সেই একই স্বল্পমেয়াদি সিগনেচার ব্যবহার করে — ফলে কেউ লিংক কপি করে নিলেও মেয়াদ শেষে সেটা আর কাজ করে না। shaka-player আর hls.js আলাদা কোড ছাড়াই এটা সামলায়।"
				/>
			</p>
			<p>
				<T
					as="span"
					en="Packaging is CENC-shaped, so studio DRM later becomes key management rather than re-packaging a library. Studio DRM itself — Widevine, FairPlay, PlayReady — is not built."
					bn="প্যাকেজিং CENC-কাঠামোয় করা, তাই পরে স্টুডিও DRM যোগ করা মানে পুরো লাইব্রেরি আবার প্যাকেজ করা নয়, শুধু কি ব্যবস্থাপনা। তবে স্টুডিও DRM — Widevine, FairPlay, PlayReady — এখনও তৈরি হয়নি।"
				/>
			</p>
		</div>
		<div>
			<p class="eyebrow"><T en="Where this actually is" bn="এখন আসলে কতদূর" /></p>
			<h2 class="sr-only"><T as="span" en="Build status" bn="তৈরির অবস্থা" /></h2>
			<ul class="status">
				<li class="yes">
					<Icon name="check" size={17} /><T
						as="span"
						en="Working today: the API, upload and URL import, the mobile ladder, byte-range CMAF, HLS and DASH output, SAMPLE-AES encryption with a key endpoint, signed playback URLs, just-in-time packaging of the higher rungs, signed webhooks, usage metering, and row-level tenant isolation in Postgres."
						bn="এখনই কাজ করে: API, আপলোড ও লিংক থেকে ইমপোর্ট, মোবাইল ল্যাডার, বাইট-রেঞ্জ CMAF, HLS ও DASH আউটপুট, কি এন্ডপয়েন্টসহ SAMPLE-AES এনক্রিপশন, সাইন করা প্লেব্যাক লিংক, উপরের ধাপগুলোর জাস্ট-ইন-টাইম প্যাকেজিং, সাইন করা ওয়েবহুক, ব্যবহারের হিসাব, আর পোস্টগ্রেসে রো-লেভেল টেন্যান্ট আলাদাকরণ।"
					/>
				</li>
				<li class="no">
					<Icon name="arrow" size={17} /><T
						as="span"
						en="Not live yet: BDIX peering and the Dhaka edge node. The origin was built to sit behind it — correct range requests, ETags and cache headers — but the edge is the next phase of work, not a running service. Per-title bitrate analysis, studio DRM and live streaming are also not built."
						bn="এখনও চালু হয়নি: BDIX পিয়ারিং আর ঢাকার এজ নোড। অরিজিন তৈরিই হয়েছে এর পেছনে বসার জন্য — ঠিকঠাক রেঞ্জ রিকোয়েস্ট, ETag আর ক্যাশ হেডারসহ — কিন্তু এজ নোডটা পরের ধাপের কাজ, চালু সার্ভিস নয়। পার-টাইটেল বিটরেট বিশ্লেষণ, স্টুডিও DRM আর লাইভ স্ট্রিমিংও তৈরি হয়নি।"
					/>
				</li>
				<li class="no">
					<Icon name="arrow" size={17} /><T
						as="span"
						en="No uptime number is published here, because nothing is in production yet. When there is one it will be measured, not promised."
						bn="এখানে কোনো আপটাইমের হিসাব দেওয়া নেই, কারণ এখনও কিছুই প্রোডাকশনে চলছে না। যখন থাকবে, সেটা মাপা সংখ্যা হবে — প্রতিশ্রুতি নয়।"
					/>
				</li>
			</ul>
		</div>
	</div>
</section>

<section class="closing">
	<div class="wrap">
		<h2><T as="span" en="If you run video in Bangladesh, talk to us early" bn="বাংলাদেশে ভিডিও নিয়ে কাজ করলে আগেভাগেই কথা বলুন" /></h2>
		<p class="lede">
			<T
				as="span"
				en="There is no signup form, no waitlist counter and no trial clock, because none of those would be honest yet. What is useful right now is a conversation about your library, your turnaround, and what your viewers' data actually costs them."
				bn="এখানে কোনো সাইনআপ ফর্ম নেই, অপেক্ষমাণ তালিকার সংখ্যা নেই, ট্রায়ালের ঘড়িও নেই — কারণ এখন এর কোনোটাই সৎ হতো না। এই মুহূর্তে কাজের জিনিস হলো একটা আলাপ: আপনার ভিডিও লাইব্রেরি, আপনার কত দ্রুত দরকার, আর আপনার দর্শকের ডেটায় আসলে কত খরচ হয়।"
			/>
		</p>
		<div class="cta">
			<a class="btn btn-primary" href="/pricing/"><T en="How billing will work" bn="বিলিং কীভাবে হবে" /></a>
			<a class="btn" href="/about/"><T en="What this project believes" bn="এই প্রকল্প কী বিশ্বাস করে" /></a>
		</div>
	</div>
</section>

<style>
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		overflow: hidden;
		clip-path: inset(50%);
		white-space: nowrap;
	}
	.hero {
		padding-block: clamp(3rem, 8vw, 6rem) clamp(2rem, 5vw, 3.5rem);
		background:
			radial-gradient(80ch 40ch at 12% -10%, var(--brass-soft), transparent 70%);
	}
	.hero h1 {
		max-width: 22ch;
	}
	.cta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.7rem;
		margin-top: 1.8rem;
	}
	.honest {
		margin-top: 2rem;
		padding-top: 1.1rem;
		border-top: 1px dashed var(--line);
		color: var(--muted);
		font-size: 0.88rem;
		max-width: 70ch;
	}
	.facts {
		padding-top: 0;
	}
	.facts .grid,
	.who-grid {
		grid-template-columns: repeat(auto-fit, minmax(228px, 1fr));
	}
	.fact .ficon {
		color: var(--brass);
		margin-bottom: 0.7rem;
	}
	.fact.save .ficon {
		color: var(--save);
	}
	.fnum {
		font-family: var(--mono);
		font-size: 2.1rem;
		line-height: 1;
		margin: 0 0 0.4rem;
		letter-spacing: -0.03em;
	}
	.fnum.word {
		font-family: var(--font);
		font-size: 1.55rem;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.fnum span {
		font-size: 0.95rem;
		color: var(--muted);
		margin-inline-start: 0.15em;
	}
	.fact.save .fnum {
		color: var(--save);
	}
	.fh {
		font-size: 1.02rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
	}
	.fact p:last-child {
		margin: 0;
		color: var(--muted);
		font-size: 0.92rem;
	}
	.ladderwrap {
		margin: 1.5rem 0;
	}
	.nowrap {
		white-space: nowrap;
	}
	.alt {
		background: var(--raise);
		border-block: 1px solid var(--line);
	}
	.steps {
		counter-reset: s;
		list-style: none;
		padding: 0;
		margin: 1.5rem 0 0;
		display: grid;
		gap: 1.2rem;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
	}
	.steps li {
		counter-increment: s;
		padding-top: 2.4rem;
		position: relative;
	}
	.steps li::before {
		content: counter(s);
		position: absolute;
		top: 0;
		inset-inline-start: 0;
		font-family: var(--mono);
		font-size: 0.8rem;
		width: 28px;
		height: 28px;
		display: grid;
		place-items: center;
		border-radius: 999px;
		background: var(--brass-soft);
		color: var(--brass);
	}
	.steps p {
		color: var(--muted);
		font-size: 0.92rem;
		margin: 0;
	}
	.who {
		display: block;
		text-decoration: none;
		color: inherit;
	}
	a.who:hover {
		border-color: var(--brass);
	}
	.who :global(svg) {
		color: var(--brass);
	}
	.who h3 {
		margin: 0.6rem 0 0.4rem;
	}
	.who p {
		color: var(--muted);
		font-size: 0.92rem;
	}
	.more {
		display: inline-flex;
		align-items: center;
		gap: 0.35rem;
		color: var(--brass);
		font-size: 0.88rem;
		font-weight: 600;
		text-decoration: none;
	}
	.two {
		display: grid;
		gap: 2.5rem;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
	}
	.two p {
		color: var(--muted);
	}
	.status {
		list-style: none;
		padding: 0;
		margin: 0;
		display: grid;
		gap: 1rem;
	}
	.status li {
		display: flex;
		gap: 0.6rem;
		font-size: 0.92rem;
		color: var(--muted);
	}
	.status :global(svg) {
		flex: 0 0 auto;
		margin-top: 0.32rem;
	}
	.status .yes :global(svg) {
		color: var(--save);
	}
	.status .no :global(svg) {
		color: var(--brass);
	}
	.closing {
		text-align: center;
	}
	.closing .lede {
		margin-inline: auto;
	}
	.closing .cta {
		justify-content: center;
	}
</style>
