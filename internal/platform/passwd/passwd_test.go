package passwd

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestHashVerify(t *testing.T) {
	h, err := Hash("correct horse battery")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if !strings.HasPrefix(h, "$argon2id$") {
		t.Fatalf("unexpected encoding: %q", h)
	}
	if err := Verify("correct horse battery", h); err != nil {
		t.Fatalf("verify correct: %v", err)
	}
	if err := Verify("wrong horse battery", h); err != ErrMismatch {
		t.Fatalf("wrong password: want ErrMismatch, got %v", err)
	}

	// Two hashes of the same password must differ: the salt is what stops one
	// rainbow table from covering every account.
	h2, _ := Hash("correct horse battery")
	if h == h2 {
		t.Fatal("identical hashes for the same password — salt is not random")
	}
}

func TestLengthBounds(t *testing.T) {
	if _, err := Hash("short"); err != ErrTooShort {
		t.Fatalf("short password: want ErrTooShort, got %v", err)
	}
	if _, err := Hash(strings.Repeat("a", MaxLength+1)); err != ErrTooLong {
		t.Fatalf("long password: want ErrTooLong, got %v", err)
	}
}

func TestVerifyRejectsGarbage(t *testing.T) {
	for _, bad := range []string{"", "not-a-hash", "$argon2id$v=19$bad$x$y",
		"$bcrypt$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA"} {
		if err := Verify("whatever", bad); err != ErrMismatch {
			t.Fatalf("%q: want ErrMismatch, got %v", bad, err)
		}
	}
}

// The dummy has to reach the argon2 call and use today's parameters. A malformed
// one still returns ErrMismatch — from the parser, in microseconds — and the login
// endpoint would answer faster for an unknown email than a wrong password, which is
// exactly the leak the dummy exists to close.
func TestDummyCostsTheSameWork(t *testing.T) {
	if err := Verify("anything at all", Dummy); err != ErrMismatch {
		t.Fatalf("dummy hash: want ErrMismatch, got %v", err)
	}
	real, err := Hash("some password here")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	params := func(encoded string) string {
		parts := strings.Split(encoded, "$")
		if len(parts) != 6 {
			t.Fatalf("not a PHC string: %q", encoded)
		}
		// Both salt and hash must decode, or Verify bails before doing the work.
		if _, err := base64.RawStdEncoding.DecodeString(parts[4]); err != nil {
			t.Fatalf("salt does not decode in %q", encoded)
		}
		if _, err := base64.RawStdEncoding.DecodeString(parts[5]); err != nil {
			t.Fatalf("hash does not decode in %q", encoded)
		}
		return parts[1] + "|" + parts[2] + "|" + parts[3]
	}
	if got, want := params(Dummy), params(real); got != want {
		t.Fatalf("dummy parameters %q differ from current %q — costs will not match", got, want)
	}
}
