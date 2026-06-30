---
name: hlivekit
description: >-
  Use the hlivekit toolkit (github.com/SamHennessy/hlive/hlivekit) for common
  HLive needs — dynamic lists of components (ComponentList / hlivekit.List),
  giving an input focus, running server logic after a browser diff applies
  (OnDiffApply), real-time fan-out updates (PubSub), element-visibility events,
  scroll/redirect helpers, and preempt-disable-on-click. Use when an HLive task
  involves any of these patterns or imports the hlivekit package.
---

# HLiveKit

`hlivekit` is HLive's companion toolkit of reusable patterns. Import alongside
the core package:

```go
import (
    l "github.com/SamHennessy/hlive"
    "github.com/SamHennessy/hlive/hlivekit"
)
```

Pick the helper that matches the need:

## Dynamic lists — `ComponentList`

For lists whose items are added/removed at runtime (search results, todo items,
table rows). It tracks items and runs their teardown when removed, avoiding leaks.

```go
list := hlivekit.List("tbody")           // shortcut for NewComponentList
// ...add it to the tree like any node...

list.AddItem(l.CM("tr",                  // items are Teardowner (CM components)
    l.T("td", key),
    l.T("td", value),
))
list.RemoveItems(item)                   // remove specific items
list.RemoveAllItems()                    // clear
```

Constructors: `hlivekit.List(name, ...)` / `hlivekit.NewComponentList(name, ...)`.
Methods: `Add`, `AddItem(items ...l.Teardowner)`, `RemoveItems`, `RemoveAllItems`.

`hlivekit.NewComponentListSimple(...)` is a lighter variant **without** the memory
cleanup logic — only use it when you manage item lifetimes yourself.

It's idiomatic to populate the list inside a `CM` component's `SetMount` (see the
hlive-core skill) so it fills when the component appears:

```go
cm := l.CM("table", list)
cm.SetMount(func(ctx context.Context) {
    for _, row := range fetchRows(ctx) {
        list.AddItem(l.CM("tr", l.T("td", row.Name)))
    }
})
```

## Focus — `Focus`

Give an input focus from the server (no JS). Add it as an attribute; pair with a
one-shot `focus` binding + `FocusRemove` so it only fires once:

```go
input := l.C("input")
input.Add(hlivekit.Focus(), l.OnOnce("focus", func(ctx context.Context, _ l.Event) {
    hlivekit.FocusRemove(input)
}))
```

## Run server logic after a diff applies — `OnDiffApply`

Fires a handler once HLive's diff has been applied in the browser. Useful for
chaining updates (animation frames) or reacting to a render completing.

```go
comp.Add(hlivekit.OnDiffApply(func(ctx context.Context, e l.Event) { /* ... */ }))
```

Also `hlivekit.OnDiffApplyOnce(...)`. See `_example/callback` and
`_example/animation`.

## Real-time updates — `PubSub`

Server-side publish/subscribe to fan out changes to many connected sessions (or
between components in one session). One `PubSub` is shared across the app.

```go
pubSub := hlivekit.NewPubSub()
page.DOM().HTML().Add(hlivekit.InstallPubSub(pubSub)) // install once per page

// subscribe a handler to topics
pubSub.Subscribe(hlivekit.NewSub(func(m hlivekit.QueueMessage) {
    // react; call l.Render(ctx)/RenderComponent to push updates
}), "my-topic")

// publish from anywhere
pubSub.Publish("my-topic", payload)
```

Key API: `NewPubSub()`, `InstallPubSub(ps)`, `ps.Subscribe(sub, topics...)`,
`ps.Unsubscribe(...)`, `ps.Publish(topic, value)`, `NewSub(onMessageFn)`. For
components that subscribe on mount there's `ComponentPubSub` (`hlivekit.CPS` /
`NewComponentPubSub` / `WrapComponentPubSub`) with `SetMountPubSub`. See
`_example/pubsub/pubsub.go`.

## Other helpers

- **`OnElementVisible(handler)`** / `OnElementVisibleOnce` — fire when an element
  scrolls into view (lazy-load, infinite scroll).
- **`ScrollIntoView(alignToTop)`** / `ScrollIntoViewRemove`,
  **`ScrollTop(val)`** / `ScrollTopRemove` — control scrolling from the server.
- **`Redirect(url)`** — issue a client-side redirect as an attribute.
- **`PreemptDisableOn(eventBinding)`** — disable a control in the browser
  *immediately* on click (before the round-trip) to prevent double-submits.

## Reference

`hlivekit/README.md` and the source: `hlivekit/componentList.go`,
`hlivekit/focus.go`, `hlivekit/diffapply.go`, `hlivekit/pubsub.go`,
`hlivekit/elementVisible.go`, `hlivekit/scrollIntoView.go`,
`hlivekit/scrollTop.go`, `hlivekit/redirect.go`,
`hlivekit/preemptDisableOnClick.go`. Examples: `_example/todo`,
`_example/pubsub`, `_example/callback`, `_example/url_params`.
