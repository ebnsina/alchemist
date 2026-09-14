<script lang="ts">
	import Seo from '$lib/Seo.svelte';
	import T from '$lib/T.svelte';
</script>

<Seo
	title="Alchemist for news sites"
	description="A clip online in under a minute beats a clip that is perfect in six. Videos become publishable before they are fully prepared, and they keep playing when the country's international links go down."
/>

<section class="shell open">
	<h1>
		<T
			as="span"
			en="A clip online in forty seconds beats a clip that is perfect in six minutes."
			bn="চল্লিশ সেকেন্ডে প্রকাশ হওয়া ক্লিপ ছয় মিনিটে নিখুঁত হওয়া ক্লিপকে হারিয়ে দেয়।"
		/>
	</h1>
	<p class="lede quiet lead">
		<T
			as="span"
			en="In a newsroom the thing that prepares your video is not a quality tool, it is a clock. Ours is built so the clock stops early: the smaller sizes are ready first, and the video can go out then."
			bn="নিউজরুমে ভিডিও তৈরি করার যন্ত্রটা আসলে কোয়ালিটির যন্ত্র নয়, ঘড়ি। আমাদেরটা এমনভাবে বানানো যাতে ঘড়ি আগেই থামে: ছোট আকারগুলো আগে তৈরি হয়, আর তখনই ভিডিওটা ছেড়ে দেওয়া যায়।"
		/>
	</p>
</section>

<section class="shell bay">
	<dl class="pts">
		<div>
			<dt><T as="span" en="Publish before it is fully prepared" bn="পুরো তৈরি হওয়ার আগেই প্রকাশ" /></dt>
			<dd>
				<T
					as="span"
					en="A video becomes watchable as soon as the smaller sizes exist, not when everything is finished. The larger sizes are made the first time a reader asks for one — once, however many arrive at the same moment."
					bn="ছোট আকারগুলো তৈরি হলেই ভিডিও দেখার উপযোগী হয়, সবকিছু শেষ হলে নয়। বড় আকারগুলো তৈরি হয় প্রথমবার কোনো পাঠক চাইলে — একবারই, একসঙ্গে যত পাঠকই আসুন।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="It keeps working when the cables cut" bn="ক্যাবল কাটলেও চলতে থাকে" /></dt>
			<dd>
				<T
					as="span"
					en="Bangladesh reaches the rest of the world through a handful of undersea cables, and a cut has slowed the whole country more than once. That tends to happen exactly when a story is breaking. Video kept inside the country does not travel those cables."
					bn="বাংলাদেশ বাকি দুনিয়ার সঙ্গে যুক্ত হাতেগোনা কয়েকটি সমুদ্রতলের ক্যাবল দিয়ে, আর একটা কাটা পড়ে একাধিকবার সারা দেশ ধীর হয়ে গেছে। ঠিক বড় খবরের সময়েই এমনটা ঘটে। দেশের ভেতরে রাখা ভিডিও ওই ক্যাবল দিয়ে যায়ই না।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Bring what you already have" bn="যা আছে তা-ই নিয়ে আসুন" /></dt>
			<dd>
				<T
					as="span"
					en="Give us a link and we fetch the file ourselves, so you can try this without rebuilding how you already publish. A whole existing library can be brought over at once."
					bn="একটা লিংক দিন, ফাইলটা আমরা নিজেরাই নিয়ে নেব — তাই এখন যেভাবে প্রকাশ করেন সেটা নতুন করে না বানিয়েও এটা যাচাই করতে পারেন। পুরোনো পুরো লাইব্রেরিটা একসঙ্গেও আনা যায়।"
				/>
			</dd>
		</div>
		<div>
			<dt><T as="span" en="Your readers are on a phone and on a budget" bn="আপনার পাঠক ফোনে, আর হিসেব করে চলেন" /></dt>
			<dd>
				<T
					as="span"
					en="Nearly nine in ten internet users here are on mobile data, by BTRC's June 2026 count. A reader standing in a crowd on patchy 4G wants the clip to start, not to be beautiful, so we favour speed over size."
					bn="বিটিআরসির ২০২৬ সালের জুনের হিসাবে এখানে প্রায় দশজনে নয়জন ইন্টারনেট ব্যবহারকারী মোবাইল ডেটায়। ভিড়ের মধ্যে দুর্বল 4G-তে দাঁড়ানো পাঠক চান ক্লিপটা চালু হোক, সুন্দর হোক তা নয় — তাই আমরা আকারের চেয়ে গতিকে এগিয়ে রাখি।"
				/>
			</dd>
		</div>
	</dl>
</section>

<section class="shell bay">
	<div class="split">
		<div class="flow">
			<h2><T as="span" en="What is not here yet" bn="যা এখনও নেই" /></h2>
			<p class="quiet">
				<T
					as="span"
					en="Live streaming is not built. If you need it — and in Bangladesh that usually means sport — say so early, because it changes the order of the work rather than slotting into it."
					bn="লাইভ স্ট্রিমিং তৈরি হয়নি। দরকার হলে — বাংলাদেশে যার মানে সাধারণত খেলা — আগেভাগে বলবেন, কারণ এটা কাজের ক্রমই বদলে দেয়, চুপচাপ কোনো ফাঁকে ঢোকে না।"
				/>
			</p>
			<div class="row">
				<a class="btn btn--solid" href="/pricing/"><T en="What it costs" bn="খরচ কত" /></a>
				<a class="btn" href="/docs/"><T en="For developers" bn="ডেভেলপারদের জন্য" /></a>
			</div>
		</div>
	</div>
</section>

<style>
	.open {
		padding-block: clamp(2.5rem, 1.5rem + 4vw, 4.5rem) 0;
	}
	.open h1 {
		max-width: 19ch;
		margin-bottom: 1.4rem;
	}
	.lead {
		max-width: 56ch;
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
	.row {
		margin-top: 1.6rem;
	}
</style>
