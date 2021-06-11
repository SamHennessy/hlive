package hlive

import (
	"bytes"
	"context"
	_ "embed"
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

//go:embed page.js
var PageJavaScript []byte

func NewPage() *Page {
	p := &Page{
		Renderer: NewRender(),
		Differ:   NewDiffer(),
		Logger:   zerolog.Nop(),

		DocType: HTML5DocType,
		HTML:    T("html", Attrs{"lang": "en"}),
		Head:    T("head"),
		Meta:    T("meta", Attrs{"charset": "utf-8"}),
		Title:   T("title"),
		Body:    T("body"),

		currentBindings: map[string]*EventBinding{},
		mountables:      map[string]struct{}{},
		unmountables:    map[string]ComponentUnmountInterface{},
	}

	p.Head.Add(p.Meta, p.Title, T("script", HTML(PageJavaScript)))
	p.HTML.Add(p.Head, p.Body)

	p.Renderer.Logger = p.Logger
	p.Differ.Logger = p.Logger

	return p
}

type Page struct {
	Upgrader websocket.Upgrader
	Renderer *Renderer
	Differ   *Differ
	Logger   zerolog.Logger

	DocType HTML
	HTML    *Tag
	Head    *Tag
	Meta    *Tag
	Title   *Tag
	Body    *Tag

	tree      interface{}
	treeLock  sync.RWMutex
	weConn    *websocket.Conn
	connected bool

	currentBindings map[string]*EventBinding
	// These with both grow and never removed from when user removed them from the tree
	// TODO: add ways to allow removal
	mountables   map[string]struct{}
	unmountables map[string]ComponentUnmountInterface
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
		p.Logger.Err(err).Msg("render page")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

type wsMsg struct {
	Typ  string            `json:"t"`
	ID   string            `json:"i,omitempty"`
	Data map[string]string `json:"d,omitempty"`
}

func (p *Page) ServerWS(w http.ResponseWriter, r *http.Request) {
	ctx := SetIsWebSocket(r.Context())

	sess := PageSess(ctx)
	if sess == nil || sess.Page == nil || sess.ID == "" {
		p.Logger.Error().Msg("server ws: empty or invalid session")
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	sess.LastActive = time.Now()

	// Upgrade request to WebSocket
	var err error

	p.weConn, err = p.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.Logger.Err(err).Msg("ws upgrade")

		return
	}

	// Won't send message until this is true
	p.connected = true

	// Send session id, it maybe new or have changed
	p.wsSend("s|id|" + sess.ID)

	defer func() {
		sess.LastActive = time.Now()
		p.connected = false

		if err := p.weConn.Close(); err != nil {
			p.Logger.Err(err).Msg("ws conn close")
		} else {
			p.Logger.Trace().Msg("ws close")
		}
	}()

	// Do an initial render
	// The unmounted components will not have an id
	// Don't think we need a lock, but could be an issue if we have the same concurrent initial request
	// We need a static render
	p.tree, err = p.copyTree(SetIsNotWebSocket(ctx), p.Render(), false)
	if err != nil {
		p.Logger.Err(err).Msg("ws static render")
	}

	// Add render function to context
	ctx = context.WithValue(ctx, CtxRender, p.executeRenderWS)
	ctx = context.WithValue(ctx, CtxRenderComponent, p.renderComponentWS)

	// Do a dynamic render
	p.executeRenderWS(ctx)

	// Start to read messages
	for {
		// This blocks until we get a message
		mt, message, err := p.weConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				p.Logger.Err(err).Msg("ws read")
			}

			// Connection properly closed
			break
		}

		sess.LastActive = time.Now()

		if mt == websocket.BinaryMessage {
			p.Logger.Error().Msg("unexpected binary message")

			continue
		}

		p.Logger.Trace().Str("msg", string(message)).Msg("ws msg recv")

		msg := wsMsg{Data: map[string]string{}}
		if err := json.Unmarshal(message, &msg); err != nil {
			p.Logger.Err(err).Str("json", string(message)).Msg("ws msg unmarshal")

			continue
		}

		switch msg.Typ {
		// Logger
		case "l":
			p.Logger.Info().Str("log", msg.Data["m"]).Msg("ws log")
		// Event
		case "e":
			// Call handler
			p.processMsgEvent(ctx, msg)
		default:
			p.Logger.Error().Str("msg", string(message)).Msg("ws msg recv: unexpected message format")
		}
	}
}

