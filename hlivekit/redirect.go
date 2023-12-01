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
		Attribute: l.NewAttribute(RedirectAttributeName, url),
	}

	return attr
}

type RedirectAttribute struct {
	*l.Attribute

	rendered bool
}

func (a *RedirectAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	page.DOM().Head().Add(l.T("script", l.HTML(RedirectJavaScript)))
}

func (a *RedirectAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	page.DOM().Head().Add(l.T("script", l.HTML(RedirectJavaScript)))
}
