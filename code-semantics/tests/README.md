# Semantic testing

I like to think that testing is a delicate subject 🪶
There's so many flavors to it, that it's just hard to agree on certain techniques or styles.
Here are some of the common buzzwords you might have heard regarding this in no specific order:
testing first, testing after, TDD, BDD, ATDD, outside in, inside out, no testing, unitary, integration, functional, acceptance, end to end... 🥵
And this is only a small –and probably best known– portion of those words. This is just way too much for me to handle.
But in order to continue reading this, you need to be open to the idea that testing is beneficial.
And to limit the scope of this page, our main focus will be unit testing.
However, the concepts can also be extended to other kinds of testing.

Hopefully, you've digested the concepts seen in the [guard clauses][guard-clauses] section.
If not, please head over there first to better understand the starting point of this page.

This is where we left off: a working `DeleteSpace` method that validates certain constraints and acts upon those.
If all the conditions are met, the space is finally removed.

```go
const defaultSpaceName = "default"

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
        return errors.New("cannot fetch the space")
    }

    if !space.isOwnedBy(whoAmI) {
        return errors.New("only owned spaces can be removed")
    }

    if !space.isEmpty() {
        return errors.New("only empty spaces can be removed")
    }

    if space.isTheDefault() {
        return errors.New("the default space cannot be removed")
    }

    return removeSpaceWithID(id)
}
```

Our goal is simple: to validate each possible scenario.
Keep in mind the goal here will never be to get to 100% of coverage with the tests.
But who knows, maybe that ends up being a side effect. But really, let's not focus on that.
However, we do want to cover each of those scenarios as those are business rules from our domain logic.
And that's important. If we decide to change our business rules, our tests will be modified as well 🤷‍♂️

Ok, back to the scenarios. By the looks of it, there might be six of them:
1. Everything works, the space gets removed
2. An error is returned when actually removing the space
3. The space is not owned by the requesting user
4. The space is not empty
5. The space happens to be the default one for this user
6. The space cannot be fetched

I listed the scenarios in this particular order as it's the one I tend to see more frequently: lay the scenarios out
starting by the success one and go from there. I think this comes from the assumption that testing the happy path
will cover the rest of them. But let's refresh our memory and think about the `DeleteSpace` method the first time we
saw it: we were reasoning about it in terms of what will work. But not that much in regards to what could go wrong.
And inevitably, things will go wrong. That's ok 👌

But looking at our `DeleteSpace` method now, where we first consider business constraints and later on the happy path,
should we reorder our potential scenarios to match the same criteria?

1. The space cannot be fetched
2. The space is not owned by the requesting user
3. The space is not empty
4. The space happens to be the default one for this user
5. An error is returned when actually removing the space
6. Everything works, the space gets removed

If we think of this order now, we can see that each scenario is an extension of the previous one.
To satisfy the second point, we need to assume that the space can successfully be fetched,
and that we already have a scenario covering for that. This is the first consideration to make as, by the looks of it,
we're making tests dependent of each other.
Some could say that we're even making tests dependant on the actual `DeleteSpace` method: if following this order,
what would happen if we were to evaluate the space emptiness before its ownership? Would our test suite start failing?
As always, the answer depends on how you structure those tests.
If each scenario is set up in a way that assumes the happy path as its starting point, and modifies the bare minimum
to guide a certain test into a different path, there's no reason to believe that by adding, subtracting or modifying
conditions our suite will start failing.
We'd be applying those changes in specific places, not influencing the rest of the scenarios.

Let's see that with an actual example: the space cannot be fetched.
To simplify the example, we'll assume that the following spaces exist:

```go
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
```

And here's what the first test would look like:

```go
func TestDeleteSpace(t *testing.T) {
    t.Run("errors when the space cannot be fetched", func(t *testing.T) {
        const spaceID = "unknown-space"

        err := DeleteSpace(spaceID, "")
        if err == nil {
            t.Fatal("an error is expected when the space cannot be fetched")
        }
    })
}
```

There are different things going on here. First of all, I took the liberty of preparing a generic `TestDeleteSpace`
suite to later write each of the scenarios as a subtest. Looking at the test in detail, it's easy to spot what we
actually care about: the unknown `spaceID` and the resulting `err`. As the `userID` is not relevant for this scenario
we simply send an empty string. Let's write the second subtest to validate the space owner.

