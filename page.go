package hlive

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

type Page struct {
	// Websocket upgrader
	Upgrader websocket.Upgrader
	// Session is this Page value's session
	Session *PageSession
	// HTML renderer
	Renderer *Renderer
	// DOM differ
	Differ *Differ
	// Root DOM elements
	DocType HTML
	HTML    Adder
	Head    Adder
	Meta    Adder
	Title   Adder
	Body    Adder
	// Page rendering pipelines
	PipelineDiff *Pipeline
	// Page HTML rendering pipeline
	PipelineRender *Pipeline
	// Internal debug logger
	logger zerolog.Logger
	// What is in the browser
	browserTree interface{}
	// Lock the page for writes
	// TODO: replace with channel
	pageLock  sync.RWMutex
	weConn    *websocket.Conn
	connected bool
	// Component caches, to prevent walking to tree to find something
	eventBindings map[string]*EventBinding
	// Buffered channel of outbound messages.
	send chan []byte
	//
	attributePluginMountedMap map[string]struct{}
	//
	HookBeforeEvent []func(ctx context.Context, e Event) (context.Context, Event)
	HookAfterEvent  []func(ctx context.Context, e Event) (context.Context, Event)
	HookAfterRender []func(context.Context, []Diff, chan<- []byte)
	HookClose       []func(context.Context, *Page)
}

func NewPage() *Page {
	p := &Page{
		Renderer: NewRender(),
		Differ:   NewDiffer(),
		logger:   zerolog.Nop(),

		DocType: HTML5DocType,
		HTML:    T("html", Attrs{"lang": "en"}),
		Head:    T("head"),
		Meta:    T("meta", Attrs{"charset": "utf-8"}),
		Title:   T("title"),
		Body:    T("body"),

		eventBindings: map[string]*EventBinding{},

		// TODO: remove buffer?
		// If I don't want to block on a ws send use a go routine?
		send: make(chan []byte, 256),

		attributePluginMountedMap: map[string]struct{}{},
	}

	p.Head.Add(p.Meta, p.Title, T("script", HTML(p.Differ.JavaScript)))
	p.HTML.Add(p.Head, p.Body)

	// Differ Pipeline
	p.PipelineDiff = NewPipeline(
		PipelineProcessorEventBindingCache(p.eventBindings),
		PipelineProcessorMount(),
		PipelineProcessorUnmount(p),
		PipelineProcessorAttributePluginMount(p),
		PipelineProcessorConvertToString(),
	)
	// Render Pipeline
	p.PipelineRender = NewPipeline(
		PipelineProcessorAttributePluginMount(p),
		PipelineProcessorStripHLiveAttrs(),
		PipelineProcessorConvertToString(),
		PipelineProcessorRenderer(p.Renderer),
	)

	return p
}

type wsMsg struct {
	Typ        string            `json:"t"`
	ID         string            `json:"i,omitempty"`
	Data       map[string]string `json:"d,omitempty"`
	File       *File             `json:"file,omitempty"`
	ValueMulti []string          `json:"vm,omitempty"`
	Selected   bool              `json:"s,omitempty"`
	Extra      map[string]string `json:"e,omitempty"`
	fileData   []byte
}

func (p *Page) SetLogger(logger zerolog.Logger) {
	p.logger = logger
	p.Renderer.SetLogger(p.logger)
	p.Differ.SetLogger(p.logger)
}

func (p *Page) IsConnected() bool {
	return p.connected
}

