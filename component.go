package hlive

import (
	"strconv"
	"strings"
)

// Componenter builds on UniqueTagger and adds the ability to handle events.
type Componenter interface {
	UniqueTagger
	// GetEventBinding returns a binding by its id
	GetEventBinding(id string) *EventBinding
	// GetEventBindings returns all event bindings for this tag
	GetEventBindings() []*EventBinding
	// RemoveEventBinding remove an event binding from this component
	RemoveEventBinding(id string)
	// IsAutoRender indicates if the page should rerender after an event binding on this tag is called
	IsAutoRender() bool
}

// Component is the default implementation of Componenter.
type Component struct {
	*Tag

	AutoRender bool
	id         string
	bindingID  uint32
	bindings   []*EventBinding
}

// C is a shortcut for NewComponent.
//
// NewComponent is a constructor for Component.
//
// You can add zero or many Attributes and Tags
func C(name string, elements ...any) *Component {
	return NewComponent(name, elements...)
}

// NewComponent is a constructor for Component.
//
// You can add zero or many Attributes and Tags.
func NewComponent(name string, elements ...any) *Component {
	c := &Component{
		Tag:        T(name),
		AutoRender: true,
	}

	c.Add(elements...)

	return c
}

// W is a shortcut for Wrap.
//
// Wrap takes a Tag and creates a Component with it.
func W(tag *Tag, elements ...any) *Component {
	return Wrap(tag, elements...)
}

// Wrap takes a Tag and creates a Component with it.
func Wrap(tag *Tag, elements ...any) *Component {
	c := &Component{
		Tag:        tag,
		AutoRender: true,
	}

	c.Add(elements...)

	return c
}

// GetID returns this component's unique ID
func (c *Component) GetID() string {
	c.Tag.mu.RLock()
	defer c.Tag.mu.RUnlock()

	return c.id
}

// SetID component's unique ID
func (c *Component) SetID(id string) {
	c.Tag.mu.Lock()
	defer c.Tag.mu.Unlock()

	c.id = id
	c.Tag.addAttributes(NewAttribute(AttrID, id))

	if value := c.bindingAttrValue(); value != "" {
		c.Tag.addAttributes(NewAttribute(AttrOn, value))
	}
}

func (c *Component) bindingAttrValue() string {
	var value string
	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == "" {
			c.bindingID++
			c.bindings[i].ID = c.id + "-" + strconv.FormatUint(uint64(c.bindingID), 10)
		}

		value += c.bindings[i].ID + "|" + c.bindings[i].Name + ","
	}

	return strings.TrimRight(value, ",")
}

// IsAutoRender indicates if this component should trigger "Auto Render"
func (c *Component) IsAutoRender() bool {
	c.Tag.mu.RLock()
	defer c.Tag.mu.RUnlock()

	return c.AutoRender
}

// GetEventBinding will return an EventBinding that exists directly on this element, it doesn't check its children.
// Returns nil is not found.
func (c *Component) GetEventBinding(id string) *EventBinding {
	c.Tag.mu.RLock()
	defer c.Tag.mu.RUnlock()

	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			return c.bindings[i]
		}
	}

	return nil
}

// GetEventBindings returns all EventBindings for this component, not it's children.
func (c *Component) GetEventBindings() []*EventBinding {
	c.Tag.mu.RLock()
	defer c.Tag.mu.RUnlock()

	return append([]*EventBinding{}, c.bindings...)
}

// RemoveEventBinding removes an EventBinding that matches the passed ID.
//
// No error if the passed id doesn't match an EventBinding.
// It doesn't check its children.
func (c *Component) RemoveEventBinding(id string) {
	c.Tag.mu.Lock()
	defer c.Tag.mu.Unlock()

	var newList []*EventBinding
	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			continue
		}

		newList = append(newList, c.bindings[i])
	}
	c.bindings = newList

	// Reset attribute
	if value := c.bindingAttrValue(); value == "" {
		c.removeAttributes(AttrOn)
	} else {
		c.Tag.addAttributes(NewAttribute(AttrOn, value))
	}
}

// Add an element to this Component.
//
// This is an easy way to add anything.
func (c *Component) Add(elements ...any) {
	if c.IsNil() {
		return
	}

	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		// NoneNodeElements
		case []any:
			for j := 0; j < len(v); j++ {
				c.Add(v[j])
			}
		case *NodeGroup:
			if v == nil {
				continue
			}

			list := v.Get()
			for j := 0; j < len(list); j++ {
				c.Add(list[j])
			}
		case *ElementGroup:
			if v == nil {
				continue
			}

			list := v.Get()
			for j := 0; j < len(list); j++ {
				c.Add(list[j])
			}
		case *EventBinding:
			if v == nil {
				continue
			}

			c.Tag.mu.Lock()
			c.on(v)
			c.Tag.mu.Unlock()
		default:
			c.Tag.Add(v)
		}
	}
}

func (c *Component) on(binding *EventBinding) {
	binding.Component = c
	c.bindings = append(c.bindings, binding)

	// See Component.SetID
	if c.id == "" {
		return
	}

	if value := c.bindingAttrValue(); value != "" {
		c.Tag.addAttributes(NewAttribute(AttrOn, value))
	}
}
