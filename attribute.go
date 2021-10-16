package hlive

import (
	"strings"
)

type Attributer interface {
	GetAttribute() *Attribute
}

type AttributePluginer interface {
	Attributer

	// Initialize will only be called once per attribute name
	Initialize(page *Page)
}

// Attrs is a helper for adding Attributes to nodes
// You can update an existing Attribute by adding new Attrs, it;s also possible to pass a string by reference.
// You can remove an Attribute by passing a nil value.
type Attrs map[string]interface{}

func (a Attrs) GetAttributes() []Attributer {
	newAttrs := make([]Attributer, 0, len(a))

	for name, val := range a {
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

// CSS a special Attribute for working with CSS classes on nodes.
// It supports turning them on and off and allowing overriding.
// All CSSs are de-duped, overriding a CSS by adding new CSS will result in the old CSS getting updated.
// You don't have to use CSS to add a class attribute, but it's the recommended way to do it.
type CSS map[string]bool

// Style is a special Attribute that allows you to work with CSS styles on nodes.
// It allows you to override
// All Styles are de-duped, overriding a Style by adding new Style will result in the old Style getting updated.
// You don't have to use Style to add a style attribute, but it's the recommended way to do it.
type Style map[string]interface{}

// NewAttribute create a new Attribute
func NewAttribute(name string, value ...string) *Attribute {
	a := Attribute{Name: strings.ToLower(name)}

	if len(value) != 0 {
		if len(value) != 1 {
			panic(ErrAttrValueCount)
		}

		a.Value = &value[0]
	}

	return &a
}

// Attribute represents an HTML attribute e.g. id="submitBtn"
type Attribute struct {
	// Name must always be lowercase
	Name  string
	Value *string
}

func (a *Attribute) SetValue(value string) {
	a.Value = &value
}

func (a *Attribute) GetValue() string {
	if a == nil || a.Value == nil {
		return ""
	}

	return *a.Value
}

// Clone creates a new Attribute using the data from this Attribute
func (a *Attribute) Clone() *Attribute {
	newA := NewAttribute(a.Name)

	if a.Value != nil {
		val := *a.Value
		newA.Value = &val
	}

	return newA
}

func (a *Attribute) GetAttribute() *Attribute {
	return a
}

func anyToAttributes(attrs ...interface{}) []Attributer {
	var newAttrs []Attributer

	for i := 0; i < len(attrs); i++ {
		if attrs[i] == nil {
			continue
		}

		switch v := attrs[i].(type) {
		case Attrs:
			newAttrs = append(newAttrs, v.GetAttributes()...)
		case Attributer:
			newAttrs = append(newAttrs, v)
		case []Attributer:
			newAttrs = append(newAttrs, v...)
		case []*Attribute:
			for j := 0; j < len(v); j++ {
				newAttrs = append(newAttrs, v[j])
			}
		default:
			panic(ErrInvalidAttribute)
		}
	}

	return newAttrs
}
