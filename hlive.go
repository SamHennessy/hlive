package hlive

import (
	"context"
	"log"

	"github.com/vmihailenco/msgpack/v5"
)

// HTML must always have a single root element, as we count it as 1 node in the tree but the browser will not if you
// have multiple root elements
type HTML string

func (e *HTML) MarshalMsgpack() ([]byte, error) {
	return []byte(*e), nil
}

func (e *HTML) UnmarshalMsgpack(b []byte) error {
	*e = HTML(b)

	return nil
}

// IsElement returns true is the pass value is a valid Element.
//
// An Element is anything that cna be rendered at HTML.
func IsElement(el any) bool {
	if IsNonNodeElement(el) {
		return true
	}

	return IsNode(el)
}

func IsNonNodeElement(el any) bool {
	switch el.(type) {
	case []Attributer, []*Attribute, Attributer, *Attribute, Attrs, ClassBool, Style, ClassList, ClassListOff, Class,
		ClassOff, *EventBinding:
		return true
	default:
		return false
	}
}

// IsNode returns true is the pass value is a valid Node.
//
// A Node is a value that could be rendered as HTML by itself. An int for example can be converted to a string which is
// valid HTML. An attribute would not be valid and doesn't make sense to cast to a string.
func IsNode(node any) bool {
	switch node.(type) {
	case nil, string, HTML, Tagger,
		[]any, *NodeGroup, []*Component, []*Tag, []Componenter, []Tagger, []UniqueTagger,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64,
		*string, *HTML:
		return true
	default:
		return false
	}
}

// G is shorthand for Group.
//
// Group zero or more Nodes together.
func G(nodes ...any) *NodeGroup {
	return Group(nodes...)
}

// Group zero or more Nodes together.
func Group(nodes ...any) *NodeGroup {
	g := &NodeGroup{}

	g.Add(nodes...)

	return g
}

// NodeGroup is a Group of Nodes
type NodeGroup struct {
	Group []any
}

func (g *NodeGroup) MarshalMsgpack() ([]byte, error) {
	return msgpack.Marshal(g.Group) //nolint:wrapcheck
}

func (g *NodeGroup) UnmarshalMsgpack(b []byte) error {
	return msgpack.Unmarshal(b, &g.Group) //nolint:wrapcheck
}

func (g *NodeGroup) Add(nodes ...any) {
	if g == nil {
		return
	}

	for i := 0; i < len(nodes); i++ {
		if !IsNode(nodes[i]) {
			log.Printf("HLive: ERROR: added a non-Node to a NodeGroup: %#v\n", nodes[i])

			continue
		}

		g.Group = append(g.Group, nodes[i])
	}
}

func (g *NodeGroup) Get() []any {
	if g == nil {
		return nil
	}

	return g.Group
}

// E is shorthand for Elements.
//
// Groups zero or more Element values.
func E(elements ...any) *ElementGroup {
	return Elements(elements...)
}

// Elements groups zero or more Element values.
func Elements(elements ...any) *ElementGroup {
	g := &ElementGroup{}

	g.Add(elements...)

	return g
}

// ElementGroup is a Group of Elements
type ElementGroup struct {
	list []any
}

func (g *ElementGroup) Add(elements ...any) {
	if g == nil {
		return
	}

	for i := 0; i < len(elements); i++ {
		if !IsElement(elements[i]) {
			log.Printf("HLive: ERROR: added a non-Element to an ElementGroup: %#v\n", elements[i])

			continue
		}

		g.list = append(g.list, elements[i])
	}
}

func (g *ElementGroup) Get() []any {
	if g == nil {
		return nil
	}

	return g.list
}

// Render will trigger a WebSocket render for the current page
func Render(ctx context.Context) {
	render, ok := ctx.Value(CtxRender).(func(context.Context))
	if !ok {
		log.Println("HLive: ERROR:", ErrRenderCtx)

		return
	}

	render(ctx)
}

// RenderComponent will trigger a WebSocket render for the current page from the passed Componenter down only
func RenderComponent(ctx context.Context, comp Componenter) {
	render, ok := ctx.Value(CtxRenderComponent).(func(context.Context, Componenter))
	if !ok {
		log.Println("HLive: ERROR:", ErrRenderCompCtx)

		return
	}

	render(ctx, comp)
}
