package hlive

// HTML must always have a single root element, as we count it as 1 node in the tree but the browser will not if you have multiple root elements
type HTML string

type RenderFunc func() interface{}

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
	case nil, string, HTML, Tagger, RenderFunc,
		[]interface{}, []*Component, []*Tag, []Componenter, []Tagger, []UniqueTagger,
		int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64,
		*int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64,
		*string, *HTML:
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
