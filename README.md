# gseq

A small, coherent generics toolkit for Go.

`gseq` brings together four focused building blocks:

- `option`: explicit optional values
- `result`: explicit success/error values
- `slice`: chainable slice helpers with parallel and iterator support
- `dict`: chainable map helpers

It is built for teams that want a tighter, more opinionated alternative to mixing `lo`, `mo`, ad-hoc helpers, and raw `iter.Seq` utilities across a codebase.

Dependency policy:

- no external runtime dependencies
- standard library only
- small enough to vendor, audit, and keep around for a long time

## Why gseq

Go already gives you good primitives, and libraries like `lo` and `mo` cover a lot of ground.

`gseq` exists for a different tradeoff:

- smaller API surface
- consistent naming and behavior across packages
- first-class support for both eager collections and lazy iterators
- explicit `Option` and `Result` types without pulling in a large FP toolbox
- built-in parallel helpers for the cases where they are actually useful
- zero third-party dependencies

If you want a broad utility library, use `lo`.
If you want a compact, cohesive foundation for collection work and explicit value handling, use `gseq`.

## Packages

```text
option  <-  slice  <-  dict
   ^          ^
result  -------
```

Import only what you need:

```go
import (
    "github.com/stefanbethge/gseq/dict"
    "github.com/stefanbethge/gseq/option"
    "github.com/stefanbethge/gseq/result"
    "github.com/stefanbethge/gseq/slice"
)
```

## Install

```bash
go get github.com/stefanbethge/gseq@latest
```

Requires Go 1.23+.

`gseq` has no third-party dependencies.

## Quick example

```go
package main

import (
    "fmt"
    "strconv"

    "github.com/stefanbethge/gseq/option"
    "github.com/stefanbethge/gseq/result"
    "github.com/stefanbethge/gseq/slice"
)

func main() {
    raw := slice.Slice[string]{"12", "x", "7", "21"}

    nums := slice.TryMap(raw, func(s string) option.Option[int] {
        return result.FromGoError(strconv.Atoi(s)).ToOption()
    })

    evenSquares := slice.Collect(
        slice.MapIter(
            slice.FilterIter(nums.Iter(), func(v int) bool { return v%2 == 0 }),
            func(v int) int { return v * v },
        ),
    )

    fmt.Println(evenSquares)
}
```

## Design principles

### Small by default

`gseq` tries to cover the common 80% cleanly. It does not aim to be the biggest utility library in Go.

### Zero-dependency core

`gseq` is intentionally standard-library-only.

That matters for teams that care about:

- dependency review and supply-chain risk
- long-term maintenance
- fast builds and simple upgrades
- avoiding utility packages that drag in more utility packages

### Explicit values over conventions

Use `Option[T]` instead of hidden sentinel values or ad-hoc `(T, bool)` plumbing when clarity matters.
Use `Result[T, E]` when you want values and failures to compose directly.

### Eager and lazy pipelines

Most collection code is easier to read with slices.
Some hot paths benefit from iterator pipelines with fewer intermediate allocations.
`gseq` supports both styles without splitting your mental model across multiple libraries.

### Parallelism where it pays off

Parallel slice operations are included, but they are not the default for everything.
Ordered output is preserved where expected, and small inputs stay sequential to avoid overhead.

## Package overview

### option

Use `option.Option[T]` for optional values.

```go
user := findUser(id)

name := option.Map(user, func(u User) string { return u.Name }).
    UnwrapOr("guest")
```

Good fit for:

- lookups that may not return a value
- staged transformations without repeated `if ok`
- APIs where `None` is a valid and expected outcome

### result

Use `result.Result[T, E]` for explicit success/failure flows.

```go
parsed := result.FromGoError(strconv.Atoi(input))

doubled := result.Map(parsed, func(n int) int { return n * 2 })
```

Good fit for:

- fallible parsing and validation
- batch processing with typed error context
- composing operations without losing the error path

### slice

Use `slice.Slice[T]` when you want fluent collection operations without giving up concrete slices.

```go
top := users.
    Filter(func(u User) bool { return u.Active }).
    SortBy(func(a, b User) bool { return a.Score > b.Score }).
    Limit(5)
```

Highlights:

- filtering, mapping, grouping, reducing
- set-like operations
- numeric helpers
- parallel transforms
- `iter.Seq` interop and lazy iterator utilities

### dict

Use `dict.Map[K, V]` for small, chainable map transformations.

```go
scores := dict.Map[string, int]{"alice": 10, "bob": 7}

curved := dict.MapValues(scores, func(_ string, v int) int { return v + 2 })
```

## Iterators

Go 1.23 introduced `iter.Seq`.
`gseq` treats that as a first-class API, not an afterthought.

```go
out := slice.Collect(
    slice.TakeIter(
        slice.MapIter(
            slice.FilterIter(items.Iter(), keep),
            transform,
        ),
        100,
    ),
)
```

Use this style when you want:

- fewer intermediate allocations
- lazy composition
- clear boundaries between data sources and collectors

## Parallel operations

`slice.MapParallel`, `slice.FilterParallel`, and related helpers are intended for:

- CPU-heavy per-item work
- independent I/O-bound work where controlled concurrency helps

They are not a blanket replacement for the sequential versions.
For small slices, `gseq` deliberately stays sequential to avoid goroutine overhead.

## When to use gseq

Use it when:

- you want a small shared utility layer for Go codebases
- you like explicit `Option` and `Result` types
- you want one consistent style for slices, maps, and iterators
- you prefer a focused library over a huge helper catalog

Do not use it when:

- your team strongly prefers plain Go conventions everywhere
- `slices`, `maps`, and a few local helpers already cover your needs
- you want the widest available utility surface and ecosystem familiarity

## Relationship to other libraries

`gseq` is not trying to replace all of:

- `github.com/samber/lo`
- `github.com/samber/mo`
- the standard library `slices`, `maps`, and `iter`

It is a tighter alternative for projects that want fewer moving parts and a more uniform API.

One practical difference is dependency footprint: `gseq` keeps the core at zero third-party dependencies instead of building on a wider helper ecosystem.

## Stability

`gseq` is versioned and intended for production use.

That said, the core promise is not "maximum feature count".
The goal is a stable, compact foundation that stays readable as a dependency over time.

## License

MIT
