package hlive

func IsElement(el interface{}) bool {
	switch el.(type) {
	case []*Attribute, *Attribute, Attrs, CSS, Style:
		return true
	default:
		return IsNode(el)
	}
}

func IsNode(node interface{}) bool {
	switch node.(type) {
	case nil, string, HTML, ComponentInterface, TagInterface, RenderFunc,
		[]interface{}, []*Component, []*Tag, []ComponentInterface, []TagInterface,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		*int, *string:
		return true
	default:
		return false
	}
}

func Tree(nodes ...interface{}) []interface{} {
	var kids []interface{}

	for i := 0; i < len(nodes); i++ {
		if !IsNode(nodes[i]) {
			panic(ErrInvalidNode)
		}

		kids = append(kids, nodes[i])
	}

	return kids
}

// HTML must always have a single root element, as we count it as 1 node in the tree but the browser will not if you have multiple root elements
type HTML string

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

func (a *Attribute) Clone() *Attribute {
	newA := NewAttribute(a.Name)
	if a.Value != nil {
		val := *a.Value
		newA.Value = &val
	}

	return newA
}

func A(name string, value ...string) *Attribute {
	return NewAttribute(name, value...)
}
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

type Attrs map[string]interface{}

type CSS map[string]bool

type Style map[string]interface{}

type RenderFunc func() []interface{}

func If(ok *bool, nodes ...interface{}) RenderFunc {
	return func() []interface{} {
		if ok != nil && *ok {
			return nodes
		}

		return nil
	}
}
