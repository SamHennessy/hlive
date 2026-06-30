---
name: hlive-project
description: >-
  Structure a real multi-page HLive application in Go — the main.go entry point
  and net/http routing, a page/ package of page factories, reusable component
  functions, static asset serving, and build/run setup. Use when starting a new
  HLive project, adding a page or route, organizing components across files, or
  deciding where state and assets live (beyond the single-file _example/ demos).
---

# Structuring an HLive project

HLive's `_example/` demos are single files to keep them readable. A real app
spreads across packages. This is the recommended layout (modeled on the
`hlive_docs` app), using **hhtml** for markup — see the **hlive-hhtml** and
**hlive-core** skills for the building blocks.

## Layout

```
myapp/
├── main.go              # entry point: routing + static assets
├── go.mod               # module github.com/you/myapp
├── justfile             # build/run recipes (optional)
├── page/                # one file per page; package "page"
│   ├── home.go          #   func Home() *l.Page
│   ├── about.go         #   func About() *l.Page
│   ├── components.go    #   shared unexported components: header(), footer()
│   └── demo/            #   sub-package for a group of related routes
│       └── clicker.go   #   func Clickers() http.Handler
├── src/                 # plain Go helper packages (no HLive), unit-testable
│   └── .../
└── assets/              # css/js/images served at /assets/
    └── main.css
```

Keep it flat — most logic lives in `page/`. Add a sub-package under `page/` only
to group a feature's routes.

## Entry point — `main.go`

Use the standard library `net/http`; HLive's `PageServer` is just an
`http.Handler`. Register each page factory, then serve static files.

```go
package main

import (
    "log"
    "net/http"

    l "github.com/SamHennessy/hlive"
    "github.com/you/myapp/page"
    "github.com/you/myapp/page/demo"
)

func main() {
    http.Handle("/", l.NewPageServer(page.Home))
    http.Handle("/about", l.NewPageServer(page.About))
    http.Handle("/demos/clickers", demo.Clickers())

    // static assets, e.g. /assets/main.css -> ./assets/main.css
    fs := http.FileServer(http.Dir("./assets"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

    log.Println("listening on :3000")
    if err := http.ListenAndServe(":3000", nil); err != nil {
        log.Fatal(err)
    }
}
```

`l.NewPageServer` takes a **page factory** `func() *l.Page`. For routes that need
extra wiring (their own handler, query params), expose a `func() http.Handler`
instead (as `demo.Clickers()` does).

## Pages — the `page` package

One **exported, PascalCase** factory per page, each returning `*l.Page`. Build
the `<head>` (title + stylesheets) and `<body>` with hhtml, composing shared
components:

```go
package page

import (
    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func Home() *l.Page {
    page := l.NewPage()
    page.DOM().Title().Add("MyApp — Home")
    page.DOM().Head().Add(Link(Rel("stylesheet"), Href("/assets/main.css")))

    page.DOM().Body().Add(
        Div(Class("container"),
            header(),
            Main(Class("main"),
                H1("Welcome"),
                P("Built with HLive."),
            ),
        ),
    )
    return page
}
```

There is no central layout wrapper; the per-page head/shell boilerplate is
repeated (or factored into a small helper you call from each page). One `*l.Page`
instance serves one connected user.

## Reusable components

Shared UI is just **unexported, camelCase** functions returning `l.Adder`
(the interface hhtml builders return), kept in `page/components.go`:

```go
func header() l.Adder {
    return Header(Class("header"),
        Div(Class("logo"), "MyApp"),
        Nav(
            Ul(Class("nav-links"),
                Li(A(Href("/"), "Home")),
                Li(A(Href("/about"), "About")),
            ),
        ),
    )
}
```

Call them inside page factories: `Body().Add(Div(header(), ...))`.

## Where state lives

- Page-local reactive values: `l.Box(...)` / `l.NewLockBox[T](...)`, captured in
  the page factory's closure and updated from `l.On` handlers.
- Feature state that spans components or goroutines (live demos, real-time
  feeds): a struct holding the boxes plus methods, often with `hlivekit.PubSub`
  for fan-out. See the **hlivekit** skill and `hlive_docs/page/demo/clicker.go`.

Keep non-UI logic (parsing, calculations, data access) in plain Go packages
under `src/` so it's unit-testable without a browser.

## Static assets

Put CSS/JS/images under `assets/` and serve them as shown in `main.go`. Link
them in each page's `<head>` with `Link(Rel("stylesheet"), Href("/assets/..."))`.

## Build & run

A `justfile` keeps commands handy (or use `go` directly):

```
build:
    go build -o ./deploy/server .

run:
    go run .
```

`go run .` from the module root is enough for local development; the server
serves `./assets` relative to its working directory.

## Reference

`hlive_docs/main.go`, `hlive_docs/page/home.go`, `hlive_docs/page/components.go`,
`hlive_docs/page/demo/clicker.go`, `hlive_docs/justfile`. (hlive_docs is a
work-in-progress and is mid-migration to hhtml — follow the structure, and prefer
hhtml for new markup.) For testing pages from a `page` package, see the
**hlive-testing** skill.
