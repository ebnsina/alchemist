package media

import "errors"

// Stable sentinel errors. The API maps these to customer-facing codes; raw ffmpeg
// output never crosses the API boundary.
var (
	ErrUnreadableSource = errors.New("unreadable_source")
	ErrNoVideoStream    = errors.New("no_video_stream")
	ErrEncodeFailed     = errors.New("encode_failed")
	ErrStitchFailed     = errors.New("stitch_failed")
	ErrPackageFailed    = errors.New("package_failed")
)
