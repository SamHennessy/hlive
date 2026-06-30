---
description: Scaffold a complete, buildable new HLive project on disk
argument-hint: "[module-path] [dir]"
---

Scaffold a brand-new, runnable HLive project — not just code to copy. Follow
the layout in the **hlive-project** skill, build markup with the **hhtml**
skill (dot-import), and use the **hlive-core** skill for the event/state API.
Verify every symbol you emit against the real source — do not invent helpers.

Arguments (optional): `$ARGUMENTS`
- **module-path** — the Go module path, e.g. `github.com/you/myapp`.
- **dir** — target directory. Default: a new directory named after the last
  path segment of the module path, created in the current working directory.

If the module path is missing, ask for it. If `dir` already exists and is
non-empty, stop and confirm with the user before writing into it — never
silently overwrite existing files.

## 1. Create the tree

```
<dir>/
├── go.mod
├── main.go
├── justfile
├── .gitignore
├── page/
│   ├── home.go
│   └── components.go
└── assets/
    ├── reset.css
    └── main.css
```

### `main.go`

Stdlib routing, no router library. Serve static assets from `./assets`.

```go
package main

import (
    "log"
    "net/http"

    "<module-path>/page"
    l "github.com/SamHennessy/hlive"
)

func main() {
    http.Handle("/", l.NewPageServer(page.Home))

    fs := http.FileServer(http.Dir("./assets"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fs))

    log.Println("listening on :3000")
    if err := http.ListenAndServe(":3000", nil); err != nil {
        log.Fatal(err)
    }
}
```

`gofmt` sorts imports within a group alphabetically by path, so
`<module-path>/page` goes before `github.com/SamHennessy/hlive` whenever the
module path sorts earlier (as it almost always will). After writing the file,
run `gofmt -l .` and reorder if it flags anything.

### `page/home.go`

Exported PascalCase page factory. Build markup with hhtml's dot-import. Show
both a static element and one interactive element (a button + `l.Box` counter)
so the round-trip is visible immediately.

```go
package page

import (
    "context"

    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func Home() *l.Page {
    count := l.Box(0)

    page := l.NewPage()
    page.DOM().Title().Add("My HLive App")
    page.DOM().Head().Add(
        Meta(Charset("UTF-8")),
        Link(Rel("stylesheet"), Href("/assets/reset.css")),
        Link(Rel("stylesheet"), Href("/assets/main.css")),
    )
    page.DOM().Body().Add(
        Div(Class("container"),
            header(),
            Main(Class("main"),
                H1("Welcome to HLive"),
                P("Edit ", Code("page/home.go"), " and refresh to see changes."),
                P("Clicks: ", count,
                    Button(Id("btn"),
                        l.On("click", func(_ context.Context, _ l.Event) {
                            count.Lock(func(v int) int { return v + 1 })
                        }),
                        "Click me",
                    ),
                ),
            ),
        ),
    )
    return page
}
```

### `page/components.go`

Unexported camelCase shared components.

```go
package page

import (
    l "github.com/SamHennessy/hlive"
    . "github.com/SamHennessy/hlive/hhtml"
)

func header() l.Adder {
    return Header(Class("header"),
        Div(Class("logo"), "My HLive App"),
    )
}
```

### `assets/reset.css` and `assets/main.css`

A small, generic reset and a minimal layout/typography stylesheet (container
max-width, basic spacing, readable font stack) — enough to look intentional,
not a design system.

### `justfile`

```
build:
    go build -o ./bin/server .

run:
    go run .
```

### `.gitignore`

```
/bin/
```

## 2. Initialize and verify it builds

Run these from `<dir>`, in order, fixing and retrying on any failure before
reporting success:

1. `go mod init <module-path>`
2. `go get github.com/SamHennessy/hlive@latest` (pulls in `hhtml` as a
   subpackage of the same module — no separate `go get` needed for it)
3. `go build -o /dev/null .` — must succeed
4. `gofmt -l .` — should print nothing; reformat any flagged file

## 3. Report

Show the created file tree, then how to run it:

```
cd <dir>
go run .
# open http://localhost:3000
```

Point to the **hlive-project**, **hlive-hhtml**, and **hlive-core** skills for
extending the app, and `/hlive-component` for adding more pages/components.
