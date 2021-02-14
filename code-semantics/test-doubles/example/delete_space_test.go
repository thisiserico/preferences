package example

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func TestDeleteSpace(t *testing.T) {
	t.Run("errors when the space cannot be fetched", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.givenASpaceThatCannotBeFetched()
		test.whenDeletingTheSpace()
		test.thenACannotFetchSpaceErrorIsReturned()
	})

	t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.givenAUserThatDoesNotOwnASpace()
		test.whenDeletingTheSpace()
		test.thenANotAnOwnedSpaceErrorIsReturned()
	})

	t.Run("errors when the space is not empty", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.givenASpaceThatIsNotEmpty()
		test.whenDeletingTheSpace()
		test.thenANonEmptySpaceErrorIsReturned()
	})

	t.Run("errors when removing the default space", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.givenTheDefaultSpaceOfAUser()
		test.whenDeletingTheSpace()
		test.thenAnUnremovableDefaultSpaceErrorIsReturned()
	})

	t.Run("errors when the space cannot be removed", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.givenASpaceThatCannotBeRemoved()
		test.whenDeletingTheSpace()
		test.thenASpaceCouldNotBeRemovedErrorIsReturned()
	})

	t.Run("does not error on success", func(t *testing.T) {
		test := setupTestDeleteSpace(t)
		t.Cleanup(test.cleanup)

		test.whenDeletingTheSpace()
		test.thenNoErrorsAreReturned()
	})
}

type testDeleteSpace struct {
	*testing.T

	ctrl  *gomock.Controller
	store *StoreDouble

	spaceID string
	ownerID string

	returnedErr error
}

func setupTestDeleteSpace(t *testing.T) *testDeleteSpace {
	ctrl := gomock.NewController(t)

	return &testDeleteSpace{
		T: t,

		ctrl:  ctrl,
		store: NewStoreDouble(ctrl),

		spaceID: "known-space",
		ownerID: "known-owner",
	}
}

func (t *testDeleteSpace) cleanup() {
	t.ctrl.Finish()
}

func (t *testDeleteSpace) givenASpaceThatCannotBeFetched() {
	t.store.
		EXPECT().
		FetchSpaceWithID(gomock.Any()).
		Return(nil)
}

func (t *testDeleteSpace) givenAUserThatDoesNotOwnASpace() {
	t.store.
		EXPECT().
		FetchSpaceWithID(gomock.Any()).
		Return(&Space{
			id:      t.spaceID,
			ownerID: "unknown-owner",
			name:    "not a default space",
		})
}

func (t *testDeleteSpace) givenASpaceThatIsNotEmpty() {
	t.store.
		EXPECT().
		FetchSpaceWithID(gomock.Any()).
		Return(&Space{
			id:        t.spaceID,
			ownerID:   t.ownerID,
			name:      "not a default space",
			resources: []resource{resource{}},
		})
}

func (t *testDeleteSpace) givenTheDefaultSpaceOfAUser() {
	t.store.
		EXPECT().
		FetchSpaceWithID(gomock.Any()).
		Return(&Space{
			id:      t.spaceID,
			ownerID: t.ownerID,
			name:    "default",
		})
}

func (t *testDeleteSpace) givenASpaceThatCannotBeRemoved() {
	t.store.
		EXPECT().
		RemoveSpaceWithID(gomock.Any()).
		Return(errSpaceCouldNotBeRemoved)
}

func (t *testDeleteSpace) whenDeletingTheSpace() {
	t.store.
		EXPECT().
		FetchSpaceWithID(t.spaceID).
		Return(&Space{
			id:      t.spaceID,
			ownerID: t.ownerID,
			name:    "not a default space",
		}).
		AnyTimes()

	t.store.
		EXPECT().
		RemoveSpaceWithID(t.spaceID).
		AnyTimes()

	t.returnedErr = DeleteSpace(t.store, t.spaceID, t.ownerID)
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
