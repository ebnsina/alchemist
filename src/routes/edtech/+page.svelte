<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="Alchemist for course platforms"
	description="Lectures that play on a cheap phone without stalling, that do not drain a student's data, and that cannot simply be downloaded and passed around."
/>

<section class="shell open">
	<h1>
		<T
			as="span"
			en="Your two real problems are copying and your students' data bill."
			bn="আপনার আসল সমস্যা দুটি — কোর্স কপি হয়ে যাওয়া আর শিক্ষার্থীর ডেটার বিল।"
		/>
	</h1>
	<p class="lede quiet lead">
		<T
			as="span"
			en="Everything else about putting a lecture online has been solved for years. These two have not, and both are harder here than in the countries the big video companies designed for."
			bn="লেকচার অনলাইনে তোলার বাকি সব সমস্যার সমাধান বহু আগেই হয়ে গেছে। এই দুটোর হয়নি — আর বড় ভিডিও কোম্পানিগুলো যেসব দেশের কথা ভেবে বানিয়েছে, সেখানকার চেয়ে এখানে দুটোই কঠিন।"
		/>
	</p>
</section>

<section class="shell bay">
	<div class="split split--even">
		<div class="flow">
			<h2><T as="span" en="Courses getting copied" bn="কোর্স কপি হয়ে যাওয়া" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Every video is scrambled, and the key to unscramble it is issued for one viewing rather than once per course. A link a student copies into a Telegram group stops working within hours instead of never."
					bn="প্রতিটি ভিডিও এলোমেলো করে রাখা হয়, আর খোলার চাবি দেওয়া হয় একবার দেখার জন্য — পুরো কোর্সের জন্য একবার নয়। কোনো শিক্ষার্থী লিংক কপি করে টেলিগ্রাম গ্রুপে দিলে সেটা কয়েক ঘণ্টার মধ্যেই অকেজো হয়ে যায় — কখনওই নয়, এমনটা নয়।"
				/>
			</p>
			<p class="cav">
				<T
					as="span"
					en="It does not stop somebody recording their own screen. Only the copy protection the film studios use does, and we have not built that. We would rather tell you than sell you the word unbreakable."
					bn="কেউ নিজের স্ক্রিন রেকর্ড করলে এটা ঠেকায় না। কেবল ফিল্ম স্টুডিওগুলোর কপি-সুরক্ষাই সেটা পারে, আর তা আমরা বানাইনি। ‘ভাঙা অসম্ভব’ শব্দটা বেচার চেয়ে আপনাকে কথাটা বলে দেওয়াই ভালো।"
				/>
			</p>
		</div>
		<div class="flow">
			<h2><T as="span" en="The data bill" bn="ডেটার বিল" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="A lecture is a person talking or a screen being shared. It squeezes down much further than film footage does, and most video services are tuned for film. An hour of lecture at our default size costs a student about 270 MB."
					bn="লেকচার মানে একজন মানুষ কথা বলছেন বা একটা স্ক্রিন দেখাচ্ছেন। সিনেমার ফুটেজের চেয়ে এটা অনেক বেশি চেপে ছোট করা যায়, অথচ বেশিরভাগ ভিডিও সার্ভিস তৈরি সিনেমার কথা ভেবে। আমাদের ডিফল্ট আকারে এক ঘণ্টার লেকচার দেখতে শিক্ষার্থীর লাগে প্রায় ২৭০ মেগাবাইট।"
				/>
			</p>
			<p class="cav">
				<T
					as="span"
					en="We expect lectures and screen recordings to come down another forty to sixty per cent once we finish sizing each video individually. That is what we expect from this kind of material, not something we have measured on your videos — that work is not finished, and we will not quote you a saving we have not measured."
					bn="প্রতিটি ভিডিওর আকার আলাদা করে ঠিক করার কাজটা শেষ হলে লেকচার আর স্ক্রিন রেকর্ডিং আরও চল্লিশ থেকে ষাট শতাংশ নামবে বলে আশা করি। এটা এই ধরনের কনটেন্ট থেকে আমাদের প্রত্যাশা, আপনার ভিডিওতে মেপে পাওয়া ফল নয় — কাজটা শেষ হয়নি, আর না মেপে আমরা কোনো সাশ্রয়ের সংখ্যা বলব না।"
				/>
			</p>
		</div>
	</div>
</section>

