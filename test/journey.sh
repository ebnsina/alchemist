#!/bin/bash
# End-to-end integration test from a third-party customer's point of view.
#
# Everything below goes through the public API with nothing but an API key: no
# database access, no internal endpoints, no privileged calls. If this passes, a
# customer can integrate without a dashboard existing.
#
#   make storage                       # terminal 1
#   python3 test/customer_servers.py   # terminal 2 (stands in for the customer)
#   make run                           # terminal 3
#   ./test/journey.sh <api-key>
set -u
B=${B:-http://localhost:8099}
KEY=$1
pass=0; fail=0
check() { if [ "$2" = "$3" ]; then echo "  ok   $1"; pass=$((pass+1)); else echo "  FAIL $1 (got '$2', want '$3')"; fail=$((fail+1)); fi }

echo "1. who am I?"
R=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/whoami)
check "tenant resolves" "$(echo "$R"|python3 -c 'import sys,json;print(json.load(sys.stdin)["ladder_profile"])')" "bd-mobile"

echo "2. register a webhook"
R=$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"https://127.0.0.1:8072/hook","events":["asset.ready","asset.failed"]}' $B/v1/webhooks)
echo "   https-only enforced: $(echo "$R" | python3 -c 'import sys,json;print(json.load(sys.stdin).get("error",{}).get("code","ACCEPTED"))')"
# The receiver is plain http in this harness, so register via the DB-free path below.

echo "3. import a video from a URL (migration off another provider)"
R=$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"http://127.0.0.1:8071/lecture.mp4"}' $B/v1/assets)
A=$(echo "$R"|python3 -c 'import sys,json;print(json.load(sys.stdin).get("asset_id",""))')
check "import accepted" "$([ -n "$A" ] && echo yes || echo no)" "yes"

echo "4. SSRF: a URL that redirects to cloud metadata must be refused"
R=$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"http://127.0.0.1:8071/redirect-internal"}' $B/v1/assets)
EVIL=$(echo "$R"|python3 -c 'import sys,json;print(json.load(sys.stdin).get("asset_id",""))')

echo "5. bad input is rejected with a stable code"
check "no url"      "$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{}' $B/v1/assets | python3 -c 'import sys,json;print(json.load(sys.stdin)["error"]["code"])')" "missing_url"
check "file scheme" "$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' -d '{"url":"file:///etc/passwd"}' $B/v1/assets | python3 -c 'import sys,json;print(json.load(sys.stdin)["error"]["code"])')" "invalid_url"
check "no api key"  "$(curl -s -o /dev/null -w '%{http_code}' $B/v1/assets)" "401"

echo "6. poll the import until it is playable"
# partially_ready is the expected end state: the low ladder is encoded and playable,
# while the expensive rungs are generated on first playback. See docs/04-roadmap.md.
for i in $(seq 1 60); do
  S=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$A | python3 -c 'import sys,json;print(json.load(sys.stdin)["state"])')
  case "$S" in ready|partially_ready|failed) break;; esac; sleep 4
done
PLAYABLE=no
if [ "$S" = "ready" ] || [ "$S" = "partially_ready" ]; then PLAYABLE=yes; fi
check "import playable ($S)" "$PLAYABLE" "yes"

echo "7. the SSRF attempt should have failed with a clear reason"
ES=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$EVIL | python3 -c 'import sys,json;d=json.load(sys.stdin);print(d["state"]+":"+str(d.get("error_code")))')
check "metadata fetch blocked" "$ES" "failed:source_url_not_allowed"

echo "8. the API says which URL to play"
PF=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$A | python3 -c 'import sys,json;p=json.load(sys.stdin).get("playback") or {};print(p.get("preferred",""))')
ENC=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$A | python3 -c 'import sys,json;p=json.load(sys.stdin).get("playback") or {};print(str(p.get("encrypted","")).lower())')
# Encrypted media is cenc and HLS cannot carry cenc, so the two answers must agree.
case "$ENC:$PF" in
  true:dash|false:hls) check "preferred matches encryption ($ENC)" ok ok;;
  *) check "preferred matches encryption" "$ENC:$PF" "true:dash or false:hls";;
esac

echo "9. playback works"
P=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$A)
HLS=$(echo "$P"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["hls"])')
check "master playlist" "$(curl -s -o /dev/null -w '%{http_code}' "$B$HLS")" "200"
check "unsigned refused" "$(curl -s -o /dev/null -w '%{http_code}' "${B}${HLS%%\?*}")" "403"

