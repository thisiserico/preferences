package application

import "github.com/thisiserico/preferences/clean-architectures/hexagonal/example/player/domain"

type PlayNext func(currentlyPlaying string) (domain.VideoID, error)

func PlayNextUseCase(store domain.Store) PlayNext {
	return func(currentlyPlaying string) (domain.VideoID, error) {
		videoID, err := domain.VideoIDFrom(currentlyPlaying)
		if err != nil {
			return videoID, err
		}

		if !store.IsPlayNextEnabled() {
			return domain.NoVideo, nil
		}

		return store.NextAfter(videoID), nil
	}
}
