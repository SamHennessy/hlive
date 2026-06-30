---
name: hlive-testing
description: >-
  Write browser tests for HLive pages using the hlivetest package
  (github.com/SamHennessy/hlive/hlivetest) and Playwright. Use when testing
  HLive UI behavior end-to-end — spinning up a test server, driving a real
  browser (Click/ClickAndWait, TextContent, GetAttribute), and asserting with
  hlivetest.Diff. Use when an HLive task mentions tests, Playwright, or
  verifying interactive behavior.
---

# HLive Testing

HLive's behavior lives in the browser round-trip, so it's tested by driving a
real browser with Playwright via the `hlivetest` package. Tests boot an HLive
server, open a browser page against it, interact, and assert on the resulting
DOM.

## Setup

Tests need Playwright browsers installed:

```shell
make install-test
# or: go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
```

Because these are heavy browser tests, guard them so `go test -short` skips them
(the harness below does this).

## The harness pattern

HLive's own suite (`hlivetest/pages/`) uses a small per-package harness. Copy it
into your test package:

```go
package pages_test

import (
    "testing"

    l "github.com/SamHennessy/hlive"
    "github.com/SamHennessy/hlive/hlivetest"
    "github.com/playwright-community/playwright-go"
)

type harness struct {
    server   *hlivetest.Server
    pwpage   playwright.Page
    teardown func()
}

func setup(t *testing.T, pageFn func() *l.Page) harness {
    t.Helper()

    if testing.Short() {
        t.Skip("skipping test in short mode.")
    }

    h := harness{
        server: hlivetest.NewServer(pageFn),
        pwpage: hlivetest.NewBrowserPage(),
    }

    h.teardown = func() {
        if err := h.pwpage.Close(); err != nil {
            t.Error(err)
        }
    }

    if _, err := h.pwpage.Goto(h.server.HTTPServer.URL); err != nil {
        t.Fatal("goto page:", err)
    }

    return h
}
```

`NewServer` takes a page **factory** (`func() *l.Page`) — the same kind of
function you hand to `l.NewPageServer`. In a structured app (see the
**hlive-project** skill) that's an exported factory from your `page` package, so
a test imports it and passes it straight in: `setup(t, page.Home)`.

Give the elements you select stable ids. With hhtml use the `Id(...)` attribute:

```go
import (
    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func Counter() *l.Page {
    count := l.Box(0)
    page := l.NewPage()
    page.DOM().Body().Add(
        Button(Id("btn"), l.On("click", func(_ context.Context, _ l.Event) {
            count.Lock(func(v int) int { return v + 1 })
        }), "+"),
        Div(Id("count"), count),
    )
    return page
}
```

## Writing a test

```go
func TestClick_OneClick(t *testing.T) {
    t.Parallel()

    h := setup(t, pages.Click()) // pages.Click() returns the page factory
    defer h.teardown()

    hlivetest.Diff(t, "0", hlivetest.TextContent(t, h.pwpage, "#count"))

    hlivetest.ClickAndWait(t, h.pwpage, "#btn") // click, then wait for the diff

    hlivetest.Diff(t, "1", hlivetest.TextContent(t, h.pwpage, "#count"))
}
```

Use `ClickAndWait` before an assertion (it waits for HLive's diff to apply). Use
plain `Click` for intermediate clicks where you'll assert only at the end:

```go
for i := 0; i < 9; i++ {
    hlivetest.Click(t, h.pwpage, "#btn")
}
hlivetest.ClickAndWait(t, h.pwpage, "#btn")
hlivetest.Diff(t, "10", hlivetest.TextContent(t, h.pwpage, "#count"))
```

## hlivetest helpers

All take `(t, pwpage, ...)` unless noted:

- `Click(selector)` — click without waiting for the diff.
- `ClickAndWait(selector)` — click, then block until the diff is applied.
- `TextContent(selector) string` — element text.
- `GetAttribute(selector, attribute) string` — an attribute value.
- `GetID(selector) string` — the element's HLive id.
- `Title(pwpage) string` — page title.
- `Diff(want, got any)` — assert equal, reporting a diff on failure (`t.Error`).
- `DiffFatal(want, got any)` — same but `t.Fatal`.
- `FatalOnErr(err)` — fail fast on an error.
- `NewServer(pageFn) *Server` / `NewBrowserPage() playwright.Page` — fixtures.

For interactions hlivetest doesn't wrap (typing, hover, selects), use the
`playwright.Page` directly (`h.pwpage.Fill(...)`, `.Hover(...)`), then assert with
`hlivetest.Diff`.

## Running

```shell
go test ./...          # full run (needs browsers)
go test -short ./...   # skips the browser tests
```

## Reference

`hlivetest/hlivetest.go`, `hlivetest/server.go`, `hlivetest/browser.go`, and the
examples `hlivetest/pages/harness_test.go`, `hlivetest/pages/click_test.go`,
`hlivetest/pages/click.go`. Larger suites live under `systemtests/`.
