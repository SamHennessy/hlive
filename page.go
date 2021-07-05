package hlive

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

		currentBindings: map[string]*EventBinding{},
		mountables:      map[string]struct{}{},
		unmountables:    map[string]Unmounter{},

		send: make(chan []byte, 256),
	}

	p.Head.Add(p.Meta, p.Title, T("script", HTML(p.Differ.JavaScript)))
	p.HTML.Add(p.Head, p.Body)

	p.Renderer.logger = p.logger
	p.Differ.logger = p.logger

	return p
}

type Page struct {
	Upgrader websocket.Upgrader
	Renderer *Renderer
	Differ   *Differ
	logger   zerolog.Logger

	DocType HTML
	HTML    *Tag
	Head    *Tag
	Meta    *Tag
	Title   *Tag
	Body    *Tag

	tree            interface{}
	treeLock        sync.RWMutex
	weConn          *websocket.Conn
	connected       bool
	currentBindings map[string]*EventBinding
	mountables      map[string]struct{}
	unmountables    map[string]Unmounter

	// Buffered channel of outbound messages.
	send chan []byte
}

func (p *Page) SetLogger(logger zerolog.Logger) {
	p.logger = logger
	p.Renderer.SetLogger(p.logger)
	p.Differ.SetLogger(p.logger)
}

func (p *Page) IsConnected() bool {
	return p.connected
}

