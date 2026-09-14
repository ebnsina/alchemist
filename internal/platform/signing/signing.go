// Package signing mints and verifies playback signatures.
//
// Verification has to be possible at the edge without touching the control plane or
// the database, so a signature carries everything needed to check it: the key id,
// the expiry, and the path prefix it covers.
package signing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SignatureLength is the truncated hex length. 128 bits is far beyond what a
// guessing attack against a short-lived URL could reach.
const SignatureLength = 32

// Keyring holds every key an edge will accept and the one currently used to sign.
// Rotation is therefore not an outage: publish the new key everywhere, then switch
// which one signs, then retire the old one.
type Keyring struct {
	keys      map[string][]byte
	activeKID string
}

// NewKeyring parses "kid:secret" entries. The first entry signs.
func NewKeyring(entries []string) (*Keyring, error) {
	if len(entries) == 0 {
		return nil, fmt.Errorf("keyring is empty")
	}
	kr := &Keyring{keys: make(map[string][]byte, len(entries))}
	for i, e := range entries {
		kid, secret, ok := strings.Cut(strings.TrimSpace(e), ":")
		if !ok || kid == "" || secret == "" {
			return nil, fmt.Errorf("keyring entry %d must be kid:secret", i)
		}
		if len(secret) < 32 {
			return nil, fmt.Errorf("keyring entry %q secret must be at least 32 bytes", kid)
		}
		if _, dup := kr.keys[kid]; dup {
			return nil, fmt.Errorf("duplicate key id %q", kid)
		}
		kr.keys[kid] = []byte(secret)
		if i == 0 {
			kr.activeKID = kid
		}
	}
	return kr, nil
}

func (k *Keyring) ActiveKID() string { return k.activeKID }

func (k *Keyring) KIDs() []string {
	out := make([]string, 0, len(k.keys))
	for kid := range k.keys {
		out = append(out, kid)
	}
	sort.Strings(out)
	return out
}

// Sign covers the asset prefix rather than an individual file, so one token
// authorizes a manifest and every segment and key request beneath it.
func (k *Keyring) Sign(prefix string, exp int64) (kid, sig string) {
	return k.activeKID, compute(k.keys[k.activeKID], prefix, exp)
}

// Verify checks a signature against the named key. An unknown key id fails closed.
func (k *Keyring) Verify(prefix, kid, sig, expRaw string) bool {
	secret, ok := k.keys[kid]
	if !ok {
		return false
	}
	exp, err := strconv.ParseInt(expRaw, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(compute(secret, prefix, exp)))
}

// compute is the canonical string the edge must reproduce exactly.
func compute(secret []byte, prefix string, exp int64) string {
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "%s|%d", prefix, exp)
	return hex.EncodeToString(mac.Sum(nil))[:SignatureLength]
}
