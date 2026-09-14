package httpx

import (
	"net/http"
	"testing"
)

func TestPrefersBangla(t *testing.T) {
	cases := map[string]bool{
		"bn":                      true,
		"bn-BD":                   true,
		"bn-BD,bn;q=0.9,en;q=0.8": true,
		"en-US,en;q=0.9":          false,
		"en":                      false,
		"":                        false,
		"fr-FR":                   false,
		// First recognised tag wins, so English ahead of Bangla stays English.
		"en-GB,bn;q=0.5": false,
		"bn;q=0.5,en-GB": true,
	}
	for header, want := range cases {
		if got := prefersBangla(header); got != want {
			t.Errorf("Accept-Language %q: got %v, want %v", header, got, want)
		}
	}
}

func TestLocaliseFallsBackToEnglish(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-Language", "bn-BD")

	if got := Localise(r, "invalid_api_key", "fallback"); got == "fallback" {
		t.Error("a code with Bangla copy fell back to English")
	}
	// An unknown code must degrade to English, never to an empty or placeholder string.
	if got := Localise(r, "code_with_no_translation", "fallback"); got != "fallback" {
		t.Errorf("unknown code returned %q, want the English fallback", got)
	}
}
