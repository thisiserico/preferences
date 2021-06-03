package domain

type Store interface {
	IsPlayNextEnabled() bool
	NextAfter(VideoID) VideoID
}
