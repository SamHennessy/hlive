package hlivekit

import (
	_ "embed"
	"strconv"

	l "github.com/SamHennessy/hlive"
)

const ScrollIntoViewAttributeName = "data-scrollIntoView"

//go:embed scrollIntoView.js
var ScrollIntoViewJavaScript []byte

func ScrollIntoView(alignToTop bool) l.Attributer {
	attr := &ScrollIntoViewAttribute{
		Attribute: l.NewAttribute(ScrollIntoViewAttributeName, strconv.FormatBool(alignToTop)),
	}

	return attr
}

func ScrollIntoViewRemove(tag l.Adder) {
	tag.Add(l.AttrsOff{ScrollIntoViewAttributeName})
}

type ScrollIntoViewAttribute struct {
	*l.Attribute

	rendered bool
}

func (a *ScrollIntoViewAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	page.DOM().Head().Add(l.T("script", l.HTML(ScrollIntoViewJavaScript)))
}

func (a *ScrollIntoViewAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	page.DOM().Head().Add(l.T("script", l.HTML(ScrollIntoViewJavaScript)))
}
