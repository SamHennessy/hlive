---
name: hlive-core
description: >-
  Build UIs with HLive, the server-side virtual-DOM framework for Go
  (github.com/SamHennessy/hlive). Use when working with HLive code — the
  `l "github.com/SamHennessy/hlive"` import, builders like l.T / l.C / l.CM,
  NewPage / NewPageServer, On(...) event bindings, l.Box reactive values, or
  questions about HLive components, rendering, the SSR-vs-WebSocket lifecycle,
  and tree-diff gotchas.
---

# HLive Core

HLive is a server-side virtual-DOM framework for Go — think Phoenix LiveView for
Go. You build the UI as a tree of Go values; HLive renders the initial HTML, then
keeps a WebSocket open and pushes DOM diffs whenever your state changes. All logic
(DB, business rules, API calls) stays in Go — you write no JavaScript.

The conventional import alias is `l`:

```go
import l "github.com/SamHennessy/hlive"
```

## Builders

For everyday page markup, prefer the **hhtml** builders (`Div(...)`,
`Button(...)`, `Class("...")`) — they're typed and autocompleting. See the
**hlive-hhtml** skill. The functions below are the underlying primitives that
hhtml wraps; use them directly for dynamic tag names, or when you haven't
imported hhtml.

| Func | Returns | Use for |
| --- | --- | --- |
| `l.T(name, ...elements)` | `*Tag` | A **static** HTML tag (no events). e.g. `l.T("div", ...)`. Void tags like `hr`/`img` have no children. |
| `l.C(name, ...elements)` | `*Component` | An **event-aware** tag — anything the user interacts with (button, input). |
| `l.W(tag, ...elements)` | `*Component` | Wrap an existing `*Tag` as a Component to attach events. |
| `l.CM(name, ...elements)` | `*ComponentMountable` | A Component with **lifecycle hooks** (Mount/Unmount). Use when the component needs to fetch data when it appears. |
| `l.G(...nodes)` | `*NodeGroup` | Group sibling nodes without a wrapper element. |
| `l.E(...elements)` | `*ElementGroup` | Group a mix of nodes + attributes + bindings. |

With hhtml you rarely choose `T` vs `C` by hand — a builder like `Button(...)`
becomes a Component automatically when it contains an event binding. With the raw
primitives, pick `T` until the element needs to respond to a browser event, then
`C`; reach for `CM` only when you need Mount/Unmount.

## Events

There are exactly two event-binding constructors: **`l.On`** and **`l.OnOnce`**.
There are no per-event helpers like `OnClick`/`OnKeyUp` — pass the DOM event name
as a string.

```go
Button(
    l.On("click", func(ctx context.Context, e l.Event) {
        // handlers ARE where you fetch from a DB / API — they run on the server
    }),
    "Save",
)

Input(
    Type("text"),
    l.On("keyup", func(ctx context.Context, e l.Event) {
        message = e.Value
    }),
)
```

(`Button`/`Input`/`Type` are hhtml builders; the binding makes them Components
automatically. The equivalent with raw primitives is `l.C("button", l.On(...))`.)

`l.OnOnce("focus", ...)` binds a handler that fires once then removes itself.

The `l.Event` struct carries (only the fields relevant to the event type are set):

- `Value string` / `Values []string` — input value(s)
- `Selected bool` — checkbox/radio checked, or select option selected
- `Key`, `CharCode`, `KeyCode`, `ShiftKey`, `AltKey`, `CtrlKey` — keyboard events
- `File *l.File` — file inputs/uploads
- `IsInitial bool` — true when a browser re-sent pre-filled field data after a
  page reload (see *Initial sync* below)

## Reactive state with Box

`l.Box(v)` wraps a value in a thread-safe container that is also a renderable
node — drop it into the tree and it renders its current value. Read/write it
safely with `Get`, `Set`, and `Lock`:

```go
count := l.Box(0)

Button(
    l.On("click", func(_ context.Context, _ l.Event) {
        count.Lock(func(v int) int { return v + 1 }) // read+write under one lock
    }),
    count, // renders the current value
)
```

For a plain variable you can also pass it **by reference** so renders pick up the
latest value: `P("Hello, ", &message)`. Box is preferred when more than one
goroutine (e.g. PubSub) can touch the value.

## Attributes and styling

```go
l.Attrs{"type": "text", "placeholder": "name"}          // plain attributes
l.Class("btn btn-primary")                              // ordered class list
l.ClassBool{"active": isActive, "disabled": !ok}        // toggle classes on/off
l.ClassList{"a", "b", "c"}                              // add a slice of classes
l.Style{"color": "red", "display": nil}                 // nil removes a rule
```

