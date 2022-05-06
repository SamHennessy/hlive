package hlivekit

import (
	_ "embed"
	"strconv"

	l "github.com/SamHennessy/hlive"
)

const ScrollTopAttributeName = "data-scrollTop"

//go:embed scrollTop.js
var ScrollTopJavaScript []byte

func ScrollTop(val int) l.Attributer {
	attr := &ScrollTopAttribute{
		l.NewAttribute(ScrollTopAttributeName, strconv.Itoa(val)),
	}

	return attr
}

func ScrollTopRemove(tag l.Adder) {
	tag.Add(l.Attrs{ScrollTopAttributeName: nil})
}

type ScrollTopAttribute struct {
	*l.Attribute
}

func (a *ScrollTopAttribute) Initialize(page *l.Page) {
	page.DOM.Head.Add(l.T("script", l.HTML(ScrollTopJavaScript)))
}

func (a *ScrollTopAttribute) InitializeSSR(page *l.Page) {
	page.DOM.Head.Add(l.T("script", l.HTML(ScrollTopJavaScript)))
}
