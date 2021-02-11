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

func DeleteSpace(id string, whoAmI string) error {
	space := fetchSpaceWithID(id)
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

	return removeSpaceWithID(id)
}

func fetchSpaceWithID(id string) *Space {
	space := &Space{
		id:        "known-space",
		ownerID:   "known-owner",
		name:      "a space name",
		resources: nil,
	}

	switch id {
	case "unknown-space":
		space = nil

	case "non-empty-space":
		space.resources = []resource{resource{}}

	case "default-space":
		space.name = defaultSpaceName
	}

	return space
}

func removeSpaceWithID(id string) error {
	if id == "fails-on-remove" {
		return errSpaceCouldNotBeRemoved
	}

	return nil
}
