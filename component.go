package hlive

import (
	"context"
	"fmt"

	"github.com/rs/xid"
)

type ComponentInterface interface {
	TagInterface

	GetID() string
	GetEventBinding(id string) *EventBinding
	GetEventBindings() []*EventBinding
	IsAutoRender() bool
}

type ComponentMountInterface interface {
	ComponentInterface

	Mount(ctx context.Context)
}

type ComponentUnmountInterface interface {
	ComponentInterface

	Unmount(ctx context.Context)
}

type Component struct {
	*Tag

	id         string
	bindings   []*EventBinding
	AutoRender bool
}

func C(name string, elements ...interface{}) *Component {
	return NewComponent(name, elements...)
}

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

func W(tag *Tag, elements ...interface{}) *Component {
	return Wrap(tag, elements...)
}

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

// GetEventBinding will return an EventBinding is it exists directly on this element, it doesn't check it's children
func (c *Component) GetEventBinding(id string) *EventBinding {
	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			return c.bindings[i]
		}
	}

	return nil
}

func (c *Component) GetEventBindings() []*EventBinding {
	return c.bindings
}

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

func (c *Component) On(bindings ...*EventBinding) {
	for i := 0; i < len(bindings); i++ {
		c.on(bindings[i])
	}
}

func (c *Component) on(binding *EventBinding) {
	binding.Component = c

	id := binding.ID
	if id == "" {
		id = xid.New().String()
	}

	var attrName string

	switch binding.Type {
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
	default:
		panic(fmt.Errorf("unknown EventType"))
	}

	if c.GetAttributeValue(attrName) != "" {
		id = c.GetAttributeValue(attrName) + "," + id
	}

	c.Add(NewAttribute(attrName, id))

	c.bindings = append(c.bindings, binding)
}

func OnClick(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = Click
	binding.Handler = handler

	return binding
}

func OnKeyDown(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = KeyDown
	binding.Handler = handler

	return binding
}

func OnKeyUp(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = KeyUp
	binding.Handler = handler

	return binding
}

func OnFocus(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = Focus
	binding.Handler = handler

	return binding
}

func OnAnimationEnd(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = AnimationEnd
	binding.Handler = handler

	return binding
}

func OnAnimationCancel(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = AnimationCancel
	binding.Handler = handler

	return binding
}

func OnMouseEnter(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = MouseEnter
	binding.Handler = handler

	return binding
}

func OnMouseLeave(handler EventHandler) *EventBinding {
	binding := NewEventBinding()
	binding.Type = MouseLeave
	binding.Handler = handler

	return binding
}
