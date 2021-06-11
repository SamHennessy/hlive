# hlive

A server-side WebSocket based dynamic template-less view layer.

## Concepts

- Rendering should not make calls to a database or API
- Rendering with the data same should produce the same result

- Need to reconcile virtual dom in Go with what's been rendered in the Browser

- URL and navigation
    - Maybe not part of view layer
    - Maybe a component?
        - But we must have a way to hook in and allow such a low level feature 

- HTML must have 1 root node
- Browsers don't view Text as a tree node. Don't Mix text and tags as siblings.

- Event
    - Type
        - Click/Select
        - ClickBool
          - Has a built-in bool so no need for a callback
        - Text input
        - Blur
        - Hover?
          - With a careful race condition prevention
          - have a global hover counter
          - What if nested elements both have a hover action, maybe it's fine
        - Gestures
        - Touch
        - History
          - On page only?
        - Keyboard
          - On page
        - Animation
          - When an animation starts and stops
          - Not sure if possible with 3rd party animation lib
    - Callback
        - Optional for text input
        - An event struct is passed to the callback
    
### Lifecycle

- HTTP Request
    - Does a throw away render
        - It's assumed that when running in a cluster the server doing this render and the server that will establish the WebSocket connection will not always be the same 
    - If a component has it Mount will not be called 
    - You can use middleware to establish an HTTP session (See session example)
- Client JavaScript
    - Connects to same domain and path but with `ws=1` appended. No params will be passed
        - If there is demand we could also send params
    - Requests a new PageSession
        - A session id is used to try to reconnect if there is a disconnect
        - Reconnection will only work with the original server so for clusters you need to plan accordingly
    

## TODO:

- Store to database
- add WS ping-pong
- WS reconnect or reload on fail   
- CSS animation chain, classes that will be in order, after the previous link triggers onanimationend
    - this is find of like giving focus but with focue you can turn it off using the onfocue event
    - On last step of the chain we can trigger a custom event
    - OnChainComplete, then use can remove the attribute
- Batch message sends and receives in the javascript (https://developer.mozilla.org/en-US/docs/Web/API/HTML_DOM_API/Microtask_guide)
- Add a queue for incoming messages on a page session
   - Maybe multiple concurrent requets is okay, maybe we just batch renders, 
- Allow the HLive JavaScript to be replaced for a single page
- Have non-Live pages, only using the templating for static pages
- What to do about when a page is reloaded, and the forms are prefilled with data?
- Create an interface that allows it to be given a close funcion that wil remove it from the mountables and unmountables maps
- Add JavaScript test for multi child delete
    - E.g.:
    - d|d|doc|1>1>0>0>1>2||
    - d|d|doc|1>1>0>0>1>3||
- Remove event listener, e.g. delete component with an onmouseleave while still hovering 

## Inspiration

### gomponents

For it's HTML API

https://github.com/maragudk/gomponents
