package hlive

type DOM struct {
	// Root DOM elements
	docType HTML
	html    Adder
	head    Adder
	meta    Adder
	title   Adder
	body    Adder
}

func NewDOM() DOM {
	dom := DOM{
		docType: HTML5DocType,
		html:    C("html", Attrs{"lang": "en"}),
		head:    C("head"),
		meta:    T("meta", Attrs{"charset": "utf-8"}),
		title:   C("title"),
		body:    C("body"),
	}

	dom.head.Add(dom.meta, dom.title)
	dom.html.Add(dom.head, dom.body)

	return dom
}

func (dom DOM) DocType() HTML {
	return dom.docType
}

func (dom DOM) HTML() Adder {
	return dom.html
}

func (dom DOM) Head() Adder {
	return dom.head
}

func (dom DOM) Meta() Adder {
	return dom.meta
}

func (dom DOM) Title() Adder {
	return dom.title
}

func (dom DOM) Body() Adder {
	return dom.body
}
