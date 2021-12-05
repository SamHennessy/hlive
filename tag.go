package hlive

import (
	"fmt"
	"strings"
	"sync"
)

// Tagger represents a static HTML tag.
type Tagger interface {
	// GetName returns a tag's name. For example <hr>'s name is hr.
	GetName() string
	// GetAttributes returns all attributes for this tag.
	GetAttributes() []Attributer
	// GetNodes returns this tags children nodes, to be rendered inside this tag.
	GetNodes() *NodeGroup
	// IsVoid indicates if this has a closing tag or not. Void tags don't have a closing tag.
	IsVoid() bool
	// IsNil returns true if pointer is nil.
	//
	// It's easy to create something like `var t *Tag` but forget to give it a value.
	// This allows us to not have panics in that case.
	IsNil() bool
}

// UniqueTagger is a Tagger that can be uniquely identified in a DOM Tree.
type UniqueTagger interface {
	Tagger
	// GetID will return a unique id
	GetID() string
}

// Adder interface for inputting elements to Tagger type values.
type Adder interface {
	// Add elements to a Tagger
	Add(elements ...interface{})
}

// UniqueAdder is an Adder that can be uniquely identified in a DOM Tree.
type UniqueAdder interface {
	Adder
	// GetID will return a unique id
	GetID() string
}

// Tag is the default HTML tag implementation.
//
// Use T or NewTag to create a value.
type Tag struct {
	name        string
	void        bool
	attributes  []Attributer
	nodes       *NodeGroup
	cssExists   map[string]bool
	cssOrder    []string
	styleValues map[string]*string
	styleOrder  []string
	lock        sync.RWMutex
}

// T is a shortcut for NewTag.
//
// NewTag creates a new Tag value.
func T(name string, elements ...interface{}) *Tag {
	return NewTag(name, elements...)
}

// NewTag creates a new Tag value.
func NewTag(name string, elements ...interface{}) *Tag {
	name = strings.ToLower(name)

	var void bool

	switch name {
	case "area", "base", "br", "col", "command", "embed", "hr", "img", "input", "keygen", "link", "meta", "param",
		"source", "track", "wbr":
		void = true
	}

	t := &Tag{
		name:        name,
		void:        void,
		cssExists:   map[string]bool{},
		styleValues: map[string]*string{},
		nodes:       G(),
	}

	addElementToTag(t, elements)

	return t
}

// IsNil returns true if pointer is nil
func (t *Tag) IsNil() bool {
	return t == nil
}

// SetName sets the tag name, e.g. for a `<div>` it's the `div` part.
func (t *Tag) SetName(name string) {
	t.name = name
}

// GetName get the tag name.
func (t *Tag) GetName() string {
	return t.name
}

// IsVoid indicates if this is a void type tag, e.g. `<hr>`.
func (t *Tag) IsVoid() bool {
	return t.void
}

// SetVoid sets the tag to be a void type tag e.g. `<hr>`.
func (t *Tag) SetVoid(void bool) {
	t.void = void
}

// GetAttributes returns a list of Attributer values that this tag has.
//
// Any Class, Style values are returned here as Attribute values.
func (t *Tag) GetAttributes() []Attributer {
	attrs := t.attributes

	if len(t.cssOrder) != 0 {
		attrs = append(attrs, NewAttribute("class", strings.Join(t.cssOrder, " ")))
	}

	if len(t.styleOrder) != 0 {
		value := ""

		for i := 0; i < len(t.styleOrder); i++ {
			name := t.styleOrder[i]
			if t.styleValues[name] == nil {
				continue
			}

			value += name + ":" + *t.styleValues[name] + ";"
		}

		attrs = append(attrs, NewAttribute("style", value))
	}

	return attrs
}

// GetAttribute returns an Attributer value by its name.
//
// This includes attribute values related to Class, and Style. If an Attributer of the passed name has not been set `nil`
// it's returned.
func (t *Tag) GetAttribute(name string) Attributer {
	attrs := t.GetAttributes()
	for i := 0; i < len(attrs); i++ {
		if attrs[i].GetAttribute().Name == name {
			return attrs[i]
		}
	}

	return nil
}

// GetAttributeValue returns a value for a given Attributer name.
//
// If an attribute has not yet been set, then an empty string is returned.
func (t *Tag) GetAttributeValue(name string) string {
	a := t.GetAttribute(name)
	if a == nil {
		return ""
	}

	return a.GetAttribute().GetValue()
}

// Add zero or more elements to this Tag.
func (t *Tag) Add(element ...interface{}) {
	if t.IsNil() {
		return
	}

	t.lock.Lock()
	addElementToTag(t, element)
	t.lock.Unlock()
}

// GetNodes returns a NodeGroup with any child Nodes that have been added to this Node.
func (t *Tag) GetNodes() *NodeGroup {
	return t.nodes
}

func (t *Tag) addNodes(nodes ...interface{}) {
	for i := 0; i < len(nodes); i++ {
		if nodes[i] == nil {
			continue
		}

		if !IsNode(nodes[i]) {
			panic(fmt.Errorf("Tag.addNodes: node: %#v: %w", nodes[i], ErrInvalidNode))
		}

		t.nodes.Add(nodes[i])
	}
}

