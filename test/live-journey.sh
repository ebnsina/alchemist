#!/bin/bash
# End-to-end test of a broadcast, including the thing that is hardest to get right and
# easiest to ship broken: an encoder that drops mid-class and comes back.
#
# Everything goes through the public API with an API key, like test/journey.sh. What it
# adds is a real publisher and a real ingest server, because every live bug worth
# having was invisible to a unit test -- a playlist that says the stream ended, a
# recording that silently contains half the class, URLs the API advertises and storage
# 404s. None of those raise an error anywhere.
#
#   make storage        # terminal 1
#   make run            # terminal 2  (ALCHEMIST_LIVE_* set in .env)
#   make live-journey KEY=<api-key>
#
# Live has to be on for the tenant: update tenant_limits set live_enabled = true.
set -u
B=${B:-http://localhost:8090}
KEY=${1:?usage: live-journey.sh <api-key>}
HOST=${LIVE_HOST:-127.0.0.1}
RUN1=${RUN1:-8}   # seconds published before the drop
RUN2=${RUN2:-8}   # seconds published after coming back
GAP=${GAP:-6}     # seconds off air; must stay under live.ReconnectGrace

pass=0; fail=0
check() { if [ "$2" = "$3" ]; then echo "  ok   $1"; pass=$((pass+1)); else echo "  FAIL $1 (got '$2', want '$3')"; fail=$((fail+1)); fi }
jq_() { python3 -c "import sys,json;d=json.load(sys.stdin);print($1)" 2>/dev/null; }

for bin in ffmpeg ffprobe mediamtx curl python3; do
  command -v $bin >/dev/null || { echo "need $bin on PATH"; exit 2; }
done
curl -sf -o /dev/null "$B/healthz" || { echo "no API at $B -- run 'make run'"; exit 2; }

MTX=""
cleanup() { [ -n "$MTX" ] && kill "$MTX" 2>/dev/null; wait "$MTX" 2>/dev/null; }
trap cleanup EXIT

# Reused if one is already up. Starting a second one fails on the bound port and dies,
# while the port check goes on passing against the first -- so the run would look fine
# and be testing somebody else's ingest server.
echo "0. ingest server"
if nc -z "$HOST" 1935 2>/dev/null; then
  echo "   reusing the mediamtx already listening on 1935"
else
  mediamtx deploy/live/mediamtx.yml >/tmp/alchemist-mediamtx.log 2>&1 &
  MTX=$!
  for _ in $(seq 20); do
    sleep 0.25
    kill -0 "$MTX" 2>/dev/null || break
    nc -z "$HOST" 1935 2>/dev/null && break
  done
  if ! kill -0 "$MTX" 2>/dev/null || ! nc -z "$HOST" 1935 2>/dev/null; then
    echo "mediamtx did not start; see /tmp/alchemist-mediamtx.log"
    tail -3 /tmp/alchemist-mediamtx.log
    exit 2
  fi
  echo "   started, listening on 1935"
fi

echo "1. create a stream"
R=$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"name":"reconnect journey","protocol":"rtmp"}' "$B/v1/live-streams")
if [ "$(echo "$R" | jq_ 'd.get("error",{}).get("code","")')" = "live_not_enabled" ]; then
  echo "   live is off for this tenant: update tenant_limits set live_enabled = true"
  exit 2
fi
SID=$(echo "$R" | jq_ 'd["id"]')
SKEY=$(echo "$R" | jq_ 'd["stream_key"]')
check "stream created" "$([ -n "$SID" ] && echo yes || echo no)" "yes"
check "key shown once" "$([ -n "$SKEY" ] && echo yes || echo no)" "yes"

echo "2. arm it"
R=$(curl -s -X POST -H "Authorization: Bearer $KEY" "$B/v1/live-streams/$SID/start")
ASSET=$(echo "$R" | jq_ 'd["asset_id"]')
SERVER=$(echo "$R" | jq_ 'd.get("ingest_server","")')
KEYFIELD=$(echo "$R" | jq_ 'd.get("ingest_stream_key","")')
check "asset minted" "$([ -n "$ASSET" ] && echo yes || echo no)" "yes"
# It was the empty string, with the key buried in the server address as a placeholder,
# so a customer following the dashboard could not connect at all.
check "stream key field is fillable" "$([ -n "$KEYFIELD" ] && echo yes || echo no)" "yes"
check "server half carries no path" "$SERVER" "rtmp://$HOST:1935"
PUBLISH="$SERVER/${KEYFIELD/YOUR_STREAM_KEY/$SKEY}"