// TODO: write tests
func (p *Page) Close(ctx context.Context) {
	for _, c := range p.unmountables {
		if c == nil {
			continue
		}

		c.Unmount(ctx)
	}
}

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := p.RenderHTML(r.Context(), w); err != nil {
		p.logger.Err(err).Msg("render page")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type wsMsg struct {
	Typ      string            `json:"t"`
	ID       string            `json:"i,omitempty"`
	Data     map[string]string `json:"d,omitempty"`
	File     *File             `json:"file,omitempty"`
	fileData []byte
}

func (p *Page) ServerWS(w http.ResponseWriter, r *http.Request) {
	ctx := SetIsWebSocket(r.Context())

	sess := PageSess(ctx)
	if sess == nil || sess.Page == nil || sess.ID == "" {
		p.logger.Error().Msg("server ws: empty or invalid session")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

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

	// Send session id, it maybe new or have changed
	p.wsSend("s|id|" + sess.ID)

	// Do an initial render
	// The unmounted components will not have an id
	// Don't think we need a lock, but could be an issue if we have the same concurrent initial request
	// We need a static render
	p.tree, err = p.copyTree(setIsNotWebSocket(ctx), p.GetNodes(), false)
	if err != nil {
		p.logger.Err(err).Msg("ws static render")
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
}

func (p *Page) wsSend(message string) {
	if p == nil || !p.IsConnected() {
		return
	}

	p.logger.Trace().Str("msg", message).Msg("ws send")

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

			if err := p.Renderer.Attribute([]*Attribute{diff.Attribute}, bb); err != nil {
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

	e := Event{
		Value:    msg.Data["value"],
		Key:      msg.Data["key"],
		KeyCode:  keyCode,
		CharCode: charCode,
		ShiftKey: shiftKey,
		AltKey:   altKey,
		CtrlKey:  ctrlKey,
		File:     msg.File,
	}

	ids := strings.Split(msg.ID, ",")
	for i := 0; i < len(ids); i++ {
		id := ids[i]

		p.logger.Trace().Str("id", id).Msg("call event handler")

		binding := p.currentBindings[id]

		if binding == nil {
			p.logger.Error().Str("id", id).Msg("unable to find binding")

			return
		}

		e.Binding = binding

		if binding.Handler == nil {
			p.logger.Error().Str("id", id).Msg("unable to find binding handler")

			delete(p.currentBindings, id)

			return
		}

		binding.Handler(ctx, e)

		// GetNodes?
		if binding.Component.IsAutoRender() {
			p.executeRenderWS(ctx)
		}

		// Once, do this after calling the handler so the developer can turn off once if they want
		if binding.Once {
			delete(p.currentBindings, id)
			binding.Component.RemoveEventBinding(id)
		}
	}
}

func (p *Page) RenderWS(ctx context.Context) ([]Diff, error) {
	p.treeLock.Lock()
	defer p.treeLock.Unlock()

	p.currentBindings = map[string]*EventBinding{}

	newTree, err := p.CopyTree(ctx, p.GetNodes())
	if err != nil {
		return nil, err
	}

	diffs, err := p.Differ.Trees("doc", "", p.tree, newTree)
	if err != nil {
		return nil, fmt.Errorf("diff old and new tag trees: %w", err)
	}

	p.tree = newTree

	return diffs, nil
}

func (p *Page) GetNodes() []interface{} {
	return Tree(p.DocType, p.HTML)
}

func (p *Page) RenderHTML(ctx context.Context, w io.Writer) error {
	var err error

	p.treeLock.Lock()
	defer p.treeLock.Unlock()

	p.tree, err = p.copyTree(ctx, p.GetNodes(), false)
	if err != nil {
		return err
	}

	return p.Renderer.HTML(w, p.tree)
}

// CopyTree Simplify the tree and create a copy
func (p *Page) CopyTree(ctx context.Context, oldTree interface{}) (interface{}, error) {
	return p.copyTree(ctx, oldTree, true)
}

func (p *Page) copyTree(ctx context.Context, oldTree interface{}, lifeCycle bool) (interface{}, error) {
	switch v := oldTree.(type) {
	case nil:
		return nil, nil
	case *string:
		if v == nil || *v == "" {
			return nil, nil
		}

		return *v, nil
	case string:
		if v == "" {
			return nil, nil
		}
		return v, nil
	case *int: // TODO: *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64
		if v == nil {
			return nil, nil
		}
		return strconv.Itoa(*v), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	// TODO: stop using printf
	case int8, int16, int32, uint, uint8, uint16, uint32:
		return fmt.Sprintf("%v", v), nil
	case *HTML:
		if v == nil || *v == "" {
			return nil, nil
		}

		return *v, nil
	case HTML:
		return v, nil
	case Tagger:
		if v == nil {
			return nil, nil
		}

		if lifeCycle {
			// Needs a lock higher up to be thread safe
			if comp, ok := v.(Componenter); ok {
				bindings := comp.GetEventBindings()
				for i := 0; i < len(bindings); i++ {
					p.currentBindings[bindings[i].ID] = bindings[i]
				}
			}

			if comp, ok := v.(Mounter); ok {
				if _, exists := p.mountables[comp.GetID()]; !exists {
					comp.Mount(ctx)

					p.mountables[comp.GetID()] = struct{}{}
				}
			}

			if compU, ok := v.(Unmounter); ok {
				if _, exists := p.unmountables[compU.GetID()]; !exists {
					p.unmountables[compU.GetID()] = compU
				}
			}

			if compTr, ok := v.(Teardowner); ok {
				compTr.SetTeardown(func() {
					delete(p.mountables, compTr.GetID())
					delete(p.unmountables, compTr.GetID())
				})
			}
		}

		kids, err := p.copyTree(ctx, v.GetNodes(), lifeCycle)

		if err != nil {
			return nil, fmt.Errorf("copy tree on tag children: %s: %w", v.GetName(), err)
		}

		// Call GetNodes before GetAttributes so users can set attributes in GetNodes
		var (
			els      []interface{}
			attrs    []*Attribute
			oldAttrs = v.GetAttributes()
		)

		for i := 0; i < len(oldAttrs); i++ {
			attrs = append(attrs, oldAttrs[i].Clone())
		}

		// Strip hlive attributes to trigger diff
		if !isWebSocket(ctx) {
			var newAttrs []*Attribute

			for i := 0; i < len(attrs); i++ {
				if strings.HasPrefix(attrs[i].Name, "data-hlive") {
					continue
				}

				newAttrs = append(newAttrs, attrs[i].Clone())
			}

			attrs = newAttrs
		}

		els = append(els, attrs)

		return T(v.GetName(), append(els, kids)), nil
	case RenderFunc:
		var newTree []interface{}

		nodes := v()
		for i := 0; i < len(nodes); i++ {
			node, err := p.copyTree(ctx, nodes[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []interface{}:
		var (
			newTree       []interface{}
			thisNodeStr   string
			thisNodeIsStr bool
			lastNodeStr   string
			lastNodeIsStr bool
		)

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			thisNodeStr, thisNodeIsStr = v[i].(string)

			// Combine strings like a browser would
			if lastNodeIsStr && thisNodeIsStr && len(newTree) > 0 {
				// replace last node
				newTree[len(newTree)-1] = lastNodeStr + thisNodeStr
			} else {
				newTree = append(newTree, node)
			}

			lastNodeStr, lastNodeIsStr = thisNodeStr, thisNodeIsStr
		}

		return newTree, nil
	case []Componenter:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Tagger:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Component:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []*Component:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []Tag:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []*Tag:
		var newTree []interface{}

		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	default:
		return nil, ErrRenderElement
	}
}

func isWebSocket(ctx context.Context) bool {
	is, _ := ctx.Value(CtxIsWS).(bool)

	return is
}

func SetIsWebSocket(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxIsWS, true)
}

func setIsNotWebSocket(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxIsWS, false)
}

func RenderWS(ctx context.Context) {
	render, ok := ctx.Value(CtxRender).(func(context.Context))

	if !ok {
		panic("ERROR: RenderWS not found in context")
	}

	render(ctx)
}

func RenderComponentWS(ctx context.Context, comp Componenter) {
	render, ok := ctx.Value(CtxRenderComponent).(func(context.Context, Componenter))
	if !ok {
		panic("ERROR: RenderComponentWS not found in context")
	}

	render(ctx, comp)
}

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

	newTreeNode, err := p.copyTree(ctx, comp, true)
	if err != nil {
		p.logger.Err(err).Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component ws: copy tree")

		return
	}

	var newTag *Tag

	newTags, ok := newTreeNode.([]interface{})
	if ok && len(newTags) != 0 {
		newTag, ok = newTags[0].(*Tag)
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

		// TODO: move to Tag
		oldTag.name = newTag.name
		oldTag.void = newTag.void
		oldTag.attributes = newTag.attributes
		oldTag.nodes = newTag.nodes
	}
}

func (p *Page) findComponentInTree(id string) *Tag {
	return p.findComponent(id, p.tree)
}

func (p *Page) findComponent(id string, tree interface{}) *Tag {
	switch v := tree.(type) {
	case []interface{}:
		for i := 0; i < len(v); i++ {
			c := p.findComponent(id, v[i])
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
				p.logger.Err(err).Msg("unexpected close error")
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
			p.logger.Info().Str("log", msg.Data["m"]).Msg("ws log")
		// Event
		case "e":
			// Call handler
			if len(msg.fileData) != 0 && msg.File != nil {
				msg.File.Data = msg.fileData
			}

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
			p.logger.Err(err).Msg("write pump: close ws connection")
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
