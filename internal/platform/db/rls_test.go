package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// Proves tenant isolation is enforced by the database, not by application queries.
// Every query below is an unfiltered "select ... from assets" with no WHERE clause.
func TestRLSIsolatesTenants(t *testing.T) {
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
	}{{"alpha", &tenantA}, {"beta", &tenantB}} {
		if err := admin.QueryRow(ctx,
			`insert into tenants (name, ladder_profile) values ($1, 'bd-mobile')
			 returning id::text`, tc.name).Scan(tc.dst); err != nil {
			t.Fatalf("seed tenant %s: %v", tc.name, err)
		}
		if _, err := admin.Exec(ctx,
			`insert into assets (tenant_id, ladder_profile) values ($1, 'bd-mobile')`,
			*tc.dst); err != nil {
			t.Fatalf("seed asset for %s: %v", tc.name, err)
		}
	}
	// LIFO: this runs before the deferred Close above. t.Cleanup would fire after it,
	// against a connection that is already shut, and the rows would survive the run.
	defer func() {
		_, _ = admin.Exec(ctx, `delete from tenants where id = any($1::uuid[])`,
			[]string{tenantA, tenantB})
	}()

	countAssets := func(tenantID string) int {
		var n int
		if err := d.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from assets`).Scan(&n)
		}); err != nil {
			t.Fatalf("count for %s: %v", tenantID, err)
		}
		return n
	}

	if got := countAssets(tenantA); got != 1 {
		t.Errorf("tenant A sees %d assets, want exactly its own 1", got)
	}
	if got := countAssets(tenantB); got != 1 {
		t.Errorf("tenant B sees %d assets, want exactly its own 1", got)
	}

	// The real test: A must not be able to read B's row even naming it directly.
	var leaked int
	if err := d.AsTenant(ctx, tenantA, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select count(*) from assets where tenant_id = $1`, tenantB).Scan(&leaked)
	}); err != nil {
		t.Fatalf("cross-tenant query: %v", err)
	}
	if leaked != 0 {
		t.Errorf("tenant A read %d of tenant B's assets; RLS is not enforcing", leaked)
	}

	// With no tenant set, the app role must see nothing at all.
	var unscoped int
	if err := d.AsTenant(ctx, "", func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select count(*) from assets`).Scan(&unscoped)
	}); err != nil {
		t.Fatalf("unscoped query: %v", err)
	}
	if unscoped != 0 {
		t.Errorf("unscoped connection read %d assets, want 0", unscoped)
	}
}
