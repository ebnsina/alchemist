package providers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// Picking the file is the only real decision an adapter makes. Choosing the smallest
// rendition would silently downgrade an entire migrated library, and nobody would
// notice until they watched one.
func TestVimeoPrefersSourceThenTallest(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files []vimeoFile
		want  string
	}{
		{"source beats a taller rendition", []vimeoFile{
			{Quality: "hd", Height: 1080, Link: "hd"},
			{Quality: "source", Height: 720, Link: "src"},
		}, "src"},
		{"no source: tallest wins", []vimeoFile{
			{Quality: "sd", Height: 540, Link: "sd"},
			{Quality: "hd", Height: 1080, Link: "hd"},
		}, "hd"},
		{"entries without a link are ignored", []vimeoFile{
			{Quality: "source", Height: 2160, Link: ""},
			{Quality: "hd", Height: 720, Link: "hd"},
		}, "hd"},
		{"nothing usable", []vimeoFile{{Quality: "hd", Link: ""}}, ""},
		{"empty", nil, ""},
	} {
		if got := bestVimeo(tc.files); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestBunnyTallestResolution(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{"240p,360p,720p", 720},
		{"1080p", 1080},
		{" 480p , 720p ", 720},
		{"", 0},
		{"garbage", 0},
	} {
		if got := tallestBunny(tc.in); got != tc.want {
			t.Errorf("tallestBunny(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// A rejected key is the one failure a customer can fix, so it must not arrive looking
// like a transport error.
func TestRejectedCredentialsAreDistinct(t *testing.T) {
	for _, code := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(code)
		}))
		_, err := do(context.Background(), "GET", srv.URL, nil)
		srv.Close()
		if err != ErrBadCredentials {
			t.Errorf("status %d gave %v, want ErrBadCredentials", code, err)
		}
	}
}

// A video Bunny has not finished, or that predates MP4 Fallback, is skipped — not
// failed. Treating it as a failure would make a partial migration look broken.
func TestBunnySkipsVideoWithNoFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"guid":"g","status":4,"hasMP4Fallback":false}`))
	}))
	defer srv.Close()

	// Point the adapter at the stub by overriding the shared client's transport.
	old := client.Transport
	client.Transport = rewriteHost{to: srv.Listener.Addr().String()}
	defer func() { client.Transport = old }()

	_, err := Bunny{}.DownloadURL(context.Background(), Creds{
		Secret: "k",
		Config: map[string]string{"library_id": "1", "pull_zone": "vz.example"},
	}, "g")
	if err != ErrNoDownload {
		t.Fatalf("got %v, want ErrNoDownload", err)
	}
}

type rewriteHost struct{ to string }

func (r rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = r.to
	return http.DefaultTransport.RoundTrip(req)
}

// The list is whatever the customer pasted. Parsing it wrongly either drops videos
// silently or turns a stray cell into a fetch, so both directions are checked.
func TestParseCSVAcceptsTheShapesPeopleHave(t *testing.T) {
	for _, tc := range []struct {
		name  string
		body  string
		want  []string
		title string
	}{
		{
			name: "bare list, one per line",
			body: "https://a.example/1.mp4\nhttps://a.example/2.mp4\n",
			want: []string{"https://a.example/1.mp4", "https://a.example/2.mp4"},
		},
		{
			name:  "header with url and title columns",
			body:  "title,url\nWeek one,https://a.example/1.mp4\n",
			want:  []string{"https://a.example/1.mp4"},
			title: "Week one",
		},
		{
			name: "header naming the column link",
			body: "link\nhttps://a.example/1.mp4\n",
			want: []string{"https://a.example/1.mp4"},
		},
		{
			name: "no header, link is not the first column",
			body: "Week one,https://a.example/1.mp4\n",
			want: []string{"https://a.example/1.mp4"},
		},
		{
			name: "blank lines and duplicates are dropped",
			body: "https://a.example/1.mp4\n\nhttps://a.example/1.mp4\n\n",
			want: []string{"https://a.example/1.mp4"},
		},
		{
			name: "non-links are ignored, not guessed at",
			body: "url\nnot a link\nftp://a.example/1.mp4\n/relative/1.mp4\nhttps://a.example/ok.mp4\n",
			want: []string{"https://a.example/ok.mp4"},
		},
		{"empty", "", nil, ""},
	} {
		got := parseCSV(tc.body)
		if len(got) != len(tc.want) {
			t.Errorf("%s: got %d videos, want %d (%v)", tc.name, len(got), len(tc.want), got)
			continue
		}
		for i := range got {
			if got[i].ID != tc.want[i] {
				t.Errorf("%s: video %d = %q, want %q", tc.name, i, got[i].ID, tc.want[i])
			}
		}
		if tc.title != "" && got[0].Title != tc.title {
			t.Errorf("%s: title = %q, want %q", tc.name, got[0].Title, tc.title)
		}
	}
}

// A bare link still needs a name a person can recognise in the list.
func TestCSVNamesFallBackToTheFilename(t *testing.T) {
	got := parseCSV("https://a.example/lectures/week%20one.mp4\n")
	if len(got) != 1 || got[0].Title != "week one.mp4" {
		t.Fatalf("got %+v, want title %q", got, "week one.mp4")
	}
}

// Paging keeps one pass bounded, and the last page must report no cursor or the
// worker queues itself forever.
func TestCSVPagesAndTerminates(t *testing.T) {
	var b strings.Builder
	for i := range csvPage + 5 {
		fmt.Fprintf(&b, "https://a.example/%d.mp4\n", i)
	}
	c := CSV{}
	first, next, err := c.List(context.Background(), Creds{Secret: b.String()}, "")
	if err != nil || len(first) != csvPage || next != strconv.Itoa(csvPage) {
		t.Fatalf("first page: %d items, cursor %q, err %v", len(first), next, err)
	}
	second, next, err := c.List(context.Background(), Creds{Secret: b.String()}, next)
	if err != nil || len(second) != 5 || next != "" {
		t.Fatalf("last page: %d items, cursor %q, err %v", len(second), next, err)
	}
}

// The id is a URL that came from a person. Handing it back without re-checking would
// make the parser the only thing standing between a stray cell and a fetch.
func TestCSVRefusesANonLinkID(t *testing.T) {
	if _, err := (CSV{}).DownloadURL(context.Background(), Creds{}, "file:///etc/passwd"); err != ErrNoDownload {
		t.Fatalf("got %v, want ErrNoDownload", err)
	}
}
