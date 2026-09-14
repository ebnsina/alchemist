// Package media wraps ffmpeg and shaka-packager. Everything here is a library call
// so the same code serves VOD batch jobs today and live segments later.
package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type Probe struct {
	DurationSec float64
	Width       int
	Height      int
	FrameRate   float64
	VideoCodec  string
	AudioCodec  string
	HasAudio    bool
	Rotation    int
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecType    string `json:"codec_type"`
		CodecName    string `json:"codec_name"`
		Width        int    `json:"width"`
		Height       int    `json:"height"`
		AvgFrameRate string `json:"avg_frame_rate"`
		SideData     []struct {
			Rotation int `json:"rotation"`
		} `json:"side_data_list"`
	} `json:"streams"`
}

// Inspect reads stream metadata. It rejects unusable sources here rather than
// failing forty minutes into an encode.
func Inspect(ctx context.Context, path string) (*Probe, error) {
	out, err := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet", "-print_format", "json",
		"-show_format", "-show_streams", path).Output()
	if err != nil {
		return nil, fmt.Errorf("%w: ffprobe could not read the file", ErrUnreadableSource)
	}

	var raw ffprobeOutput
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("%w: unexpected ffprobe output", ErrUnreadableSource)
	}

	p := &Probe{}
	p.DurationSec, _ = strconv.ParseFloat(raw.Format.Duration, 64)

	for _, s := range raw.Streams {
		switch s.CodecType {
		case "video":
			if p.Width != 0 {
				continue // first video stream wins
			}
			p.Width, p.Height, p.VideoCodec = s.Width, s.Height, s.CodecName
			p.FrameRate = parseRate(s.AvgFrameRate)
			for _, sd := range s.SideData {
				if sd.Rotation != 0 {
					p.Rotation = sd.Rotation
				}
			}
		case "audio":
			p.HasAudio, p.AudioCodec = true, s.CodecName
		}
	}

	if p.Width == 0 || p.Height == 0 {
		return nil, fmt.Errorf("%w: no decodable video stream", ErrNoVideoStream)
	}
	if p.DurationSec <= 0 {
		return nil, fmt.Errorf("%w: duration is zero or unknown", ErrUnreadableSource)
	}
	if p.FrameRate <= 0 || p.FrameRate > 240 {
		return nil, fmt.Errorf("%w: implausible frame rate %.2f", ErrUnreadableSource, p.FrameRate)
	}
	return p, nil
}

// parseRate handles ffprobe's "30000/1001" rational form.
func parseRate(s string) float64 {
	num, den, ok := strings.Cut(s, "/")
	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0
	}
	if !ok {
		return n
	}
	d, err := strconv.ParseFloat(den, 64)
	if err != nil || d == 0 {
		return 0
	}
	return n / d
}
