package example

import "testing"

func TestDeleteSpace(t *testing.T) {
	t.Run("errors when the space cannot be fetched", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.givenASpaceThatCannotBeFetched()
		test.whenDeletingTheSpace()
		test.thenACannotFetchSpaceErrorIsReturned()
	})

	t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.givenAUserThatDoesNotOwnASpace()
		test.whenDeletingTheSpace()
		test.thenANotAnOwnedSpaceErrorIsReturned()
	})

	t.Run("errors when the space is not empty", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.givenASpaceThatIsNotEmpty()
		test.whenDeletingTheSpace()
		test.thenANonEmptySpaceErrorIsReturned()
	})

	t.Run("errors when removing the default space", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.givenTheDefaultSpaceOfAUser()
		test.whenDeletingTheSpace()
		test.thenAnUnremovableDefaultSpaceErrorIsReturned()
	})

	t.Run("errors when the space cannot be removed", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.givenASpaceThatCannotBeRemoved()
		test.whenDeletingTheSpace()
		test.thenASpaceCouldNotBeRemovedErrorIsReturned()
	})

	t.Run("does not error on success", func(t *testing.T) {
		test := setupTestDeleteSpace(t)

		test.whenDeletingTheSpace()
		test.thenNoErrorsAreReturned()
	})
}

type testDeleteSpace struct {
	*testing.T

	spaceID string
	ownerID string

	returnedErr error
}

func setupTestDeleteSpace(t *testing.T) *testDeleteSpace {
	return &testDeleteSpace{
		T: t,

		spaceID: "known-space",
		ownerID: "known-owner",
	}
}

func (t *testDeleteSpace) givenASpaceThatCannotBeFetched() {
	t.spaceID = "unknown-space"
}

func (t *testDeleteSpace) givenAUserThatDoesNotOwnASpace() {
	t.ownerID = "unknown-owner"
}

func (t *testDeleteSpace) givenASpaceThatIsNotEmpty() {
	t.spaceID = "non-empty-space"
}

func (t *testDeleteSpace) givenTheDefaultSpaceOfAUser() {
	t.spaceID = "default-space"
}

func (t *testDeleteSpace) givenASpaceThatCannotBeRemoved() {
	t.spaceID = "fails-on-remove"
}

func (t *testDeleteSpace) whenDeletingTheSpace() {
	t.returnedErr = DeleteSpace(t.spaceID, t.ownerID)
}

func (t *testDeleteSpace) thenACannotFetchSpaceErrorIsReturned() {
	if t.returnedErr != errCannotFetchSpace {
		t.Fatalf("an error is expected when a space cannot be fetched, got %#v", t.returnedErr)
	}
}

func (t *testDeleteSpace) thenANotAnOwnedSpaceErrorIsReturned() {
	if t.returnedErr != errNotAnOwnedSpace {
		t.Fatalf("an error is expected when the space is not owned by the user removing it, got %#v", t.returnedErr)
	}
}

func (t *testDeleteSpace) thenANonEmptySpaceErrorIsReturned() {
	if t.returnedErr != errNonEmptySpace {
		t.Fatalf("an error is expected when the space is not empty, got %#v", t.returnedErr)
	}
}

func (t *testDeleteSpace) thenAnUnremovableDefaultSpaceErrorIsReturned() {
	if t.returnedErr != errUnremovableDefaultSpace {
		t.Fatalf("an error is expected when removing the default space, got %#v", t.returnedErr)
	}
}

func (t *testDeleteSpace) thenASpaceCouldNotBeRemovedErrorIsReturned() {
	if t.returnedErr != errSpaceCouldNotBeRemoved {
		t.Fatalf("an error is expected when a space cannot be removed, got %#v", t.returnedErr)
	}
}

func (t *testDeleteSpace) thenNoErrorsAreReturned() {
	if t.returnedErr != nil {
		t.Fatalf("no errors were expected, got %#v", t.returnedErr)
	}
}
