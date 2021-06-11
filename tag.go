package hlive

import (
	"fmt"
	"strings"
)

type TagInterface interface {
	Render() []interface{}
	GetName() string
	IsVoid() bool
	GetAttributes() []*Attribute
}

type Tag struct {
	name        string
	void        bool
	attributes  []*Attribute
	nodes       []interface{}
	cssExists   map[string]bool
	cssOrder    []string
	styleNames  []string
	styleValues map[string]string
}

func T(name string, elements ...interface{}) *Tag {
	return NewTag(name, elements...)
}
func NewTag(name string, elements ...interface{}) *Tag {
	name = strings.ToLower(name)
	var void bool
	switch name {
	case "area", "base", "br", "col", "command", "embed", "hr", "img", "input", "keygen", "link", "meta", "param", "source", "track", "wbr":
		void = true
	}

	t := &Tag{
		name:        name,
		void:        void,
		cssExists:   map[string]bool{},
		styleValues: map[string]string{},
	}

	for i := 0; i < len(elements); i++ {
		if !IsElement(elements[i]) {
			panic(fmt.Errorf("element: %#v: %w", elements[i], ErrRenderElement))
		}

		addElementToTag(t, elements[i])
	}

	return t
}

func (t *Tag) SetName(name string) {
	t.name = name
}

func (t *Tag) GetName() string {
	return t.name
}

func (t *Tag) IsVoid() bool {
	return t.void
}

func (t *Tag) SetVoid(void bool) {
	t.void = void
}

func (t *Tag) GetAttributes() []*Attribute {
	attrs := t.attributes

	if len(t.cssOrder) != 0 {
		attrs = append(attrs, NewAttribute("class", strings.Join(t.cssOrder, " ")))
	}

	if len(t.styleNames) != 0 {
		value := ""
		for i := 0; i < len(t.styleNames); i++ {
			name := t.styleNames[i]
			value += name + ":" + t.styleValues[name] + ";"
		}

		attrs = append(attrs, NewAttribute("style", value))
	}

	return attrs
}

func (t *Tag) GetAttribute(name string) *Attribute {
	attrs := t.GetAttributes()
	for i := 0; i < len(attrs); i++ {
		if attrs[i].Name == name {
			return attrs[i]
		}
	}

	return nil
}

func (t *Tag) GetAttributeValue(name string) string {
	a := t.GetAttribute(name)
	if a == nil || a.Value == nil {
		return ""
	}

	return *a.Value
}

func (t *Tag) Add(element ...interface{}) {
	addElementToTag(t, element)
}

func (t *Tag) Render() []interface{} {
	return t.nodes
}

func (t *Tag) AddNodes(nodes ...interface{}) {
	for i := 0; i < len(nodes); i++ {
		if nodes[i] == nil {
			continue
		}

		if !IsNode(nodes[i]) {
			panic(fmt.Errorf("node: %#v: %w", nodes[i], ErrInvalidNode))
		}

		t.nodes = append(t.nodes, nodes[i])
	}
}

func (t *Tag) SetAttributes(attrs ...interface{}) {
	attributes := anyToAttributes(attrs...)

	for i := 0; i < len(attributes); i++ {
		hit := false
		for j := 0; j < len(t.attributes); j++ {
			if t.attributes[j].Name == attributes[i].Name {
				hit = true
				t.attributes[j].Value = attributes[i].Value
			}
		}

		if !hit {
			t.attributes = append(t.attributes, attributes[i])
		}
	}
}

func (t *Tag) RemoveAttributes(names ...string) {
	var newAttrs []*Attribute

	for j := 0; j < len(t.attributes); j++ {
		attr := t.attributes[j]
		hit := false

		for i := 0; i < len(names); i++ {
			if names[i] == attr.Name {
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
	if !IsElement(v) {
		panic(fmt.Errorf("element: %#v: %w", v, ErrInvalidElement))
	}

	switch v := v.(type) {
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
	case []ComponentInterface:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case []TagInterface:
		for i := 0; i < len(v); i++ {
			addElementToTag(t, v[i])
		}
	case *Attribute, []*Attribute, Attrs:
		t.SetAttributes(v)
	case CSS:
		addCSS(t, v)
	case Style:
		addStyle(t, v)
	default:
		t.AddNodes(v)
	}
}

func addStyle(t *Tag, v Style) {
	removeList := map[string]bool{}

	for name, value := range v {
		if _, exists := t.styleValues[name]; exists {
			if value == nil {
				removeList[name] = true
			} else {
				str, _ := value.(string)
				t.styleValues[name] = str
			}
		} else if value != nil {
			removeList[name] = false
			str, _ := value.(string)
			t.styleValues[name] = str
			t.styleNames = append(t.styleNames, name)

		}
	}

	var newNames []string
	for i := 0; i < len(t.styleNames); i++ {
		name := t.styleNames[i]
		if removeList[name] {
					delete(t.styleValues, name)

					continue
		}

		newNames = append(newNames, name)
	}

	t.styleNames = newNames
}

func addCSS(t *Tag, v CSS) {
	// This does allow duplicates in the same hlive.CSS element

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

func anyToAttributes(attrs ...interface{}) []*Attribute {
	var newAttrs []*Attribute

	for i := 0; i < len(attrs); i++ {
		if attrs[i] == nil {
			continue
		}

		switch v := attrs[i].(type) {
		case *Attribute:
			newAttrs = append(newAttrs, v)
		case []*Attribute:
			newAttrs = append(newAttrs, v...)
		case Attrs:
			newAttrs = append(newAttrs, attrsToAttributes(v)...)
		default:
			panic(ErrInvalidAttribute)
		}
	}

	return newAttrs
}

func attrsToAttributes(attrs Attrs) []*Attribute {
	var newAttrs []*Attribute
	for name, val := range attrs {
		attr := NewAttribute(name)
		switch v := val.(type) {
		case string:
			attr.SetValue(v)
		case *string:
			attr.Value = v
		}

		newAttrs = append(newAttrs, attr)
	}

	return newAttrs
}