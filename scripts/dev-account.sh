#!/bin/bash
# A local test account with the quotas lifted, so a day of poking at the dashboard
# does not stop at "you have 4 videos processing, which is your limit".
#
# LOCAL ONLY. This is deliberately a script and not a migration: migrations run in
# production, and seeding a known address with a known password into a production
# database would be a way in, not a convenience. Nothing here runs unless somebody
# types it.
#
#   make run            # terminal 1
#   make dev-account
#
# Override anything: EMAIL=me@example.com PASSWORD=... ./scripts/dev-account.sh
set -euo pipefail

B=${B:-http://localhost:8090}
DB=${DB:-alchemist}
EMAIL=${EMAIL:-sina@alchemist.io}
PASSWORD=${PASSWORD:-alchemist-dev-password}
ORG=${ORG:-Alchemist Dev}
ORIGIN=${ORIGIN:-http://localhost:5173}

# Refuse to run against anything that is not obviously a local machine. The limits
# this sets would be a gift to whoever found the password.
case "$B" in
	http://localhost:*|http://127.0.0.1:*) ;;
	*) echo "refusing: $B is not local. This account is for development only." >&2; exit 1 ;;
esac

echo "-> account for $EMAIL"
BODY=$(printf '{"org":"%s","email":"%s","password":"%s"}' "$ORG" "$EMAIL" "$PASSWORD")
RES=$(curl -s -H "Origin: $ORIGIN" -H 'Content-Type: application/json' -d "$BODY" "$B/v1/auth/signup")

if echo "$RES" | grep -q '"api_key"'; then
	KEY=$(echo "$RES" | python3 -c 'import sys,json;print(json.load(sys.stdin)["api_key"])')
	echo "   created. API key (shown once): $KEY"
elif echo "$RES" | grep -q 'email_taken'; then
	echo "   already exists, leaving the password alone"
else
	echo "   could not create it: $RES" >&2
	exit 1
fi

TENANT=$(psql -tAq -d "$DB" -c "select tenant_id from users where email = '$EMAIL'")
if [ -z "$TENANT" ]; then
	echo "   no tenant for $EMAIL -- is the API pointed at $DB?" >&2
	exit 1
fi

# Raised, not removed. A number the code still checks is easier to reason about than
# a bypass branch that could one day be reachable in production.
psql -q -d "$DB" <<SQL
insert into tenant_limits (tenant_id, max_concurrent_jobs, max_source_bytes, max_ingest_hours_mo)
values ('$TENANT', 1000000, 1099511627776, null)
on conflict (tenant_id) do update
  set max_concurrent_jobs = excluded.max_concurrent_jobs,
      max_source_bytes    = excluded.max_source_bytes,
      max_ingest_hours_mo = null,
      updated_at          = now();
SQL

echo "-> limits lifted for tenant $TENANT"
psql -d "$DB" -c "select max_concurrent_jobs, pg_size_pretty(max_source_bytes) as max_file, coalesce(max_ingest_hours_mo::text,'no monthly cap') as monthly from tenant_limits where tenant_id = '$TENANT';"
echo "-> sign in at $ORIGIN/login/  --  $EMAIL / $PASSWORD"
