package live

import "testing"

// The URL is the whole of what a customer is handed, and each of the three is wrong
// in its own way if the branch slips: a WHIP URL with a key in the query is refused
// with a 401 that looks exactly like a wrong key, and an encoder URL without the
// placeholder publishes to a path nothing authorised.
func TestPublishURLPerSource(t *testing.T) {
	for _, c := range []struct{ protocol, want string }{
		{"camera", "http://edge:8889/abc/whip"},
		{"srt", "srt://edge:8890?streamid=publish:abc:publisher:YOUR_STREAM_KEY"},
		{"rtmp", "rtmp://edge:1935/abc?user=publisher&pass=YOUR_STREAM_KEY"},
	} {
		if got := publishURL(c.protocol, "edge", "abc"); got != c.want {
			t.Errorf("publishURL(%q) = %q, want %q", c.protocol, got, c.want)
		}
	}
}
