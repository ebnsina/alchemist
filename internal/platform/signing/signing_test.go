package signing

import (
	"testing"
	"time"
)

func ring(t *testing.T) *Keyring {
	t.Helper()
	kr, err := NewKeyring([]string{
		"k2:second-generation-secret-at-least-32b",
		"k1:first-generation-secret-at-least-32by",
	})
	if err != nil {
		t.Fatal(err)
	}
	return kr
}

func TestSignVerifyRoundTrip(t *testing.T) {
	kr := ring(t)
	exp := time.Now().Add(time.Hour).Unix()
	kid, sig := kr.Sign("/playback/t/a", exp, "", "", "")

	if kid != "k2" {
		t.Errorf("signed with %q, want the first entry k2", kid)
	}
	if len(sig) != SignatureLength {
		t.Errorf("signature length %d, want %d", len(sig), SignatureLength)
	}
	if !kr.Verify("/playback/t/a", kid, sig, itoa(exp), "", "", "") {
		t.Error("valid signature rejected")
	}
}

// The binding is the anti-sharing feature: a link minted for one student must stop
// working the moment somebody edits their id or their watermark out of it.
func TestBindingCannotBeStripped(t *testing.T) {
	kr := ring(t)
	exp := time.Now().Add(time.Hour).Unix()
	expS := itoa(exp)
	kid, sig := kr.Sign("/playback/t/a", exp, "student-88", "rafi-01712", "")

	if !kr.Verify("/playback/t/a", kid, sig, expS, "student-88", "rafi-01712", "") {
		t.Error("bound signature rejected with its own binding")
	}
	for _, tc := range []struct{ name, viewer, label string }{
		{"viewer removed", "", "rafi-01712"},
		{"label removed", "student-88", ""},
		{"both removed", "", ""},
		{"viewer swapped", "student-99", "rafi-01712"},
		{"label swapped", "student-88", "someone-else"},
	} {
		if kr.Verify("/playback/t/a", kid, sig, expS, tc.viewer, tc.label, "") {
			t.Errorf("%s was accepted", tc.name)
		}
	}

	// An unbound link is not a bound one with empty fields, or the two would be
	// interchangeable and stripping the binding would be free.
	_, plain := kr.Sign("/playback/t/a", exp, "", "", "")
	if plain == sig {
		t.Error("bound and unbound links signed identically")
	}
}

// Rotation must not invalidate links already handed out under the previous key.
func TestRetiredKeyStillVerifies(t *testing.T) {
	old, _ := NewKeyring([]string{"k1:first-generation-secret-at-least-32by"})
	exp := time.Now().Add(time.Hour).Unix()
	kid, sig := old.Sign("/playback/t/a", exp, "", "", "")

	if !ring(t).Verify("/playback/t/a", kid, sig, itoa(exp), "", "", "") {
		t.Error("link signed before rotation stopped working")
	}
}

func TestRejections(t *testing.T) {
	kr := ring(t)
	future := itoa(time.Now().Add(time.Hour).Unix())
	_, sig := kr.Sign("/playback/t/a", time.Now().Add(time.Hour).Unix(), "", "", "")

	cases := []struct {
		name                  string
		prefix, kid, sig, exp string
	}{
		{"unknown key id", "/playback/t/a", "k9", sig, future},
		{"tampered signature", "/playback/t/a", "k2", "00000000000000000000000000000000", future},
		{"different asset", "/playback/t/OTHER", "k2", sig, future},
		{"expired", "/playback/t/a", "k2", sig, itoa(time.Now().Add(-time.Minute).Unix())},
		{"non numeric expiry", "/playback/t/a", "k2", sig, "soon"},
		{"empty signature", "/playback/t/a", "k2", "", future},
	}
	for _, tc := range cases {
		if kr.Verify(tc.prefix, tc.kid, tc.sig, tc.exp, "", "", "") {
			t.Errorf("%s was accepted", tc.name)
		}
	}
}

func TestMalformedKeyringRejected(t *testing.T) {
	for _, entries := range [][]string{
		{},
		{"nocolon"},
		{"k1:tooshort"},
		{"k1:first-generation-secret-at-least-32by", "k1:duplicate-id-secret-at-least-32byte"},
	} {
		if _, err := NewKeyring(entries); err == nil {
			t.Errorf("malformed keyring %v was accepted", entries)
		}
	}
}

func itoa(i int64) string {
	return time.Unix(i, 0).Format("") + fmtInt(i)
}

func fmtInt(i int64) string {
	if i == 0 {
		return "0"
	}
	var b []byte
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		b = append([]byte{byte('0' + i%10)}, b...)
		i /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
