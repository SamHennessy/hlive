package hlive

// HTML must always have a single root element, as we count it as 1 node in the tree but the browser will not if you have multiple root elements
type HTML string

type RenderFunc func() []interface{}

func If(ok *bool, nodes ...interface{}) RenderFunc {
	return func() []interface{} {
		if ok != nil && *ok {
			return nodes
		}

		return nil
	}
}

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
	case nil, string, HTML, Componenter, Tagger, RenderFunc,
		[]interface{}, []*Component, []*Tag, []Componenter, []Tagger,
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