func (p *Page) Close(ctx context.Context) {
	for i := 0; i < len(p.HookClose); i++ {
		p.HookClose[i](ctx, p)
	}
}

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := p.serverHTTP(w, r); err != nil {
		p.logger.Err(err).Msg("server http")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (p *Page) serverHTTP(w http.ResponseWriter, r *http.Request) error {
	p.pageLock.Lock()

	_, err := p.runRenderPipeline(r.Context(), w)

	p.pageLock.Unlock()

	return err
}

func (p *Page) runRenderPipeline(ctx context.Context, w io.Writer) (*NodeGroup, error) {
start:
	tree, err := p.PipelineRender.run(ctx, w, p.GetNodes())

	// A plugin being added will do this
	if errors.Is(err, ErrDOMInvalidated) {
		goto start
	}

	return tree, err
}

func (p *Page) runDiffPipeline(ctx context.Context, w io.Writer) (*NodeGroup, error) {
start:
	tree, err := p.PipelineDiff.run(ctx, w, p.GetNodes())

	// Plugins will do this
	if errors.Is(err, ErrDOMInvalidated) {
		goto start
	}

	return tree, err
}

func (p *Page) ServerWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess := PageSess(ctx)
	if sess == nil || sess.Page == nil || sess.ID == "" {
		p.logger.Error().Msg("server ws: empty or invalid session")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	p.Session = sess

	sess.LastActive = time.Now()

	// Upgrade request to WebSocket
	var err error

	p.weConn, err = p.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Err(err).Msg("ws upgrade")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	// Won't send message until this is true
	p.connected = true

	// Send session id, it may be new or have changed
	p.wsSend("s|id|" + sess.ID)

	// Do an initial render
	if p.browserTree == nil {
		p.logger.Trace().Msg("initial render")
		// We need a static render
		// p.browserTree, err = p.PipelineRender.run(ctx, io.Discard, p.GetNodes())
		p.browserTree, err = p.runRenderPipeline(r.Context(), io.Discard)
		if err != nil {
			p.logger.Err(err).Msg("ws static render: html pipeline")
		}
	}

	// Add render function to context
	ctx = context.WithValue(ctx, CtxRender, p.executeRenderWS)
	ctx = context.WithValue(ctx, CtxRenderComponent, p.renderComponentWS)

	// Do a dynamic render
	p.executeRenderWS(ctx)

	// Start to read messages
	go p.readPump(ctx, sess)
	go p.writePump()
}

func (p *Page) executeRenderWS(ctx context.Context) {
	// Do a dynamic render
	diffs, err := p.RenderWS(ctx)
	if err != nil {
		p.logger.Err(err).Msg("ws render")
	}
	// Any DOM updates?

	if len(diffs) != 0 {
		p.wsSend(p.diffsToMsg(diffs))
	}

	for i := 0; i < len(p.HookAfterRender); i++ {
		p.HookAfterRender[i](ctx, diffs, p.send)
	}
}

func (p *Page) wsSend(message string) {
	if p == nil || !p.IsConnected() {
		return
	}

	p.logger.Debug().Str("msg", message).Msg("ws send")

	p.send <- []byte(message)
}

// Create and deletes should only happen at the end of a tag or attr list?
func (p *Page) diffsToMsg(diffs []Diff) string {
	message := ""
	// Reverse order to help with offset changes when adding new nodes
	for i := 0; i < len(diffs); i++ {
		diff := diffs[i]
		// Diff
		message += "d|"
		message += string(diff.Type) + "|"
		message += diff.Root + "|"
		message += diff.Path + "|"

		bb := bytes.NewBuffer(nil)

		var el interface{}

		if diff.Type == DiffDelete && diff.Attribute == nil {
			message += "|"
		} else if diff.Text != nil {
			// We use node.textContent in the browser which doesn't require us to encode
			el = HTML(*diff.Text)
			message += "t|"
		} else if diff.HTML != nil {
			el = *diff.HTML
			message += "h|"
		} else if diff.Attribute != nil {
			message += "a|"

			if err := p.Renderer.Attribute([]Attributer{diff.Attribute}, bb); err != nil {
				p.logger.Err(err).Msg("diffs to msg: render attribute")
			}
		} else if diff.Tag != nil {
			el = diff.Tag
			message += "h|"
		}

		if err := p.Renderer.HTML(bb, el); err != nil {
			p.logger.Err(err).Msg("diffs to msg: render children")
		}

		message += base64.StdEncoding.EncodeToString(bb.Bytes())

		message += "\n"
	}

	return message
}

func (p *Page) processMsgEvent(ctx context.Context, msg wsMsg) {
	keyCode, _ := strconv.Atoi(msg.Data["keyCode"])
	charCode, _ := strconv.Atoi(msg.Data["charCode"])
	shiftKey, _ := strconv.ParseBool(msg.Data["shiftKey"])
	altKey, _ := strconv.ParseBool(msg.Data["altKey"])
	ctrlKey, _ := strconv.ParseBool(msg.Data["ctrlKey"])
	isInitial, _ := strconv.ParseBool(msg.Data["init"])

	e := Event{
		Value:     msg.Data["value"],
		Values:    msg.ValueMulti,
		IsInitial: isInitial,
		Key:       msg.Data["key"],
		KeyCode:   keyCode,
		CharCode:  charCode,
		ShiftKey:  shiftKey,
		AltKey:    altKey,
		CtrlKey:   ctrlKey,
		File:      msg.File,
		Selected:  msg.Selected,
		Extra:     msg.Extra,
	}

	ids := strings.Split(msg.ID, ",")
	for i := 0; i < len(ids); i++ {
		id := ids[i]

		p.logger.Trace().Str("id", id).Msg("call event handler")

		binding := p.eventBindings[id]

		if binding == nil {
			p.logger.Error().Str("id", id).Msg("unable to find binding")

			return
		}

		e.Binding = binding

		if binding.Handler == nil {
			p.logger.Error().Str("id", id).Msg("unable to find binding handler")

			delete(p.eventBindings, id)

			return
		}

		// Hook
		for j := 0; j < len(p.HookBeforeEvent); j++ {
			ctx, e = p.HookBeforeEvent[j](ctx, e)
		}

		e.Binding.Handler(ctx, e)

		// Hook
		for j := 0; j < len(p.HookAfterEvent); j++ {
			ctx, e = p.HookAfterEvent[j](ctx, e)
		}

		// Once, do this after calling the handler so the developer can change their mind
		if e.Binding.Once {
			delete(p.eventBindings, id)
			binding.Component.RemoveEventBinding(id)
		}

		// Auto Render?
		if e.Binding.Component.IsAutoRender() {
			p.executeRenderWS(ctx)
		}
	}
}

func (p *Page) RenderWS(ctx context.Context) ([]Diff, error) {
	p.pageLock.Lock()
	defer p.pageLock.Unlock()

	// TODO: replace discard with something useful
	tree, err := p.runDiffPipeline(ctx, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("run pipeline: %w", err)
	}

	diffs, err := p.Differ.Trees("doc", "", p.browserTree, tree)
	if err != nil {
		return nil, fmt.Errorf("diff old and new tag trees: %w", err)
	}

	p.browserTree = tree

	return diffs, nil
}

func (p *Page) GetNodes() *NodeGroup {
	return G(p.DocType, p.HTML)
}

// Will fail if plugins invalidate the tree
func (p *Page) renderComponentWS(ctx context.Context, comp Componenter) {
	if p == nil {
		p.logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component on dead page")

		return
	}

	oldTag := p.findComponentInTree(comp.GetID())
	if oldTag == nil {
		p.logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).
			Msg("render component ws: can't find component in tree")

		return
	}

	// TODO: replace discard
	newTreeNode, err := p.PipelineDiff.runNode(ctx, io.Discard, comp)

	var newTag *Tag

	newTags, ok := newTreeNode.(*NodeGroup)
	if ok && len(newTags.Get()) != 0 {
		newTag, ok = newTags.Get()[0].(*Tag)
	} else {
		newTag, ok = newTreeNode.(*Tag)
	}

	if !ok {
		p.logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).
			Msg("render component ws: new node is not a tag, component, or group")

		return
	}

	diffs, err := p.Differ.Trees(comp.GetID(), "", oldTag, newTag)
	if err != nil {
		p.logger.Err(err).Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("diff trees")

		return
	}

	if len(diffs) != 0 {
		p.wsSend(p.diffsToMsg(diffs))

		oldTag.name = newTag.name
		oldTag.void = newTag.void
		oldTag.attributes = newTag.attributes
		oldTag.nodes = newTag.nodes
	}
}