```go
t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
    const spaceID = "known-space"
    const userID = "unknown-owner"

    err := DeleteSpace(spaceID, userID)
    if err == nil {
        t.Fatal("an error is expected when the space is not owned by the user removing it")
    }
})
```

The style is pretty similar to the previous one. In this case, we do care about the user removing the space.
On the other hand, the `spaceID` we request is only useful to force the `fetchSpaceWithID` to return exactly what we
need for this scenario. With the test as we have it now, can we ensure it's testing the case we have at hand? Nope 🙅‍♀️
We do check that an error is returned, but our test could be failing because of another guard clause.
Let's address that issue.
One of the ways for doing this in `go` is to provide a bunch of global errors on top of the file.
Without getting into details, for sure other languages handle this in a different way. If you're using exceptions,
throwing them as we saw in the [guard clauses][guard-clauses] chapter should be enough.

```go
var (
    errCannotFetchSpace        = errors.New("cannot fetch the space")
    errNotAnOwnedSpace         = errors.New("only owned spaces can be removed")
    errNonEmptySpace           = errors.New("only empty spaces can be removed")
    errUnremovableDefaultSpace = errors.New("the default space cannot be removed")
)
```

Make sure these are returned within the `DeleteSpace` method. We can now use them in our tests as well:

```go
...

if err != errCannotFetchSpace {
    t.Fatalf("an error is expected when the space cannot be fetched, got %v", err)
}

...

if err != errNotAnOwnedSpace {
    t.Fatalf("an error is expected when the space is not owned by the user removing it, got %#v", err)
}

...
```

Sweet, our tests are now asserting exactly the case they were meant to test.
Let's test the two remaining validation scenarios, which should look pretty similar.

```go
t.Run("errors when the space is not empty", func(t *testing.T) {
    const spaceID = "non-empty-space"
    const userID = "known-owner"

    err := DeleteSpace(spaceID, userID)
    if err != errNonEmptySpace {
        t.Fatalf("an error is expected when the space is not empty, got %#v", err)
    }
})

t.Run("errors when removing the default space", func(t *testing.T) {
    const spaceID = "default-space"
    const userID = "known-owner"

    err := DeleteSpace(spaceID, userID)
    if err != errUnremovableDefaultSpace {
        t.Fatalf("an error is expected when removing the default space, got %#v", err)
    }
})
```

Hmmm... This works, but I'm starting to feel a lot of repetition 🤔
If I were to write this as production code, would I improve it somehow?
I might! I'd probably apply some of the maintainable code principles, like avoiding unnecessary repetition!
We're writing those constants over and over, deleting a space always work in the same way and the asserts all look the
same. Let's simplify all that. There are different ways of doing so, but in this case I'll extract certain operations
into methods.

```go
func TestDeleteSpace(t *testing.T) {
    const spaceID = "known-space"
    const ownerID = "known-owner"

    assertErrorEquals := func(want, got error) {
        if want != got {
            t.Fatalf("unexpected error, want %#v, got %#v", want, got)
        }
    }

    t.Run("errors when the space cannot be fetched", func(t *testing.T) {
        const spaceID = "unknown-space"

        assertErrorEquals(errCannotFetchSpace, DeleteSpace(spaceID, ownerID))
    })

    t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
        const ownerID = "unknown-owner"

        assertErrorEquals(errNotAnOwnedSpace, DeleteSpace(spaceID, ownerID))
    })

    t.Run("errors when the space is not empty", func(t *testing.T) {
        const spaceID = "non-empty-space"

        assertErrorEquals(errNonEmptySpace, DeleteSpace(spaceID, ownerID))
    })

    t.Run("errors when removing the default space", func(t *testing.T) {
        const spaceID = "default-space"

        assertErrorEquals(errUnremovableDefaultSpace, DeleteSpace(spaceID, ownerID))
    })
}
```

Well, there's certainly less repetition now.
Besides, the constants on each subtest are letting us know what's important for the scenario.
On the other hand, we still need to process each of the lines of each subtest and translate that into natural language
inside our brains 🧠 Wait, this rings a bell; we've talked about natural language before!
And something like a path in a map with some instructions...
Exactly, in cases where there's logic to process, it's normally easier if the code tell us exactly what the intention was.