<section class="shell bay">
	<h2 class="wide"><T as="span" en="Things that matter in a classroom and nowhere else" bn="যেগুলো কেবল ক্লাসরুমেই জরুরি" /></h2>
	<dl class="pts">
		<div>
			<dt><T as="span" en="Audio on its own" bn="শুধু অডিও" /></dt>
			<dd>
				<T
					as="span"
					en="Most of a lecture is the talking. A student revising can listen instead of watch, for roughly a tenth of the data. On a metered connection that is the difference between finishing the course and rationing it."
					bn="লেকচারের বেশিরভাগটাই আসলে কথা। রিভিশনের সময় শিক্ষার্থী দেখার বদলে শুনতে পারেন, ডেটা লাগে প্রায় দশ ভাগের এক ভাগ। মিটারড কানেকশনে এটাই ঠিক করে দেয় — কোর্সটা শেষ হবে, নাকি মেপে মেপে দেখতে হবে।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Publishable before it is finished" bn="পুরো শেষ হওয়ার আগেই প্রকাশ করা যায়" /></dt>
			<dd>
				<T
					as="span"
					en="A video becomes watchable as soon as the smaller sizes are ready. A teacher uploading at eleven the night before class does not sit and wait for the rest."
					bn="ছোট আকারগুলো তৈরি হলেই ভিডিও দেখার উপযোগী হয়ে যায়। ক্লাসের আগের রাত এগারোটায় আপলোড করা শিক্ষককে বাকিটার জন্য বসে থাকতে হয় না।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Bangla that displays properly" bn="বাংলা যেন ঠিকভাবে দেখায়" /></dt>
			<dd>
				<T
					as="span"
					en="Bangla titles, file names and subtitles survive from your upload to the student's screen. Conjunct letters keep their shape instead of turning into empty boxes, which is a real and common failure."
					bn="বাংলা শিরোনাম, ফাইলের নাম আর সাবটাইটেল আপনার আপলোড থেকে শিক্ষার্থীর পর্দা পর্যন্ত অক্ষত থাকে। যুক্তাক্ষর ভেঙে ফাঁকা বাক্স হয়ে যায় না — এটা সত্যিই ঘটে, আর প্রায়ই ঘটে।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="A data saver the student can see" bn="শিক্ষার্থীর চোখে পড়ার মতো ডেটা সেভার" /></dt>
			<dd>
				<T
					as="span"
					en="The player has a visible button that holds the video at the smaller size and shows what an hour will cost. Hidden in a settings menu it is a feature; in plain sight it is a reason to trust you."
					bn="প্লেয়ারে চোখে পড়ার মতো একটি বোতাম আছে, যেটি ভিডিওকে ছোট আকারেই আটকে রাখে আর দেখায় এক ঘণ্টায় কত খরচ হবে। সেটিংস মেনুতে লুকানো থাকলে এটা নিছক ফিচার; সামনে থাকলে এটা আপনাকে বিশ্বাস করার কারণ।"
				/>
			</dd>
		</div>
	</dl>
	<div class="tail flow">
		<p class="fine">
			<T
				as="span"
				en="Before you plan a term around this: the servers inside Bangladesh are the next piece of work and are not running yet. Everything on this site about videos starting instantly and not counting against a data allowance describes what those servers do. Today they would be served from abroad."
				bn="একটা সেমিস্টারের পরিকল্পনা করার আগে: দেশের ভেতরের সার্ভারগুলো পরের ধাপের কাজ, এখনও চালু হয়নি। এই সাইটে ভিডিও সঙ্গে সঙ্গে চালু হওয়া আর ডেটা লিমিটের হিসাবে না আসা নিয়ে যা যা বলা, সবই ওই সার্ভারগুলোর কাজ। আজ ভিডিও যেত বিদেশ থেকেই।"
			/>
		</p>
		<div class="row">
			<a class="btn btn--solid" href="/pricing/"><T en="What it costs" bn="খরচ কত" /></a>
			<a class="btn" href="/docs/"><T en="For developers" bn="ডেভেলপারদের জন্য" /></a>
		</div>
	</div>
</section>

<style>
	.open {
		padding-block: clamp(2.5rem, 1.5rem + 4vw, 4.5rem) 0;
	}
	.open h1 {
		max-width: 17ch;
		margin-bottom: 1.4rem;
	}
	.lead {
		max-width: 56ch;
	}
	.cav {
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		font-size: var(--step-0);
		color: var(--graphite);
	}
	.wide {
		max-width: 22ch;
		margin-bottom: 2rem;
	}
	.pts {
		margin: 0;
		display: grid;
		gap: 0 clamp(1.5rem, 4vw, 3.5rem);
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
	}
	.pts dt {
		font-weight: 600;
		padding-top: 1rem;
		border-top: 1px solid var(--rule);
		margin-bottom: 0.35rem;
	}
	.pts dd {
		margin: 0 0 1.6rem;
		color: var(--graphite);
		font-size: var(--step-0);
	}
	.tail {
		margin-top: 1.2rem;
		padding-top: 1.4rem;
		border-top: 1px solid var(--rule-firm);
	}
</style>
