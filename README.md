# HLive

Server-side virtual DOM

HLive is a server-side WebSocket based dynamic template-less view layer for Go.

HLive is a fantastic tool for creating complex and dynamic browser-based user interfaces for developers who want to keep all the logic in Go.

All the power and data available on the backend with the responsive feel of a pure JavaScript solution. 

It's a great use case for admin interfaces and internal company tools.

## Notice

**The first version of the API is under active development. Change is likely. Your feedback is welcome.**

Please help the project by building something and giving us your feedback.

## Table of contents
- [Quick Start Tutorial](#quick-start-tutorial)
  * [Step 1: Static Page](#step-1--static-page)
  * [Step 2: Interactive Page](#step-2--interactive-page)
- [Examples](#examples)
  * [Simple](#simple)
    + [Click](#click)
    + [Hover](#hover)
    + [Diff Apply](#diff-apply)
  * [Advanced](#advanced)
    + [Animation](#animation)
    + [Clock](#clock)
    + [File Upload](#file-upload)
    + [Initial Sync](#initial-sync)
    + [Local Render](#local-render)
    + [Session](#session)
    + [To Do List](#to-do-list)
    + [URL Parameters](#url-parameters)
- [Concepts](#concepts)
  * [Tag](#tag)
  * [Attribute](#attribute)
    + [CSS Classes](#css-classes)
    + [Style Attribute](#style-attribute)
  * [Tag Children](#tag-children)
  * [Components](#components)
    + [EventBinding](#eventbinding)
    + [EventHandler](#eventhandler)
  * [Node](#node)
  * [Element](#element)
  * [Page](#page)
    + [HTML vs WebSocket](#html-vs-websocket)
  * [PageSession](#pagesession)
  * [PageServer](#pageserver)
  * [Middleware](#middleware)
  * [PageSessionStore](#pagesessionstore)
  * [HTTP vs WebSocket Render](#http-vs-websocket-render)
  * [Tree and Tree Copy](#tree-and-tree-copy)
  * [WebSocket Render and Tree Diffing](#websocket-render-and-tree-diffing)
  * [First WebSocket Render](#first-websocket-render)
  * [AutoRender and Manuel Render](#autorender-and-manuel-render)
  * [Local Render](#local-render-1)
  * [Differ](#differ)
  * [Render](#render)
  * [HTML Type](#html-type)
  * [JavaScript](#javascript)
  * [Virtual DOM, Browser DOM](#virtual-dom--browser-dom)
  * [Lifecycle](#lifecycle)
- [Known Issues](#known-issues)
  * [Invalid HTML](#invalid-html)
  * [Browser Quirks](#browser-quirks)
- [Inspiration](#inspiration)
  * [Phoenix LiveView](#phoenix-liveview)
  * [gomponents](#gomponents)
  * [ReactJS and JSX](#reactjs-and-jsx)
- [Similar Projects](#similar-projects)
  * [GoLive](#golive)
  * [live](#live)
- [TODO](#todo)
  * [API Change](#api-change)
  * [Bugs](#bugs)
  * [Internal improvements](#internal-improvements)
    + [Groups](#groups)
    + [Page Pipeline](#page-pipeline)
    + [Other](#other)
  * [Tests](#tests)
  * [Docs](#docs)
  * [Security](#security)
  * [New Features/Improvements](#new-features-improvements)
- [HHot](#hhot)
  * [HHot Ideas](#hhot-ideas)
  * [Older ideas](#older-ideas)
    + [Serializable tree/component state](#serializable-tree-component-state)
    + [Do more with the HTTP Server render](#do-more-with-the-http-server-render)
    + [Multi file upload using WS and HTTP](#multi-file-upload-using-ws-and-http)
    + [Visibility](#visibility)
- [Contributing](#contributing)
  * [Run tests](#run-tests)
    + [Setup](#setup)
    + [Run](#run)

## Quick Start Tutorial

### Step 1: Static Page

Import HLive using the optional alias `l`:

```go
package main

import l "github.com/SamHennessy/hlive"
```

Let's create our first page:

```go
func home() *l.Page {
	page := l.NewPage()
	page.DOM.Body.Add("Hello, world.")

	return page
}
```

Next we use a `PageServer` to add it to an HTTP router:

```go
func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("Listing on :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("Error: http listen and serve:", err)
	}
}
```

Your editor should add the extra imports `http` and `log` for you.

You can now run it, for example:
```shell
go run ./tutorial/helloworld/helloworld.go
```

In a browser go to http://localhost:3000 you should see this:

![Hello world step 1](./_tutorial/helloworld/img/step1.png)

### Step 2: Interactive Page
HLive is all about interactive content. We're going to add a text input field to let us type our own hello message.

We need to replace our existing `home` function. We need a string to hold our message:

```go
func home() *l.Page {
	var message string
```

Now we're going to create a `Component`. `Component`'s are HTML tags that can react to browser events. We are going to base our `Component` on the `input` HTML tag.

```go
	input := l.C("input")
```

We want to set the input to a text type. We do this adding a`Attrs` map to our `Component`.

```go
	input.Add(l.Attrs{"type": "text"})
```

Here we add an `EventBinding` to listen to "keyup" JavaScript events. When triggered, the handler function will be called. Our handler will update `message`. It does this by using the data passed in the `Event` parameter.

```go
	input.Add(l.On("keyup", func(ctx context.Context, e l.Event) {
		message = e.Value
	}))
```

We create a new `Page` like before:

```go
	page := l.NewPage()
```

Here we add our `input` to the body but first we wrap it in a `div` tag.

```go
	page.DOM.Body.Add(l.T("div", input))
```

Next, we will display our message. Notice that we're passing `message` by reference. That's key for making this example work. We'll also add an "hr" tag to stop it being squashed todeather.

```go
	page.DOM.Body.Add(l.T("hr"))
	page.DOM.Body.Add("Hello, ", &message)
```

Finally, we return the `Page` we created.
```go
	return page
}
```

Let's see that all together, but this time I'm going to use some shortcuts. Can you spot the differences?

```go
func home() *l.Page {
	var message string

	input := l.C("input",
		l.Attrs{"type": "text"},
		l.OnKeyUp(func(ctx context.Context, e l.Event) {
			message = e.Value
		}),
	)

	page := l.NewPage()
	page.DOM.Body.Add(
		l.T("div", input),
		l.T("hr"),
		"Hello, ", &message,
	)

	return page
}
```

Run it and type something into the input. The page should update to display what you typed.

![Hello world step 2](./_tutorial/helloworld/img/step2.gif)

## Examples

The examples can be run from the root of the project using `go run <path_to_example>`. For example:

```shell
go run _example/click/click.go
```

### Simple

#### Click

[_example/click/click.go](./_example/click/click.go)

Click a button see a counter update.

https://user-images.githubusercontent.com/119867/131120937-64091d27-3232-4820-ab20-e579c86cfb92.mp4

#### Hover

[_example/hover/hover.go](./_example/hover/hover.go)

Hover over an element and see another element change

#### Diff Apply

[_example/callback/callback.go](./_example/callback/callback.go)

Trigger a Diff Apply event when a DOM change is applied in the browser. Use it to trigger server side logic.

### Advanced

#### Animation

[_example/animation/animation.go](./_example/animation/animation.go)

Create a continuously changing animation by chaining Diff Apply callbacks.

#### Clock

[_example/clock/clock.go](./_example/clock/clock.go)

Push browser DOM changes from the server without the need for a user to interact with the page.

#### File Upload

[_example/fileUpload/fileUpload.go](./_example/fileUpload/fileUpload.go)

Use a file input to get information about a file before uploading it. Then trigger a file upload from the server when you're ready.

The file is uploaded via WebSocket as a binary (not base64 encoded) object.

#### Initial Sync

[_example/initialSync/initialSync.go](./_example/initialSync/initialSync.go)

Some browsers, such as FireFox, will not clear data from form fields when the page is reloaded. To the user there is data in the field and if they submit a form they expect that data to be recognised. 

Initial sync is a client side process that will send this data to the server after a page refresh. You can check for this behavior in your event handlers. 

This example also shows how to get multiple values from inputs that support that.

#### Local Render

[_example/localRender/localRender.go](./_example/localRender/localRender.go)

By default, all Components are rendered after each Event Binding that a user triggers. 

You can disable this by turning Auto Render off for a component. You can then render that manually but this will rerender the whole page.

If you only want to re-render a single component, and it's children you can do that instead. It's easy to introduce subtle bugs when using this feature.

#### Session

[_example/session/session.go](./_example/session/session.go)

An example of how to implement a user session using middleware and cookies. It also shows our to pass data from middleware to Components.

Using middleware in HLive is just like any Go app.

#### To Do List
[_example/todo/todo.go](./_example/todo/todo.go)

A simple To Do list app.

#### URL Parameters

[_example/urlParams/urlParams.go](./_example/urlParams/urlParams.go)

Passing URL params to Components is not straightforward in HLive. Here is an example of how to do it. 

This is due to the HLive having a two-step process of loading a page and Components are primarily designed to get data from Events.

## Concepts

### Tag

A static HTML tag. A Tag has a name (e.g., an `<p></p>`'s name is `p`). A Tag can have zero or more Attributes. A Tag can have child Tags nested inside it. A Tag may be Void, which means it doesn't have a closing tag (e.g., `<hr>`). Void tags can't have child Tags.

### Attribute

An Attribute has a name and a value.  (e.g., `href="https://example.com"` or `disabled=""`).

#### CSS Classes

The HLive implementation of Tag has an optional special way to work with the `class` attribute. These types are all designed to make toggling CSS classes on and off easy. 

HLive's `ClassBool` is a `map[string]bool` type. The key is a CSS class, and the value enables the class for rending if true. This allows you to turn a class on and off. (e.g. `l.ClassBool{"foo": true, "bar": true, "fizz": true}`). The order of the class names in a single `ClassBool` is NOT respected. If the order of class names is significant, you can add them as separate `ClassBool` elements, and the order will be respected. You can add a new `ClassBool` elements with the same class name, and the original `ClassBool` element will be updated.

Even better is `Class`, this is a string type that converts into a CSSBool. (e.g. `l.Class("foo bar fizz")`). The order of the class names is respected. Each class can still be turned off individually using a `ClassBool` of the `ClassOff` string type.

`ClassList` and `ClassListOff` are string slices that will enable or disable respectively CSS classes. (e.g. `l.ClassList{"foo", "bar", "fizz"}`)

#### Style Attribute

The HLive implementation of Tag has an optional special way to work with the `style` attribute.

HLive's `Style` is a `map[string]interface{}` type. The key is the CSS style rule, and the value is the value of the rule. The value can be a `string` or `nil`. If `nil`, the style rule gets removed.

The order of the style rules in a single `Style` is NOT respected. If the order of rules is significant, you can add them as separate `Style` elements, and the order will be respected.

### Tag Children

`Tag` has `func GetNodes() *l.NodeGroup`. This will return can children a `Tag` has.

This function is called many times and not always when it's time to render. Calls to `GetNodes` must
be [deterministic](https://en.wikipedia.org/wiki/Deterministic_algorithm). If you've not made a change to the `Tag`
the output is expected to be the same.

This function should not get or change data. For example, no calls to a remote API or database should happen in this function.

### Components

A `Compnent` wraps a `Tag`. It adds the ability to bind events that primarily happens in the browser to itself.

#### EventBinding

An `EventBinding` is a combination of an `EventType` (e.g., click, focus, mouseenter), with a `Component` and an `EventHandler`.

#### EventHandler

The `EventHandler` is a `func(ctx context.Context, e Event)` type.

These handlers are where you can fetch data from remote APIs or databases.

Depending on the `EventType` you'll have data in the `Event` parameter.

### Node

A Node is something that can be rendered into an HTML tag. For example, a string, `Tag`, or `Component`. An
`Attribute` is not a Node as it can't be rendered to a complete HTML tag.

### Element

An Element is anything associated with a `Tag` or `Component`. This means that in addition to nodes, `Attribute` and `EventBinding` are also Elements.

### Page

A `Page` is the root element in HLive. There will be a single page instance for a single connected user.

`Page` has HTML5 boilerplate pre-defined. This boilerplate also includes HLive's JavaScript.

#### HTML vs WebSocket

When a user requests a page, there are two requests. First is the initial request that generates the pages HTML. Then the second request is to establish a WebSocket connection.

HLive considers the initial HTML is can be though of as the Server Side Rendering phase (SSR). This SSR request will not be used when processing WebSocket requests. This render is a good candidate for use in a CDN.

When an HLive SSR page is loaded in a browser, the HLive JavaScript library will kick into action.

The first thing the JavaScript will do is establish a WebSocket connection to the server. This connection is made using the same URL with `?hlive=1` added to the URL. Due to typical load balancing strategies, the server that HLive establishes a Websocket connection to may not be the one that generated the SSR Page.

### PageSession

When the JavaScript establishes the WebSocket connection, the backend will create a new session and send down the session id to the browser.

A `PageSession` represents a single instance of a `Page`. There will be a single WebSocket connection to a `PageSession`.

### PageServer

The `PageServer` is what handles incoming HTTP requests. It's an `http.Handler`, so it can be used in your router of choice. When `PageServer` receives a request, if the request has the `hlive=1` query parameter, it will start the WebSocket flow. It will create a new instance of your `Page`. It will then make a new `PageSession`. Finally, it will pass the request to `Page` `ServerWS` function.

If not, then it will create a new `Page`, generate a complete a SSR page render and return that and discard that `Page`.

### Middleware

It's possible to wrap `PageServer` in middleware. You can add data to the context like normal. The context will be passed to your `Component`'s `Mount` function if it has one.

### PageSessionStore

To manage your all the `PageSession`s `PageServer` uses a `PageSessionStore`. By default, each page gets its own `PageSessionStore`, but it's recommended that you have a single `PageSessionStore` that's shared by all your `Page`s on a server.

`PageSessionStore` can control the number of active `PageSession`s you have at one time. This control can prevent your servers from becoming overloaded. Once the `PageSession` limit is reached, `PageSessionStore` will make incoming WebSocket requests wait for an existing connection to disconnect.

### HTTP vs WebSocket Render

`Mount` is not called on SSR requests but is called on WebSocket requests.

### Tree and Tree Copy

Tree describes a Node and all it's child Nodes.

Tree copy is a critical process that takes your `Page`'s Tree and makes a simplified clone of it. Once done, the only elements in the cloned Tree are `Tag`s and `Attribute`s.

### WebSocket Render and Tree Diffing

When it's time to do a WebSocket render, no HTML is rendered *(1)*. What happens is a new Tree Copy is created from the `Page`. This Tree is compared to the Tree that's in that should be in the browser. The differences are calculated, and instructions are sent to the browser on updating its DOM with our new Tree.

*(1) except Attributes, but that's just convenient data format.*

### First WebSocket Render

When a WebSocket connection is successfully established, we need to do 2 `Page` renders. The first is to duplicate what should be in the browser. This render will be creating a Tree Copy as if it were going to be an SSR render. This Tree is then set as the "current" Tree. Then a WebSocket Tree Copy is made. This copy will contain several attributes not present in the HTML Tree. Also, each `Component` in the Tree that implements `Mounter` will be called with the context, meaning the Tree may also have more detail based on any data fetched. This render will then be diffed against the "current" Tree and the diff instructions sent to the browser like normal.

For an initial, successful `Page` load there will be 3 renders, 2 HTML renders and a WebSocket render.

### AutoRender and Manuel Render

By default, HLive's `Component` will do a render every time an `EventBinding` is triggered.

This behaviour can be turned off on `Component` by setting `AutoRender` to `false`.

If you set `AutoRender` to `false` you can manually trigger a WebSocket render by calling `hlive.Render(ctx context.Context)` with the context passed to your handler.

### Local Render

If you want only to render a single `Component` and not the whole page, you can call `hlive.RenderComponent(ctx context.Context, comp Componenter)` you will also want to set any relevant `Component`s to `AutoRender` `false`.

### Differ

TODO: What is it and how does it work

### Render

TODO: What is it

### HTML Type

HLive's `HTML` type is a special `string` type that will render what you've set. One rule is that the HTML in `HTML` have a single root node.

### JavaScript

The goal of HLive is not to require the developer to need to write any JavaScript. As such, we have unique solutions for things like giving fields focus.

Nothing is preventing the developer from adding their JavaScript. If JavaScript changes the DOM in the browser, you could cause HLive's diffing to stop working. This is also [true in libraries like ReactJS](https://reactjs.org/docs/integrating-with-other-libraries.html#integrating-with-dom-manipulation-plugins).

### Virtual DOM, Browser DOM

HLive is blind to what the actual state of the browser's DOM is. It assumes that it what it has set it to.

### Lifecycle

TODO

## Known Issues

### Invalid HTML

If you use invalid HTML typically by using HTML where you should not, the browser will ignore the HTML and not add it to the browsers DOM. If the element were something like a `span` tag then it may not be perceivable that it's happened. If this happens then the path finding for these tags, and it's children will not work or will work strangely. 

We don't have HTML validation rules in HLive, so there is no way of warning you of this being the problem. 

### Browser Quirks

Browsers are complex things and sometimes act in unexpected ways. For example, if you have a table without a table body tag (`tbody`) some browsers will add a `tbody` to the DOM. This breaks HLives element path finding. Another example is that if you have multiple text nodes next to each other, some browsers will combine them. 

We'll try and account for this where we can by mimicking the browser's behavior when doing a Tree Copy. We've done this be the text quirk but not the `tbody` quirk yet. 



## Inspiration

### Phoenix LiveView

For the concept of server-side rendering for dynamic applications.

https://hexdocs.pm/phoenix_live_view/Phoenix.LiveView.html

### gomponents

For it's HTML API.

https://github.com/maragudk/gomponents

### ReactJS and JSX

For its component approach and template system.

https://reactjs.org/

## Similar Projects

### GoLive
[https://github.com/brendonmatos/golive](https://github.com/brendonmatos/golive)

Live views for GoLang with reactive HTML over WebSockets

### live

[https://github.com/jfyne/live](https://github.com/jfyne/live)

Live views and components for golang

## TODO

### API Change

### Bugs

- Initial sync seems to be triggering when it shouldn't
  - Maybe when the value attribute doesn't exist?
- Set the z-index higher than Bulma menu for default disconnect layer
  - Need to test
- Need to reflect in the browser virtual DOM that a select option has become selected when a user selects it
  - So that we can reset the selection (e.g. move dropdowns)
- Can read a POST but can't pass POST data to a render (display errors)
  - Makes Auth logins an issue
  - Workaround it to go a redirect with an url param

### Internal improvements

#### Groups

- Add the Grouper interface
  - func GetGroup() []interface{}
  - Add the NoneNodeElementsGroup

#### Page Pipeline

- HTTP request w, r

#### Other

- Batch message sends and receives in the javascript (https://developer.mozilla.org/en-US/docs/Web/API/HTML_DOM_API/Microtask_guide)
- If we want to batch sends from the server I think be a problem with out-of-order changes.
- Add log level to client side logging
- Send config for debug, and log level down to client side
- Remove the data- prefix from my attributes?
  - No one else seems to care
- Add a queue for incoming messages on a page session
  - Maybe multiple concurrent requests is okay, maybe we just batch renders?
- Use a channel with a single reader to process page events
- How does it work with grammarly?
- How does it work with Last Pass?
- Get rid of the `Page.treeLock` by using a channel?

### Tests

- Add JavaScript test for multi child delete
    - E.g.:
      - d|d|doc|1>1>0>0>1>2||
      - d|d|doc|1>1>0>0>1>3||
- `Page.Close`

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

### Security

- Add a CSRF token
  - https://github.com/gorilla/csrf

### New Features/Improvements

- Look for a CSS class to show on a failed reconnect
- Allow adding mount and unmount function as elements?
- How well does [Alpine.JS](https://alpinejs.dev/) work with HLive
  - https://dockyard.com/blogs/optimizing-user-experience-with-liveview

- Can we add a way for a component like List to inform tree copy not to bother doing a diff and just do a full HTML replacement
- Cluster
  - Proxy reconnect, find the session on another node and proxy the connection to that node
    - https://github.com/koding/websocketproxy
- Message bus
  - Cluster level
  - Plugin based protocol
  - Simple example protocol
    - Websocket?
    - manual node config
  - Prevent race conditions on the page dom
- ComponentList
  - Operations by ID
    - Get by ID
    - Remove By ID
- User friendly HTTP error pages
  - Display a request ID if it exits
- Map component ids and event binding ids to a sequential number
  - Would need to map out and in
  - Would provide for smaller ids
  - Could make debugging hard

- Add can take a `func() string` this would be kept in the tree and re-run on each render
  - Could be expensive

## HHot

A highly opinionated web framework that use hot reload and code generation.

- You install the hhot binary and when in dev mode you get a web UI to create new things, change config etc.
- Place models in certain folder, and we'll generate the code to wire them up. Same with HLive Pages
- Use a DI registry to pass config, logger, DB, Cache, etc. to each page.
- Use off-the-shelf ORM where possible.
- Form -> Model -> Database flow
- Data grid
- Pagination
- Navigation
- User management
- Permissions
- Batteries included but swappable

### HHot Ideas
- hhot-create-app
- Add the HTTP request to the context by default?
- Create a middleware package
  - HTTP Request
  - Security
- Add HTML dsl (Kit?)
- Cluster
  - User management, who's online
- Page level data store
  - `map[string]interface{}`
  - `map[string]Adder`
  - Available in the context
  - Mutex lock
  - Middleware?
- Asset management
  - Allow plugins/component libraries to work with this
    - Developer would connect them
  - Use an api to add the JS and CSS to the page HTML
    - Cache busting
  - Watch for change?
    - If change update cache buster
  - Development mode, production mode
- CSS Build Pipeline
  - Crete an example production quality CSS build pipeline using Tailwind and esbuild that can purse the unused CSS.
  - Don't use dynamic css class names
    - https://tailwindcss.com/docs/optimizing-for-production
- Forms
  - Make having forms easy
  - Full form validation
  - Use message bus
- Form -> Model -> Database flow
- Automatic TLS via Let’s Encrypt
- Logging

### Older ideas

- Limit execution by having a worker pool controlled by the page session
  - That way we have a way to limit RAM and CPU

#### Serializable tree/component state

** Maybe this should just be done with traditional user sessions? **

- Can we make it so that tree/component state is, serializable?
- If so we can store it in a database and allow reconnecting later.
- If load was getting high we can swap sessions in and out of memory
- We would need a way to capture and store then rehydrate
- How do we remount event listeners?
  - Need to map state to a function that can rehydrate it
    - Needs to support cold start (we have the data, but we've never seen the component)
  - Maybe we would need to use reflection
    - Only the first time 
- Can't have pointers, outside the components state
- Maybe all event driven?
- Encrypt the data by default
- Could this be a special kind of Page Session?

#### Do more with the HTTP Server render

- Think server side render
- Get the HTTP request easier
- Easy to cache the response

- CDN ready
  - Think about what would work well in a CDN
  - Product page
    - Good
      - Pics
      - Description
    - Bad
      - Personal recommendations
      - Recently viewed items


#### Multi file upload using WS and HTTP

- Need a count of files
- Group them together in an event?
- Make a channel?
- File upload progress

#### Visibility

- Is a component visible?
- Trigger event when visible?
- Scroll events
  - Page position
  - Viewport

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
