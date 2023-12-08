package hlive

import (
	"fmt"
	"strings"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// TODO: add tests
type GetTagger interface {
	GetTagger() Tagger
}

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
	// SetID Components will be assigned a unique id
	SetID(id string)
}

// Adder interface for inputting elements to Tagger type values.
type Adder interface {
	// Add elements to a Tagger
	Add(elements ...any)
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
	styleValues map[string]*LockBox[string]
	styleOrder  []string
	mu          sync.RWMutex
}

func (t *Tag) MarshalMsgpack() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	attrs := make([]*Attribute, len(t.attributes))
	for i := 0; i < len(t.attributes); i++ {
		attrs[i], _ = t.attributes[i].(*Attribute)
	}

	styleValues := make([]string, 0, len(t.styleValues))
	for _, l := range t.styleValues {
		styleValues = append(styleValues, l.Get())
	}

	return msgpack.Marshal([8]any{ //nolint:wrapcheck
		t.name,
		t.void,
		attrs,
		t.nodes,
		t.cssExists, // TODO: get from cssOrder
		t.cssOrder,
		styleValues,
		t.styleOrder,
	})
}

func (t *Tag) UnmarshalMsgpack(b []byte) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var values [8]any
	err := msgpack.Unmarshal(b, &values)
	if err != nil {
		return fmt.Errorf("msgpack.Unmarshal: %w", err)
	}

	t.name, _ = values[0].(string)
	t.void, _ = values[1].(bool)

	attributes, _ := values[2].([]any)
	for _, val := range attributes {
		attribute, _ := val.(Attributer)
		t.attributes = append(t.attributes, attribute)
	}

	t.nodes, _ = values[3].(*NodeGroup)

	cssList, _ := values[4].(map[string]any)
	for key, val := range cssList {
		t.cssExists[key], _ = val.(bool)
	}

	if values[5] != nil {
		t.cssOrder, _ = values[5].([]string)
	}

	if values[6] != nil && values[7] != nil {
		styles, _ := values[6].([]string)
		t.styleOrder, _ = values[7].([]string)

		if len(styles) == len(t.styleOrder) {
			for i := 0; i < len(styles); i++ {
				t.styleValues[t.styleOrder[i]] = NewLockBox(styles[i])
			}
		}
	}

	return nil
}

// T is a shortcut for NewTag.
//
// NewTag creates a new Tag value.
func T(name string, elements ...any) *Tag {
	return NewTag(name, elements...)
}

// NewTag creates a new Tag value.
func NewTag(name string, elements ...any) *Tag {
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
		styleValues: map[string]*LockBox[string]{},
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
	t.mu.Lock()
	t.name = name
	t.mu.Unlock()
}

// GetName get the tag name.
func (t *Tag) GetName() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.name
}

// IsVoid indicates if this is a void type tag, e.g. `<hr>`.
func (t *Tag) IsVoid() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.void
}

// SetVoid sets the tag to be a void type tag e.g. `<hr>`.
func (t *Tag) SetVoid(void bool) {
	t.mu.Lock()
	t.void = void
	t.mu.Unlock()
}

