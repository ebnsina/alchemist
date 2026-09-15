package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// A stream key is a publishing credential. If RLS on live_streams is wrong, one
// customer lists another's streams and their key hashes -- so this asserts the
// database refuses it, rather than trusting that every query carries a WHERE clause.
func TestRLSIsolatesLiveStreams(t *testing.T) {
	url := os.Getenv("ALCHEMIST_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("ALCHEMIST_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()

	d, err := db.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer d.Close()

	admin, err := pgx.Connect(ctx, os.Getenv("ALCHEMIST_TEST_ADMIN_URL"))
	if err != nil {
		t.Fatalf("admin connect: %v", err)
	}
	defer admin.Close(ctx)

	var tenantA, tenantB string
	for _, tc := range []struct {
		name string
		dst  *string
	}{{"live-alpha", &tenantA}, {"live-beta", &tenantB}} {
		if err := admin.QueryRow(ctx,
			`insert into tenants (name, ladder_profile) values ($1, 'bd-mobile')
			 returning id::text`, tc.name).Scan(tc.dst); err != nil {
			t.Fatalf("seed tenant %s: %v", tc.name, err)
		}
		if _, err := admin.Exec(ctx,
			`insert into live_streams (tenant_id, name, key_hash, protocol)
			 values ($1, $2, sha256($3::bytea), 'srt')`, *tc.dst, tc.name, []byte(tc.name)); err != nil {
			t.Fatalf("seed stream for %s: %v", tc.name, err)
		}
	}
	defer func() {
		_, _ = admin.Exec(ctx, `delete from tenants where id = any($1::uuid[])`,
			[]string{tenantA, tenantB})
	}()

	count := func(tenantID, where string, args ...any) int {
		var n int
		if err := d.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from live_streams `+where, args...).Scan(&n)
		}); err != nil {
			t.Fatalf("count for %q: %v", tenantID, err)
		}
		return n
	}

	if got := count(tenantA, ""); got != 1 {
		t.Errorf("tenant A sees %d live streams, want exactly its own 1", got)
	}
	if got := count(tenantA, `where tenant_id = $1`, tenantB); got != 0 {
		t.Errorf("tenant A read %d of tenant B's live streams; RLS is not enforcing", got)
	}
	// A worker or a background job with no tenant in scope must read nothing, which
	// is why stream keys resolve through resolve_stream_key() instead.
	if got := count("", ""); got != 0 {
		t.Errorf("unscoped connection read %d live streams, want 0", got)
	}
}