Ordering note: within a single `ClassBool`/`Style` the order is **not** preserved;
add separate elements when order matters. Re-adding the same class key updates the
existing one.

With hhtml, set plain attributes with the typed funcs (`Type("text")`,
`Placeholder("name")`, `Class("btn")`) instead of `l.Attrs`; keep the core
`l.ClassBool`/`l.ClassList`/`l.ClassOff`/`l.Style` types for **toggling** classes
and styles on/off.

## Page and server

```go
import (
    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func home() *l.Page {
    page := l.NewPage()
    page.DOM().Title().Add("My App")
    page.DOM().Head().Add(Link(Rel("stylesheet"), Href("/assets/main.css")))
    page.DOM().Body().Add( /* your tree, built with hhtml */ )
    return page
}

func main() {
    http.Handle("/", l.NewPageServer(home)) // PageServer is an http.Handler
    http.ListenAndServe(":3000", nil)
}
```

`page.DOM()` is a **method** that returns the document; `.Title()`, `.Head()`,
`.Body()`, and `.HTML()` are methods on it. (Some older README snippets show
`page.DOM.Body` as a field — that's stale; always call `page.DOM().Body()`.)

## Rendering model

- **AutoRender** is on by default: every triggered event binding re-renders and
  diffs the page automatically.
- To render manually, set a component's AutoRender off and call
  `hlive.Render(ctx)` (re-renders the whole page).
- To re-render just one component and its children, call
  `hlive.RenderComponent(ctx, comp)`. This is powerful but easy to get subtly
  wrong — only reach for it when whole-page renders are a measured problem.

The `ctx` passed to your handler is what `Render`/`RenderComponent` need.

## Lifecycle (CM components)

```go
cm := l.CM("table", /* children */)
cm.SetMount(func(ctx context.Context) {
    // runs after the component mounts (WebSocket phase) — fetch data here
})
cm.SetUnmount(func(ctx context.Context) { /* cleanup before removal */ })
```

`Mount` is the place to load per-session data — including values stashed in
`ctx` by middleware.

## Gotchas (read before debugging weird diffs)

- **SSR vs WebSocket.** Each page load is two requests: an initial HTML render
  (SSR — `Mount` is **not** called) and then a WebSocket connection (where
  `Mount` runs and live diffing begins). The WS connects to the same URL with
  `?hlive=1`. One `Page` instance = one connected user.
- **`GetNodes` must be deterministic.** A `Tag`'s children are read many times,
  not only at render. Do **no** I/O and make no state changes when building
  children — fetch data in event handlers or `Mount`, not while constructing the
  tree.
- **HLive is blind to the real browser DOM.** It assumes the DOM matches what it
  last set. If your own JavaScript mutates the DOM, diffing can break.
- **Invalid HTML / browser quirks break path-finding.** Browsers silently drop
  invalid HTML and "fix" some structures (e.g. injecting `<tbody>`, merging
  adjacent text nodes). HLive locates elements by path, so these surprises can
  make updates land in the wrong place. Keep markup valid and well-formed.
- **Initial sync.** Some browsers (e.g. Firefox) keep form field values across a
  reload. HLive resends that data; an input needs an event binding to receive it,
  and you can detect it with `e.IsInitial`.

## Worked example (interactive counter)

```go
import (
    "context"

    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func home() *l.Page {
    count := l.Box(0)

    page := l.NewPage()
    page.DOM().Title().Add("Click")
    page.DOM().Body().Add(
        P("Clicks: ",
            Button(
                l.On("click", func(_ context.Context, _ l.Event) {
                    count.Lock(func(v int) int { return v + 1 })
                }),
                count,
            ),
        ),
    )
    return page
}
```

## Reference

Source of truth for current API is the runnable examples and core source, not
prose docs:

- `_example/click/click.go`, `_example/todo/todo.go`, `_example/url_params/url_params.go`
- `component.go`, `componentMountable.go`, `tag.go`, `page.go`, `dom.go`, `event.go`, `hlive.go`

For building markup with the typed hhtml builders, see the **hlive-hhtml** skill.
For laying out a multi-page app, see the **hlive-project** skill. For dynamic
lists, focus, real-time updates, and diff-apply callbacks, see the **hlivekit**
skill. For browser tests, see the **hlive-testing** skill.
