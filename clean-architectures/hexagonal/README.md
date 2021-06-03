# Hexagonal architecture

You might have already seen this before with other names: the onion architecture, ports and adapters...
I'll refer to this as hexagonal ⬢
But what is this, exactly? It's one of the clean architectures. But as that doesn't mean much, let's say
that it's a layered way to organize your software with changeability in mind.
There's a hard distinction between some of the layers: domain, application and infrastructure 🧅

It normally gets represented as a hexagon, but I'll use a rectangle for no particular reason:

```
    +-------------------------------+
    |        infrastructure         |
 ---+  +-------------------------+  +---
       |    application layer    |
 ---+  |  +-------------------+  |  +---
    |  |  |  business domain  |  |  |
 ---+  |  +-------------------+  |  +---
       |                         |
 ---+  +-------------------------+  +---
    |                               |
    +--+ +---------+ +---------+ +--+
       | |         | |         | |
```

There's a hard rule in this model that we'll see with examples, but let's put an emphasis on it now:
layers are to be traversed from the outside to the inside and never the other way around.
That is, the business domain will never have a reference to an infrastructure piece.
To illustrate that with a common example: the pieces of code handling your business logic will never
have a reference to a postgres piece 🤯 Your business logic most likely doesn't need to have references
to third party services (unless your business is actually based on that, of course).

That's something that I've noticed is hard to accept many times. And following this rule might not be that good of
an idea if you have performance constraints. And with those I mean the kind of issues where you need to run
hundreds of instances of a running service, shard a database on the client side or other issues I haven't faced yet.
But most of us don't face those issues, even if your throughput is high 😉

But you're probably wondering what goes into each layer. Let's try to sum that up!

Business domain implies anything that is used to express your business in code, but not use cases.
In fact, this layer can sometimes be seen as a slice of domain services, model and "core".
I rather not use those terms due to the ambiguity they create, not to mention that different languages
handle these concepts in different ways (classes, functions, structs, methods, packages...).
At the same time, we must not fall into the trap of thinking that our business gets modelled around a database
schema. So then, what goes into this layer? The building blocks that you use to make sense of your business,
common language that you use not only with technicals but with any stakeholder, the pieces that change when
your business changes. We'll get to see these concepts in detain when we talk about domain driven design,
as they are of outmost importance there. But let's not head ahead of ourselves just yet 🏃‍♀️

But then, why not use cases? Are those not part of the business? Well, they are, but those don't model your
business but the functionality you provide. Take both netflix and youtube as examples.
I'm sure 🤓 they both have a `video` entity, which is part of the business domain.
But they also have functionality to `play something next` once the current stream has finished:
that's business functionality, and is more prone to change than the `video` itself.
In fact, they both seem to tweak this functionality often. Can you imagine modifying your building blocks
any time a feature needs to slightly change? That would be unsustainable. On the other hand, writing a new
use case and being able to switch between them using a feature toggle? Great success! 👍👍

And lastly, the infrastructure layer.
To summarize a lot, any piece you interact with and you don't own goes into this layer.
Sometimes even things you own but are not specific to this domain in particular! In general, any IO.
Think about specific code to handle database calls, framework dependant code, queues interaction, http handlers,
3rd party services interaction... You name it.
There will always be exceptions, of course. But the bulk of this kind of code lives as infrastructure pieces.

Now, going back to the diagram above... If there's a dependency rule from the outside to the inside,
how can my code talk to pieces of infrastructure? 🤔
Here's where the dependency inversion principle comes into play. Or to put it lightly,
the pieces you own couple to a contract that also lives in your domain, never to concrete implementations.
While specific implementations are injected to that domain, in a way that always leaves the domain in control of those pieces.

And what's the benefit of doing all this, exactly? 🧐 Changeability! 💰
By putting your domain first, and letting the rest be collaborators, you ensure that you can safely tweak your dependencies
without your domain having to receive updates.

Ok, enough talk, show me the code. This time, we'll assume we run a video streaming platform 🎬🎥
Specifically, we'll focus on a single feature we've already mentioned: `play next`.
And because we really care about distinguishing different aspects of our business, and the audience of those,
we'll put this specific feature in the `player` context. If we were to implement more functionality,
like `show recommendations` or `rate video`, those would probably live in the same context.
On the other hand, features like `confirm age` or `set viewing options`, even thought they might use
–and this is important– _similar_ `video` entities, they are not the same building blocks!
Using the same piece of code for everything leads to a really fat model that is extremely hard to maintain,
offers no clear distinction about functionality and prevents for it to express semantics.
From top of my head, I could envision having some `legal` and `editor` contexts.
But of course, this is not as easy as that and you really need to spend a while thinking about how to split
these in the right way.

Sadly, there's no silver bullet on how to structure the scaffolding for an hexagonal architecture.
I'll follow an approach that maps one to one with the layers we've mentioned, mainly for simplicity.
I'd like to mention, however, that with time I've learned not to depend on this specific split
and prefer another one slightly less constraining. This makes a lot of sense specially on a `go` codebase,
but it might just as well not work well with other languages.

We'll approach the problem in an outside-in fashion: starting by the use case itself.
This is where we are so far, in terms of structure:

```
clean-architectures/hexagonal/example
├── editor
├── legal
└── player
    ├── application
    │   └── play_next.go
    ├── domain
    │   └── video.go
    └── infra
```

And the basic code to make that work:

```go
// application/play_next.go

func PlayNext(currentlyPlaying string) (domain.VideoID, error) {
        return domain.NoVideo, nil
}

// domain/video.go

var NoVideo = VideoID("")

type VideoID string
```