Each of these subtests are rather easy to read and understand. That's not always the case.
More often than not, tests occupy several lines to set them up and assert results.
But in –probably– all of them, we can distinguish three different segments:
- The scenario setup.
- The action execution.
- And the expectations check.

I was blown away years ago when I first heard in a talk by [@fiunchinho][fiunchinho] what I'm about to share with you:
those three points above can be seen as the acceptance criteria for a user story! Say woooot! 🤯

This is what a user story template might look like:

```txt
User story:

As a {type of user},
I want {some functionality}
so that {goal of the functionality}.


Acceptance criteria:

Given {user story scenario},
When {performing action}
Then {observable outcome}
```

Should we try to map a couple of those existing subtests into acceptance criterias?

```txt
Given a user that doesn't own a space,
When deleting the space
Then a not owned space error is be returned

Given the default space of a user,
When deleting the space
Then an unremovable default space error is be returned
```

Interesting, don't you think? If you're wondering why this looks so familiar, chances are that you've seen `gherkin` 🥒
before: an ordinary language parser normally used when writing behaviour driven development tests.
But worry not, we're not gonna use that here.

What we'll do, though, is to rewrite our tests, this time using natural language.
First, we're gonna need to define some helper methods:

```go
func TestDeleteSpace(t *testing.T) {
    const spaceID = "known-space"
    const ownerID = "known-owner"

    var scenarioSpaceID, scenarioUserID string
    var returnedError error

    testSetup := func() {
        scenarioSpaceID = spaceID
        scenarioUserID = ownerID
    }

    givenAUserThatDoesNotOwnASpace := func() {
        scenarioUserID = "unknown-owner"
    }

    givenTheDefaultSpaceOfAUser := func() {
        scenarioSpaceID = "default-space"
    }

    whenDeletingTheSpace := func() {
        returnedError = DeleteSpace(scenarioSpaceID, scenarioUserID)
    }

    thenANotAnOwnedSpaceErrorIsReturned := func() {
        if returnedError != errNotAnOwnedSpace {
            t.Fatalf("an error is expected when the space is not owned by the user removing it, got %#v", returnedError)
        }
    }

    thenAnUnremovableDefaultSpaceErrorIsReturned := func() {
        if returnedError != errUnremovableDefaultSpace {
            t.Fatalf("an error is expected when removing the default space, got %#v", returnedError)
        }
    }
}
```

And secondly, replace the previous logic with them:

```go
t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
    testSetup()

    givenAUserThatDoesNotOwnASpace()
    whenDeletingTheSpace()
    thenANotAnOwnedSpaceErrorIsReturned()
})

t.Run("errors when removing the default space", func(t *testing.T) {
    testSetup()

    givenTheDefaultSpaceOfAUser()
    whenDeletingTheSpace()
    thenAnUnremovableDefaultSpaceErrorIsReturned()
})
```

Now, this is easier to read 🧘‍♂️ There're good, not so good and bad things going on here.

Let's talk about the goods first. I don't need to understand every single line of code in the test to know what the
intention was at first glance. Navigating test scenarios is simpler now and common language from your business has been
introduced. There are also a couple side effects that I also think are priceless.
One of them is that, if you like writing tests first, you can provide a version of those methods that do nothing to get
feedback early on for those tasks that might not contain all the necessary details. The other one is that, when
requiring two versions –implementations– of the same concept using different technologies underneath, the exact same
test scenarios can be used. They would use different helper methods, but their names wouldn't change.

The not so goods. First, a setup method is being used. In more complex scenarios, some details might slip through into
that method. But the moment we take the happy path as the base scenario, that's inevitable. There's a way to avoid
falling into that, which is to provide a single `given` step for each of the pre-conditions that the scenario needs.
But in the long run, those deviate the attention of the intention by having to read way too many given conditions.
My preference is to stick to a single given, but this of course will depend on many factors and is worth bringing up
as a team to reach consensus. Secondly, more helper methods will appear. If you remember a previous version of these
tests, we were using an `assertErrorEquals` method to compare two errors, as opposed to now, where we use one for each
error check. This is avoidable by allowing to pass arguments to the helpers.
However, I feel those end up complicating the helpers and taking consistency away from this approach. When introduced,
it becomes harder to see where the limit is on the arguments being passed around. They also make your scenarios more
dependant on the helpers, possibly ending up in scenarios that send random data to those arguments as they are simply
not relevant. It's a totally valid way of doing this, but this is one of those where consistency actually helps.
Regardless, it's also a good chat to have among team members.

