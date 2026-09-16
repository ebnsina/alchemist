package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/passwd"
)

// staffOrg is the tenant the platform administrator's own account lives in. It is an
// ordinary tenant with nothing in it: the flag, not the tenant, is what grants
// anything, and impersonation is how staff reach a customer's data.
const staffOrg = "Alchemist"

// createAdmin makes or promotes the platform administrator.
//
// The password comes from the environment, never from an argument: an argument is in
// the shell history and visible in ps to every user on the box. Run twice with the
// same address it promotes the account that is already there and leaves its password
// alone, so re-running after a deploy is safe and does not create a second user.
func createAdmin(log *slog.Logger, args []string) {
	if len(args) != 1 || !strings.Contains(args[0], "@") {
		log.Error("usage: alchemist create-admin <email>  (password in ALCHEMIST_STAFF_PASSWORD)")
		os.Exit(2)
	}
	email := strings.ToLower(strings.TrimSpace(args[0]))

	dbURL := os.Getenv("ALCHEMIST_DATABASE_URL")
	password := os.Getenv("ALCHEMIST_STAFF_PASSWORD")
	for k, v := range map[string]string{
		"ALCHEMIST_DATABASE_URL": dbURL, "ALCHEMIST_STAFF_PASSWORD": password,
	} {
		if strings.TrimSpace(v) == "" {
			log.Error("configuration", "err", fmt.Sprintf("%s is not set", k))
			os.Exit(1)
		}
	}

	hash, err := passwd.Hash(password)
	if err != nil {
		log.Error("password", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	database, err := db.New(ctx, dbURL)
	if err != nil {
		log.Error("database", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	var userID, tenantID string
	var created bool
	if err := database.Pool().QueryRow(ctx,
		`select user_id::text, tenant_id::text, created from staff_grant($1, $2, $3)`,
		email, hash, staffOrg).Scan(&userID, &tenantID, &created); err != nil {
		log.Error("granting platform admin", "err", err)
		os.Exit(1)
	}
	if created {
		log.Info("platform administrator created", "email", email,
			"user_id", userID, "tenant_id", tenantID)
		return
	}
	log.Info("platform administrator confirmed; the account already existed and its password is unchanged",
		"email", email, "user_id", userID, "tenant_id", tenantID)
}
