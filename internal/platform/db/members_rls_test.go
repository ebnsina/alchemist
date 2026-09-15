package db_test

import (
	"context"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/ebnsina/alchemist/internal/platform/db"
)

// Invites and branding are tenant-scoped like everything else, and the same way:
// by policy, not by a WHERE clause somebody has to remember to write.
func TestRLSIsolatesInvitesAndBranding(t *testing.T) {
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
	for i, dst := range []*string{&tenantA, &tenantB} {
		if err := admin.QueryRow(ctx,
			`insert into tenants (name, ladder_profile) values ($1, 'bd-mobile')
			 returning id::text`, []string{"inv-alpha", "inv-beta"}[i]).Scan(dst); err != nil {
			t.Fatalf("seed tenant: %v", err)
		}
		hash := sha256.Sum256([]byte("token-" + *dst))
		if _, err := admin.Exec(ctx,
			`insert into invites (tenant_id, email, token_hash, expires_at)
			 values ($1, $2, $3, now() + interval '7 days')`,
			*dst, "person-"+*dst+"@example.com", hash[:]); err != nil {
			t.Fatalf("seed invite: %v", err)
		}
		if _, err := admin.Exec(ctx,
			`insert into tenant_branding (tenant_id, logo_key) values ($1, $2)`,
			*dst, "brand/"+*dst+"/logo.png"); err != nil {
			t.Fatalf("seed branding: %v", err)
		}
	}
	// LIFO: this runs before the deferred Close above. t.Cleanup would fire after it,
	// against a connection that is already shut.
	defer func() {
		_, _ = admin.Exec(ctx, `delete from tenants where id = any($1::uuid[])`,
			[]string{tenantA, tenantB})
	}()

	count := func(tenantID, table string) int {
		var n int
		if err := d.AsTenant(ctx, tenantID, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, `select count(*) from `+table).Scan(&n)
		}); err != nil {
			t.Fatalf("count %s for %s: %v", table, tenantID, err)
		}
		return n
	}

	for _, table := range []string{"invites", "tenant_branding"} {
		if got := count(tenantA, table); got != 1 {
			t.Errorf("tenant A sees %d rows of %s, want its own 1", got, table)
		}
		if got := count("", table); got != 0 {
			t.Errorf("unscoped connection read %d rows of %s, want 0", got, table)
		}
	}
}

// Redeeming happens before any tenant is in scope, so it has to be a definer
// function — under RLS a plain query reads zero rows and reports a valid token as
// unknown. The expiry is enforced there too, not by the caller.
func TestInviteRedeemCreatesUserAndRefusesExpired(t *testing.T) {
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

	var tenant string
	if err := admin.QueryRow(ctx,
		`insert into tenants (name, ladder_profile) values ('redeem', 'bd-mobile')
		 returning id::text`).Scan(&tenant); err != nil {
		t.Fatalf("seed tenant: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(ctx, `delete from tenants where id = $1::uuid`, tenant)
	}()

	live := sha256.Sum256([]byte("live-token"))
	dead := sha256.Sum256([]byte("dead-token"))
	if _, err := admin.Exec(ctx,
		`insert into invites (tenant_id, email, role, token_hash, expires_at) values
		   ($1, 'new@example.com', 'admin', $2, now() + interval '1 day'),
		   ($1, 'old@example.com', 'member', $3, now() - interval '1 day')`,
		tenant, live[:], dead[:]); err != nil {
		t.Fatalf("seed invites: %v", err)
	}

	var userID, gotTenant, email string
	err = d.AsTenant(ctx, "", func(tx pgx.Tx) error {
		return tx.QueryRow(ctx,
			`select user_id::text, tenant_id::text, email::text
			   from invite_redeem($1, 'argon2-hash')`, live[:]).Scan(&userID, &gotTenant, &email)
	})
	if err != nil {
		t.Fatalf("redeem a live invite: %v", err)
	}
	if gotTenant != tenant || email != "new@example.com" {
		t.Fatalf("redeemed into tenant %s as %s", gotTenant, email)
	}

	// The role travels with the invite. Redeeming must not silently make a member.
	var role string
	if err := d.AsTenant(ctx, tenant, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `select role from users where id = $1::uuid`, userID).Scan(&role)
	}); err != nil {
		t.Fatalf("read role: %v", err)
	}
	if role != "admin" {
		t.Errorf("role = %q, want admin from the invite", role)
	}

	// Second use of the same token must find nothing.
	if err := d.AsTenant(ctx, "", func(tx pgx.Tx) error {
		var u string
		return tx.QueryRow(ctx,
			`select user_id::text from invite_redeem($1, 'x')`, live[:]).Scan(&u)
	}); err != pgx.ErrNoRows {
		t.Errorf("reusing a redeemed token returned %v, want no rows", err)
	}

	if err := d.AsTenant(ctx, "", func(tx pgx.Tx) error {
		var u string
		return tx.QueryRow(ctx,
			`select user_id::text from invite_redeem($1, 'x')`, dead[:]).Scan(&u)
	}); err != pgx.ErrNoRows {
		t.Errorf("expired token returned %v, want no rows", err)
	}

}
