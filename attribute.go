package hlive

// Attrs is a helper for adding Attributes to nodes
// You can update an existing Attribute by adding new Attrs but you can't remove an Attribute, use RemoveAttribute.
type Attrs map[string]interface{}

// CSS a special Attribute for working with CSS classes on nodes.
// It supports turning them on and off and allowing overriding.
// All CSSs are de-duped, overriding a CSS by adding new CSS will result in the old CSS getting updated.
// You don't have to use CSS to add a class attribute but it's the recommended way to do it.
type CSS map[string]bool

// Style is a special Attribute that allows you to work with CSS styles on nodes.
// It allows you to override
// All Styles are de-duped, overriding a Style by adding new Style will result in the old Style getting updated.
// You don't have to use Style to add a style attribute but it's the recommended way to do it.
type Style map[string]interface{}

// NewAttribute create a new Attribute
func NewAttribute(name string, value ...string) *Attribute {
	a := Attribute{Name: name}

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
	Name  string
	Value *string
}

func (a *Attribute) SetValue(value string) {
	a.Value = &value
}

func (a *Attribute) GetValue() string {
	if a.Value == nil {
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
	newAttrs := make([]*Attribute, 0, len(attrs))

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
