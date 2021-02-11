# Guard clauses

Heads up! There's absolutely nothing new in this doc 🙅‍♀️
What you'll see here is something that you're already doing, specially when validating "things".
But if you've read some other guide in this repo before, you'd have noticed the emphasis I put around making things
easy to understand while providing common language not just for developers, but for any stakeholder.

We need to make some assumptions for the example we'll use in this scenario.
Assume a feature that allows users to group resources in spaces (think of them as folders).
Thus each user will have an unknown number of spaces. By default, a space named `default` always need to exist.
We want to implement a feature to delete spaces. It might look similar to this:

```go
func DeleteSpace(id string, whoAmI string) error {
    space := fetchSpaceWithID(id)

    if space != nil && space.ownerID == whoAmI {
        if len(space.resources) == 0 {
            if space.name != "default" {
                return removeSpaceWithID(id)
            } else {
                return errors.New("the default space cannot be removed")
            }
        } else {
            return errors.New("only empty spaces can be removed")
        }
    } else {
        return errors.New("only owned spaces can be removed")
    }
}
```

The code above is cumbersome for the purpose of illustrating how guard clauses fit in.
However, this kind of code is rather frequent when dealing with operations that require several validations.
It's easy to spot by the amount of indentation levels: 4 in this case.

In `go` specifically, you handle situations like this by using nullable objects
(like is the case of the `space` variable) and returning the `error` type.
In other languages, you'd throw exceptions instead. No matter what you use, the same principles apply.

Let's tackle the first condition, which actually contains a mistake.
Notice how we return the same error (`only owned spaces can be removed`) no matter whether the space cannot be fetched
or I'm not the owner of such space.

```go
space := fetchSpaceWithID(id)
if space == nil {
    return errors.New("cannot fetch the space")
}

if space.ownerID != whoAmI {
    return errors.New("only owned spaces can be removed")
}
```

A guard clause is no other than running a validation upfront to later continue evaluating the success scenario.
By doing so, we normally end up inverting the conditions and reducing one level of indentation on each of them.
At this point, the main condition has become easier to reason about:
there's one less element in our mental stack to worry about 🥳

Now that we know how to write guard clauses, let's do the same for the rest of the method.

```go
...

if len(space.resources) > 0 {
    return errors.New("only empty spaces can be removed")
}

if space.name == "default" {
    return errors.New("the default space cannot be removed")
}

return removeSpaceWithID(id)
```

The difference regarding indentation is incredible. This code is much more natural to read and reason about.
And our brain 🧠 is thanking us for having to process less information at once.
But there's a huge difference that might be overlooked at first:
errors or exceptions are provided right after checking conditions, and not several steps later.
Come back to the initial code for a second and check the first condition and its associated error (within its `else`).
If this were to be a real world example, with much more code in between, would it make sense to return that error
so many lines later? Definitely not. Even matching ifs with their respective elses would be tough!

Is there a way to improve these guard clauses? Absolutely!
At this point the whole method is simplified in terms of logic, but it's still not using common language.
Assuming you've already checked the [semantic types][semantic-types] page out, you might already be seeing how we're
accessing attributes bounded to the space. By doing so, we're shooting ourselves in the foot by complicating
future refactors that have to deal with those attributes.
Besides, some of the checks we've seen might already be required in other operations.
Wouldn't it be great to have them isolated?

```go
type resource struct{}

type Space struct {
    id string
    ownerID string
    name string
    resources []resource
}
```

First things first: this is what our `Space` struct actually looks like. We don't really care what `resource` contains.
Our first condition required us to own a space in order to remove it:

```go
func (s Space) isOwnedBy(userID string) bool {
    return s.ownerID == userID
}

func DeleteSpace(id string, whoAmI string) error {
    space := fetchSpaceWithID(id)
    if space == nil {
        return errors.New("cannot fetch the space")
    }

    if !space.isOwnedBy(whoAmI) {
        return errors.New("only owned spaces can be removed")
    }

    ...
}
```

Much better! Not only we don't directly depend on `ownerID` anymore, but we've provided semantics to the condition.

