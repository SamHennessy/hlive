package hlive

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/cornelk/hashmap"
	"github.com/rs/zerolog"
)

type Page struct {
	// HTML renderer
	Renderer *Renderer
	// DOM differ
	Differ *Differ
	// Page rendering pipelines
	PipelineDiff *Pipeline
	// Page HTML rendering pipeline
	PipelineSSR *Pipeline
	// Internal debug logger
	logger zerolog.Logger
	// Virtual DOM
	DOM DOM
	// What we think is the browser done is now
	DOMBrowser interface{}
	// Lock the page for writes
	// TODO: replace with channel?
	pageLock sync.RWMutex
	// weConn    *websocket.Conn
	connected bool
	// sessID is the WebSocket connection session id
	// only one page will have this at a time but is can be passed from page to page if connection is kept open
	sessID string
	// Component caches, to prevent walking to tree to find something
	eventBindings *hashmap.HashMap
	// Buffered channel of outbound messages.
	send chan<- MessageWS
	// Buffered channel of inbound messages.
	receive <-chan MessageWS
	// done with the main loop
	done chan bool
	//
	// Hooks
	//
	// Before each event
	HookBeforeEvent []func(ctx context.Context, e Event) (context.Context, Event)
	// After each event
	HookAfterEvent []func(ctx context.Context, e Event) (context.Context, Event)
	// After each render
	HookAfterRender []func(context.Context, []Diff, chan<- MessageWS)
	HookClose       []func(context.Context, *Page)
	// Before we do the initial render and send to the browser
	HookBeforeMount []func(context.Context, *Page)
	// After we do the initial render and send to the browser
	HookMount []func(context.Context, *Page)
	// When we close the page
	HookUnmount []func(context.Context, *Page)
}

func NewPage() *Page {
	p := &Page{
		Renderer:      NewRenderer(),
		Differ:        NewDiffer(),
		DOM:           *NewDOM(),
		logger:        zerolog.Nop(),
		eventBindings: hashmap.New(10), // No real reason for 10
	}

	p.DOM.Head.Add(T("script", HTML(p.Differ.JavaScript)))

	// Differ Pipeline
	p.PipelineDiff = NewPipeline(
		PipelineProcessorAttributePluginMount(p),
		PipelineProcessorEventBindingCache(p.eventBindings),
		PipelineProcessorMount(),
		PipelineProcessorUnmount(p),
		PipelineProcessorConvertToString(),
	)
	// Server Side Render Pipeline
	p.PipelineSSR = NewPipeline(
		PipelineProcessorAttributePluginMountSSR(p),
		PipelineProcessorStripHLiveAttrs(),
		PipelineProcessorConvertToString(),
		PipelineProcessorRenderer(p.Renderer),
	)

	return p
}