func (p *Page) executeRenderWS(ctx context.Context) {
	// Do a dynamic render
	diffs, err := p.RenderWS(ctx)
	if err != nil {
		p.Logger.Err(err).Msg("ws render")
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

	p.Logger.Trace().Str("msg", message).Msg("ws send")
	if err := p.weConn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		p.Logger.Err(err).Msg("ws write")
	}
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
			el = *diff.Text
			message += "t|"
		} else if diff.HTML != nil {
			el = *diff.HTML
			message += "h|"
		} else if diff.Attribute != nil {
			message += "a|"

			if err := p.Renderer.Attribute([]*Attribute{diff.Attribute}, bb); err != nil {
				p.Logger.Err(err).Msg("diffs to msg: render attribute")
			}
		} else if diff.Tag != nil {
			el = diff.Tag
			message += "h|"
		}

		if err := p.Renderer.HTML(bb, el); err != nil {
			p.Logger.Err(err).Msg("diffs to msg: render children")
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
	}

	ids := strings.Split(msg.ID, ",")
	for i := 0; i < len(ids); i++ {
		id := ids[i]

		p.Logger.Trace().Str("id", id).Msg("call event handler")
		// TODO: Maybe maintain a map of bindings in the current tree
		//       Add and remove binding on each render
		// binding := GetEventBindingFromTree(id, p.Render())
		binding := p.currentBindings[id]

		if binding == nil {
			p.Logger.Error().Str("id", id).Msg("unable to find binding")

			return
		}

		e.Binding = binding

		if binding.Handler == nil {
			p.Logger.Error().Str("id", id).Msg("unable to find binding handler")

			return
		}

		binding.Handler(ctx, e)

		// Render?
		if binding.Component.IsAutoRender() {
			p.executeRenderWS(ctx)
		}
	}
}

