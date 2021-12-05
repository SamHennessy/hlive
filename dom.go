package hlive

type DOM struct {
	// Root DOM elements
	DocType HTML
	HTML    Adder
	Head    Adder
	Meta    Adder
	Title   Adder
	Body    Adder
}

func NewDOM() *DOM {
	dom := &DOM{
		DocType: HTML5DocType,
		HTML:    C("html", Attrs{"lang": "en"}),
		Head:    C("head"),
		Meta:    T("meta", Attrs{"charset": "utf-8"}),
		Title:   C("title"),
		Body:    C("body"),
	}

	dom.Head.Add(dom.Meta, dom.Title)
	dom.HTML.Add(dom.Head, dom.Body)

	return dom
}
