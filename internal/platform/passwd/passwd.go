// Package passwd hashes and verifies account passwords with argon2id.
//
// argon2id rather than bcrypt: bcrypt silently truncates at 72 bytes and has no
// memory cost, so a GPU farm is limited only by its clock. The parameters below are
// the RFC 9106 second recommended set — 64 MiB, one pass, four lanes — which costs
// roughly 50ms on a small VPS core and makes a large offline attack expensive in
// memory rather than only in time.
package passwd

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// MinLength follows NIST SP 800-63B: length is the only rule worth enforcing.
	MinLength = 10
	// MaxLength bounds the work an unauthenticated request can ask for. Without it
	// a megabyte password is a free denial of service.
	MaxLength = 128

	timeCost    = 1
	memoryKiB   = 64 * 1024
	parallelism = 4
	saltLen     = 16
	keyLen      = 32
)

var (
	ErrTooShort = fmt.Errorf("password must be at least %d characters", MinLength)
	ErrTooLong  = fmt.Errorf("password must be at most %d characters", MaxLength)
	// ErrMismatch is returned for both a wrong password and a malformed stored hash,
	// so a caller cannot tell the two apart from the outside.
	ErrMismatch = errors.New("password does not match")
)

// Hash returns the PHC-encoded hash for a password.
func Hash(password string) (string, error) {
	if len(password) < MinLength {
		return "", ErrTooShort
	}
	if len(password) > MaxLength {
		return "", ErrTooLong
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey([]byte(password), salt, timeCost, memoryKiB, parallelism, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memoryKiB, timeCost, parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum)), nil
}

// Verify reports whether password matches the encoded hash.
func Verify(password, encoded string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return ErrMismatch
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return ErrMismatch
	}
	var mem uint32
	var time uint32
	var lanes uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &time, &lanes); err != nil {
		return ErrMismatch
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ErrMismatch
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ErrMismatch
	}
	got := argon2.IDKey([]byte(password), salt, time, mem, lanes, uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return ErrMismatch
	}
	return nil
}

// Dummy is a valid hash of a value nobody knows. Verifying against it when an email
// is unknown keeps the work — and so the response time — the same as a real miss.
const Dummy = "$argon2id$v=19$m=65536,t=1,p=4$" +
	"c29tZXNhbHRzb21lc2FsdA$Zm9vYmFyYmF6cXV4Zm9vYmFyYmF6cXV4Zm9vYmFyYmF6cXV4MDA"
