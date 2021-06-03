package domain

import "errors"

var NoVideo = VideoID("")

type VideoID string

func VideoIDFrom(raw string) (VideoID, error) {
	if raw == "" {
		return NoVideo, errors.New("invalid video ID")
	}

	return VideoID(raw), nil
}
