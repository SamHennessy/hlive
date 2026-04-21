package hlivekit

import (
	_ "embed"

	l "github.com/SamHennessy/hlive"
)

func OnElementVisible(handler l.EventHandler) *l.ElementGroup {
	eb := l.On(ElementVisibleEvent, handler)
	attr := &ElementVisibleAttribute{
		Attribute: l.NewAttribute(ElementVisibleAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

// TODO: how we remove the attribute once done?
func OnElementVisibleOnce(handler l.EventHandler) *l.ElementGroup {
	eb := l.OnOnce(ElementVisibleEvent, handler)
	attr := &ElementVisibleAttribute{
		Attribute: l.NewAttribute(ElementVisibleAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

const (
	ElementVisibleEvent         = "helementvisible"
	ElementVisibleAttributeName = "data-helementVisible"
)

//go:embed elementVisible.js
var ElementVisibleJavaScript []byte

func ElementVisible() l.Attributer {
	attr := &ElementVisibleAttribute{
		Attribute: l.NewAttribute(ElementVisibleAttributeName, ""),
	}

	return attr
}

func ElementVisibleRemove(tag l.Adder) {
	tag.Add(l.AttrsOff{ElementVisibleAttributeName})
}

type ElementVisibleAttribute struct {
	*l.Attribute

	rendered bool
}

func (a *ElementVisibleAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	page.DOM().Head().Add(l.T("script", l.HTML(ElementVisibleJavaScript)))
}

func (a *ElementVisibleAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	page.DOM().Head().Add(l.T("script", l.HTML(ElementVisibleJavaScript)))
}