// AddAttributes will add zero or more attributes types (Attributer, Attribute, Attrs, Style, ClassBool).
//
// Adding an attribute with the same name will override an existing attribute.
func (t *Tag) AddAttributes(attrs ...interface{}) {
	if t.IsNil() {
		return
	}

	t.lock.Lock()
	t.addAttributes(attrs...)
	t.lock.Unlock()
}

func (t *Tag) addAttributes(attrs ...interface{}) {
	newAttributes := anyToAttributes(attrs...)
	newAttrsCount := len(newAttributes)
	for i := 0; i < newAttrsCount; i++ {
		hit := false

		// TODO: why does this panic sometimes?
		// Maybe when we had nil tags possible
		for j := 0; j < len(t.attributes); j++ {
			if t.attributes[j].GetAttribute().Name == newAttributes[i].GetAttribute().Name {
				hit = true
				t.attributes[j].GetAttribute().Value = newAttributes[i].GetAttribute().Value
			}
		}

		if !hit {
			t.attributes = append(t.attributes, newAttributes[i])
		}
	}

	for i := 0; i < len(t.attributes); i++ {
		if t.attributes[i].GetAttribute().Value == nil {
			t.removeAttributes(t.attributes[i].GetAttribute().Name)
		}
	}
}

func (t *Tag) RemoveAttributes(names ...string) {
	if t.IsNil() {
		return
	}

	t.lock.Lock()
	t.removeAttributes(names...)
	t.lock.Unlock()
}

// RemoveAttributes remove zero or more Attributer value by their name.
func (t *Tag) removeAttributes(names ...string) {
	var newAttrs []Attributer

	for j := 0; j < len(t.attributes); j++ {
		attr := t.attributes[j]
		hit := false

		// TODO: Why is this possible?
		if attr == nil {
			continue
		}

		for i := 0; i < len(names); i++ {
			if names[i] == attr.GetAttribute().Name {
				hit = true

				break
			}
		}

		if !hit {
			newAttrs = append(newAttrs, attr)
		}
	}

	t.attributes = newAttrs
}

func addElementToTag(t *Tag, v interface{}) {
	if _, ok := v.(*EventBinding); ok {
		panic("You've added an event binding to a Tag. You can only add these to a Component. " +
			"You can turn any Tag into a Component by using the Wrap or W functions.")
	}

	if !IsElement(v) {
		panic(fmt.Errorf("element: %#v: %w", v, ErrInvalidElement))
	}

	switch v := v.(type) {
	case *NodeGroup:
		if v == nil {
			return
		}

		g := v.Get()
		for i := 0; i < len(g); i++ {
			addElementToTag(t, g[i])
		}
	case *ElementGroup:
		if v == nil {
			return
		}

		g := v.Get()
		for i := 0; i < len(g); i++ {
			addElementToTag(t, g[i])
		}
	case []interface{}:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []*Component:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []*Tag:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []Componenter:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []Tagger:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []UniqueTagger:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []Attributer, Attributer, *Attribute, []*Attribute, Attrs:
		t.addAttributes(v)
	case Class:
		classes := strings.Split(string(v), " ")
		for i := 0; i < len(classes); i++ {
			if classes[i] == "" {
				continue
			}

			addCSS(t, ClassBool{classes[i]: true})
		}
	case ClassOff:
		classes := strings.Split(string(v), " ")
		for i := 0; i < len(classes); i++ {
			if classes[i] == "" {
				continue
			}

			addCSS(t, ClassBool{classes[i]: false})
		}
	case ClassList:
		for i := 0; i < len(v); i++ {
			addCSS(t, ClassBool{v[i]: true})
		}
	case ClassListOff:
		for i := 0; i < len(v); i++ {
			addCSS(t, ClassBool{v[i]: false})
		}
	case ClassBool:
		addCSS(t, v)
	case Style:
		addStyle(t, v)
	default:
		t.addNodes(v)
	}
}

func addStyle(t *Tag, v Style) {
	removeList := map[string]bool{}

	for name, value := range v {
		if _, exists := t.styleValues[name]; exists {
			if value == nil {
				removeList[name] = true
			} else {
				strP, ok := value.(*string)
				if ok {
					t.styleValues[name] = strP
				} else {
					str, _ := value.(string)
					t.styleValues[name] = &str
				}
			}
		} else if value != nil {
			removeList[name] = false

			strP, ok := value.(*string)
			if ok {
				t.styleValues[name] = strP
			} else {
				str, _ := value.(string)
				t.styleValues[name] = &str
			}

			t.styleOrder = append(t.styleOrder, name)
		}
	}

	var newNames []string

	for i := 0; i < len(t.styleOrder); i++ {
		name := t.styleOrder[i]
		if removeList[name] {
			delete(t.styleValues, name)

			continue
		}

		newNames = append(newNames, name)
	}

	t.styleOrder = newNames
}

// This does allow duplicates in the same hlive.ClassBool element.
func addCSS(t *Tag, v ClassBool) {
	// Update the map
	for class, enable := range v {
		if enable {
			if !t.cssExists[class] {
				t.cssExists[class] = true
				t.cssOrder = append(t.cssOrder, class)
			}
		} else {
			delete(t.cssExists, class)
		}
	}

	// Update the order
	// Loop over order and remove all that no longer exists
	var newOrder []string

	for i := 0; i < len(t.cssOrder); i++ {
		if t.cssExists[t.cssOrder[i]] {
			newOrder = append(newOrder, t.cssOrder[i])
		}
	}

	t.cssOrder = newOrder
}
