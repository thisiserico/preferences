# Test doubles

Heads up, this is a spin off of the [semantic testing][tests] chapter.
Make sure you're familiar with the concepts there to be able to follow along here.

The short version: test doubles 👯‍♀️ are a mechanism to replace a piece of software with a slightly altered
version of that piece, normally to be used for dependencies in conjunction with tests.
The double, depending on its kind, will let you either guide your tests, assert a bunch of expectations in regards to
that piece or simply make your build pass, among others.

The long version: it's just way too long to fit in this page.
But moving forward, I'll have to assume you've got an average understanding of what a `stub` –the ones that let you
guide tests– and a `mock` –the ones that let you assert behaviour– are.

As we already spent a bunch of time with the `DeleteSpace` scenario, let's stick to it.
However, our end goal is to end up using a new `Store` dependency in order to fetch and remove spaces from.
This is what our store will look like:

```go
//go:generate mockgen -self_package=github.com/thisiserico/preferences/code-semantics/test-doubles/example -package=example -destination=./double_store_test.go -mock_names=Store=StoreDouble github.com/thisiserico/preferences/code-semantics/test-doubles/example Store

type Store interface {
    FetchSpaceWithID(id string) *Space
    RemoveSpaceWithID(id string) error
}
```

We made it an interface for different reasons.
First of all, because otherwise we can't really create a double out of it in `go`.
And secondly, because we don't yet know what the actual contents of the methods will be.
Even in that case, chances are we'll keep the interface, as we can sense that the contents will depend on an actual
implementation. We'll discuss this further when we talk about [hexagonal architecture][hexagonal].

See that `//go:generate` comment above the interface declaration? Yes, it looks cumbersome.
Feel free to ignore that as it's not that relevant for the understanding of the example.
Just know that it's the way we have in `go` to autogenerate code when running `go generate ./...` 🤷‍♂️
It ends up creating a `StoreDouble` struct that we can use as, guess what, a store double.
Other languages differ but there's always a way to create your doubles.

And now let's prepare our `DeleteSpace` method to make use of the interface:

```go
func DeleteSpace(store Store, id string, whoAmI string) error {
    // Nothing changes inside this method for now.
}
```

Notice that, although we're injecting the store into the method, we're not using it yet!
Right after introducing this new parameter, our tests should start failing as the method signature doesn't match with
what we're sending into it. Let's fix that:

```go
func (t *testDeleteSpace) whenDeletingTheSpace() {
    t.returnedErr = DeleteSpace(nil, t.spaceID, t.ownerID)
}
```

That's it! Because there's a unique place for us to trigger the action being tested, only a line of code is
required. And because our method expects and interface, we can simply pass along a `nil` –which is actually called
a dummy test double, a double that does nothing but satisfy a dependency. But hey, we didn't have to modify our
scenarios, those that look like acceptance criteria. Seems reasonable as our actual business logic didn't change.

But honestly, we've been risking it here... We've introduced a code change and updated our tests afterwards.
If each test is telling us exactly how to setup the scenario and what the expectations are, why not work on those
upfront? That way, we'd transform a passing test into a failing one and work on the actual code to make it pass again.
Seems like an interesting approach to me, as we'd know exactly when our change went from invalid to valid.
Testing first is an interesting technique, but that's not what this chapter is about.
Do whatever feels comfortable to you at this point. However, that's the approach I'll follow here.

It's time to start making use of the store. The way we were guiding our tests to either succeed or fail was by using
certain space and owner identifiers. That's exactly what we need to change, but not just yet.
Before that, we need to make the store double available within our tests.

```go
type testDeleteSpace struct {
    *testing.T

    ctrl  *gomock.Controller
    store *StoreDouble

    spaceID string
    ownerID string

    returnedErr error
}
```

Two struct attributes have been added: some kind of controller and the actual store double.
The controller is used by `gomock` to assert the expectations we set up on the double.
It does so when using a `Finish` method of its own, that needs to be called at the end of each scenario.
That means that we might need to use a clean up method, just like we use a setup one, on each scenario.
Here's one as example, the rest of them need to change accordingly.

```go
t.Run("errors when the space cannot be fetched", func(t *testing.T) {
    test := setupTestDeleteSpace(t)
    t.Cleanup(test.cleanup)

    test.givenASpaceThatCannotBeFetched()
    test.whenDeletingTheSpace()
    test.thenACannotFetchSpaceErrorIsReturned()
})
```

And you can see we need a `cleanup` custom method now. Besides that, before calling `Finish` on the controller,
we're gonna need for it to be initialized. Therefore, here's what our setup and clean up methods will look like:

```go
func setupTestDeleteSpace(t *testing.T) *testDeleteSpace {
    ctrl := gomock.NewController(t)

    return &testDeleteSpace{
        T: t,

        ctrl: ctrl,

        spaceID: "known-space",
        ownerID: "known-owner",
    }
}

func (t *testDeleteSpace) cleanup() {
    t.ctrl.Finish()
}
```

Well, too much setting up controllers and not that much using the store so far 🙅‍♀️
Just one more step in that regards and we can move on to the actual test changes.
Let's also initialize the store double so it's ready to be used.

