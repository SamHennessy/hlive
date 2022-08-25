package hlivekit

import (
	_ "embed"

	l "github.com/SamHennessy/hlive"
)

const FocusAttributeName = "data-hlive-focus"

//go:embed focus.js
var FocusJavaScript []byte

func Focus() l.Attributer {
	attr := &FocusAttribute{
		Attribute: l.NewAttribute(FocusAttributeName, ""),
	}

	return attr
}

func FocusRemove(tag l.Adder) {
	tag.Add(l.Attrs{FocusAttributeName: nil})
}

type FocusAttribute struct {
	*l.Attribute

	rendered bool
}

func (a *FocusAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	page.DOM.Head.Add(l.T("script", l.HTML(FocusJavaScript)))
}

func (a *FocusAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	page.DOM.Head.Add(l.T("script", l.HTML(FocusJavaScript)))
}
