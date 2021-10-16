package hlive

import (
	"strings"

	"github.com/teris-io/shortid"
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

	id         string
	bindings   []*EventBinding
	AutoRender bool
}

// C is a shortcut for NewComponent.
//
// NewComponent is a constructor for Component.
//
// You can add zero or many Attributes and Tags
func C(name string, elements ...interface{}) *Component {
	return NewComponent(name, elements...)
}

// NewComponent is a constructor for Component.
//
// You can add zero or many Attributes and Tags.
func NewComponent(name string, elements ...interface{}) *Component {
	c := &Component{
		Tag:        T(name),
		id:         shortid.MustGenerate(),
		AutoRender: true,
	}

	c.Add(NewAttribute(AttrID, c.GetID()), elements)

	return c
}

// W is a shortcut for Wrap.
//
// Wrap takes a Tag and creates a Component with it.
func W(tag *Tag, elements ...interface{}) *Component {
	return Wrap(tag, elements...)
}

// Wrap takes a Tag and creates a Component with it.
func Wrap(tag *Tag, elements ...interface{}) *Component {
	c := &Component{
		Tag:        tag,
		id:         shortid.MustGenerate(),
		AutoRender: true,
	}

	c.Add(NewAttribute(AttrID, c.GetID()))
	c.Add(elements...)

	return c
}

// GetID returns this components unique ID
func (c *Component) GetID() string {
	return c.id
}

// IsAutoRender indicates if this component should trigger "Auto Render"
func (c *Component) IsAutoRender() bool {
	return c.AutoRender
}

// GetEventBinding will return an EventBinding that exists directly on this element, it doesn't check its children.
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
//
// No error if the passed id doesn't match an EventBinding.
// It doesn't check its children.
func (c *Component) RemoveEventBinding(id string) {
	var newList []*EventBinding

	for i := 0; i < len(c.bindings); i++ {
		if c.bindings[i].ID == id {
			continue
		}

		newList = append(newList, c.bindings[i])
	}

	c.bindings = newList

	// Reset attribute
	var value string

	for i := 0; i < len(c.bindings); i++ {
		value += c.bindings[i].ID + "|" + c.bindings[i].Name + ","
	}

	value = strings.TrimRight(value, ",")

	if value == "" {
		c.RemoveAttributes(AttrOn)
	} else {
		c.Add(Attrs{AttrOn: value})
	}
}

// Add an element to this Component.
//
// This is an easy way to add anything.
func (c *Component) Add(elements ...interface{}) {
	for i := 0; i < len(elements); i++ {
		switch v := elements[i].(type) {
		// NoneNodeElements
		case []interface{}:
			for j := 0; j < len(v); j++ {
				c.Add(v[j])
			}
		case *NodeGroup:
			list := v.Get()
			for j := 0; j < len(list); j++ {
				c.Add(list[j])
			}
		case *ElementGroup:
			list := v.Get()
			for j := 0; j < len(list); j++ {
				c.Add(list[j])
			}
		case *EventBinding:
			c.on(v)
		default:
			c.Tag.Add(v)
		}
	}
}

func (c *Component) on(binding *EventBinding) {
	binding.Component = c

	id := binding.ID
	if id == "" {
		id = shortid.MustGenerate()
	}

	value := id + "|" + binding.Name

	// Support multiple bindings per type
	if c.GetAttributeValue(AttrOn) != "" {
		value = c.GetAttributeValue(AttrOn) + "," + value
	}

	c.Add(NewAttribute(AttrOn, value))

	c.bindings = append(c.bindings, binding)
}
