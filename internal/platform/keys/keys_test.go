package keys

import (
	"bytes"
	"testing"
)

func TestWrapRoundTrip(t *testing.T) {
	w, err := NewWrapper("a-test-key-encryption-key-32-bytes!!")
	if err != nil {
		t.Fatal(err)
	}
	_, key, err := Generate()
	if err != nil {
		t.Fatal(err)
	}

	wrapped, nonce, err := w.Wrap(key, "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(wrapped, key) {
		t.Fatal("wrapped key is the plaintext key")
	}

	got, err := w.Unwrap(wrapped, nonce, "asset-1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, key) {
		t.Error("round trip did not recover the key")
	}

	// Binding the asset id prevents lifting a key onto a different asset.
	if _, err := w.Unwrap(wrapped, nonce, "asset-2"); err == nil {
		t.Error("key unwrapped under the wrong asset id")
	}
}

func TestShortKEKRejected(t *testing.T) {
	if _, err := NewWrapper("too-short"); err == nil {
		t.Error("a short key-encryption key was accepted")
	}
}
