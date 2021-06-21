package hlive

import (
	"context"
	"strings"

	"github.com/rs/xid"
)

// Componenter builds on Tagger and adds the ability to handle events.
type Componenter interface {
	// Tagger means this is a tag
	Tagger

	// GetID will return a unique id for this tag
	GetID() string
	// GetEventBinding returns a binding by it's id
	GetEventBinding(id string) *EventBinding
	// GetEventBindings returns all event bindings for this tag
	GetEventBindings() []*EventBinding
	// RemoveEventBinding remove an event binding from this component
	RemoveEventBinding(id string)
	// IsAutoRender indicates if the page should rerender after an event binding on this tag is called
	// TODO: move this to it's own interface?
	IsAutoRender() bool
}

// Mounter wants to be notified after it's mounted.
type Mounter interface {
	// GetID will return a unique id for this tag
	GetID() string
	// Mount is called after a component is mounted
	Mount(ctx context.Context)
}

// Unmounter wants to be notified before it's unmounted.
type Unmounter interface {
	// GetID will return a unique id for this tag
	GetID() string
	// Unmount is called before a component is unmounted
	Unmount(ctx context.Context)
}

// Teardowner wants to be able to manual control when it needs to be removed from a Page.
// If you have a Mounter or Unmounter that will be permanently removed from a Page they must call the passed
// function to clean up their references.
type Teardowner interface {
	// GetID will return a unique id for this tag
	GetID() string
	// SetTeardown set teardown function
	SetTeardown(teardown func())
}

// Component is an the default implementation of Componenter.
type Component struct {
	*Tag

	id         string
	bindings   []*EventBinding
	AutoRender bool
}

// C is a shortcut for NewComponent.
func C(name string, elements ...interface{}) *Component {
	return NewComponent(name, elements...)
}

// NewComponent is a constructor for Component.
// You can add zero or many Attributes and Tags
func NewComponent(name string, elements ...interface{}) *Component {
	c := &Component{
		Tag:        T(name),
		id:         xid.New().String(),
		AutoRender: true,
	}

	c.Add(NewAttribute(AttrID, c.GetID()))
	c.Add(elements...)

	return c
}

// W is a shortcut for Wrap.
func W(tag *Tag, elements ...interface{}) *Component {
	return Wrap(tag, elements...)
}

// Wrap takes a Tag and creates a Component with it.
func Wrap(tag *Tag, elements ...interface{}) *Component {
	c := &Component{
		Tag:        tag,
		id:         xid.New().String(),
		AutoRender: true,
	}

	c.Add(NewAttribute(AttrID, c.GetID()))
	c.Add(elements...)

	return c
}

func (c *Component) GetID() string {
	return c.id
}

func (c *Component) IsAutoRender() bool {
	return c.AutoRender
}

// GetEventBinding will return an EventBinding that exists directly on this element, it doesn't check it's children.
// Returns nil is not found.
func (c *Component) GetEventBinding(id string) *EventBinding {
	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			return c.bindings[i]
		}
	}

	return nil
}

// GetEventBindings returns all EventBindings for this component, not it's children.
func (c *Component) GetEventBindings() []*EventBinding {
	return c.bindings
}

// RemoveEventBinding removes an EventBinding that matches the passed ID.
// No error if the passed id doesn't match an EventBinding.
// It doesn't check it's children.
func (c *Component) RemoveEventBinding(id string) {
	var newList []*EventBinding

	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			c.removeIDFromAttr(id, eventToAttr(c.bindings[i].Type))

			continue
		}

		newList = append(newList, c.bindings[i])
	}

	c.bindings = newList
}

// Add an element to this Component.
// This is an easy way to add anything.
func (c *Component) Add(elements ...interface{}) {
	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		case []interface{}:
			for j := 0; j < len(v); j++ {
				c.Add(v[j])
			}
		case *EventBinding:
			c.On(v)
		default:
			c.Tag.Add(v)
		}
	}
}

// On allows you to add one or more EventBinding to this component
// TODO: do we need this, as Add also does this
func (c *Component) On(bindings ...*EventBinding) {
	for i := 0; i < len(bindings); i++ {
		c.on(bindings[i])
	}
}

func eventToAttr(et EventType) string {
	var attrName string

	switch et {
	case Click:
		attrName = AttrOnClick
	case KeyDown:
		attrName = AttrOnKeyDown
	case KeyUp:
		attrName = AttrOnKeyUp
	case Focus:
		attrName = AttrOnFocus
	case AnimationEnd:
		attrName = AttrOnAnimationEnd
	case AnimationCancel:
		attrName = AttrOnAnimationCancel
	case MouseEnter:
		attrName = AttrOnMouseEnter
	case MouseLeave:
		attrName = AttrOnMouseLeave
	case DiffApply:
		attrName = AttrOnDiffApply
	case Upload:
		attrName = AttrOnUpload
	case Change:
		attrName = AttrOnChange
	default:
		panic(ErrEventType)
	}

	return attrName
}

func (c *Component) removeIDFromAttr(id, attrName string) {
	var newIDs []string

	oldIDs := strings.Split(c.GetAttribute(attrName).GetValue(), ",")
	for j := 0; j < len(oldIDs); j++ {
		if oldIDs[j] == id {
			continue
		}

		newIDs = append(newIDs, oldIDs[j])
	}

	if len(newIDs) == 0 {
		c.RemoveAttributes(attrName)
	} else {
		c.Add(NewAttribute(attrName, strings.Join(newIDs, ",")))
	}
}

func (c *Component) on(binding *EventBinding) {
	binding.Component = c

	id := binding.ID
	if id == "" {
		id = xid.New().String()
	}

	var attrName string

	if binding.Name != "" {
		attrName = "data-hlive-on"
		id = id + "|" + binding.Name
	} else {
		attrName = eventToAttr(binding.Type)
	}

	// Support multiple bindings per type
	if c.GetAttributeValue(attrName) != "" {
		id = c.GetAttributeValue(attrName) + "," + id
	}

	c.Add(NewAttribute(attrName, id))

	c.bindings = append(c.bindings, binding)
}

func onHelper(eventType EventType, handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = eventType
	binding.Handler = handler

	return binding
}

func OnClick(handler EventHandler) *EventBinding {
	return onHelper(Click, handler)
}

func OnKeyDown(handler EventHandler) *EventBinding {
	return onHelper(KeyDown, handler)
}

func OnKeyUp(handler EventHandler) *EventBinding {
	return onHelper(KeyUp, handler)
}

func OnFocus(handler EventHandler) *EventBinding {
	return onHelper(Focus, handler)
}

func OnAnimationEnd(handler EventHandler) *EventBinding {
	return onHelper(AnimationEnd, handler)
}

func OnAnimationCancel(handler EventHandler) *EventBinding {
	return onHelper(AnimationCancel, handler)
}

func OnMouseEnter(handler EventHandler) *EventBinding {
	return onHelper(MouseEnter, handler)
}

func OnMouseLeave(handler EventHandler) *EventBinding {
	return onHelper(MouseLeave, handler)
}

// OnDiffApply is a special event that will trigger when a diff is applied.
// This means that it will trigger itself when first added. This will allow you to know when a change in the tree has
// made it to the browser. You can then, if you wish, immediately remove it from the tree to prevent more triggers.
func OnDiffApply(handler EventHandler) *EventBinding {
	return onHelper(DiffApply, handler)
}

func OnUpload(handler EventHandler) *EventBinding {
	return onHelper(Upload, handler)
}

func OnChange(handler EventHandler) *EventBinding {
	return onHelper(Change, handler)
}

func On(name string, handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Handler = handler
	binding.Name = name

	return binding
}
