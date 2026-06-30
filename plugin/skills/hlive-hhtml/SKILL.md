---
name: hlive-hhtml
description: >-
  Build HLive page markup with the hhtml package
  (github.com/SamHennessy/hlive/hhtml) — typed, autocompleting Go builders for
  HTML5 tags and attributes (Div, Button, A, Input, Class, Href, ...). Use when
  writing or editing HLive page/component markup, when you see the
  `. "github.com/SamHennessy/hlive/hhtml"` dot-import, or when choosing between
  hhtml builders and the raw l.T/l.C primitives.
---

# Building markup with hhtml

`hhtml` is the **preferred way to write HLive markup**. Instead of passing tag
and attribute names as strings to `l.T`/`l.C`, you call typed Go functions —
`Div(...)`, `Button(...)`, `Class("...")`, `Href("...")` — so you get editor
autocomplete, no string typos, and embedded docs for each tag.

## Import idiom

hhtml is designed for a **dot import**, used alongside the core package:

```go
import (
    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)
```

The dot import lets you write `Div(...)` instead of `hhtml.Div(...)`. Behavior
(events, state, render control) still comes from the `l` package — hhtml is
markup only.

## Tag builders

Every HTML5 tag has a builder with the signature:

```go
func TagName(elements ...any) hlive.Adder
```

Capitalized tag name → function: `Html`, `Head`, `Body`, `Title`, `Meta`,
`Link`, `Div`, `Span`, `P`, `H1`–`H6`, `A`, `Nav`, `Header`, `Footer`, `Aside`,
`Main`, `Section`, `Article`, `Ul`/`Ol`/`Li`, `Form`, `Input`, `Button`,
`Label`, `Select`/`Option`, `Textarea`, `Table`/`Thead`/`Tbody`/`Tr`/`Td`/`Th`,
`Img`, `Script`, `Style`, and ~100 more.

```go
Div(Class("card"),
    H1("Hello"),
    P("Welcome to ", A(Href("/docs"), "the docs"), "."),
)
```

A builder returns the `hlive.Adder` interface (written `l.Adder` with the
conventional `l` alias) — the **same type** `l.T`/`l.C` return — so hhtml tags
drop straight into `page.DOM().Body().Add(...)` and mix freely with raw
`l.T`/`l.C` calls.

## Attribute functions

Every attribute has a function with the signature:

```go
func AttrName(value string) *hlive.Attribute
```

e.g. `Class`, `Id`, `Href`, `Src`, `Type`, `Value`, `Name`, `Placeholder`,
`Rel`, `Charset`, `Content`, `For`. Pass them as elements:

```go
Input(Type("text"), Name("email"), Placeholder("you@example.com"), Id("email"))
Link(Rel("stylesheet"), Href("/assets/main.css"))
```

### Collision rule — the `Attr` suffix

When a tag and an attribute share a name, the **tag keeps the clean name** and
the **attribute gets an `Attr` suffix**:

| Clean name (tag) | Attribute function |
| --- | --- |
| `Title(...)` → `<title>` element | `TitleAttr("...")` → `title="..."` |
| `Style(...)` → `<style>` element | `StyleAttr("...")` → `style="..."` |
| `Form(...)` → `<form>` element | `FormAttr("...")` → `form="..."` |

```go
Head(
    Meta(Charset("UTF-8")),
    Title("My App"),                         // the <title> element
    Link(Rel("stylesheet"), Href("/app.css")),
)

Span(TitleAttr("tooltip text"), "hover me")  // the title attribute
```

## Behavior stays on the core `l` package

hhtml does **not** export `On` or state helpers. Attach events and state with the
core package. `tagBuilder` detects an event binding in the elements and
automatically produces a Component (`l.C`) instead of a static Tag (`l.T`):

```go
count := l.Box(0)

Button(
    Class("btn"),
    Id("counter"),
    l.On("click", func(_ context.Context, _ l.Event) {
        count.Lock(func(v int) int { return v + 1 })
    }),
    "Clicks: ", count,            // an l.Box renders its value
) // → becomes a Component because it contains an l.On binding
```

Use `l.OnOnce` for one-shot bindings. See the **hlive-core** skill for the event
model, `l.Box`/`l.NewLockBox` state, and `l.Render`/`l.RenderComponent`.

### Classes: setting vs toggling

`Class("a b")` (hhtml) just **sets** the `class` attribute string. To toggle
classes on/off reactively, use the core class types:

```go
Div(
    l.ClassBool{"active": isActive, "open": isOpen}, // toggle on/off
    l.Class("base ordered"),                         // ordered, additive
)
// remove a class with the l.ClassOff string type
btn.Add(l.ClassOff("active"))
```

## Mixing with raw builders

hhtml and `l.T`/`l.C` interoperate (same `hlive.Adder`). Reach for raw `l.T`
only when the tag name is dynamic, or for a one-off where you haven't dot-imported
hhtml:

```go
Div(Class("box"),
    l.T("custom-element", l.Attrs{"data-x": "1"}), // dynamic / non-standard tag
    P("standard tag via hhtml"),
)
```

## Generated code — do not edit

`hhtml/tags.go` and `hhtml/attrs.go` are **generated** (`hhtml/cmd/tags` and
`hhtml/cmd/attrs`) and carry a `DO NOT EDIT` header. To add/change tags or
attributes, change the generators and regenerate — never hand-edit the output.

## Reference

`hhtml/tags.go`, `hhtml/attrs.go`, `hhtml/builder.go` (the `l.C`-vs-`l.T`
auto-selection). Real usage: `hlive_docs/page/demo/clicker.go` (dot import +
hhtml layout + `l.On` handlers). For structuring a whole app see the
**hlive-project** skill.
