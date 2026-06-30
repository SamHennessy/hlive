# HLive Claude Code plugin

Expert guidance and scaffolding for [HLive](https://github.com/SamHennessy/hlive),
the server-side virtual-DOM framework for Go, inside [Claude Code](https://code.claude.com).

## What you get

- **`hlive-core` skill** — the core API: builders (`l.T`/`l.C`/`l.CM`), `l.On`
  event handling, `l.Box` reactive state, page/server setup, the render model,
  and the SSR-vs-WebSocket lifecycle and tree-diff gotchas.
- **`hlive-hhtml` skill** — building markup with the typed `hhtml` builders
  (`Div`, `Button`, `Class`, `Href`, …) instead of raw tag strings.
- **`hlive-project` skill** — structuring a real multi-page app: `main.go`
  routing, a `page/` package, reusable components, and static assets.
- **`hlivekit` skill** — the toolkit: `ComponentList` for dynamic lists, `Focus`,
  `OnDiffApply`, `PubSub` for real-time updates, visibility/scroll/redirect
  helpers.
- **`hlive-testing` skill** — writing Playwright browser tests with `hlivetest`.
- **`/hlive-component` command** — scaffold a component (plain or mountable) with
  events wired up.
- **`/hlive-new-project` command** — scaffold a complete, buildable new HLive
  project on disk (go.mod, main.go, a `page/` package, assets, justfile),
  verified with a real `go build`.

The skills activate automatically when you work on HLive code; you don't need to
invoke them.

## Install

This plugin is published from the HLive repo's marketplace:

```
/plugin marketplace add SamHennessy/hlive
/plugin install hlive@hlive
```

To develop against a local checkout, point the marketplace at the repo path:

```
/plugin marketplace add /path/to/hlive
/plugin install hlive@hlive
```

## Layout

```
plugin/
├── .claude-plugin/plugin.json
├── skills/
│   ├── hlive-core/SKILL.md
│   ├── hlive-hhtml/SKILL.md
│   ├── hlive-project/SKILL.md
│   ├── hlivekit/SKILL.md
│   └── hlive-testing/SKILL.md
└── commands/
    ├── hlive-component.md
    └── hlive-new-project.md
```

The skills are written against the runnable code in `_example/` and
`hlivetest/pages/` as the source of truth for the current API.
