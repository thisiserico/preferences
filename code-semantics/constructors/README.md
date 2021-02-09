# Named constructors

Many languages have class constructors. Some provide magic methods like `__construct`, others assume a method matching
the class name as the constructor and others, for better or worse, don't expose any known technique to deal with this.
And those are the ones I personally like the most, as they force you to think what is it that you want to denote
when creating an object. Those constructors take the assumption away that an object needs to be created in a single way,
opening the door to express semantics not only with the class or struct itself, but in the way it will be used.

If you came from the [semantic types][semantic-types] page, you've already seen some of this when, for example,
creating a `Discount`:

```go
type Discount struct {
    inCents int
    asPercentage int
}

func DiscountAsPercentage(percentage int) Discount {
    return Discount{asPercentage: percentage}
}
```

Even though we might want to apply discounts as a whole unit or as a percentage in our store, by hiding the `Discount`
internals and enforcing the use of certain exported methods, we naturally reduce the mental overhead to understand how
discounts work. And that's our first named constructor 🎯

What about discounts in cents? Well, another constructor.

```go
func DiscountInCents(cents int) Discount {
    return Discount{inCents: cents}
}
```

But is it possible to model a discount as a combination of an amount in cents and a percentage? No it's not!
How do I know? Because there's no constructor for such case, as simple as that.
The lack of a constructor, in this case, implies that such case doesn't need to be taken into account at all.
No matter your experience on this particular domain, you won't assume that such case exists.
And that's a beautiful side effect, as this team's code is letting you know what the boundaries are, reducing the
amount of questions others in the team will need to answer.

Ok, let's move on to another example: the snooze in the alarm clock ⏰
And for the sake of the example, let's keep it simple.

```go
type AlarmEntry struct {
    event string
    snoozes bool
    snoozeInterval time.Duration
}

func CreateNewAlarmEntry(event string, snoozes bool, snoozeInterval time.Duration) AlarmEntry {
    return AlarmEntry{
        event: event,
        snoozes: snoozes,
        snoozeInterval: snoozeInterval,
    }
}
```

By seeing this, you can start making assumptions on how this code will naturally evolve over time:
the more settings an alarm entry allows, the more this named constructor will grow.
That's not necessarily true, as we might start exposing methods to facilitate the set up of an entry.
But even in that case, we can agree that the snooze configuration can be simplified while reducing side effects.
Or would you, at first glance, know what to pass as `snoozeInterval` when `snoozes` is `false`? I wouldn't.

```go
type Snooze struct {
    enabled bool
    interval time.Duration
}

type AlarmEntry struct {
    event string
    snooze Snooze
}
```

Well, not much has changed, but this is a step forward not to encapsulate every single alarm entry detail directly
into the entry itself. But the main benefit is still not present. Let's provide a way to create those snoozes:

```go
func NoSnooze() Snooze {
    return Snooze{enabled: false}
}

func SnoozeInterval(interval time.Duration) Snooze {
    return Snooze{
        enabled: true,
        interval: interval,
    }
}

_ = CreateNewAlarmEntry("routine...", NoSnooze())
_ = CreateNewAlarmEntry("party time 🥳", SnoozeInterval(time.Minute))
```

There's no room for mistakes now: either I have the snooze configured or not, and in the case of having it,
I only need to define its interval. But is this one of those fancy alarm clocks that tries by all means for you not
oversleep, not letting you define snoozes larger than 5 minutes? Well, the `AlarmEntry` object doesn't need to know
about that:

```go
const maximumSnoozeIntervalInMinutes = 5

func SnoozeInterval(interval time.Duration) (Snooze, error) {
    if interval.Minutes() > maximumSnoozeIntervalInMinutes {
        return Snooze{}, errors.New("snooze interval way too large")
    }

    return Snooze{
        enabled: true,
        interval: interval,
    }
}
```

Now we're not only giving semantics to the snooze, but also validating our business rules directly in the object that
actually deals with that specific rule 💯

This technique has so many benefits when applied for the sake of clarity.
But it's a game changed when dealing with refactors! Take the typical `Pagination` example:

```go
type Pagination struct {
    page int
    itemsPerPage int
}

func LoadItemsFromPage(page, itemsPerPage int) Pagination {
    return Pagination{
        page: page,
        itemsPerPage: itemsPerPage,
    }
}
```

We've all seen this at some point.
But in some scenarios, a pagination that starts after a certain known ID might be useful.
When asked to swap for that behaviour, we can either introduce a different `Pagination` struct,
knowing lots of existing methods will have to change.
Or we can try an incremental approach where the existing pagination is not removed, the new one gets implemented
on top of that one to be tried out, and we end up removing the old one once the new strategy is successfully running.

```go
type Pagination struct {
    page int
    loadAfterID string
    itemsPerPage int
}

func LoadItemsAfterID(id string, itemsPerPage int) Pagination {
    return Pagination{
        loadAfterID: id,
        itemsPerPage: itemsPerPage,
    }
}
```

Refactoring the existing codebase is now much simpler, as the changes are minimal:
only where the pagination is defined and where it's actually evaluated. Everything in between stays the same! ✌️

There will be times when this seems totally unnecessary.
But even in those cases, it might help the next maintainer to see something other than `NewSomething`.
Give it a try!

[semantic-types]: ../types
