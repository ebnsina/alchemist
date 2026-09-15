package providers

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// CSV: the customer brings the list. This is how anything else gets across — a host
// that releases no original through its API, a provider we have no adapter for, or a
// library already sitting on a file server. The links are fetched through the same guarded
// path as a single URL import, so nothing here reaches somewhere it should not.
type CSV struct{}

func (CSV) Name() string  { return "csv" }
func (CSV) Label() string { return "A list of links" }

func (CSV) NeedsConfig() []ConfigField { return nil }

func (CSV) SecretSpec() Secret {
	return Secret{
		Label: "Your list",
		Hint:  "One link per line, or a CSV with a url column and an optional title.",
		Kind:  "text",
	}
}

// csvPage bounds one pass, matching the other adapters: the worker takes a page and
// queues the next, so a list of ten thousand makes steady progress.
const csvPage = 100

// parseCSV accepts both shapes people actually have: a bare list of URLs, and a CSV
// exported from somewhere with a header. Anything that is not an http(s) link is
// dropped rather than guessed at.
func parseCSV(body string) []Video {
	out := []Video{}
	seen := map[string]bool{}

	r := csv.NewReader(strings.NewReader(body))
	r.FieldsPerRecord = -1 // ragged rows are normal in hand-made files
	rows, err := r.ReadAll()
	if err != nil {
		// Not valid CSV at all — fall back to one link per line, which is what a
		// column pasted out of a spreadsheet looks like.
		rows = nil
		for _, line := range strings.Split(body, "\n") {
			rows = append(rows, []string{strings.TrimSpace(line)})
		}
	}

	urlCol, titleCol := 0, -1
	start := 0
	// A header only counts as one if its first row holds no link.
	if len(rows) > 0 && !hasLink(rows[0]) {
		for i, cell := range rows[0] {
			switch strings.ToLower(strings.TrimSpace(cell)) {
			case "url", "link", "source", "source_url":
				urlCol = i
			case "title", "name", "filename":
				titleCol = i
			}
		}
		start = 1
	}

	for _, row := range rows[start:] {
		if urlCol >= len(row) {
			continue
		}
		link := strings.TrimSpace(row[urlCol])
		// A row whose named column is empty may still be a bare list with the link
		// somewhere else on the line.
		if !isLink(link) {
			link = firstLink(row)
		}
		if link == "" || seen[link] {
			continue
		}
		seen[link] = true

		title := ""
		if titleCol >= 0 && titleCol < len(row) {
			title = strings.TrimSpace(row[titleCol])
		}
		if title == "" {
			title = nameFromURL(link)
		}
		out = append(out, Video{ID: link, Title: title})
	}
	return out
}

func isLink(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func hasLink(row []string) bool { return firstLink(row) != "" }

func firstLink(row []string) string {
	for _, cell := range row {
		if c := strings.TrimSpace(cell); isLink(c) {
			return c
		}
	}
	return ""
}

// The last path segment is the closest thing to a name a bare link carries.
func nameFromURL(link string) string {
	u, err := url.Parse(link)
	if err != nil {
		return link
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if last := parts[len(parts)-1]; last != "" {
		if unescaped, err := url.PathUnescape(last); err == nil {
			return unescaped
		}
		return last
	}
	return u.Host
}

func (c CSV) List(_ context.Context, creds Creds, cursor string) ([]Video, string, error) {
	all := parseCSV(creds.Secret)

	offset := 0
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("csv: bad cursor %q", cursor)
		}
		offset = n
	}
	if offset >= len(all) {
		return nil, "", nil
	}
	end := min(offset+csvPage, len(all))
	next := ""
	if end < len(all) {
		next = strconv.Itoa(end)
	}
	return all[offset:end], next, nil
}

// The id is the link, so there is nothing to resolve. It is re-checked rather than
// trusted: the list is stored as given, and a row that is not a link never becomes
// a fetch.
func (c CSV) DownloadURL(_ context.Context, _ Creds, id string) (string, error) {
	if !isLink(id) {
		return "", ErrNoDownload
	}
	return id, nil
}