And lastly, the bads. Which in this case we can limit to one: subtests are sharing a global state (ノಠ益ಠ)ノ彡┻━┻
Keeping the scenarios as we have them now prevents us not only from parallelizing the tests execution,
but also lead to undesired side effects that become just way to hard to debug when tests leave modified scenarios.

Luckily for us, the solution for this problem is rather simple. And it uses principles that we've already seen in the
[semantic types][types] chapter! In that chapter, we suggested the idea that we can model behaviour within types for fun
and profit. Well, tests can follow the same principles 🤷‍♂️  Let's try to do that step by step.

```go
type testDeleteSpace struct {
    *testing.T

    spaceID string
    ownerID string

    returnedErr error
}

func (t *testDeleteSpace) givenAUserThatDoesNotOwnASpace() {
    t.ownerID = "unknown-owner"
}

func (t *testDeleteSpace) givenTheDefaultSpaceOfAUser() {
    t.spaceID = "default-space"
}

func (t *testDeleteSpace) whenDeletingTheSpace() {
    t.returnedErr = DeleteSpace(t.spaceID, t.ownerID)
}

func (t *testDeleteSpace) thenANotAnOwnedSpaceErrorIsReturned() {
    if t.returnedErr != errNotAnOwnedSpace {
        t.Fatalf("an error is expected when the space is not owned by the user removing it, got %#v", t.returnedErr)
    }
}

func (t *testDeleteSpace) thenAnUnremovableDefaultSpaceErrorIsReturned() {
    if t.returnedErr != errUnremovableDefaultSpace {
        t.Fatalf("an error is expected when removing the default space, got %#v", t.returnedErr)
    }
}
```

Alright, so far so good. There's a new type –declared in the test file– that encapsulates the scenario requirements
and the helper methods we'll need. Using composition, the `*testing.T` type is also embedded so that we can `t.Fatal`
when needed. Let's write our setup method, which will do as little as possible.

```go
func setupTestDeleteSpace(t *testing.T) *testDeleteSpace {
    return &testDeleteSpace{
        T: t,

        spaceID: "known-space",
        ownerID: "known-owner",
    }
}
```

Nothing new there. There's no need for constants anymore, as this is the entry point for all the tests.
But it's pretty much what we already had. Lastly, let's make use of these new methods.

```go
func TestDeleteSpace(t *testing.T) {
    t.Run("errors when the space is not owned by the user removing it", func(t *testing.T) {
        test := setupTestDeleteSpace(t)

        test.givenAUserThatDoesNotOwnASpace()
        test.whenDeletingTheSpace()
        test.thenANotAnOwnedSpaceErrorIsReturned()
    })

    t.Run("errors when removing the default space", func(t *testing.T) {
        test := setupTestDeleteSpace(t)

        test.givenTheDefaultSpaceOfAUser()
        test.whenDeletingTheSpace()
        test.thenAnUnremovableDefaultSpaceErrorIsReturned()
    })
}
```

Look at that, no more global state! Let's put the table back ┬─┬ノ( º _ ºノ)
Remember that this is what you'll read on top of your testing file. And notice the consequences: if you need to
change the way your actual production code works, but the domain logic remains the same, only the helper methods
will be modified. The actual scenario definitions, the ones that tell you what needs to happen, are left untouched.

We used to have a couple more scenarios before these changes. Let's put them back using this style.
Here're the helpers:

