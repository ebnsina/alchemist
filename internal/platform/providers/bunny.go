package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Bunny Stream. Two things are needed beyond the key: the library id, and MP4
// Fallback switched on for that library — Bunny only generates a downloadable file
// when it is, and only for videos uploaded afterwards. The fallback also caps at
// 1080p, so a 4K master on Bunny comes across as 1080p. That is Bunny's ceiling, not
// ours, and the dashboard says so rather than letting it surprise anyone later.
type Bunny struct{}

func (Bunny) Name() string  { return "bunny" }
func (Bunny) Label() string { return "Bunny Stream" }

func (Bunny) NeedsConfig() []ConfigField {
	return []ConfigField{
		{
			Key:   "library_id",
			Label: "Video library ID",
			Hint:  "The number in the Stream library URL.",
		},
		{
			Key:   "pull_zone",
			Label: "Pull zone hostname",
			Hint:  "Where the files are served from, like vz-1a2b3c.b-cdn.net.",
		},
	}
}

func (Bunny) SecretSpec() Secret {
	return Secret{
		Label: "Stream API key",
		Hint:  "The library's API key, from its API tab.",
		Kind:  "key",
	}
}

type bunnyVideo struct {
	GUID                 string `json:"guid"`
	Title                string `json:"title"`
	Length               int    `json:"length"`
	Status               int    `json:"status"`
	HasMP4Fallback       bool   `json:"hasMP4Fallback"`
	AvailableResolutions string `json:"availableResolutions"`
}

type bunnyPage struct {
	TotalItems  int          `json:"totalItems"`
	CurrentPage int          `json:"currentPage"`
	Items       []bunnyVideo `json:"items"`
}

func (b Bunny) List(ctx context.Context, c Creds, cursor string) ([]Video, string, error) {
	lib := c.Config["library_id"]
	if lib == "" {
		return nil, "", fmt.Errorf("bunny: no library id")
	}
	page := 1
	if cursor != "" {
		n, err := strconv.Atoi(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("bunny: bad cursor %q", cursor)
		}
		page = n
	}

	res, err := do(ctx, "GET",
		fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos?page=%d&itemsPerPage=100",
			url.PathEscape(lib), page),
		map[string]string{"AccessKey": c.Secret, "accept": "application/json"})
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	var p bunnyPage
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, "", fmt.Errorf("bunny list: %w", err)
	}

	out := make([]Video, 0, len(p.Items))
	for _, it := range p.Items {
		out = append(out, Video{ID: it.GUID, Title: it.Title, Duration: float64(it.Length)})
	}
	// Bunny pages by number and reports the total, so "no more" is a full page short.
	next := ""
	if len(p.Items) > 0 && page*100 < p.TotalItems {
		next = strconv.Itoa(page + 1)
	}
	return out, next, nil
}

func (b Bunny) DownloadURL(ctx context.Context, c Creds, id string) (string, error) {
	lib, zone := c.Config["library_id"], strings.TrimSuffix(c.Config["pull_zone"], "/")
	if lib == "" || zone == "" {
		return "", fmt.Errorf("bunny: library id and pull zone are both required")
	}

	res, err := do(ctx, "GET",
		fmt.Sprintf("https://video.bunnycdn.com/library/%s/videos/%s",
			url.PathEscape(lib), url.PathEscape(id)),
		map[string]string{"AccessKey": c.Secret, "accept": "application/json"})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var v bunnyVideo
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return "", fmt.Errorf("bunny video: %w", err)
	}
	// Status 4 is finished. Anything else has no file to hand over yet.
	if !v.HasMP4Fallback || v.Status != 4 {
		return "", ErrNoDownload
	}
	height := tallestBunny(v.AvailableResolutions)
	if height == 0 {
		return "", ErrNoDownload
	}
	scheme := "https://"
	if strings.HasPrefix(zone, "http") {
		scheme = ""
	}
	return fmt.Sprintf("%s%s/%s/play_%dp.mp4", scheme, zone, id, height), nil
}

// "240p,360p,720p" — take the tallest, since that is the best file Bunny will serve.
func tallestBunny(list string) int {
	best := 0
	for _, part := range strings.Split(list, ",") {
		n, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(part), "p"))
		if err == nil && n > best {
			best = n
		}
	}
	return best
}
