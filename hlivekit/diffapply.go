package hlivekit

import (
	_ "embed"

	l "github.com/SamHennessy/hlive"
)

//go:embed diffapply.js
var DiffApplyScript []byte

// DiffApply is a special event that will trigger when a diff is applied.
// This means that it will trigger itself when first added. This will allow you to know when a change in the tree has
// made it to the browser. You can then, if you wish, immediately remove it from the tree to prevent more triggers.
// You can also add it as a OnOnce and it wil remove itself.

func OnDiffApply(handler l.EventHandler) *l.ElementGroup {
	eb := l.On(DiffApplyEvent, handler)
	attr := &DiffApplyAttribute{
		Attribute: l.NewAttribute(DiffApplyAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

// TODO: how we remove the attribute once done?
func OnDiffApplyOnce(handler l.EventHandler) *l.ElementGroup {
	eb := l.OnOnce(DiffApplyEvent, handler)
	attr := &DiffApplyAttribute{
		Attribute: l.NewAttribute(DiffApplyAttributeName, eb.ID),
	}

	return l.E(eb, attr)
}

const (
	DiffApplyEvent         = "diffapply"
	DiffApplyAttributeName = "data-hlive-on-diffapply"
)

type DiffApplyAttribute struct {
	*l.Attribute

	rendered bool
}

func (a *DiffApplyAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	page.DOM().Head().Add(l.T("script", l.HTML(DiffApplyScript)))
}

func (a *DiffApplyAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	page.DOM().Head().Add(l.T("script", l.HTML(DiffApplyScript)))
}
