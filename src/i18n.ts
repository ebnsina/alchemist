// Bangla is the default because the audience is. English is one tap away.

export type Lang = 'bn' | 'en';

export const LANGS: Lang[] = ['bn', 'en'];

type Dict = Record<string, string>;

const bn: Dict = {
  play: 'চালান',
  pause: 'থামান',
  replay: 'আবার চালান',
  mute: 'শব্দ বন্ধ',
  unmute: 'শব্দ চালু',
  volume: 'শব্দের মাত্রা',
  fullscreen: 'পূর্ণ পর্দা',
  exitFullscreen: 'পূর্ণ পর্দা থেকে বের হন',
  settings: 'সেটিংস',
  quality: 'ছবির মান',
  auto: 'স্বয়ংক্রিয়',
  captions: 'সাবটাইটেল',
  captionsOff: 'বন্ধ',
  language: 'ভাষা',
  seek: 'সময় বেছে নিন',
  loading: 'ভিডিও আসছে…',
  buffering: 'বাফার হচ্ছে…',

  dataSaver: 'ডেটা সাশ্রয়',
  dataSaverOn: 'ডেটা সাশ্রয় চালু',
  dataSaverOff: 'ডেটা সাশ্রয় বন্ধ',
  dataSaverHint: 'ছবির মান ৩৬০পি-তে রাখে, যাতে আপনার ডেটা কম খরচ হয়।',
  perHour: 'ঘণ্টায় প্রায়',
  megabytes: 'এমবি',

  errTitle: 'ভিডিও চালানো যাচ্ছে না',
  errExpired: 'এই ভিডিওর লিংকের মেয়াদ শেষ। পাতাটি আবার লোড করুন।',
  errNetwork: 'ইন্টারনেট সংযোগ পাওয়া যাচ্ছে না। সংযোগ দেখে আবার চেষ্টা করুন।',
  errUnsupported: 'আপনার ব্রাউজারে এই ভিডিও চলবে না। ক্রোম বা ফায়ারফক্সের নতুন সংস্করণে চেষ্টা করুন।',
  errNotFound: 'ভিডিওটি পাওয়া যায়নি।',
  errDrm: 'এই ডিভাইসে সুরক্ষিত ভিডিও চালানো যাচ্ছে না।',
  errKeySystem: 'এই ব্রাউজারে সুরক্ষিত ভিডিও চলে না। কম্পিউটারে ক্রোম, ফায়ারফক্স বা এজ দিয়ে একই লিংক খুলুন।',
  errGeneric: 'কিছু একটা ভুল হয়েছে। আবার চেষ্টা করুন।',
  retry: 'আবার চেষ্টা করুন',
  offline: 'আপনি এখন অফলাইনে। সংযোগ ফিরলে ভিডিও চলতে থাকবে।',
};

const en: Dict = {
  play: 'Play',
  pause: 'Pause',
  replay: 'Play again',
  mute: 'Mute',
  unmute: 'Unmute',
  volume: 'Volume',
  fullscreen: 'Full screen',
  exitFullscreen: 'Leave full screen',
  settings: 'Settings',
  quality: 'Picture quality',
  auto: 'Automatic',
  captions: 'Subtitles',
  captionsOff: 'Off',
  language: 'Language',
  seek: 'Jump to a time',
  loading: 'Getting the video…',
  buffering: 'Loading more…',

  dataSaver: 'Data Saver',
  dataSaverOn: 'Data Saver is on',
  dataSaverOff: 'Data Saver is off',
  dataSaverHint: 'Keeps the picture at 360p so you use less data.',
  perHour: 'about',
  megabytes: 'MB an hour',

  errTitle: 'This video will not play',
  errExpired: 'This video link has expired. Reload the page to keep watching.',
  errNetwork: 'We cannot reach the internet. Check your connection and try again.',
  errUnsupported: 'This browser cannot play the video. Try a recent Chrome or Firefox.',
  errNotFound: 'We could not find this video.',
  errDrm: 'This device cannot play protected video.',
  errKeySystem: 'This browser cannot play protected video. Open the same link in Chrome, Firefox or Edge on a computer.',
  errGeneric: 'Something went wrong. Please try again.',
  retry: 'Try again',
  offline: 'You are offline right now. The video will continue when you reconnect.',
};

const DICTS: Record<Lang, Dict> = { bn, en };

/** Picks bn or en from an explicit choice, then the browser, then bn. */
export function negotiateLang(explicit?: string | null, navLangs?: readonly string[]): Lang {
  const wanted = [explicit, ...(navLangs ?? [])];
  for (const raw of wanted) {
    if (!raw) continue;
    const base = String(raw).toLowerCase().split(/[-_]/)[0];
    if (base === 'bn' || base === 'bd') return 'bn';
    if (base === 'en') return 'en';
  }
  return 'bn';
}

export function makeT(lang: Lang) {
  return (key: keyof typeof bn | string): string =>
    DICTS[lang][key] ?? DICTS.en[key] ?? String(key);
}

export type T = ReturnType<typeof makeT>;

/** Bangla digits for Bangla UI; Intl does the work, we do not hand-roll it. */
export function formatNumber(lang: Lang, n: number, opts?: Intl.NumberFormatOptions): string {
  return new Intl.NumberFormat(lang === 'bn' ? 'bn-BD' : 'en-US', opts).format(n);
}

/** Timecodes stay in Latin digits in both languages: they sit in a monospace column. */
export function formatTime(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds < 0) seconds = 0;
  const total = Math.floor(seconds);
  const h = Math.floor(total / 3600);
  const m = Math.floor((total % 3600) / 60);
  const s = total % 60;
  const pad = (v: number) => String(v).padStart(2, '0');
  return h > 0 ? `${h}:${pad(m)}:${pad(s)}` : `${m}:${pad(s)}`;
}
