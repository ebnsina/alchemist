// Package keys wraps content encryption keys for storage.
package keys

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
)

// Wrapper encrypts content keys with a key-encryption key held in the environment.
// A database dump therefore does not decrypt the library on its own.
type Wrapper struct{ aead cipher.AEAD }

func NewWrapper(kek string) (*Wrapper, error) {
	if len(kek) < 32 {
		return nil, fmt.Errorf("key-encryption key must be at least 32 bytes")
	}
	sum := sha256.Sum256([]byte(kek))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Wrapper{aead: aead}, nil
}

// Generate returns a fresh 128-bit content key and its identifier.
func Generate() (keyID, key []byte, err error) {
	keyID = make([]byte, 16)
	key = make([]byte, 16)
	if _, err = rand.Read(keyID); err != nil {
		return nil, nil, err
	}
	if _, err = rand.Read(key); err != nil {
		return nil, nil, err
	}
	return keyID, key, nil
}

// Wrap encrypts a content key. The asset id is bound in as additional authenticated
// data, so a wrapped key cannot be moved between assets.
func (w *Wrapper) Wrap(key []byte, assetID string) (wrapped, nonce []byte, err error) {
	nonce = make([]byte, w.aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return w.aead.Seal(nil, nonce, key, []byte(assetID)), nonce, nil
}

func (w *Wrapper) Unwrap(wrapped, nonce []byte, assetID string) ([]byte, error) {
	key, err := w.aead.Open(nil, nonce, wrapped, []byte(assetID))
	if err != nil {
		return nil, fmt.Errorf("unwrap content key: %w", err)
	}
	return key, nil
}
