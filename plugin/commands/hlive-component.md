---
description: Scaffold an HLive component (plain or mountable) with event handlers wired up
argument-hint: "[name] [tag] [click|keyup|...] [--mountable]"
---

Scaffold a new HLive component or page. Build markup with the **hhtml** builders
(see the **hlive-hhtml** skill) and follow the project layout in the
**hlive-project** skill. Use the **hlive-core** skill for the event/state/render
API, and verify any API you emit against the real source — do not invent helpers.

Arguments (all optional): `$ARGUMENTS`
- **name** — the component/function name (e.g. `counter`, `searchBox`).
- **tag** — the base HTML tag (e.g. `button`, `input`, `div`). Default `div`.
- **event(s)** — DOM event names to wire (e.g. `click`, `keyup`).
- **`--mountable`** — generate a `l.CM` component with a `SetMount` data-fetch hook.

If any of these are missing or ambiguous, ask the user briefly before generating.

Generate idiomatic Go that:

1. Imports both packages — `l "github.com/SamHennessy/hlive"` and the hhtml
   dot import `. "github.com/SamHennessy/hlive/hhtml"`.
2. Returns a builder function (e.g. `func newCounter() l.Adder`) for a reusable
   component, or an exported page factory `func Name() *l.Page` placed in the
   `page` package when scaffolding a whole page.
3. Builds markup with hhtml tag builders (`Button(...)`, `Div(...)`, `Input(...)`).
   An event binding makes the tag a Component automatically; pass `--mountable`
   intent to use `l.CM(tag, ...)` + `comp.SetMount(func(ctx context.Context) { /* fetch data */ })`
   when the component must load data on appearance.
4. Wires each requested event with `l.On("<event>", func(ctx context.Context, e l.Event) { ... })`,
   reading the right `l.Event` field for the event type (`e.Value` for input,
   `e.Selected` for checkbox/radio, `e.Key` for keyboard, `e.File` for uploads).
5. Holds any mutable state in an `l.Box(...)` (thread-safe, renderable) and updates
   it with `count.Lock(func(v T) T { return ... })`, or passes a value by
   reference (`&v`) for simple display.
6. Gives elements you'd select in a test a stable id via hhtml's `Id("...")`.

Match the surrounding file's style. Keep handlers as the place for any I/O —
never fetch data while building child nodes (`GetNodes` must stay deterministic).

After generating, show how to add it to a page (`page.DOM().Body().Add(newX())`)
and offer to write a matching `hlivetest` browser test (see the **hlive-testing**
skill).

### Reference example shape

```go
import (
    "context"

    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func newCounter() l.Adder {
    count := l.Box(0)

    return Button(
        Id("counter"),
        l.On("click", func(_ context.Context, _ l.Event) {
            count.Lock(func(v int) int { return v + 1 })
        }),
        "Clicks: ", count,
    )
}
```