```go
func (t *testDeleteSpace) givenASpaceThatCannotBeFetched() {
    t.spaceID = "unknown-space"
}

func (t *testDeleteSpace) givenASpaceThatIsNotEmpty() {
    t.spaceID = "non-empty-space"
}

func (t *testDeleteSpace) thenACannotFetchSpaceErrorIsReturned() {
    if t.returnedErr != errCannotFetchSpace {
        t.Fatalf("an error is expected when a space cannot be fetched, got %#v", t.returnedErr)
    }
}

func (t *testDeleteSpace) thenANonEmptySpaceErrorIsReturned() {
    if t.returnedErr != errNonEmptySpace {
        t.Fatalf("an error is expected when the space is not empty, got %#v", t.returnedErr)
    }
}
```

And the actual scenarios:

```go
t.Run("errors when the space cannot be fetched", func(t *testing.T) {
    test := setupTestDeleteSpace(t)

    test.givenASpaceThatCannotBeFetched()
    test.whenDeletingTheSpace()
    test.thenACannotFetchSpaceErrorIsReturned()
})

t.Run("errors when the space is not empty", func(t *testing.T) {
    test := setupTestDeleteSpace(t)

    test.givenASpaceThatIsNotEmpty()
    test.whenDeletingTheSpace()
    test.thenANonEmptySpaceErrorIsReturned()
})
```

What a ride! It's been an interesting journey that took us from some lines of code to... Well, this.
And this I actually want to address, as it's often material for debate: is it really worth to extract all these
behaviour into methods? In my opinion, it is. But it's just that, an opinion. The same can be said from production
code methods. If everything can fit in a slightly large method, why extract bits and pieces into others?
It's not just semantics, it's also reusability. For these tests, we're not really reusing anything other than the
`whenDeletingTheSpace` method, so it's mainly semantics. It would be a different story if our methods were accepting
input parameters, as reusability would also come into play. But even without that, it's hard not to agree that this
would simplify the understanding of a system for a person not so familiar with it.
It's at this point when you can say that the tests document the domain logic of the business.

But wait, we're not done yet! We've been playing around with four scenarios, but we initially identified six: we're
missing the ones that end up executing the `removeSpaceWithID` method.
Same as we did before, let's hardcode the spaces what will succeed and the ones that won't.

```go
func removeSpaceWithID(id string) error {
    if id == "fails-on-remove" {
        return errSpaceCouldNotBeRemoved
    }

    return nil
}
```

Now that our tests follow a certain structure, how would you go about adding new tests?
Simple, we'll write the scenario first and don't care about the actual helpers.

```go
func (t *testDeleteSpace) givenASpaceThatCannotBeRemoved() {
}

func (t *testDeleteSpace) thenASpaceCouldNotBeRemovedErrorIsReturned() {
}

func (t *testDeleteSpace) thenNoErrorsAreReturned() {
}

func TestDeleteSpace(t *testing.T) {
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
```

The tests pass, but they are not doing anything. However, we can now see the overall picture of their scenarios,
reason about them, ask for feedback, change the keyboard to your pairing partner, change the driver in a mob session,
go for a walk and come back to the save point of your game. When I see myself in situations like those, I simply
force one of the steps to fail and know exactly where I left off.

The remaining work is now easy to complete. And by the way, if you feel more comfortable having the
`givenASpaceThatCanBeRemoved` step, feel free to add it.

```go
func (t *testDeleteSpace) givenASpaceThatCannotBeRemoved() {
    t.spaceID = "fails-on-remove"
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
```

And now we're done 💯
This ended up being a rather simple and boring test. It's completely expected: the code itself didn't have a lot of
complexity, mainly validations. If we were to have a piece of code with more potential execution flows, the tests
would have ended up being more complex as well. Therefore we might have had even more helpers, or at least the same
amount but slightly longer. But even in those cases, the same principles apply! I'm sure you also had to deal with one
of those tests that you just can't understand. Maybe if they had been expressed in a more semantical way, the effort
wouldn't have been that high. Just remember, this technique can be extrapolated to lots of scenarios, testing
types and languages.

You might have noticed that our production code has no dependencies. Keeping it that way actually helped iterate the
code and tests throughout this explanation. But I do want to briefly tackle what changes our code and tests would
require to fit test doubles in 👯 Check it out in the [test doubles][test-doubles] chapter.

[fiunchinho]: https://github.com/fiunchinho
[guard-clauses]: ../guard-clauses
[types]: ../types
[test-doubles]: ../test-doubles
