// One injected stylesheet, scoped under .alc. Kept as a string so the SDK ships as
// a single ES module with no CSS import ceremony for the host.

export const CSS = `
/* Google Sans Flex, latin only. Inlined rather than linked from
   fonts.googleapis.com so a cold BD mobile connection pays one round trip instead of
   two, and so the latin-ext / vietnamese / math / syriac subsets Google also serves
   can never be pulled. 35 KB woff2, cached a year.

   The unicode-range is what protects the Bangla rendering: Bengali codepoints are not
   in it, so the browser never tries this face for them and falls straight through to
   Noto Sans Bengali, which the system already has and which shapes conjuncts
   correctly. font-display:swap means text is readable before the font arrives. */
@font-face{
  font-family:"Google Sans Flex";
  font-style:normal;
  font-weight:400 700;
  font-stretch:100%;
  font-display:swap;
  src:url(https://fonts.gstatic.com/s/googlesansflex/v22/t5sEIQcYNIWbFgDgAAzZ34auoVyXkJCOvp3SFWJbN5hF8Ju1x6sKCyp0l9sI40swNJwInycYAJzz0m7kJ4qFQOJBOjLvDSndo0SKMpKSTzwliVdHAy4bxTDHg_ugnAakp8ubq8BIo1pdkkXZj4igdvKMDV8.woff2) format("woff2");
  unicode-range:U+0000-00FF,U+0131,U+0152-0153,U+02BB-02BC,U+02C6,U+02DA,U+02DC,U+0304,U+0308,U+0329,U+2000-206F,U+20AC,U+2122,U+2191,U+2193,U+2212,U+2215,U+FEFF,U+FFFD;
}
.alc{
  /* Alchemist: brass on near-black. Transmutation, not another blue video player. */
  --alc-bg:#0C0C0E;
  --alc-raise:#17171B;
  --alc-line:#2A2A31;
  --alc-ink:#F4F1EC;
  --alc-muted:#A19C94;
  --alc-brass:#E8A33D;
  --alc-brass-ink:#1A1206;
  --alc-save:#3FBF8F;
  --alc-danger:#E8785F;
  --alc-radius:10px;
  --alc-font:"Google Sans Flex","Noto Sans Bengali",system-ui,-apple-system,"Segoe UI",Roboto,"Noto Sans",sans-serif;
  --alc-mono:"Geist Mono",ui-monospace,SFMono-Regular,Menlo,"Roboto Mono",monospace;

  position:relative;display:block;width:100%;background:var(--alc-bg);color:var(--alc-ink);
  font-family:var(--alc-font);font-size:14px;line-height:1.45;
  border-radius:var(--alc-radius);overflow:hidden;isolation:isolate;
  -webkit-tap-highlight-color:transparent;
}
.alc *,.alc *::before,.alc *::after{box-sizing:border-box}
.alc:fullscreen,.alc:-webkit-full-screen{border-radius:0;height:100%}

.alc video{display:block;width:100%;height:100%;background:#000;object-fit:contain}
.alc-stage{position:relative;width:100%;aspect-ratio:16/9;background:#000}
.alc:fullscreen .alc-stage{aspect-ratio:auto;height:100%}

/* ---- control bar ---- */
.alc-scrim{
  position:absolute;inset:auto 0 0 0;height:56%;pointer-events:none;
  background:linear-gradient(to top,rgba(4,4,6,.92) 0%,rgba(4,4,6,.62) 42%,rgba(4,4,6,0) 100%);
  opacity:0;transition:opacity .18s ease;
}
.alc-bar{
  position:absolute;inset:auto 0 0 0;padding:0 10px 9px;display:flex;flex-direction:column;gap:4px;
  opacity:0;transform:translateY(6px);transition:opacity .18s ease,transform .18s ease;
}
.alc[data-chrome="on"] .alc-scrim{opacity:1}
.alc[data-chrome="on"] .alc-bar{opacity:1;transform:none;pointer-events:auto}
.alc[data-chrome="off"] .alc-bar{pointer-events:none}
.alc[data-chrome="off"]{cursor:none}

.alc-row{display:flex;align-items:center;gap:2px;min-height:38px}
.alc-spacer{flex:1 1 auto}

/* ---- buttons ---- */
.alc-btn{
  appearance:none;border:0;background:transparent;color:var(--alc-ink);
  width:38px;height:38px;padding:7px;border-radius:8px;cursor:pointer;
  display:inline-flex;align-items:center;justify-content:center;flex:0 0 auto;
  transition:background .12s ease,color .12s ease;
}
.alc-btn svg{width:100%;height:100%;display:block}
.alc-btn:hover{background:rgba(255,255,255,.09)}
.alc-btn[aria-pressed="true"],.alc-btn[aria-expanded="true"]{background:rgba(232,163,61,.18);color:var(--alc-brass)}
.alc-btn:focus-visible,.alc-pill:focus-visible,.alc-item:focus-visible,.alc-seek:focus-visible{
  outline:2px solid var(--alc-brass);outline-offset:2px;
}

/* ---- Data Saver: a headline control, never a settings-menu afterthought ---- */
.alc-pill{
  appearance:none;cursor:pointer;display:inline-flex;align-items:center;gap:7px;
  height:32px;padding:0 11px 0 8px;margin-left:4px;
  border:1px solid var(--alc-line);border-radius:999px;
  background:rgba(255,255,255,.05);color:var(--alc-muted);
  font:inherit;font-size:12.5px;font-weight:600;white-space:nowrap;
  transition:background .14s ease,border-color .14s ease,color .14s ease;
}
.alc-pill svg{width:17px;height:17px;flex:0 0 auto}
.alc-pill:hover{background:rgba(255,255,255,.1);color:var(--alc-ink)}
.alc-pill[aria-pressed="true"]{
  background:rgba(63,191,143,.16);border-color:rgba(63,191,143,.55);color:var(--alc-save);
}
.alc-pill-rate{font-family:var(--alc-mono);font-size:11px;font-weight:500;opacity:.85;letter-spacing:-.01em}
@media (max-width:430px){.alc-pill-label{display:none}}

/* ---- seek bar ---- */
.alc-seek{
  position:relative;display:block;width:100%;height:16px;padding:0;margin:0 2px;
  border:0;background:transparent;cursor:pointer;touch-action:none;
}
.alc-track{position:absolute;left:0;right:0;top:6px;height:4px;border-radius:2px;background:rgba(255,255,255,.22)}
.alc-buffered{position:absolute;left:0;top:0;height:100%;border-radius:2px;background:rgba(255,255,255,.3)}
.alc-played{position:absolute;left:0;top:0;height:100%;border-radius:2px;background:var(--alc-brass)}
.alc-knob{
  position:absolute;top:50%;width:12px;height:12px;margin-left:-6px;border-radius:50%;
  background:var(--alc-brass);transform:translateY(-50%) scale(0);transition:transform .13s ease;
}
.alc-seek:hover .alc-knob,.alc-seek:focus-visible .alc-knob,.alc[data-scrubbing] .alc-knob{transform:translateY(-50%) scale(1)}

.alc-preview{
  position:absolute;bottom:26px;transform:translateX(-50%);display:none;
  padding:3px;border-radius:8px;background:var(--alc-raise);border:1px solid var(--alc-line);
  box-shadow:0 8px 24px rgba(0,0,0,.55);pointer-events:none;z-index:3;
}
.alc-preview[data-show]{display:block}
.alc-preview-img{border-radius:5px;background-repeat:no-repeat;background-color:#000}
.alc-preview-time{
  display:block;text-align:center;font-family:var(--alc-mono);font-size:11px;
  padding-top:3px;color:var(--alc-muted);
}

.alc-time{
  font-family:var(--alc-mono);font-size:12px;color:var(--alc-muted);
  padding:0 8px;white-space:nowrap;font-variant-numeric:tabular-nums;
}
.alc-time b{color:var(--alc-ink);font-weight:500}

/* ---- volume ---- */
.alc-vol{display:flex;align-items:center}
.alc-vol input{
  width:0;opacity:0;transition:width .16s ease,opacity .16s ease;
  accent-color:var(--alc-brass);height:16px;margin:0;cursor:pointer;
}
.alc-vol:hover input,.alc-vol:focus-within input{width:68px;opacity:1;margin:0 6px 0 2px}
@media (hover:none){.alc-vol input{display:none}}

/* ---- menu ---- */
.alc-menu{
  position:absolute;right:10px;bottom:56px;min-width:208px;max-width:min(280px,calc(100% - 20px));
  background:var(--alc-raise);border:1px solid var(--alc-line);border-radius:12px;
  box-shadow:0 14px 40px rgba(0,0,0,.6);padding:6px;z-index:4;
  max-height:min(300px,60%);overflow-y:auto;overscroll-behavior:contain;
}
.alc-menu[hidden]{display:none}
.alc-menu-title{
  display:flex;align-items:center;gap:6px;padding:6px 8px 8px;
  font-size:11px;font-weight:700;letter-spacing:.06em;text-transform:uppercase;color:var(--alc-muted);
}
.alc-item{
  appearance:none;border:0;background:transparent;color:var(--alc-ink);font:inherit;
  display:flex;align-items:center;gap:10px;width:100%;padding:9px 10px;border-radius:8px;
  cursor:pointer;text-align:start;
}
.alc-item:hover{background:rgba(255,255,255,.07)}
.alc-item[aria-checked="true"]{color:var(--alc-brass)}
.alc-item svg{width:17px;height:17px;flex:0 0 auto;opacity:0}
.alc-item[aria-checked="true"] svg{opacity:1}
.alc-item-note{margin-inline-start:auto;font-family:var(--alc-mono);font-size:11px;color:var(--alc-muted)}
.alc-brand{display:flex;align-items:baseline;justify-content:space-between;gap:.5rem;
  margin-top:.35rem;padding:.5rem .7rem 0;border-top:1px solid rgba(255,255,255,.1)}
.alc-brand-name{font-size:11px;font-weight:600;letter-spacing:.02em;color:rgba(255,255,255,.62)}
.alc-brand-version{font-family:var(--alc-mono,ui-monospace,monospace);font-size:10px;
  font-variant-numeric:tabular-nums;color:rgba(255,255,255,.38)}
.alc-hint{padding:2px 10px 8px;font-size:12px;color:var(--alc-muted)}

/* ---- overlays ---- */
.alc-centre{
  position:absolute;inset:0;display:flex;align-items:center;justify-content:center;
  z-index:2;pointer-events:none;
}
.alc-big{
  pointer-events:auto;appearance:none;border:0;cursor:pointer;
  width:66px;height:66px;border-radius:50%;padding:19px;
  background:var(--alc-brass);color:var(--alc-brass-ink);
  box-shadow:0 6px 26px rgba(0,0,0,.5);transition:transform .14s ease;
}
.alc-big:hover{transform:scale(1.06)}
.alc-big svg{width:100%;height:100%}
.alc-big:focus-visible{outline:3px solid #fff;outline-offset:3px}

.alc-spinner{
  width:38px;height:38px;border-radius:50%;
  border:3px solid rgba(255,255,255,.2);border-top-color:var(--alc-brass);
  animation:alc-spin .8s linear infinite;
}
@keyframes alc-spin{to{transform:rotate(360deg)}}
@media (prefers-reduced-motion:reduce){
  .alc-spinner{animation-duration:2s}
  .alc-scrim,.alc-bar,.alc-knob,.alc-big{transition:none}
}

.alc-panel{
  position:absolute;inset:0;z-index:5;display:flex;flex-direction:column;
  align-items:center;justify-content:center;gap:10px;text-align:center;
  padding:26px 22px;background:rgba(9,9,11,.94);
}
.alc-panel[hidden]{display:none}
.alc-panel-icon{width:34px;height:34px;color:var(--alc-danger)}
.alc-panel h2{margin:0;font-size:16px;font-weight:700;letter-spacing:-.01em}
.alc-panel p{margin:0;max-width:38ch;font-size:13.5px;color:var(--alc-muted)}
.alc-action{
  appearance:none;border:0;cursor:pointer;font:inherit;font-weight:650;font-size:13px;
  margin-top:4px;padding:9px 18px;border-radius:999px;
  background:var(--alc-brass);color:var(--alc-brass-ink);
}
.alc-action:hover{filter:brightness(1.07)}
.alc-action:focus-visible{outline:2px solid #fff;outline-offset:2px}

.alc-toast{
  position:absolute;left:50%;top:12px;transform:translateX(-50%);z-index:6;
  display:flex;align-items:center;gap:8px;max-width:calc(100% - 24px);
  padding:7px 13px;border-radius:999px;font-size:12.5px;font-weight:600;
  background:rgba(23,23,27,.95);border:1px solid var(--alc-line);color:var(--alc-ink);
  opacity:0;transition:opacity .18s ease;pointer-events:none;
}
.alc-toast[data-show]{opacity:1}
.alc-toast svg{width:16px;height:16px;color:var(--alc-save);flex:0 0 auto}
.alc-toast[data-tone="warn"] svg{color:var(--alc-danger)}

.alc-sr{
  position:absolute!important;width:1px;height:1px;padding:0;margin:-1px;
  overflow:hidden;clip:rect(0 0 0 0);white-space:nowrap;border:0;
}

/* The viewer's own label, carried inside the signature. It names the account a leak
   came from and discourages casual resharing; it does not stop a screen recorder. */
.alc-wm{
  position:absolute;left:0;top:0;z-index:1;max-width:58%;
  padding:2px 6px;pointer-events:none;user-select:none;-webkit-user-select:none;
  white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
  font-family:var(--alc-mono);font-size:11px;font-variant-numeric:tabular-nums;
  color:rgba(255,255,255,.32);text-shadow:0 1px 2px rgba(0,0,0,.6);
  transition:transform 1.6s ease-in-out;
}
@media (prefers-reduced-motion:reduce){.alc-wm{transition:none}}

/* Captions sit above the bar when the chrome is up, and stay legible on any frame. */
.alc .shaka-text-container,.alc video::cue{font-family:var(--alc-font)}
.alc video::cue{background:rgba(0,0,0,.72);color:#fff;font-size:.95em}
`;

let injected = false;
export function injectStyles(doc: Document = document): void {
  if (injected && doc === document) return;
  if (doc.getElementById('alchemist-player-css')) return;
  const style = doc.createElement('style');
  style.id = 'alchemist-player-css';
  style.textContent = CSS;
  doc.head.appendChild(style);
  injected = true;
}
