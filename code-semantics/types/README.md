# Semantic types

This is one of those that leads to confusion when seen at first.
Types are becoming more and more popular, but often times I've seen how the only use we give to them is to avoid castings
and the like. I personally like to use types to hide implementation details on my internals as well, which is a not so
extended practice in `go`. But overall, if something, their key benefit is that they provide specific semantics to our
domain logic.

Assume an e-commerce and take the following example:

```go
type Product struct {
    ID string
    Name string
    Price int
    Discount int
}
```

With a quick glance, you can see how this struct would be used in different scenarios to maybe list products,
calculate prices with discounts, etc. Actually, let's use it:

```go
product := Product{
    ID: "uuid",
    Name: "rocket to the moon",
    Price: 2432,
    Discount: 10,
}
```

Easy enough 👌 But only for those who know exactly how every piece on this software works!
Does the ID needs to be of a certain type? `uuid` is understandable, but what if my business requires for those to be
something like `p:uuid`? Does the name has any constraint –like minimum or maximum size–, are any words forbidden,
does it need to be the actual public name or something else? 🤔 What are the units of the price? Certainly a rocket to
the moon would cost more than $24 and $2432... Are those millions? And why am I assuming those are dollars and not
euros? Good thing we've got a discount of $10! Or 10%?

Our product definition can for sure model an e-commerce, but chances are that we'll confuse one of those properties
sooner or later, potentially leading to a very cheap visit to the moon.

Would hiding the internal representation of a product help us in any way here? Let's find out.

```go
type Product struct {
    id string
    name string
    price int
    discount int
}

func NewProduct(id, name string, priceInCents, discountPercentage int) Product {
    return Product{
        id: id,
        name: name,
        price: priceInCents,
        discount: discountPercentage,
    }
}

product := NewProduct("uuid", "mug", 600, 100)
```

The moment we cannot directly manipulate the struct attributes from outside the package, we're forced to use a
constructor for it. The constructor signature is already giving us some valuable information regarding what each
parameter is. But we can still make mistakes with this method, get confused by it or mess up the input parameters order.
Are we giving the mug for free? Or did someone confused that `discountPercentage` for `discountInPriceUnits`?

Maybe we can see this clearly with a different example.
Did you ever had to deal with an endless data or service migration or an unfinished refactor that led to two IDs
representing the same concept? (╯°□°)╯︵ ┻━┻ Let's navigate this example starting with exported struct attributes.

```go
type User struct {
    UserID int
    AccountID int
}

user := User{
    UserID: 24,
    AccountID: 32,
}
```

On a first iteration, we could hide the internal representation for the struct, keeping the same behaviour:

```go
type User struct {
    userID int
    accountID int
}

func NewUser(userIDFromDeprecatedSystem, accountIDFromNewSystem int) User {
    return User{
        userID: userIDFromDeprecatedSystem,
        accountID: accountIDFromNewSystem,
    }
}

user := NewUser(32, 24)
```

There, done 🚀 But wait, did you spot the mistake?
Having constructors and not exposing internals is good and all, but it's just way too easy to introduce undesired
mistakes. Which leads to the necessity to use custom types, not only to avoid these mistakes, but also to give semantics
to that same code.

Using this same `User` example, a possible approach might look like this:

```go
type User struct {
    userID DeprecatedUserID
    accountID NewUserID
}

type DeprecatedUserID string

type NewUserID string

func NewUser(userID DeprecatedUserID, accountID NewUserID) User {
    return User{
        userID: userID,
        accountID: accountID,
    }
}

user := NewUser(DeprecatedUserID(24), NewUserID(32))
```

Yes, at first glance it looks less straightforward. But pay attention to the last line of the example: as a newcomer
to a team, would you be confused by it? Probably not. Even if you're not completely familiar with ongoing migrations,
you'll know how to use the `User` struct.

This same concept becomes extremely useful when exporting interfaces. Let's see that with the typical `Store`
(aka `Repository`) interface:

```go
type Store interface {
    FindProducts(brand, category string, page, limit int) ([]Product, error)
}
```

