// Package fetch downloads customer-supplied URLs without letting them reach the
// inside of our network.
package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"syscall"
	"time"
)

var (
	ErrBlockedAddress = errors.New("blocked_address")
	ErrBadScheme      = errors.New("unsupported_scheme")
	ErrTooLarge       = errors.New("source_too_large")
	ErrTooManyHops    = errors.New("too_many_redirects")
	ErrUnreachable    = errors.New("source_unreachable")
)

const (
	maxRedirects = 5
	dialTimeout  = 10 * time.Second
)

// blocked covers every range that must never be reachable from a user-supplied URL.
// The one that matters most in practice is 169.254.0.0/16: cloud instance metadata
// lives at 169.254.169.254, and reaching it usually means handing over credentials.
var blocked = []string{
	"0.0.0.0/8",      // unspecified
	"10.0.0.0/8",     // private
	"100.64.0.0/10",  // carrier-grade NAT
	"127.0.0.0/8",    // loopback
	"169.254.0.0/16", // link-local, includes cloud metadata
	"172.16.0.0/12",  // private
	"192.0.0.0/24",   // IETF protocol assignments
	"192.168.0.0/16", // private
	"198.18.0.0/15",  // benchmarking
	"224.0.0.0/4",    // multicast
	"240.0.0.0/4",    // reserved, includes broadcast
	"::/128",         // unspecified
	"::1/128",        // loopback
	"fc00::/7",       // unique local
	"fe80::/10",      // link-local
	"ff00::/8",       // multicast
	"64:ff9b::/96",   // IPv4/IPv6 translation
}

var blockedNets = func() []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(blocked))
	for _, c := range blocked {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			panic("fetch: bad CIDR " + c)
		}
		nets = append(nets, n)
	}
	return nets
}()

// devAllowlist holds exact "ip:port" pairs that bypass the address rules. It exists
// so a local harness can serve a source file from loopback, which the guard would
// otherwise refuse -- correctly.
//
// Deliberately an exact allowlist rather than a blanket "permit private ranges"
// switch: with a blanket switch, a test proving that a redirect to 169.254.169.254
// is refused would silently stop proving anything. Entries here are exact, so every
// address not listed stays blocked.
var devAllowlist = map[string]bool{}

// SetDevAllowlist configures the bypass. Production passes nothing.
func SetDevAllowlist(entries []string) {
	devAllowlist = make(map[string]bool, len(entries))
	for _, e := range entries {
		if e = strings.TrimSpace(e); e != "" {
			devAllowlist[e] = true
		}
	}
}

// DevAllowlistSize reports how many bypass entries are active, so the caller can warn.
func DevAllowlistSize() int { return len(devAllowlist) }

// IsBlocked reports whether an address must not be connected to.
func IsBlocked(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() {
		return true
	}
	// match every IPv4 address and block the entire internet. Normalise to 4 bytes
	// instead: that also collapses ::ffff:127.0.0.1 onto the 127.0.0.0/8 rule, so
	// the mapped form cannot be used to slip past the IPv4 ranges.
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	for _, n := range blockedNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// safeDialer validates the address at connect time rather than only resolving it
// first. Checking up front and connecting afterwards is a TOCTOU window: a hostname
// can resolve to a public address for the check and a private one for the connection
// (DNS rebinding). Control runs after resolution with the address actually about to
// be dialled, which closes that window.
func safeDialer() *net.Dialer {
	return &net.Dialer{
		Timeout: dialTimeout,
		Control: func(network, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return ErrBlockedAddress
			}
			if devAllowlist[address] {
				return nil
			}
			ip := net.ParseIP(host)
			if IsBlocked(ip) {
				return fmt.Errorf("%w: %s", ErrBlockedAddress, host)
			}
			return nil
		},
	}
}

// Client builds an HTTP client that cannot be pointed at internal infrastructure.
func Client(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext:           safeDialer().DialContext,
			TLSHandshakeTimeout:   dialTimeout,
			ResponseHeaderTimeout: 30 * time.Second,
			DisableKeepAlives:     true,
		},
		// Every hop is re-validated, because only the first URL was ever vetted and a
		// redirect is entirely under the remote server's control.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return ErrTooManyHops
			}
			return checkScheme(req.URL)
		},
	}
}

func checkScheme(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%w: %s", ErrBadScheme, u.Scheme)
	}
	return nil
}

// checkSizeLimit enforces the cap. Split out so the rule can be tested without
// standing up a server the dial guard would refuse to reach anyway.
func checkSizeLimit(got, max int64) error {
	if got > max {
		return ErrTooLarge
	}
	return nil
}

// ToFile downloads src to dst, refusing anything larger than maxBytes.
func ToFile(ctx context.Context, src, dst string, maxBytes int64, timeout time.Duration) (int64, error) {
	u, err := url.Parse(src)
	if err != nil {
		return 0, fmt.Errorf("%w: malformed url", ErrUnreachable)
	}
	if err := checkScheme(u); err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	req.Header.Set("User-Agent", "Alchemist/1.0")

	resp, err := Client(timeout).Do(req)
	if err != nil {
		if errors.Is(err, ErrBlockedAddress) {
			return 0, ErrBlockedAddress
		}
		if errors.Is(err, ErrTooManyHops) {
			return 0, ErrTooManyHops
		}
		return 0, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%w: upstream returned %d", ErrUnreachable, resp.StatusCode)
	}
	// Trust Content-Length only to fail early; the real limit is enforced below,
	// because a lying or absent header must not let an unbounded body through.
	if resp.ContentLength > 0 && resp.ContentLength > maxBytes {
		return 0, ErrTooLarge
	}

	f, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	// Read one byte past the limit: hitting it proves the body was oversized rather
	// than exactly at the cap.
	n, err := io.Copy(f, io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return n, fmt.Errorf("%w: %v", ErrUnreachable, err)
	}
	if err := checkSizeLimit(n, maxBytes); err != nil {
		os.Remove(dst)
		return n, err
	}
	return n, nil
}
