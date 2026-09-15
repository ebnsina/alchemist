DB_ADMIN := postgres://$(USER)@localhost:5432/alchemist?sslmode=disable
DB_APP   := postgres://alchemist_app:alchemist@localhost:5432/alchemist?sslmode=disable
SEAWEED  := .local/seaweed

# Local port map: alchemist 8090; SeaweedFS uses 8080 (volume), 8888 (filer),
# 9000 (S3), 9333 (master). Do not put alchemist on 8080 -- it collides silently and
# the health check answers from whatever got there first.
build:
	go build -o bin/ ./cmd/...

run: build
	set -a; . ./.env; set +a; ./bin/alchemist

# Dev object storage. Production runs the same SeaweedFS, with erasure coding.
storage:
	mkdir -p $(SEAWEED)/data
	weed server -dir=$(SEAWEED)/data -s3 -s3.port=9000 \
	  -s3.config=$(SEAWEED)/s3.json -volume.max=64 -master.volumeSizeLimitMB=1024 -ip=127.0.0.1

db-reset:
	dropdb --if-exists alchemist && createdb alchemist
	for m in internal/platform/db/migrations/*.sql; do psql -q -d alchemist -v ON_ERROR_STOP=1 -f "$$m" || exit 1; done
	go run github.com/riverqueue/river/cmd/river@latest migrate-up --database-url "$(DB_ADMIN)"
	psql -q -d alchemist -c "grant select,insert,update,delete on all tables in schema public to alchemist_app; grant usage,select on all sequences in schema public to alchemist_app;"

# A local account with the quotas lifted, for testing the dashboard without hitting
# the four-concurrent-job limit every few minutes. Never run against a real host.
dev-account:
	./scripts/dev-account.sh

test:
	ALCHEMIST_TEST_DATABASE_URL="$(DB_APP)" ALCHEMIST_TEST_ADMIN_URL="$(DB_ADMIN)" go test ./...

# Local edge harness: real cache config, minus TLS and njs (no njs in Homebrew nginx).
edge:
	nginx -c $(PWD)/deploy/edge/nginx.test.conf

edge-stop:
	- nginx -c $(PWD)/deploy/edge/nginx.test.conf -s stop

# Stand-ins for a third-party customer: an origin hosting their video and a webhook
# receiver. Run alongside `make run` with ALCHEMIST_FETCH_ALLOWLIST set.
customer-servers:
	python3 test/customer_servers.py

.PHONY: build run storage db-reset test edge edge-stop customer-servers
