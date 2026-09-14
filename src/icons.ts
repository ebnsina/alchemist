// Hugeicons-style strokes, inlined. The icon package is 72 MB across 12,000 files;
// twelve 24px paths are not worth that on a player whose whole point is small bytes.
// ponytail: hand-authored in the Hugeicons stroke grid (24px, 1.6 stroke, round caps).
// Swap in the real set by replacing the path bodies below.

const wrap = (body: string) =>
  `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true" focusable="false">${body}</svg>`;

export const icons = {
  play: wrap('<path d="M6.5 5.3c0-1 1.1-1.6 2-1.1l9.3 5.6c.8.5.8 1.7 0 2.2l-9.3 5.6c-.9.5-2-.1-2-1.1z" fill="currentColor" stroke="none"/>'),
  pause: wrap('<path d="M8 4.5v15M16 4.5v15"/>'),
  replay: wrap('<path d="M20 12a8 8 0 1 1-2.6-5.9"/><path d="M20 3.5V8h-4.5"/>'),
  volume: wrap('<path d="M4 9.5h2.6L11 5.6c.6-.6 1.6-.2 1.6.7v11.4c0 .9-1 1.3-1.6.7L6.6 14.5H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1Z"/><path d="M16.2 9a4 4 0 0 1 0 6"/><path d="M18.8 6.6a7.5 7.5 0 0 1 0 10.8"/>'),
  volumeLow: wrap('<path d="M4 9.5h2.6L11 5.6c.6-.6 1.6-.2 1.6.7v11.4c0 .9-1 1.3-1.6.7L6.6 14.5H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1Z"/><path d="M16.2 9a4 4 0 0 1 0 6"/>'),
  mute: wrap('<path d="M4 9.5h2.6L11 5.6c.6-.6 1.6-.2 1.6.7v11.4c0 .9-1 1.3-1.6.7L6.6 14.5H4a1 1 0 0 1-1-1v-3a1 1 0 0 1 1-1Z"/><path d="m16.5 9.5 4 5M20.5 9.5l-4 5"/>'),
  fullscreen: wrap('<path d="M3.5 8.5V5a1.5 1.5 0 0 1 1.5-1.5h3.5M20.5 8.5V5A1.5 1.5 0 0 0 19 3.5h-3.5M3.5 15.5V19A1.5 1.5 0 0 0 5 20.5h3.5M20.5 15.5V19a1.5 1.5 0 0 1-1.5 1.5h-3.5"/>'),
  fullscreenExit: wrap('<path d="M9 3.5V7a2 2 0 0 1-2 2H3.5M15 3.5V7a2 2 0 0 0 2 2h3.5M9 20.5V17a2 2 0 0 0-2-2H3.5M15 20.5V17a2 2 0 0 1 2-2h3.5"/>'),
  settings: wrap('<circle cx="12" cy="12" r="3"/><path d="M12 2.8c.6 0 1.1.4 1.3 1l.3 1.2c.5.2 1 .5 1.4.8l1.2-.4c.6-.2 1.2 0 1.5.5l.8 1.4c.3.5.2 1.2-.3 1.6l-.9.8c0 .3.1.6.1.9s0 .6-.1.9l.9.8c.5.4.6 1 .3 1.6l-.8 1.4c-.3.5-.9.7-1.5.5l-1.2-.4c-.4.3-.9.6-1.4.8l-.3 1.2c-.2.6-.7 1-1.3 1h-1.6c-.6 0-1.1-.4-1.3-1l-.3-1.2c-.5-.2-1-.5-1.4-.8l-1.2.4c-.6.2-1.2 0-1.5-.5l-.8-1.4c-.3-.6-.2-1.2.3-1.6l.9-.8a6 6 0 0 1 0-1.8l-.9-.8c-.5-.4-.6-1-.3-1.6l.8-1.4c.3-.5.9-.7 1.5-.5l1.2.4c.4-.3.9-.6 1.4-.8l.3-1.2c.2-.6.7-1 1.3-1z"/>'),
  captions: wrap('<rect x="2.8" y="4.8" width="18.4" height="14.4" rx="3"/><path d="M10 10.2a2.4 2.4 0 1 0 0 3.6M17.5 10.2a2.4 2.4 0 1 0 0 3.6"/>'),
  dataSaver: wrap('<path d="M12 21c4.4 0 8-3.4 8-7.7 0-5-4.2-8.8-6.9-10.8a1.8 1.8 0 0 0-2.2 0C8.2 4.5 4 8.3 4 13.3 4 17.6 7.6 21 12 21Z"/><path d="M12 17.2V9.8M9.2 12.4 12 9.6l2.8 2.8"/>'),
  language: wrap('<circle cx="12" cy="12" r="9"/><path d="M3.2 9.5h17.6M3.2 14.5h17.6"/><path d="M12 3a15 15 0 0 1 0 18 15 15 0 0 1 0-18Z"/>'),
  check: wrap('<path d="m4.5 12.5 5 5 10-11"/>'),
  back: wrap('<path d="M14.5 5 8 12l6.5 7"/>'),
  alert: wrap('<circle cx="12" cy="12" r="9"/><path d="M12 7.5v5.5M12 16.3v.2"/>'),
  offline: wrap('<path d="M3 3l18 18"/><path d="M8.6 15.4a4.8 4.8 0 0 1 6.8 0M5.2 12a9.6 9.6 0 0 1 3.2-2.1M18.8 12a9.6 9.6 0 0 0-6.9-2.6M2 8.6A14.4 14.4 0 0 1 7 5.7M22 8.6a14.4 14.4 0 0 0-7.9-3.4"/><path d="M12 18.8v.2"/>'),
} as const;

export type IconName = keyof typeof icons;