func (p *Page) RenderWS(ctx context.Context) ([]Diff, error) {
	p.treeLock.Lock()
	defer p.treeLock.Unlock()

	p.currentBindings = map[string]*EventBinding{}

	newTree, err := p.CopyTree(ctx, p.Render())
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

// TODO: this was really hard to write, gonna leave it here for now, just in case
// func (p *Page) TreeLifecycle(ctx context.Context, diffs ...Diff) {
// 	var comps []string
// 	compC := map[string]interface{}{}
// 	compD := map[string]interface{}{}
//
// 	for i := 0; i < len(diffs); i++ {
// 		d := diffs[i]
// 		switch d.Type {
// 		case DiffCreate:
// 			if d.Attribute != nil && d.Attribute.name == AttrID && d.Attribute.Value != nil {
// 				comps = append(comps, *d.Attribute.Value)
// 				compC[*d.Attribute.Value] = nil
// 			}
//
// 			if d.Tag != nil {
// 				ids := getCompIDsFromTag(d.Tag)
// 				for j := 0; j < len(ids); j++ {
// 					comps = append(comps, ids[j])
// 					compC[ids[j]] = nil
// 				}
// 			}
// 		case DiffUpdate:
// 			// New Part
// 			if d.Attribute != nil && d.Attribute.name == AttrID && d.Attribute.Value != nil {
// 				comps = append(comps, *d.Attribute.Value)
// 				compC[*d.Attribute.Value] = nil
// 			}
// 			// Old Part
// 			oldAttr, ok := d.Old.(Attribute)
// 			if ok && oldAttr.name == AttrID && oldAttr.Value != nil {
// 				comps = append(comps, *oldAttr.Value)
// 				compD[*oldAttr.Value] = nil
// 			}
//
// 			// Tags are not updated
// 		case DiffDelete:
// 			if d.Attribute != nil {
// 				if d.Attribute.name == AttrID && d.Attribute.Value != nil {
// 					comps = append(comps, *d.Attribute.Value)
// 					compD[*d.Attribute.Value] = nil
// 				}
// 			} else if d.Tag != nil {
// 				ids := getCompIDsFromTag(d.Tag)
// 				for j := 0; j < len(ids); j++ {
// 					comps = append(comps, ids[j])
// 					compD[ids[j]] = nil
// 				}
// 			} else if d.Old != nil {
// 				switch v := d.Old.(type) {
// 				case TagInterface:
// 					ids := getCompIDsFromTag(v)
// 					for j := 0; j < len(ids); j++ {
// 						comps = append(comps, ids[j])
// 						compD[ids[j]] = nil
// 					}
// 				}
// 			}
// 		}
// 	}
//
// 	// There could be duplicates in comps
// 	for i := 0; i < len(comps); i++ {
// 		id := comps[i]
// 		_, ce := compC[id]
// 		_, de := compD[id]
//
// 		// Only delete
// 		if !ce && de {
// 			comp, exits := p.unmountables[id]
// 			if exits {
// 				delete(p.unmountables, id)
// 				if err := comp.Unmount(ctx); err != nil {
// 					p.Logger.Err(err).Msg("unmount component")
// 				}
// 			}
// 		}
// 	}
// }

func (p *Page) Render() []interface{} {
	return Tree(p.DocType, p.HTML)
}

func (p *Page) RenderHTML(ctx context.Context, w io.Writer) error {
	var err error

	p.treeLock.Lock()
	defer p.treeLock.Unlock()

	p.tree, err = p.copyTree(ctx, p.Render(), false)
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
	case HTML:
		return v, nil
	case TagInterface:
		if v == nil {
			return nil, nil
		}

		if lifeCycle {
			// Needs a lock higher up to be thread safe
			if comp, ok := v.(ComponentInterface); ok {
				bindings := comp.GetEventBindings()
				for i := 0; i < len(bindings); i++ {
					p.currentBindings[bindings[i].ID] = bindings[i]
				}
			}

			if comp, ok := v.(ComponentMountInterface); ok {
				if _, exists := p.mountables[comp.GetID()]; !exists {
					comp.Mount(ctx)
					p.mountables[comp.GetID()] = struct{}{}
				}
			}

			if compU, ok := v.(ComponentUnmountInterface); ok {
				if _, exists := p.unmountables[compU.GetID()]; !exists {
					p.unmountables[compU.GetID()] = compU
				}
			}
		}

		kids, err := p.copyTree(ctx, v.Render(), lifeCycle)
		if err != nil {
			return nil, fmt.Errorf("copy tree on tag children: %s: %w", v.GetName(), err)
		}

		// Call Render before GetAttributes so users can set attributes in Render
		var els []interface{}
		oldAttrs := v.GetAttributes()
		var attrs []*Attribute

		for i := 0; i < len(oldAttrs); i++ {
			attrs = append(attrs, oldAttrs[i].Clone())
		}

		// Strip hlive attributes to trigger diff
		if !IsWebSocket(ctx) {
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
		var newTree []interface{}
		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []ComponentInterface:
		var newTree []interface{}
		for i := 0; i < len(v); i++ {
			node, err := p.copyTree(ctx, v[i], lifeCycle)
			if err != nil {
				return nil, err
			}

			newTree = append(newTree, node)
		}

		return newTree, nil
	case []TagInterface:
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

func IsWebSocket(ctx context.Context) bool {
	isWebSocket, _ := ctx.Value(CtxIsWS).(bool)

	return isWebSocket
}

func SetIsWebSocket(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxIsWS, true)
}

func SetIsNotWebSocket(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxIsWS, false)
}

func RenderWS(ctx context.Context) {
	render, ok := ctx.Value(CtxRender).(func(context.Context))

	if !ok {
		panic("ERROR: RenderWS not found in context")
	}

	render(ctx)
}

func RenderComponentWS(ctx context.Context, comp ComponentInterface) {
	render, ok := ctx.Value(CtxRenderComponent).(func(context.Context, ComponentInterface))
	if !ok {
		panic("ERROR: RenderComponentWS not found in context")
	}

	render(ctx, comp)
}

func (p *Page) renderComponentWS(ctx context.Context, comp ComponentInterface) {
	if p == nil {
		p.Logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component on dead page")

		return
	}

	oldTag := p.findComponentInTree(comp.GetID())
	if oldTag == nil {
		p.Logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component ws: can't find component in tree")

		return
	}

	newTreeNode, err := p.copyTree(ctx, comp, true)
	if err != nil {
		p.Logger.Err(err).Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component ws: copy tree")

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
		p.Logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("render component ws: new node is not a tag, component, or group")

		return
	}

	diffs, err := p.Differ.Trees(comp.GetID(), "", oldTag, newTag)
	if err != nil {
		p.Logger.Err(err).Str("id", comp.GetID()).Str("name", comp.GetName()).Msg("diff trees")

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