Think of it this way: if I give you a map 🗺️ and ask you to go from point A to point B, you'll find your way there.
It will take you more or less time depending on how well you orient yourself and navigate with a map,
but you'll get there. On the other hand, if I give you that same map with the actual path from A to B highlighted on it,
you'll have one less thing to worry about. If I give you not only the map and the highlighted path, but also clear
instructions on how to navigate it at all times, there's absolutely nothing for you to worry about.
And you can share the map and instructions with others, no matter their orientation skills,
that you all will navigate the path in the exact same way 📍
It's providing common language and understanding for you all.
This is exactly what happens when replacing a condition for a semantic method.
It's forcing everyone to be on the same page when reading a line of code.
It's taking away the need to read the actual `isOwnedBy` method.
It's opening the door for anyone to modify that actual method in any way required in the future minimizing the impact.
Long story short, it's making the `Space` object more maintainable.

Let's see how the rest of the conditions would look like after encapsulating their logic:

```go
const defaultSpaceName = "default"

func (s Space) isEmpty() bool {
    return len(s.resources) == 0
}

func (s Space) isTheDefault() bool {
    return s.name == defaultSpaceName
}

func DeleteSpace(id string, whoAmI string) error {
    ...

    if !space.isEmpty() {
        return errors.New("only empty spaces can be removed")
    }

    if space.isTheDefault() {
        return errors.New("the default space cannot be removed")
    }

    return removeSpaceWithID(id)
}
```

Nice 🚀 All the details are hidden now, we're only left with actual domain logic of the business!
At the same time, complex conditions that have to deal with opposites (`!=`) and comparisons (`<` and `>`)
are no longer a problem. I don't know about you, but I always struggle with those.
There's one more benefit when tackling "unhappy paths" upfront: it will be easier to reason about tests.
Don't forget to check out the ["semantic testing"][tests] chapter to see how ✌️

For this particular example, several validations were involved. In other scenarios, there might be just one.
But even if it's just for one, extracting a condition into a guard clause can be beneficial.
Always put yourself in the shoes of a person less familiar with the domain you're working with.
Make the code easier to understand for them, not faster to parse by a compiler.
The times where code was meant to be optimized for a machine are gone.
Nowadays, high level languages are meant for humans; assembler language is for machines.

I wanted to tackle one last –and maybe less important– point regarding the style of guard clauses.
On the examples above, we've been using booleans to act upon in conditions.
But we still evaluate a condition. Others prefer to have assert-like methods that throw exceptions.
Here's how those could look like in some languages:

```go
// There's no such thing as throw nor exceptions in go, but let's assume there are.

const defaultSpaceName = "default"

func (s Space) assertIsOwnedBy(userID) {
    if s.ownerID == userID {
        return
    }

    throw SpaceNotOwnedByUser{}
}

func (s Space) assertIsEmpty() {
    if len(s.resources) == 0 {
        return
    }

    throw SpaceNotEmpty{}
}

func (s Space) assertIsNotDefault() {
    if s.name != defaultSpaceName {
        return
    }

    throw SpaceIsDefault{}
}

func DeleteSpace(id string, whoAmI string) error {
    space := fetchSpaceWithID(id)
    if space == nil {
        throw NonExistingSpace{}
    }

    space.assertIsOwnedBy(whoAmI)
    space.assertIsEmpty()
    space.assertIsNotDefault()

    return removeSpaceWithID(id)
}
```

This is a perfectly valid way to write guard clauses, and in fact lots of people prefer this way of doing so.
In the same way, we could port this style to `go` using `error` instead of `bool`:

```go
func (s Space) assertIsOwnedBy(userID) error {
    if s.ownerID == userID {
        return nil
    }

    return errors.New("space not owned by user")
}

func DeleteSpace(id string, whoAmI string) error {
    ...

    if err := space.assertIsOwnedBy(whoAmI); err != nil {
        return err
    }

    ...
}
```

However, this approach might not fit all cases.
In the delete operation, not owning a space is bad enough as to stop the operation.
On the other hand, there might be operations that simply take a different route in the case of not owning a space.
The execution might not necessarily need to be stopped, but handled in a different way.
At the same time, the error or exception to provide might be different depending on such operation.
Because of that, I personally feel more comfortable modelling guard clauses in the way you've seen from the beginning,
but as always, each context and scenario needs to be evaluated individually.
There's no right or wrong here, just different styles and preferences.
As long as everything is agreed upfront, the benefits still apply.

[semantic-types]: ../types
[tests]: ../tests
