package hlivekit

import (
	_ "embed"

	l "github.com/SamHennessy/hlive"
)

const RedirectAttributeName = "data-redirect"

//go:embed redirect.js
var RedirectJavaScript []byte

func Redirect(url string) l.Attributer {
	attr := &RedirectAttribute{
		l.NewAttribute(RedirectAttributeName, url),
	}

	return attr
}

type RedirectAttribute struct {
	*l.Attribute
}

func (a *RedirectAttribute) Initialize(page *l.Page) {
	page.DOM.Head.Add(l.T("script", l.HTML(RedirectJavaScript)))
}

func (a *RedirectAttribute) InitializeSSR(page *l.Page) {
	page.DOM.Head.Add(l.T("script", l.HTML(RedirectJavaScript)))
}