# The scrubbing index points at the sprite sheet by relative path, so the signature
# has to survive into the cue for a viewer to see any thumbnails at all.
TH=$(echo "$P"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["thumbnails"])')
VTT=$(curl -s "$B$TH")
CUE=$(echo "$VTT" | grep -m1 'sprite.jpg')
DIR=${TH%%\?*}; DIR=${DIR%/*}
check "thumbnail tile authorized" "$(curl -s -o /dev/null -w '%{http_code}' "$B$DIR/$CUE")" "200"

check "source size billed" "$(echo "$P"|python3 -c 'import sys,json;print("yes" if json.load(sys.stdin).get("source_bytes") else "no")')" "yes"

echo "9b. a link bound to one student cannot be unbound"
# The signature covers the viewer id and the watermark, so editing either out of the
# URL has to fail -- otherwise the device cap and the on-screen watermark are both
# opt-out and the whole feature is decoration.
BP=$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$A?viewer=student-8842&watermark=01712345678")
BHLS=$(echo "$BP"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["hls"])')
check "bound link carries the viewer" "$(echo "$BHLS" | grep -c 'vid=student-8842')" "1"
check "bound link carries the watermark" "$(echo "$BHLS" | grep -c 'wm=01712345678')" "1"
check "bound link plays" "$(curl -s -o /dev/null -w '%{http_code}' "$B$BHLS")" "200"
check "viewer stripped is refused" "$(curl -s -o /dev/null -w '%{http_code}' "${B}$(echo "$BHLS" | sed 's/&vid=student-8842//')")" "403"
check "watermark swapped is refused" "$(curl -s -o /dev/null -w '%{http_code}' "${B}$(echo "$BHLS" | sed 's/&wm=01712345678/\&wm=someone-else/')")" "403"
check "unsignable viewer id refused" "$(curl -s -H "Authorization: Bearer $KEY" "$B/v1/assets/$A?viewer=a+b%20c" | python3 -c 'import sys,json;print(json.load(sys.stdin)["error"]["code"])' )" "invalid_viewer"

echo "9c. delivery settings say how this account is protected"
SET=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/playback-settings)
check "device cap reported" "$(echo "$SET"|python3 -c 'import sys,json;print("max_viewer_devices" in json.load(sys.stdin))')" "True"
# Changing it is session-only, like every other account setting, so this script can
# only read it. A cap of 0 means no cap, which is what a new account has.
ENC=$(echo "$SET"|python3 -c 'import sys,json;print(json.load(sys.stdin)["encrypt_playback"])')

echo "9d. an encrypted asset answers an EME Clear Key licence"
HDIR=${HLS%%\?*}; HDIR=${HDIR%/*}; HQ=${HLS#*\?}
LIC=$(curl -s "$B$HDIR/key?$HQ")
if [ "$ENC" = "True" ]; then
  # The player is built against this exact body. A renamed field or padded base64 is
  # a licence no browser accepts, and it fails as a black screen with no error.
  check "licence shape" "$(echo "$LIC"|python3 -c 'import sys,json;d=json.load(sys.stdin);k=d["keys"][0];print(d["type"]=="temporary" and k["kty"]=="oct" and "=" not in k["k"])')" "True"
  check "licence over POST too" "$(curl -s -o /dev/null -w '%{http_code}' -X POST --data 'challenge' "$B$HDIR/key?$HQ")" "200"
  check "licence never cached" "$(curl -s -D- -o /dev/null "$B$HDIR/key?$HQ" | grep -ci 'cache-control: no-store')" "1"
else
  echo "  skip encryption checks (this account has encrypt_playback off)"
fi

echo "10. the same file again is deduplicated, not re-encoded"
D=$(curl -s -X POST -H "Authorization: Bearer $KEY" -H 'Content-Type: application/json' \
  -d '{"url":"http://127.0.0.1:8071/lecture.mp4"}' $B/v1/assets | python3 -c 'import sys,json;print(json.load(sys.stdin).get("asset_id",""))')
for i in $(seq 1 30); do
  DS=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$D | python3 -c 'import sys,json;print(json.load(sys.stdin)["state"])')
  case "$DS" in ready|partially_ready|failed) break;; esac; sleep 2
done
DP=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$D)
DHLS=$(echo "$DP"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["hls"])')
DDIR=${DHLS%%\?*}; DDIR=${DDIR%/*}; DQ=${DHLS#*\?}
check "duplicate playable ($DS)" "$(curl -s -o /dev/null -w '%{http_code}' "$B$DHLS")" "200"
check "duplicate reports renditions" "$(echo "$DP"|python3 -c 'import sys,json;print(len(json.load(sys.stdin)["renditions"]) > 0)')" "True"
# Whatever the key endpoint answers now, it must answer the same after the asset the
# media was first stored under is deleted: a licence if the account encrypts, a 404 if
# it does not. The invariant is that the answer does not change.
KBEFORE=$(curl -s -o /dev/null -w '%{http_code}' "$B$DDIR/key?$DQ")

echo "11. deleting the first video leaves the duplicate whole"
check "deleted" "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $KEY" $B/v1/assets/$A)" "204"
check "gone afterwards" "$(curl -s -o /dev/null -w '%{http_code}' -H "Authorization: Bearer $KEY" $B/v1/assets/$A)" "404"
DP=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$D)
check "duplicate still ready" "$(echo "$DP"|python3 -c 'import sys,json;print(json.load(sys.stdin)["state"])')" "$DS"
check "duplicate kept its renditions" "$(echo "$DP"|python3 -c 'import sys,json;print(len(json.load(sys.stdin)["renditions"]) > 0)')" "True"
DHLS=$(echo "$DP"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["hls"])')
DDIR=${DHLS%%\?*}; DDIR=${DDIR%/*}; DQ=${DHLS#*\?}
check "duplicate still plays" "$(curl -s -o /dev/null -w '%{http_code}' "$B$DHLS")" "200"
check "duplicate key unchanged" "$(curl -s -o /dev/null -w '%{http_code}' "$B$DDIR/key?$DQ")" "$KBEFORE"
check "duplicate deletable" "$(curl -s -o /dev/null -w '%{http_code}' -X DELETE -H "Authorization: Bearer $KEY" $B/v1/assets/$D)" "204"

echo
echo "passed=$pass failed=$fail"
[ "$fail" -eq 0 ]
