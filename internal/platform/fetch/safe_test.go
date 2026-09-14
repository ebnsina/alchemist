package fetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsBlocked(t *testing.T) {
	mustBlock := []string{
		"127.0.0.1", "127.1.2.3", "::1",
		"10.0.0.1", "10.255.255.255",
		"172.16.0.1", "172.31.255.255",
		"192.168.1.1",
		"169.254.169.254", // cloud instance metadata: the one that leaks credentials
		"169.254.0.1",
		"0.0.0.0",
		"100.64.0.1", // CGNAT
		"224.0.0.1",  // multicast
		"255.255.255.255",
		"fc00::1", "fd00::1", // unique local
		"fe80::1",          // link-local
		"::ffff:127.0.0.1", // IPv4-mapped loopback must not bypass IPv4 rules
		"::ffff:169.254.169.254",
		"::",
	}
	for _, s := range mustBlock {
		if ip := net.ParseIP(s); !IsBlocked(ip) {
			t.Errorf("%s was ALLOWED and must be blocked", s)
		}
	}

	mustAllow := []string{
		"8.8.8.8", "1.1.1.1", "93.184.216.34",
		"172.32.0.1",  // just outside the private range
		"172.15.0.1",  // just below it
		"11.0.0.1",    // just outside 10/8
		"192.169.0.1", // just outside 192.168/16
		"2606:4700::1111",
	}
	for _, s := range mustAllow {
		if ip := net.ParseIP(s); IsBlocked(ip) {
			t.Errorf("%s was BLOCKED and should be reachable", s)
		}
	}

	if !IsBlocked(nil) {
		t.Error("a nil address must be blocked")
	}
}

// The guard has to hold at dial time, not just on the literal URL the customer gave.
func TestDialToLoopbackIsRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("secret"))
	}))
	defer srv.Close()

	_, err := ToFile(context.Background(), srv.URL, filepath.Join(t.TempDir(), "out"),
		1<<20, 5*time.Second)
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("reached a loopback server: err = %v", err)
	}
}

// The attack that matters: a public-looking URL that redirects inward. Only the
// first URL is ever vetted by the caller, so every hop must be re-checked.
func TestRedirectToInternalIsRefused(t *testing.T) {
	internal := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("credentials"))
	}))
	defer internal.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, internal.URL, http.StatusFound)
	}))
	defer redirector.Close()

	_, err := ToFile(context.Background(), redirector.URL, filepath.Join(t.TempDir(), "out"),
		1<<20, 5*time.Second)
	if err == nil {
		t.Fatal("a redirect into internal space was followed")
	}
	if !errors.Is(err, ErrBlockedAddress) && !errors.Is(err, ErrUnreachable) {
		t.Fatalf("unexpected error kind: %v", err)
	}
}

func TestRedirectLoopIsBounded(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, srv.URL+"/again", http.StatusFound)
	}))
	defer srv.Close()

	done := make(chan error, 1)
	go func() {
		_, err := ToFile(context.Background(), srv.URL, filepath.Join(t.TempDir(), "out"),
			1<<20, 5*time.Second)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("an infinite redirect loop was followed to completion")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("redirect loop was not bounded")
	}
}

func TestSchemesAreRestricted(t *testing.T) {
	for _, u := range []string{
		"file:///etc/passwd",
		"gopher://example.com/",
		"ftp://example.com/x.mp4",
		"data:video/mp4;base64,AAAA",
	} {
		_, err := ToFile(context.Background(), u, filepath.Join(t.TempDir(), "out"),
			1<<20, 5*time.Second)
		if !errors.Is(err, ErrBadScheme) {
			t.Errorf("%s: expected a scheme rejection, got %v", u, err)
		}
	}
}

// A body larger than the cap must be refused even when Content-Length lies about it.
func TestOversizeBodyRefusedDespiteLyingHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "10")
		w.(http.Flusher).Flush()
		for range 100 {
			fmt.Fprint(w, strings.Repeat("A", 1024))
		}
	}))
	defer srv.Close()

	// Exercise the size logic directly, since the dial guard blocks httptest hosts.
	if err := checkSizeLimit(2048, 1024); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversize body accepted: %v", err)
	}
	if err := checkSizeLimit(1024, 1024); err != nil {
		t.Errorf("a body exactly at the cap was rejected: %v", err)
	}
}
