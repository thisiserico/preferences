# Builder and mother patterns

When we saw the [test doubles][test-doubles] 👯, by the end of the chapter, we realized that we had to create many
spaces, when all them looked almost the same. The idea was to create spaces on the fly, each of them simulating one
of the cases we had to test, in regards to validation. Although not ideal, as there was too much repetition,
we managed with that. For today's chapter, we're gonna give that yet another iteration to make our tests that rely
on spaces more robust.

To refresh our memories, this is our starting point:

```go
type resource struct{}

type Space struct {
    id        string
    ownerID   string
    name      string
    resources []resource
}
```

Let's call this struct our "model", a unit that represents an entity in our business domain.
When testing, we'll have to instantiate models pretty often: to use them as input arguments,
to compare against a returned value, to force returns when using stubs... Creating them is rather easy:

```go
space := Space{
    id:        "an-id",
    ownerID:   "some-owner-id",
    name:      "a space name",
    resources: []resource{resource{}},
}
```

The problem comes when dealing with the same situation we've been mentioning over and over: when there's a need to
modify that code, that being because of a refactor or because our business rules have changed.
Even a small model like our `Space` can suffer from this: identifiers might change, an owner could become a principal,
a name may require a machine readable version, resources might become a collection of resources and other spaces...
You get the idea. What do we do in those situations? Is it worth to go to all the space definitions in our tests and
modify each single one of them? At first sight, that seems not only a potential solution for misuse,
but also an inefficient use our team's time 💸

_Enter, stage left_: the builder pattern. The idea for this pattern is simple: an empty builder entity is created,
a method is provided as a setter for –normally– each of the properties and a build method needs to be called last,
which will end up creating the actual model with either default and safe values or with the values that were provided.

```go
type SpaceBuilder struct {
    id        *string
    ownerID   *string
    name      *string
    resources []resource
}

func (b SpaceBuilder) WithID(id string) SpaceBuilder {
    b.id = &id
    return b
}

func (b SpaceBuilder) WithResources(list []resource) SpaceBuilder {
    b.resources = list
    return b
}

func (b SpaceBuilder) Build() Space {
    space := Space{
        id:        "some-id",
        ownerID:   "some-owner-id",
        name:      "a space name",
        resources: []resource{},
    }

    if b.id != nil {
        space.id = *b.id
    }

    if b.resources != nil {
        space.resources = b.resources
    }

    return space
}
```

This pattern enables some interesting uses, as it allows tests to tweak the properties only relevant for them.
At the same time, random but safe values can be created on the fly when calling the builder method.
On the other hand, specially when dealing with a model with several attributes in it, it becomes less than ideal to
maintain the builder.

_Enter, stage right_: the mother pattern. This patterns follows a similar approach than the builder, but scoping the
amount of possibilities by providing tailored views of the model.

```go
func BuildDefaultSpace() Space {
    return Space{
        id:        "some-id",
        ownerID:   "some-owner-id",
        name:      "default",
        resources: []resource{},
    }
}

func BuildNonEmptySpace() Space {
    return Space{
        id:        "some-id",
        ownerID:   "some-owner-id",
        name:      "a space name",
        resources: []resource{resource{}},
    }
}
```

The benefit of this one is that it's slightly easier to read when on a test. Instead of having to process each of the
attributes being set, you can just assume it works the way the [named constructor][constructors] indicates.
Modifying the actual `Space` doesn't end up in tons of test files being modified, just the one where the mother
methods exist. However, what happens when we need a default space that is not empty? Do we create yet another mother
method to accommodate the case? Do we use `Space` methods to tweak its internal representation? It's hard to decide.

_Enter swing_: the builder, the mother and [functional options][functional-options]. In order to alleviate the
drawbacks of both the builder and the mother patterns, we can use a combination of both, together with the functional
options pattern. The result would be to provide a limited amount of mother methods that accept an unknown number of
functional options to build a `Space` from.

```go
type BuilderOption func(Space) Space

func UsingID(id string) BuilderOption {
    return func(s Space) Space {
        s.id = id
        return s
    }
}

func BeingDefault() BuilderOption {
    return func(s Space) Space {
        s.name = "default"
        return s
    }
}

func NonEmpty() BuilderOption {
    return func(s Space) Space {
        s.resources = []resource{resource{}}
        return s
    }
}

func FilledWith(list []resource) BuilderOption {
    return func(s Space) Space {
        s.resources = list
        return s
    }
}

func SpaceOwnedBy(userID string, opts ...BuilderOption) Space {
    return rootSpace().tweak(opts...)
}

func rootSpace() Space {
    return Space{
        id:        "some-id",
        ownerID:   "some-owner-id",
        name:      "a space name",
        resources: []resource{resource{}},
    }
}

func (s Space) tweak(opts ...BuilderOption) Space {
    for _, option := range opts {
        s = option(s)
    }

    return s
}
```

In this example, we're gonna be interested in having spaces with a certain owner. But at the same time, other tweaks
might be necessary for certain tests. But as we don't want to end up with methods like
`SpaceOwnedByKnownUserThatIsNotEmpty` and the like, we simply accept a bunch of variadic options.
This pattern enables a really interesting use: generic configurations can be provided globally,
and tests can use them right out of the box in conjunction with other options or, even better, mother functions:

```go
var removableSpaceOptions = []BuilderOption{NonDefault(), Empty()}

func TestSomethingRegardingARemovableSpace(t *testing.T) {
    _ = SpaceOwnedBy("owner-id", removableSpaceOptions...)
}
```

It doesn't differ that much from the builder pattern, as you can still end up with a model definition that spawns
several lines. Nor from the mother pattern, as multiple constructors might be needed for the sake of clarity.
But it does expand the possibilities of how a model is created without having to maintain yet a slightly altered
copy of the original model, while giving some freedom to use the pattern that better fits each case 🤷‍♂️

[test-doubles]: ../test-doubles
[constructors]: ../constructors
[functional-options]: ../functional-options