// GetAttributes returns a list of Attributer values that this tag has.
//
// Any Class, Style values are returned here as Attribute values.
func (t *Tag) GetAttributes() []Attributer {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Copy the slice
	attrs := append([]Attributer{}, t.attributes...)

	if len(t.cssOrder) != 0 {
		val := strings.Join(t.cssOrder, " ")
		attrs = append(attrs, NewAttribute("class", val))
	}

	if len(t.styleOrder) != 0 {
		value := ""

		for i := 0; i < len(t.styleOrder); i++ {
			name := t.styleOrder[i]
			if t.styleValues[name] == nil {
				continue
			}

			value += name + ":" + t.styleValues[name].Get() + ";"
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
		if attrs[i].GetName() == name {
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

	return a.GetValue()
}

// Add zero or more elements to this Tag.
func (t *Tag) Add(element ...any) {
	if t.IsNil() {
		return
	}

	t.mu.Lock()
	addElementToTag(t, element)
	t.mu.Unlock()
}

// GetNodes returns a NodeGroup with any child Nodes that have been added to this Node.
func (t *Tag) GetNodes() *NodeGroup {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return G(t.nodes.Get()...)
}

// AddAttributes will add zero or more attributes types (Attributer, Attribute, Attrs, Style, ClassBool).
//
// Adding an attribute with the same name will override an existing attribute.
func (t *Tag) AddAttributes(attrs ...any) {
	if t.IsNil() {
		return
	}

	t.mu.Lock()
	t.addAttributes(attrs...)
	t.mu.Unlock()
}

func (t *Tag) addAttributes(attrs ...any) {
	newAttributes := anyToAttributes(attrs...)
	for i := 0; i < len(newAttributes); i++ {
		hit := false
		// Replace existing attribute?
		for j := 0; j < len(t.attributes); j++ {
			if t.attributes[j].GetName() == newAttributes[i].GetName() {
				hit = true

				t.attributes[j] = newAttributes[i]
			}
		}

		if !hit {
			t.attributes = append(t.attributes, newAttributes[i])
		}
	}
}

func (t *Tag) RemoveAttributes(names ...string) {
	if t.IsNil() {
		return
	}

	t.mu.Lock()
	t.removeAttributes(names...)
	t.mu.Unlock()
}

// RemoveAttributes remove zero or more Attributer value by their name.
func (t *Tag) removeAttributes(names ...string) {
	var newAttrs []Attributer

	for j := 0; j < len(t.attributes); j++ {
		attr := t.attributes[j]
		hit := false

		for i := 0; i < len(names); i++ {
			if strings.ToLower(names[i]) == attr.GetName() {
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

func addElementToTag(t *Tag, v any) {
	switch v := v.(type) {
	// Common error
	case *EventBinding:
		LoggerDev.Error().Str("callers", CallerStackStr()).Msg(
			"You've added an event binding to a Tag. You can only add these to a Component. " +
				"You can fix this by replacing l.T(\"div\"...) with l.C(\"div\"...)." +
				"You can also turn any Tag into a Component by using the Wrap function.")

		return
	// Groups
	case []any:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []*Component:
		t.nodes.mu.Lock()
		for i := 0; i < len(v); i++ {
			t.nodes.group = append(t.nodes.group, v[i])
		}
		t.nodes.mu.Unlock()
	case []*Tag:
		t.nodes.mu.Lock()
		for i := 0; i < len(v); i++ {
			t.nodes.group = append(t.nodes.group, v[i])
		}
		t.nodes.mu.Unlock()
	case []Componenter:
		t.nodes.mu.Lock()
		for i := 0; i < len(v); i++ {
			t.nodes.group = append(t.nodes.group, v[i])
		}
		t.nodes.mu.Unlock()
	case []Tagger:
		t.nodes.mu.Lock()
		for i := 0; i < len(v); i++ {
			t.nodes.group = append(t.nodes.group, v[i])
		}
		t.nodes.mu.Unlock()
	case []UniqueTagger:
		t.nodes.mu.Lock()
		for i := 0; i < len(v); i++ {
			t.nodes.group = append(t.nodes.group, v[i])
		}
		t.nodes.mu.Unlock()
	case *NodeGroup:
		t.nodes.Add(v)
	case *ElementGroup:
		for i := 0; i < len(v.group); i++ {
			addElementToTag(t, v.group[i])
		}
	// Attributes
	case AttrsOff:
		t.removeAttributes(v...)
	case Attrs, AttrsLockBox, *Attribute, Attributer, []Attributer, []*Attribute:
		t.addAttributes(v)
	// Singles
	case nil:
		return
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
	case StyleLockBox:
		addStyleLockBox(t, v)
	case StyleOff:
		removeStyle(t, v)
	case GetTagger:
		t.nodes.Add(v.GetTagger())
	default:
		t.nodes.Add(v)
	}
}

func addStyle(t *Tag, styleMap Style) {
	for name, value := range styleMap {
		if lockBox, exists := t.styleValues[name]; exists {
			lockBox.Set(value)
		} else {
			t.styleValues[name] = NewLockBox(value)
			t.styleOrder = append(t.styleOrder, name)
		}
	}
}

func addStyleLockBox(t *Tag, styleMap StyleLockBox) {
	for name, value := range styleMap {
		if _, exists := t.styleValues[name]; exists {
			t.styleValues[name] = value
		} else {
			t.styleValues[name] = value
			t.styleOrder = append(t.styleOrder, name)
		}
	}
}

func removeStyle(t *Tag, offList StyleOff) {
	for i := 0; i < len(offList); i++ {
		delete(t.styleValues, offList[i])
	}

	var newOrder []string
	for i := 0; i < len(t.styleOrder); i++ {
		if _, exist := t.styleValues[t.styleOrder[i]]; exist {
			newOrder = append(newOrder, t.styleOrder[i])
		}
	}

	t.styleOrder = newOrder
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