```go
// Below the ctrl: ctrl, line in the setup method goes the following:
store: NewStoreDouble(ctrl),

// And because we already want to forget about it, this is the line in the when method:
t.returnedErr = DeleteSpace(t.store, t.spaceID, t.ownerID)
```

Our store double is ready to be used! Our dummy is still a dummy, but it's now prepared to actually work.
Bear in mind our scenarios are still untouched.
We've only been changing the way our internals work, why would they have to change? 🤔

Time to change the way our code works. But let's not get too cocky and start changing everything.
As the operation less used is the actual space removal, we'll start by that one.
Two scenarios impact that one, the one testing what happens if the operation fails and the happy path.
For the first one, we need to guide our execution into believing it needs to fail.
Whilst in the second one, we need to make sure that the store method is being called the way we expect it to.
That can be translated into a stub and a mock respectively.
In both cases, our scenario expectations –the `thens`– don't need to change.
The business rules for them both are still the same.
What we do need to change is the way the context is set up –the `givens`.
The change for the first scenario is easy to locate: within its own `given`.
On the other hand, as we don't have a given for the happy path, we'll have to improvise and modify our `when`.

```go
func (t *testDeleteSpace) givenASpaceThatCannotBeRemoved() {
    t.store.
        EXPECT().
        RemoveSpaceWithID(gomock.Any()).
        Return(errSpaceCouldNotBeRemoved)
}

func (t *testDeleteSpace) whenDeletingTheSpace() {
    t.store.
        EXPECT().
        RemoveSpaceWithID(t.spaceID).
        AnyTimes()

    t.returnedErr = DeleteSpace(t.store, t.spaceID, t.ownerID)
}
```

We need to guide our failing scenario to get into failure mode. As a stub, we're not that interested in what the input
arguments are for the `RemoveSpaceWithID` method, only its return value. On the other hand, regarding the happy path
scenario, we do care about what input parameters we're passing into it. Therefore, we use a mock and set up an
expectation by indicating the space ID that is supposed to be sent. As this last block in the `when` will be run
on each scenario, even when a previous double is configured, we need to indicate that the double should be prepared
only when not done before. `AnyTimes` is the way to do so when using `gomock`. Other libraries will differ.
I tend to encapsulate these default behaviours in an `assignDefaultDoublesBehaviour` method, if that helps.

After these changes, our tests went back to failing; we need to address this by using our store dependency:

```go
return store.RemoveSpaceWithID(id)
```

We changed the last line of our `DeleteSpace` method and can now remove the unused `removeSpaceWithID` method.
Back to passing tests 🥳 Take a moment to understand what happened here. We're virtually doing the same we were doing
before: guide execution paths. Only this time, it's the test double the one leading those paths.

It's now time to guide the tests to get into the appropriate cases when fetching the space.
To see how our double will look for that operation, we can start by the generic case for the happy path,
the one we'll place in the `when` (or that `assignDefaultDoublesBehaviour` method if we use it).

```go
t.store.
    EXPECT().
    FetchSpaceWithID(t.spaceID).
    Return(&Space{
        id:        t.spaceID,
        ownerID:   t.ownerID,
        name:      "not a default space",
        resources: nil,
    }).
    AnyTimes()
```

That's what we need right on top within the `whenDeletingTheSpace` test helper. We make it come into play only when no
other expectations have been declared with the `AnyTimes` indicator. Notice that in this case, it's a fully fledged double:
both a mock and a stub. A mock as we set up an expectation regarding the output parameter and a stub forcing a
concrete return value. Which in this case, represents a space with no restrictions to be removed.

Similar changes are needed in the rest of the givens.

```go
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
```

Each of the stubs prepares a `Space` that, starting with the basic one that would be used for the happy path,
decorates it in a way to make the particular scenario fail. This could be simplified using the
[builder or mother patterns][builder-mother], but let's leave that for another chapter.
It's worth mentioning that more complex scenarios might end up preparing more than one double.
But the same behaviour can be achieved by using the same patterns and principles.

Our tests are now failing, as our production code is still using the old `fetchSpaceWithID` method.
Let's get rid of that outdated method and use our store instead:

```go
space := store.FetchSpaceWithID(id)
```

And with that, we reach our destination: production code using injected dependencies and tests guiding the execution
through doubles ✌️✌️ And once again, without modifying any of our scenarios! 💥
We can now see why that is. If our business logic doesn't change, our scenarios don't need to change either.
You might have heard that before: tests only need to be modified when business rules change.
Or a similar one: refactoring code is not supposed to modify our tests.
But it was always hard to see that in practice. Not anymore, as we made it easy to distinguish what the scenario is
and what our internals are. We'll see more on this when we discuss clean architectures 😉

The most avid gophers might have been asking why would we return a `*Space` instead of `(Space, error)` when fetching.
The answer is simple: to keep things consistent with other languages that only allow a single return parameter.
However, as this is not the ideal way to handle this situations in `go`, here's a small challenge for you:
refactor the existing code so that `FetchSpaceWithID` returns a space and an error.
And of course, tackle the tests first.

We've mentioned the builder and mother patterns above. And that's a chapter worth checking.
However, head over to the [functional options][functional-options] one first, as it will come in handy.

[tests]: ../tests
[functional-options]: ../functional-options
[builder-mother]: ../builder-mother
[hexagonal]: ../../clean-architectures/hexagonal
