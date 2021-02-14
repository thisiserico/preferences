package example

import "errors"

const defaultSpaceName = "default"

var (
	errCannotFetchSpace        = errors.New("cannot fetch the space")
	errNotAnOwnedSpace         = errors.New("only owned spaces can be removed")
	errNonEmptySpace           = errors.New("only empty spaces can be removed")
	errUnremovableDefaultSpace = errors.New("the default space cannot be removed")
	errSpaceCouldNotBeRemoved  = errors.New("the space could not be removed")
)

//go:generate mockgen -self_package=github.com/thisiserico/how-to/code-semantics/test-doubles/example -package=example -destination=./double_store_test.go -mock_names=Store=StoreDouble github.com/thisiserico/how-to/code-semantics/test-doubles/example Store

type Store interface {
	FetchSpaceWithID(id string) *Space
	RemoveSpaceWithID(id string) error
}

type resource struct{}

type Space struct {
	id        string
	ownerID   string
	name      string
	resources []resource
}

func (s Space) isOwnedBy(userID string) bool {
	return s.ownerID == userID
}

func (s Space) isEmpty() bool {
	return len(s.resources) == 0
}

func (s Space) isTheDefault() bool {
	return s.name == defaultSpaceName
}

func DeleteSpace(store Store, id string, whoAmI string) error {
	space := store.FetchSpaceWithID(id)
	if space == nil {
		return errCannotFetchSpace
	}

	if !space.isOwnedBy(whoAmI) {
		return errNotAnOwnedSpace
	}

	if !space.isEmpty() {
		return errNonEmptySpace
	}

	if space.isTheDefault() {
		return errUnremovableDefaultSpace
	}

	return store.RemoveSpaceWithID(id)
}
