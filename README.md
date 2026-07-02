# HLive

Build interactive, real-time web UIs in pure Go. No JavaScript, no npm, no build step — just Go code that talks to the browser over a WebSocket.

If you know Go, you already have everything you need to build a modern, interactive web app. HLive lets you skip the second language: no JavaScript to write, no npm dependencies to audit, no frontend build pipeline to maintain, and no state to keep in sync by hand between a Go backend and a JS frontend. You write one program, in Go. HLive renders it as HTML, opens a WebSocket, and pushes only the DOM changes your Go code actually made — live, on every keystroke or click, without you writing a single line of client-side code.

[Learn more and see live demos](https://hlive.thefam.uk/)

## Contributing

Contributions welcome

### Run tests

#### Setup

Install Play wright Go

```shell
make install-test
```

or

```shell
go run github.com/playwright-community/playwright-go/cmd/playwright install --with-deps
```

#### Run

```shell
go test ./...
```


## TODO (out of date)

## v0.2.0

- Race conditions in examples
- Update docs based on API changes
- Add SSR example

### API Change

- Add missing data from browser events, like:
  - https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent
  - https://developer.mozilla.org/en-US/docs/Web/API/PointerEvent
  - https://developer.mozilla.org/en-US/docs/Web/API/Touch_events
  - https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent

### Bugs

- Need to reflect in the browser virtual DOM that a select option has become selected when a user selects it
  - So that we can reset the selection (e.g. move dropdowns)
- Can read a POST but can't pass POST data to a render (display errors)
  - Makes Auth logins an issue
  - Workaround it to go a redirect with an url param
- Preempt disable on click prevents form submit in Chrome

### Internal improvements

#### Groups

- Add the Grouper interface
  - func GetGroup() []interface{}
  - Add the NoneNodeElementsGroup

#### Page Pipeline

- HTTP request w, r

#### Performance

- Use binary for websocket
  - Skips UTF8 processing
  - Is this worth the trouble?
- Alternative WebSocket lib?
  - https://github.com/nhooyr/websocket
- Add special logic for class tags
  - Add, remove classes
- Add special logic for style tags
  - Only update property value if needed

- Batch message sends and receives in the javascript (https://developer.mozilla.org/en-US/docs/Web/API/HTML_DOM_API/Microtask_guide)

- Move file upload out of Page.js

#### Other

- Add log level to client side logging
- Send config for debug, and log level down to client side
  - Set via an attribute

#### Can we make a hash of a simplified DOM tree?

- If that page hash is not found in the cache then we need a fallback
  - Force a browser reload with a new hash?
- Need a way to know that the version of HLive has changed, if so need a hard page reload and cache bypass 

#### Add support for Wails
- would be a JS binding for reading incoming messages what just blocks when waiting for a message
- another binding sending messages

### Tests

- Add JavaScript test for multi child delete
    - E.g.:
      - d|d|doc|1>1>0>0>1>2||
      - d|d|doc|1>1>0>0>1>3||
- `Page.Close`

- How well does [Alpine.JS](https://alpinejs.dev/) work with HLive
  - https://dockyard.com/blogs/optimizing-user-experience-with-liveview

#### Performance

- Need a way to test performance improvement ideas
- Why are large tables of data slow to page?
  - It's faster to delete all the rows first
  - Can we add a way for a component like List to inform tree copy not to bother doing a diff and just do a full HTML replacement
  - If we check for `hid` and they are different, then do an HTML replace

### Docs

- Add initial page sync to concepts
  - An input needs to have binding for this to work
- Add one page one user to concepts
- How to debug in browser
- How on mount render order issues
  - Try to update an element that has already been processed the diff will not be noticed
  - Use the dirty tree error?
- Logging
- Plugins
- Preempt pattern
- Event bubbling 
- Prevent default
- Stop propagation 
- Explain performance goals
  - Explain why WASM is not a good fit for the goals
- From the beginning tech intro - https://www.reddit.com/r/golang/comments/w5v4oe/comment/ihcm8i9/?utm_source=share&utm_medium=web2x&context=3
- Page hooks

### Security

- Add a CSRF token
  - https://github.com/gorilla/csrf
  - Is this needed?

### New Features/Improvements

- Look for a CSS class to show on a failed reconnect
  - Set current z-index higher than Bulma menu for default disconnect layer
- Allow adding mount and unmount function as elements?

- Add support for "key" to allow better diff logic for lists
  - Use `hid`

- Add a `func() *l.NodeGroup` value
  - Reduce code count
  - Does it solve any real issues

- ComponentList
  - Operations by ID
    - Get by ID
    - Remove By ID

- User friendly HTTP error pages
  - Display a request ID if it exits

- Add can take a `func() string` this would be kept in the tree and re-run on each render
  - Could be expensive

#### Multi file upload using WS and HTTP

- Need a count of files
- Group them together in an event?
- Make a channel?
- File upload progress?

#### Visibility

- Is a component visible?
- Trigger event when visible?
- Scroll events
  - Page position
  - Viewport
