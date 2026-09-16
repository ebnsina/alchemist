#!/usr/bin/env bash
# Builds a local stand-in for one ready Alchemist asset: BD-mobile ladder, aligned
# GOPs, single-file CMAF addressed by byte range, SAMPLE-AES (cbcs) encrypted.
# Same shape the real pipeline writes, so what plays here plays there.
#
# Needs ffmpeg and shaka-packager on PATH.
set -euo pipefail

OUT=${1:-fixture/cmaf/t_dev/a_dev}
GOP=4
FPS=24
KEY_ID=${KEY_ID:-9eb4050de44b4802932e27d75083e0bd}
KEY=${KEY:-166634c675823c356a6a1ea56f4da1cd}

rm -rf "$OUT"; mkdir -p "$OUT" "$(dirname "$OUT")/../.."
SRC=$(mktemp -d)/src.mp4
trap 'rm -rf "$(dirname "$SRC")"' EXIT

echo "1/4  source clip"
ffmpeg -hide_banner -loglevel error -y \
  -f lavfi -i "testsrc2=size=1280x720:rate=$FPS:duration=24" \
  -f lavfi -i "sine=frequency=440:duration=24" \
  -c:v libx264 -preset veryfast -crf 20 -pix_fmt yuv420p \
  -c:a aac -b:a 96k -shortest "$SRC"

# H.264 Main for the mobile rungs: universal hardware decode on the BD fleet.
rung() {
  local h=$1 kbps=$2 profile=$3
  ffmpeg -hide_banner -loglevel error -y -i "$SRC" \
    -an -c:v libx264 -profile:v "$profile" -preset veryfast \
    -vf "scale=-2:$h" -b:v "${kbps}k" -maxrate "$((kbps*3/2))k" -bufsize "$((kbps*2))k" \
    -r $FPS -g $((GOP*FPS)) -keyint_min $((GOP*FPS)) -sc_threshold 0 \
    -movflags +faststart "$OUT/${h}p.mp4"
}
echo "2/4  ladder — 144p 240p 360p 480p 720p"
rung 144 150 baseline; rung 240 300 main; rung 360 600 main
rung 480 1000 main;    rung 720 1800 main
ffmpeg -hide_banner -loglevel error -y -i "$SRC" -vn -c:a aac -b:a 64k -ac 1 "$OUT/audio.mp4"

echo "3/4  package — CMAF byte-range, SAMPLE-AES cbcs"
args=()
for h in 144 240 360 480 720; do
  args+=("in=$OUT/${h}p.mp4,stream=video,output=$OUT/${h}p.cmfv,playlist_name=$OUT/${h}p.m3u8")
done
args+=("in=$OUT/audio.mp4,stream=audio,output=$OUT/audio.cmfa,playlist_name=$OUT/audio.m3u8,hls_group_id=audio,hls_name=AUDIO")
packager "${args[@]}" \
  --hls_master_playlist_output "$OUT/master.m3u8" \
  --mpd_output "$OUT/manifest.mpd" \
  --segment_duration $GOP \
  --enable_raw_key_encryption \
  --keys "label=:key_id=$KEY_ID:key=$KEY" \
  --protection_scheme cbcs \
  --clear_lead 0 \
  --hls_key_uri "key" \
  --quiet
rm -f "$OUT"/*.mp4

echo "4/4  poster and scrub sprite"
ffmpeg -hide_banner -loglevel error -y -i "$SRC" -frames:v 1 -q:v 3 "$OUT/poster.jpg"
ffmpeg -hide_banner -loglevel error -y -i "$SRC" \
  -vf "fps=1/1,scale=160:-2,tile=10x3" -frames:v 1 -q:v 4 "$OUT/sprite.jpg"
python3 - "$OUT/sprite.vtt" <<'PY'
import sys
w, h, cols, n, step = 160, 90, 10, 24, 1.0
t = lambda s: "%02d:%02d:%06.3f" % (int(s)//3600, int(s)%3600//60, s%60)
out = ["WEBVTT", ""]
for i in range(n):
    out += ["%s --> %s" % (t(i*step), t((i+1)*step)),
            "sprite.jpg#xywh=%d,%d,%d,%d" % ((i%cols)*w, (i//cols)*h, w, h), ""]
open(sys.argv[1], "w").write("\n".join(out))
PY

printf 'raw key for the origin stand-in: %s\n' "$KEY"
echo "done — $OUT"
