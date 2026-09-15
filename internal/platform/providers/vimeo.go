package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// Vimeo. The token needs the `public`, `private` and `video_files` scopes and a paid
// plan; without video_files the download and progressive fields simply come back
// absent, which reads as "no downloadable file" rather than as an error.
//
// Links are 302s to a CDN with a few hours' TTL, so they are resolved per video at
// import time and never stored.
type Vimeo struct{}

func (Vimeo) Name() string  { return "vimeo" }
func (Vimeo) Label() string { return "Vimeo" }

func (Vimeo) NeedsConfig() []ConfigField { return nil }

func (Vimeo) SecretSpec() Secret {
	return Secret{
		Label: "Vimeo access token",
		Hint:  "Needs the public, private and video_files scopes, on a paid plan.",
		Kind:  "key",
	}
}

type vimeoFile struct {
	Quality   string `json:"quality"`
	Rendition string `json:"rendition"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Size      int64  `json:"size"`
	Link      string `json:"link"`
}

type vimeoVideo struct {
	URI      string      `json:"uri"`
	Name     string      `json:"name"`
	Duration float64     `json:"duration"`
	Download []vimeoFile `json:"download"`
	Play     struct {
		Progressive []vimeoFile `json:"progressive"`
	} `json:"play"`
}

type vimeoPage struct {
	Data   []vimeoVideo `json:"data"`
	Paging struct {
		Next string `json:"next"`
	} `json:"paging"`
}

// The API returns "/videos/12345"; the bare number is what we key on.
func vimeoID(uri string) string {
	return strings.TrimPrefix(uri, "/videos/")
}

func (v Vimeo) List(ctx context.Context, c Creds, cursor string) ([]Video, string, error) {
	path := cursor
	if path == "" {
		path = "/me/videos?per_page=50&fields=uri,name,duration"
	}
	res, err := do(ctx, "GET", "https://api.vimeo.com"+path, map[string]string{
		"Authorization": "bearer " + c.Secret,
		"Accept":        "application/vnd.vimeo.*+json;version=3.4",
	})
	if err != nil {
		return nil, "", err
	}
	defer res.Body.Close()

	var page vimeoPage
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		return nil, "", fmt.Errorf("vimeo list: %w", err)
	}
	out := make([]Video, 0, len(page.Data))
	for _, d := range page.Data {
		out = append(out, Video{ID: vimeoID(d.URI), Title: d.Name, Duration: d.Duration})
	}
	return out, page.Paging.Next, nil
}

func (v Vimeo) DownloadURL(ctx context.Context, c Creds, id string) (string, error) {
	if _, err := strconv.Atoi(id); err != nil {
		return "", fmt.Errorf("vimeo: %q is not a video id", id)
	}
	res, err := do(ctx, "GET",
		"https://api.vimeo.com/videos/"+url.PathEscape(id)+"?fields=download,play.progressive",
		map[string]string{
			"Authorization": "bearer " + c.Secret,
			"Accept":        "application/vnd.vimeo.*+json;version=3.4",
		})
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var vid vimeoVideo
	if err := json.NewDecoder(res.Body).Decode(&vid); err != nil {
		return "", fmt.Errorf("vimeo video: %w", err)
	}

	// `download` is the original where the plan allows it. `play.progressive` is the
	// transcoded ladder and is the fallback: worse than the source, but a real file,
	// and better than refusing to move the video at all.
	if link := bestVimeo(vid.Download); link != "" {
		return link, nil
	}
	if link := bestVimeo(vid.Play.Progressive); link != "" {
		return link, nil
	}
	return "", ErrNoDownload
}

// Largest wins: source first if it is there, otherwise the tallest rendition.
func bestVimeo(files []vimeoFile) string {
	usable := make([]vimeoFile, 0, len(files))
	for _, f := range files {
		if f.Link != "" {
			usable = append(usable, f)
		}
	}
	if len(usable) == 0 {
		return ""
	}
	sort.SliceStable(usable, func(i, j int) bool {
		if usable[i].Quality == "source" != (usable[j].Quality == "source") {
			return usable[i].Quality == "source"
		}
		if usable[i].Height != usable[j].Height {
			return usable[i].Height > usable[j].Height
		}
		return usable[i].Size > usable[j].Size
	})
	return usable[0].Link
}