publish() {
  ffmpeg -hide_banner -loglevel error -re \
    -f lavfi -i "testsrc2=size=640x360:rate=30" -f lavfi -i "sine=frequency=440" \
    -c:v libx264 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -sc_threshold 0 \
    -c:a aac -b:a 96k -ac 2 -ar 48000 -t "$1" -f flv "$PUBLISH" >/dev/null 2>&1
}

echo "3. publish for ${RUN1}s, then drop"
publish "$RUN1"
sleep 3

echo "4. while off air"
R=$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$ASSET")
check "asset is a broadcast" "$(echo "$R" | jq_ 'd["state"]')" "live"
# Live writes a playlist and segments and nothing else. Advertising a manifest, a
# poster and a scrubbing index beside them is three URLs that 404.
check "hls offered" "$(echo "$R" | jq_ '"yes" if d["playback"].get("hls") else "no"')" "yes"
check "no dash offered" "$(echo "$R" | jq_ '"yes" if d["playback"].get("dash") else "no"')" "no"
check "no poster offered" "$(echo "$R" | jq_ '"yes" if d["playback"].get("poster") else "no"')" "no"
check "prefers hls" "$(echo "$R" | jq_ 'd["playback"]["preferred"]')" "hls"

HLS=$(echo "$R" | jq_ 'd["playback"]["hls"]')
PL=$(curl -s "$B$HLS")
# ffmpeg writes ENDLIST whenever its process exits, including on a drop. A player that
# has seen it does not come back when the segments resume.
check "playlist does not end the stream" "$(echo "$PL" | grep -c 'EXT-X-ENDLIST')" "0"
check "segments published" "$([ "$(echo "$PL" | grep -c '\.m4s')" -gt 0 ] && echo yes || echo no)" "yes"

echo "5. deleting a live stream is refused"
check "delete refused" \
  "$(curl -s -X DELETE -H "Authorization: Bearer $KEY" "$B/v1/live-streams/$SID" | jq_ 'd["error"]["code"]')" \
  "stream_on_air"

echo "6. wait ${GAP}s, then the encoder comes back"
sleep "$GAP"
publish "$RUN2"
sleep 3

R=$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$ASSET")
HLS=$(echo "$R" | jq_ 'd["playback"]["hls"]')
PL=$(curl -s "$B$HLS")
# Proof the second run appended rather than starting a new playlist over the first.
check "reconnect marked, not restarted" "$(echo "$PL" | grep -c 'EXT-X-DISCONTINUITY')" "1"
check "broadcast survived the drop" "$(echo "$R" | jq_ 'd["state"]')" "live"

echo "7. stop it"
curl -s -X POST -H "Authorization: Bearer $KEY" "$B/v1/live-streams/$SID/stop" >/dev/null

echo "8. the recording converts"
STATE=""
for _ in $(seq 60); do
  sleep 3
  STATE=$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$ASSET" | jq_ 'd["state"]')
  case "$STATE" in ready|partially_ready|failed) break ;; esac
done
check "recording became a video" "$(case "$STATE" in ready|partially_ready) echo yes ;; *) echo "$STATE" ;; esac)" "yes"

R=$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$ASSET")
DUR=$(echo "$R" | jq_ 'int(d.get("duration_seconds") or 0)')
WANT=$((RUN1 + RUN2))
# The one that matters. Appending the bytes of both runs makes a file whose timeline
# runs backwards, and everything downstream believes the first run's duration: ten
# seconds of broadcast came back as six, with no error anywhere.
check "recording holds both runs (~${WANT}s, got ${DUR}s)" \
  "$([ "$DUR" -ge $((WANT - 4)) ] && echo yes || echo no)" "yes"

# One realtime rung, and not the cheapest: a core goes on encoding at all rather than
# on the pixel count, and 144p is unreadable for the slides this is used for.
check "recorded at the live cap" \
  "$(echo "$R" | jq_ 'max(r["height"] for r in d["renditions"]) if d["renditions"] else 0')" "360"
check "dash is back on a finished recording" \
  "$(echo "$R" | jq_ '"yes" if d["playback"].get("dash") else "no"')" "yes"

echo "9. now it deletes"
check "delete accepted" \
  "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $KEY" "$B/v1/live-streams/$SID")" \
  "204"

echo
echo "$pass passed, $fail failed"
[ "$fail" -eq 0 ]
