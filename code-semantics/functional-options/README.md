# Functional options

There are times when something needs to be configured: a library, a server, an object... But we might need different
configuration for each of use of the same concept. The reasons for that vary, but they range from cases where different
configuration provide complete different behaviour to cases where software is being actively maintained and more options
are being added. Regardless of the case, most of them fit the same criteria: code that needs to be open to extension
while maintaining backwards compatibility 🔄

There are two very common approaches to this. The first one, to accept a configuration object where clients can
indicate the options relevant for them. They might look something like this:

```go
type FirstAPIConfig struct {
    ClientName     string
    APIKey         string
    RequestTimeout time.Duration
    DumpDebugLogs  bool
    IsDryRun       bool
}
```

This approach is easy to work with, as only one configuration attribute needs to be used, it's easily discovered and
it can be reused if the library exposing it provides different functionality.
The problem is that, as the configuration attributes are exposed, we're shooting 🔫 ourselves in the foot in the
future, as we won't be able to modify the object attributes if we need to change the behaviour of the library.
More so, when the usage of one rules out the need for another.

The second common approach is completely different, as it relies on methods to tweak the behaviour of the software
at hand:

```go
func NewSecondAPI(clientName, apiKey string) *SecondAPI {
    return &SecondAPI{
        clientName: clientName,
        aPIKey:     apiKey,
    }
}

func (api *SecondAPI) WithRequestTimeout(dur time.Duration) {
    api.requestTimeout = dur
}
```

But it's not ideal either. When using this pattern, options that are required are requested as input parameters,
and optional ones can be defined through the use of methods.
But it's still hard to know whether an option will become required at some point, or the opposite.
At the same time, when using this second approach, the resulting code makes it look like the first statement,
the new in our example, is more important than the following ones.
When in reality, all of them are just as important, only the library is not written in a way to make that obvious.

And with that we get to the functional options approach, which can be seen as an iteration on the option above.
The core concept is the same, and entitles similar problems in that regards: required attributes are explicitly
asked for. On the other hand, optional arguments can be specified making use of variadic capabilities.

```go
type Option func(*ThirdAPI)

func NewThirdAPI(clientName, apiKey string, opts ...Option) *ThirdAPI {
    api := &ThirdAPI{
        clientName: clientName,
        aPIKey:     apiKey,
    }

    for _, option := range opts {
        option(api)
    }

    return api
}
```

As you can see, the pattern relies on an `Option` type, that happens to define a method.
And an unknown number of options can be specified! What those options are supposed to do is to manipulate the main
object as any of the other options might, but they do so in a way that the code can be easily extended or modified.
At the same time, constraints regarding two conflicting attributes can be controlled.
Not only that, semantic is added to each of the configuration options like so:

```go
func WithRequestTimeout(dur time.Duration) Option {
    return func(api *ThirdAPI) {
        api.requestTimeout = dur
    }
}

func WithDebugLogsEnabled() Option {
    return func(api *ThirdAPI) {
        api.dumpDebugLogs = true
    }
}

func RunningOnDryRunMode() Option {
    return func(api *ThirdAPI) {
        api.isDryRun = true
    }
}
```

Interesting concept, isn't it? At the same time, it might not fit all the scenarios.
My recommendation is to use with caution, and avoid its usage if we're just starting to implement the first version of
a library that we don't yet know how it will evolve ⚠️

Most of the time we see this pattern, it happens to be to configure libraries and similar ideas:
servers, API libraries, pubsub mechanisms...
However, I'm here to talk about testing, and this pattern enables a really interesting usage when testing.
Check out the [builder and mother patterns][builder-mother] to see yet another use of functional options.

[builder-mother]: ../builder-mother