type websocketMessage struct {
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

func (p *Page) GetSessionID() string {
	return p.sessID
}

func (p *Page) Close(ctx context.Context) {
	for i := 0; i < len(p.HookClose); i++ {
		p.HookClose[i](ctx, p)
	}

	if p.connected {
		p.connected = false
	}

	if p.done != nil {
		p.done <- true
	}
}

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := p.serverHTTP(w, r); err != nil {
		p.logger.Err(err).Msg("server http")
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
	tree, err := p.PipelineSSR.run(ctx, w, p.GetNodes())

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

func (p *Page) ServeWS(ctx context.Context, sessID string, send chan<- MessageWS, receive <-chan MessageWS) error {
	p.send = send
	p.receive = receive
	p.done = make(chan bool)

	defer func() {
		if p.done != nil {
			close(p.done)
		}
		p.done = nil
	}()

	var err error

	p.connected = true

	// Send session id, it may be new or have changed
	p.sessID = sessID
	p.wsSend("s|id|" + sessID)

	// Do an initial render
	if p.DOMBrowser == nil {
		p.logger.Trace().Msg("initial render")
		// We need a static render
		// p.DOMBrowser, err = p.PipelineSSR.run(ctx, io.Discard, p.GetNodes())
		p.DOMBrowser, err = p.runRenderPipeline(ctx, io.Discard)
		if err != nil {
			p.logger.Err(err).Msg("ws static render: html pipeline")
		}
	}

	// Add render function to context
	ctx = context.WithValue(ctx, "Unique", strconv.Itoa(rand.Int()))
	ctx = context.WithValue(ctx, CtxRender, p.executeRenderWS)
	ctx = context.WithValue(ctx, CtxRenderComponent, p.renderComponentWS)

	// TODO: add tests
	for i := 0; i < len(p.HookBeforeMount); i++ {
		p.HookBeforeMount[i](ctx, p)
	}

	// Do a dynamic render
	p.executeRenderWS(ctx)

	// TODO: add tests
	for i := 0; i < len(p.HookMount); i++ {
		p.HookMount[i](ctx, p)
	}

	defer func() {
		for i := 0; i < len(p.HookUnmount); i++ {
			p.HookUnmount[i](ctx, p)
		}
	}()

	taskQueue := make(chan func())
	defer func() { close(taskQueue) }()
	go func() {
		for task := range taskQueue {
			task()
		}
	}()

	for {
		select {
		case <-p.done:
			return nil
		case messageWS, ok := <-p.receive:
			if !ok {
				return nil
			}

			// We can't block here else we can't close and events here can trigger a close
			go func() {
				if !p.connected {
					return
				}

				taskQueue <- func() {
					message := messageWS.Message
					msg := websocketMessage{Data: map[string]string{}}

					if messageWS.IsBinary {
						msgParts := bytes.SplitN(message, []byte("\n\n"), 2)

						if len(msgParts) != 2 {
							p.logger.Error().Msg("invalid binary message")

							return
						}

						message = msgParts[0]
						msg.fileData = msgParts[1]
					}

					p.logger.Debug().Str("msg", string(message)).Msg("ws msg recv")

					if err := json.Unmarshal(message, &msg); err != nil {
						p.logger.Err(err).Str("json", string(message)).Msg("ws msg unmarshal")

						return
					}

					switch msg.Typ {
					// log
					case "l":
						p.logger.Info().Str("log", msg.Data["m"]).Str("sess", sessID).Msg("ws log")
					// Event
					case "e":
						if len(msg.fileData) != 0 && msg.File != nil {
							msg.File.Data = msg.fileData
						}

						// Call handler
						go p.processMsgEvent(ctx, msg)
					default:
						p.logger.Error().Str("msg", string(message)).Msg("ws msg recv: unexpected message format")
					}
				}
			}()
		}
	}
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
	defer func() {
		if r := recover(); r != nil {
			p.logger.Error().Interface("data", r).Msg("wsSend recover")
		}
	}()

	if p == nil || !p.IsConnected() {
		return
	}

	p.logger.Debug().Str("msg", message).Msg("ws send")

	p.send <- MessageWS{Message: []byte(message)}
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

func (p *Page) processMsgEvent(ctx context.Context, msg websocketMessage) {
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

		val, _ := p.eventBindings.GetStringKey(id)
		binding, _ := val.(*EventBinding)

		if binding == nil {
			p.logger.Error().Str("id", id).Msg("unable to find binding")

			return
		}

		e.Binding = binding

		if binding.Handler == nil {
			p.logger.Error().Str("id", id).Msg("binding handler nil")

			p.eventBindings.Del(id)

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
			p.pageLock.Lock()
			p.eventBindings.Del(id)
			binding.Component.RemoveEventBinding(id)
			p.pageLock.Unlock()
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

	diffs, err := p.Differ.Trees("doc", "", p.DOMBrowser, tree)
	if err != nil {
		return nil, fmt.Errorf("diff old and new tag trees: %w", err)
	}

	p.DOMBrowser = tree

	return diffs, nil
}

func (p *Page) GetNodes() *NodeGroup {
	return G(p.DOM.DocType, p.DOM.HTML)
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
	if err != nil {
		p.logger.Error().Str("id", comp.GetID()).Str("name", comp.GetName()).
			Msg("render component ws: pipeline run node")

		return
	}

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
	return p.findComponent(id, p.DOMBrowser)
}

func (p *Page) findComponentInTree(id string) *Tag {
	return p.findComponent(id, p.DOMBrowser)
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
