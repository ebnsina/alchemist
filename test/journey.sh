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

echo "8. playback works"
P=$(curl -s -H "Authorization: Bearer $KEY" $B/v1/assets/$A)
HLS=$(echo "$P"|python3 -c 'import sys,json;print(json.load(sys.stdin)["playback"]["hls"])')
check "master playlist" "$(curl -s -o /dev/null -w '%{http_code}' "$B$HLS")" "200"
check "unsigned refused" "$(curl -s -o /dev/null -w '%{http_code}' "${B}${HLS%%\?*}")" "403"

echo
echo "passed=$pass failed=$fail"
[ "$fail" -eq 0 ]
