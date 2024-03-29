package hlive

import (
	"context"
	"fmt"
	"sync"

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
	case []Attributer, []*Attribute, *Attribute, Attributer, Attrs, AttrsLockBox, AttrsOff,
		ClassBool, Style, ClassList, ClassListOff, Class, ClassOff,
		*EventBinding:
		return true
	default:
		return false
	}
}

type NodeBoxer interface {
	GetNode() any
}

type NodeBox[V any] struct {
	*LockBox[V]
}

func (b NodeBox[V]) GetNode() any {
	return b.Get()
}

func Box[V any](node V) *NodeBox[V] {
	if !IsNode(node) {
		LoggerDev.Error().
			Str("callers", CallerStackStr()).
			Str("node", fmt.Sprintf("%#v", node)).
			Msg("invalid node")
	}

	return &NodeBox[V]{NewLockBox(node)}
}

type LockBox[V any] struct {
	mu  sync.RWMutex
	val V
}

func (b *LockBox[V]) Set(val V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.val = val
}

func (b *LockBox[V]) Get() V {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.val
}

type LockBoxer interface {
	GetLockedAny() any
}

func (b *LockBox[V]) GetLockedAny() any {
	return b.Get()
}

func (b *LockBox[V]) Lock(f func(val V) V) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.val = f(b.val)
}

func NewLockBox[V any](val V) *LockBox[V] {
	return &LockBox[V]{val: val}
}

// IsNode returns true is the pass value is a valid Node.
//
// A Node is a value that could be rendered as HTML by itself. An int for example can be converted to a string which is
// valid HTML. An attribute would not be valid and doesn't make sense to cast to a string.
func IsNode(node any) bool {
	switch node.(type) {
	// TODO: Need *HTML for encoding, maybe new read only Tag will help
	case nil, string, HTML, *HTML, Tagger,
		[]any, *NodeGroup, []*Component, []*Tag, []Componenter, []Tagger, []UniqueTagger,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		NodeBoxer, LockBoxer:
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
	group []any
	mu    sync.RWMutex
}

func (g *NodeGroup) MarshalMsgpack() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return msgpack.Marshal(g.group) //nolint:wrapcheck
}

func (g *NodeGroup) UnmarshalMsgpack(b []byte) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return msgpack.Unmarshal(b, &g.group) //nolint:wrapcheck
}

func (g *NodeGroup) Add(nodes ...any) {
	if g == nil {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("nil call")

		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for i := 0; i < len(nodes); i++ {
		if !IsNode(nodes[i]) {
			LoggerDev.Error().
				Str("callers", CallerStackStr()).
				Str("node", fmt.Sprintf("%#v", nodes[i])).
				Msg("invalid node")

			continue
		}

		g.group = append(g.group, nodes[i])
	}
}

// Get returns all nodes, dereferences any valid pointers
func (g *NodeGroup) Get() []any {
	if g == nil {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("nil call")

		return nil
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	var newGroup []any

	for i := 0; i < len(g.group); i++ {
		node := g.group[i]
		if node == nil {
			continue
		}

		switch v := node.(type) {
		case NodeBoxer:
			newGroup = append(newGroup, v.GetNode())
		case LockBoxer:
			newGroup = append(newGroup, v.GetLockedAny())
		default:
			newGroup = append(newGroup, node)
		}
	}

	return newGroup
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
	group []any
	mu    sync.RWMutex
}

func (g *ElementGroup) Add(elements ...any) {
	if g == nil {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("nil call")

		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	for i := 0; i < len(elements); i++ {
		if !IsElement(elements[i]) {
			LoggerDev.Error().
				Str("callers", CallerStackStr()).
				Str("element", fmt.Sprintf("%#v", elements[i])).
				Msg("invalid element")

			continue
		}

		g.group = append(g.group, elements[i])
	}
}

func (g *ElementGroup) Get() []any {
	if g == nil {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("nil call")

		return nil
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	var newGroup []any

	for i := 0; i < len(g.group); i++ {
		node := g.group[i]
		if node == nil {
			continue
		}

		switch v := node.(type) {
		case NodeBoxer:
			newGroup = append(newGroup, v.GetNode())
		case LockBoxer:
			newGroup = append(newGroup, v.GetLockedAny())
		default:
			newGroup = append(newGroup, node)
		}
	}

	return newGroup
}

// Render will trigger a WebSocket render for the current page
func Render(ctx context.Context) {
	render, ok := ctx.Value(CtxRender).(func(context.Context))
	if !ok {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("render not found in context")

		return
	}

	render(ctx)
}

// RenderComponent will trigger a WebSocket render for the current page from the passed Componenter down only
func RenderComponent(ctx context.Context, comp Componenter) {
	render, ok := ctx.Value(CtxRenderComponent).(func(context.Context, Componenter))
	if !ok {
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg("component render not found in context")

		return
	}

	render(ctx, comp)
}
