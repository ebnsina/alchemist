package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/ebnsina/alchemist/internal/platform/db"
	"github.com/ebnsina/alchemist/internal/platform/passwd"
)

// staffOrg is the tenant the platform administrator's own account lives in. It is an
// ordinary tenant with nothing in it: the flag, not the tenant, is what grants
// anything, and impersonation is how staff reach a customer's data. Live is switched
// on for it, since staff testing their own product should not be sold it.
const staffOrg = "Alchemist"

// ask reads one line from the terminal, with the answer echoed.
func ask(in *bufio.Reader, prompt string) (string, error) {
	fmt.Print(prompt)
	line, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// askSecret reads without echoing. Typed in the clear it would sit in the scrollback
// of whatever terminal, and over somebody's shoulder.
func askSecret(prompt string) (string, error) {
	fmt.Print(prompt)
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(b), err
}

// prompt asks for the address and password, twice for the password because a typo in
// something that cannot be reset yet locks the only administrator out of the product.
func prompt(email string) (string, string, error) {
	in := bufio.NewReader(os.Stdin)
	var err error
	for email == "" || !strings.Contains(email, "@") {
		if email, err = ask(in, "Email: "); err != nil {
			return "", "", err
		}
		if !strings.Contains(email, "@") {
			fmt.Println("That does not look like an email address.")
			email = ""
		}
	}
	for {
		first, err := askSecret("Password: ")
		if err != nil {
			return "", "", err
		}
		if len(first) < 10 {
			fmt.Println("Use at least 10 characters, the same as signing up.")
			continue
		}
		again, err := askSecret("Again: ")
		if err != nil {
			return "", "", err
		}
		if first != again {
			fmt.Println("Those did not match.")
			continue
		}
		return strings.ToLower(email), first, nil
	}
}

// createAdmin makes or promotes the platform administrator.
//
// Asked for at the terminal, or read from the environment when there is no terminal
// to ask at, so a deploy script still works. Never an argument: an argument is in the
// shell history and visible in ps to every user on the box. Run twice with the same
// address it promotes the account that is already there and leaves its password
// alone, so re-running after a deploy is safe and does not create a second user.
func createAdmin(log *slog.Logger, args []string) {
	if len(args) > 1 {
		log.Error("usage: alchemist create-admin [email]")
		os.Exit(2)
	}
	email := ""
	if len(args) == 1 {
		email = strings.ToLower(strings.TrimSpace(args[0]))
	}

	dbURL := os.Getenv("ALCHEMIST_DATABASE_URL")
	if strings.TrimSpace(dbURL) == "" {
		log.Error("configuration", "err", "ALCHEMIST_DATABASE_URL is not set")
		os.Exit(1)
	}

	password := os.Getenv("ALCHEMIST_STAFF_PASSWORD")
	if term.IsTerminal(int(os.Stdin.Fd())) && password == "" {
		var err error
		if email, password, err = prompt(email); err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) || err.Error() == "EOF" {
				fmt.Println()
			}
			log.Error("cancelled", "err", err)
			os.Exit(1)
		}
	}
	if email == "" || !strings.Contains(email, "@") {
		log.Error("usage: alchemist create-admin <email>  (no terminal to ask at)")
		os.Exit(2)
	}
	if strings.TrimSpace(password) == "" {
		log.Error("configuration", "err", "ALCHEMIST_STAFF_PASSWORD is not set and there is no terminal to ask at")
		os.Exit(1)
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
			"user_id", userID, "tenant_id", tenantID, "live_enabled", true)
		return
	}
	log.Info("platform administrator confirmed; the account already existed and its password is unchanged",
		"email", email, "user_id", userID, "tenant_id", tenantID, "live_enabled", true)
}