Imagine the disaster when confusing the order of the attributes, specially when you add yet a new criterion and need
to modify callers of this method. Or when you decide to change the pagination strategy of the whole site. How would
you feel implementing those when having an interface like so? We can evaluate a more typed approach:

```go
type Store interface {
    FindProducts(Brand, Category, Pagination) ([]Product, error)
}
```

For sure, more types also means longer files. But in the long run, the business benefits 💰
Not only you're using common language no matter who you're talking to, but also human mistakes are automatically
reduced and future refactors become much more simple to handle. This directly translates in time spent working on
features or paying debt™️.

But what is this common language or semantics I'm referring to? Let's come back to the `Product` struct to illustrate
that clearly. I took the liberty of modifying it a bit, following the same principles we saw in the `User` struct:

```go
type ID string

type Name string

type Price struct{
    cents int
}

type Discount int

type Product struct {
    id ID
    name Name
    price Price
    discount Discount
}

func NewProduct(i ID, n Name, p Price, d Discount) Product {
    return Product{
        id: i,
        name: n,
        price: p,
        discount: d,
    }
}

product := NewProduct(
    ID("uuid"),
    Name("mug"),
    Price{cents: 600},
    Discount(100),
)
_ = product.FinalPrice()
```

We might have reduced one of the issues, but this code is still not talking to us.
For this particular scenario, would it make sense to consider both price and discount tied together? The internals of
the struct are already hidden, nothing prevent us from modelling such thing:

```go
type Price struct {
    cents int,
    discount int
}

func PriceWithDiscount(basePriceIncents, discount int) Price {
    return Price {
        cents: basePriceIncents,
        discount: discount,
    }
}

product := NewProduct(
    ID("uuid"),
    Name("mug"),
    PriceWithDiscount(600, 100),
)
_ = product.FinalPrice()
```

At this point, not much has changed when consuming the product's final price.
However, when creating the product itself, there's no confusion: there's a price and, by the looks of it, it might
have a discount. Nothing prevents us from having a method that creates a `Price` without a discount ✌️
But at the same time, we took a step backwards as it's still not clear what `Price` or `Discount` represent.
Iterating over this one more time might take us to a better place:

```go
type Discount struct {
    inCents int
    asPercentage int
}

func DiscountAsPercentage(percentage int) Discount {
    return Discount{asPercentage: percentage}
}

type Price struct {
    cents int,
    discount Discount,
}

func PriceWithDiscount(basePrice int, discount Discount) Price {
    return Price {
        cents: basePrice,
        discount: discount,
    }
}

product := NewProduct(
    ID("uuid"),
    Name("mug"),
    PriceWithDiscount(600, DiscountAsPercentage(100)),
)
_ = product.FinalPrice()
```

The price itself could use some attention.
If we were to work on it, we'd end up applying the same principles we have up until this point.
On the other hand, there's no doubts that we're getting the mug for free 💸

At this point, your possibilities are almost endless to represent your domain.
For instance, imagine having a global `var FullPrice Discount = DiscountAsPercentage(100)` 🤷‍♂️
You get the idea. And although possibilities are endless, the limit on this needs to be agreed upon.
Does everything needs to be a type or only more complex concepts? It depends.
Do we value consistency across the codebase around this or we better evaluate each occurrence individually? It depends.
Is a `boolean` easy enough to understand or we rather have semantics around it? It depends.
Each team and business are different, but types can be a very powerful tool in the right scenarios.

Let's address the elephant 🐘 in the room. Does this have an impact on performance? Well, a bit.
Compilers are smart enough to optimize your builds and each language is a world of its own,
but overall some extra bits need to be allocated in runtime. Now, is the impact on performance that high?
Probably not. We're talking nanoseconds here! On the other hand, does it have an impact on your team?
I want to believe so, specially for people like me: the average programmer.
We're talking about hours or days in some cases...
Every context and scenario is different and each needs to be evaluated separately, but unless those nanoseconds mean a
huge difference to you, this is something worth considering.

Is there a next step on this? Absolutely! [Named constructors][named-constructors], which is something you've already
seen here, is a good step forward.

[named-constructors]: ../constructors