Let's stop here for a second and talk a bit about that `currentlyPlaying` string.
Chances are that you might be thinking that passing a string in is not safe enough: any random piece of string
could make it there, as it's not clear what that parameter expects: is it an ID? Maybe an URL?
And I'd agree with you 👌 In fact, I used to pass along actual types –as in `domain.VideoID`– to make it clear.
But after a long time doing so, I realized that I was shooting myself in the foot:
every time I wanted to refactor anything that belonged to my domain, I had to go to pieces of infrastructure
(remember those talk to the application layer) and modify those as well.
My domain was not closed to modification! Now that I pass in native types,
I can literally refactor all the domain internally that nothing changes on the outermost layer:
changes are limited to the context I'm working on.
There's a second benefit when not using custom types there, and that is domain validations.

```go
// application/play_next.go

func PlayNext(currentlyPlaying string) (domain.VideoID, error) {
        videoID, err := domain.VideoIDFrom(currentlyPlaying)
        if err != nil {
                return videoID, err
        }

        return domain.NoVideo, nil
}

// domain/video.go

func VideoIDFrom(raw string) (VideoID, error) {
        if raw == "" {
                return NoVideo, errors.New("invalid video ID")
        }

        return VideoID(raw), nil
}
```

Anything that comes from the outside world needs to be validated 👮‍♀️
If we were passing a `VideoID` in, the validation should happen even before calling the `PlayNext` method.
And by doing so, we'd be writing business logic outside of our context!
To make this obvious, imagine a case where an "invalid" video ID is a valid option. What would make more sense:
for the use case to know, test and handle the situation, or spreading the business logic between the use case
and an http handler? 🤷‍♂️

For those two reasons, I choose to pass native types along, and let the use cases be the ones converting those
into building blocks.

Ok, next up: querying for some data! We're all busy people, I don't have time to spin up a database to write these.
What to do, then? The answer is somewhere above ;)

```go
// domain/store.go

type Store interface {
        IsPlayNextEnabled() bool
        NextAfter(VideoID) VideoID
}
```

You guess that right! We'll be coupling our use case to the domain only, by making use of an interface 👍
And how the million dollar question: how to inject that store into the use case in the first place?
Well, there're different alternatives. Let's evaluate the following:

```go
// application/service.go

type Service struct {
        store domain.Store
}

func NewService(store domain.Store) *Service {
        return &Service{
                store: store,
        }
}

func (s *Service) PlayNext(currentlyPlaying string) (domain.VideoID, error) {
        return domain.NoVideo, nil
}
```

This just works™️
At the same time, we know the drill: we'll end up adding any use case method to this one struct.
Once we have multiples –with potentially several dependencies and such– would you be able to distinguish,
on a quick glance, what `PlayNext` depends on? Probably not!
Or at least I haven't been able to answer that question before quickly enough.
Not only that, when adding more dependencies to our "generic" application service,
we need to start modifying several tests that make use of them.
When doing so, it feels like we're focusing on several things at once as opposed to just one: our use case.

Let's evaluate a different alternative to this:

```go
// application/play_next.go

type PlayNext func(currentlyPlaying string) (domain.VideoID, error)

func PlayNextUseCase(store domain.Store) PlayNext {
        return func(currentlyPlaying string) (domain.VideoID, error) {
                return domain.NoVideo, nil
        }
}
```

Try picturing the test for this use case in your mind for a second 🧠
You know what your use case depends upfront and you can write the thing focusing on a single piece.
Moreover, whatever piece of code using this will have a direct reference to the `PlayNext` contract.
And lastly, you won't have to worry about potentially leaking functionality when you want to expose a use case,
but not all, into a different consuming method (eg. a CLI to use on a terminal or a pubsub consumer,
even though it might not apply to this example).

And with that, we can finish the functionality.
Disclaimer for the gophers out there `ʕ◔ϖ◔ʔ`: don't worry about error handling.
I'm trying to keep the example as simple as possible 😉

```go
// application/play_next.go

func PlayNextUseCase(store domain.Store) PlayNext {
        return func(currentlyPlaying string) (domain.VideoID, error) {
                videoID, err := domain.VideoIDFrom(currentlyPlaying)
                if err != nil {
                        return videoID, err
                }

                if !store.IsPlayNextEnabled() {
                        return domain.NoVideo, nil
                }

                return store.NextAfter(videoID), nil
        }
}
```

There we go, a completed use case that clearly states the intention without getting lost into nuance.
Some of you might be wondering what happens when the use case is slightly more complex than this.
What if more domain logic needs to be used?
What to do when a use case might need information that lives in a different context?
The answer to the first question is simple: it will live in the domain layer, you just use it.
However, the answer to the second question is slightly more hairy, and is often times topic of debate:
you shouldn't have to ask any piece of data to other contexts, the specific context that needs the data
should already have it available! Take the `is play next enabled` setting. You can see how there's a chance
that the `analytics` context consumes this specific data point. That being the case means than,
whenever this setting changes, it needs to make its way into both contexts.
But isn't that introducing duplication and more lines of code? Indeed it is.
But we're not trying to optimize in terms of lines of code.
We're aiming instead for easier to change code that happens to be understandable by any person, now or in the future.
There are ways to simplify this situations (CQRS, event driven architectures, etc) but I won't get into those just yet.

And with that, the chapter on hexagonal is completed! 🥳

Wait, what!? That's it? Is this the whole thing? Indeed.
Hexagonal architecture is nothing else than layer your code, which we've done.
Putting specific pieces in the right layer (use cases as part of application,
business logic into the domain one) and forget about the intricacies of the infrastructure itself.
If you're wondering why is it that we didn't even mention the infrastructure layer, the answer is simple:
that's the layer that completely depends on each business. It wouldn't make sense to make assumptions
about what works best for your business, would it?