func (p *Page) GetBrowserNodeByID(id string) *Tag {
	return p.findComponent(id, p.browserTree)
}

func (p *Page) findComponentInTree(id string) *Tag {
	return p.findComponent(id, p.browserTree)
}

func (p *Page) findComponent(id string, tree interface{}) *Tag {
	switch v := tree.(type) {
	case *NodeGroup:
		g := v.Get()
		for i := 0; i < len(g); i++ {
			c := p.findComponent(id, g[i])
			if c != nil {
				return c
			}
		}
	case *Tag:
		if v.GetAttributeValue(AttrID) == id {
			return v
		}

		return p.findComponent(id, v.nodes)
	}

	return nil
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (p *Page) readPump(ctx context.Context, sess *PageSession) {
	defer func() {
		sess.LastActive = time.Now()
		p.connected = false

		if err := p.weConn.Close(); err != nil {
			p.logger.Err(err).Msg("ws conn close")
		} else {
			p.logger.Trace().Msg("ws close")
		}
	}()

	// c.conn.SetReadLimit(maxMessageSize)
	if err := p.weConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		p.logger.Err(err).Msg("read pump set read deadline")
	}

	p.weConn.SetPongHandler(func(string) error {
		p.logger.Trace().Msg("ws pong")

		if err := p.weConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			p.logger.Err(err).Msg("pong handler: set read deadline")
		}

		return nil
	})

	for {
		mt, message, err := p.weConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				p.logger.Debug().Err(err).Msg("unexpected close error")
			}

			break
		}

		sess.LastActive = time.Now()

		msg := wsMsg{Data: map[string]string{}}

		if mt == websocket.BinaryMessage {
			msgParts := bytes.SplitN(message, []byte("\n\n"), 2)

			if len(msgParts) != 2 {
				p.logger.Error().Msg("invalid binary message")

				continue
			}

			message = msgParts[0]
			msg.fileData = msgParts[1]
		}

		p.logger.Debug().Str("msg", string(message)).Msg("ws msg recv")

		if err := json.Unmarshal(message, &msg); err != nil {
			p.logger.Err(err).Str("json", string(message)).Msg("ws msg unmarshal")

			continue
		}

		switch msg.Typ {
		// logger
		case "l":
			p.logger.Info().Str("log", msg.Data["m"]).Str("sess", sess.ID).Msg("ws log")
		// Event
		case "e":
			if len(msg.fileData) != 0 && msg.File != nil {
				msg.File.Data = msg.fileData
			}

			// Call handler
			p.processMsgEvent(ctx, msg)
		default:
			p.logger.Error().Str("msg", string(message)).Msg("ws msg recv: unexpected message format")
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (p *Page) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()

		if err := p.weConn.Close(); err != nil {
			p.logger.Trace().Err(err).Msg("write pump: close ws connection")
		}
	}()

	for {
		select {
		case message, ok := <-p.send:
			if err := p.weConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				p.logger.Err(err).Msg("write pump: message set write deadline")
			}

			if !ok {
				// Send channel closed.
				if err := p.weConn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					p.logger.Err(err).Msg("write pump: write close message")
				}

				return
			}

			w, err := p.weConn.NextWriter(websocket.TextMessage)
			if err != nil {
				p.logger.Err(err).Msg("write pump: create writer")

				return
			}

			if _, err := w.Write(message); err != nil {
				p.logger.Err(err).Msg("write pump: write first message")
			}

			// Add queued chat messages to the current websocket message.
			n := len(p.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					p.logger.Err(err).Msg("write pump: write queued message")
				}

				if _, err := w.Write(<-p.send); err != nil {
					p.logger.Err(err).Msg("write pump: write message delimiter")
				}
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			p.logger.Trace().Msg("ws ping")

			if err := p.weConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				p.logger.Err(err).Msg("write pump: ping tick: set write deadline")
			}

			if err := p.weConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
