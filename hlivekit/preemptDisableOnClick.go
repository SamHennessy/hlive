package hlivekit

import (
	"context"
	_ "embed"

	l "github.com/SamHennessy/hlive"
)

const (
	PreemptDisableAttributeName = "data-hlive-pre-disable"
)

//go:embed preemptDisableOnClick.js
var PreemptDisableOnClickJavaScript []byte

// Once?
func PreemptDisableOn(eb *l.EventBinding) *l.ElementGroup {
	sourceAttr := &PreemptDisableAttribute{
		Attribute: l.NewAttribute(PreemptDisableAttributeName, eb.Name),
	}

	ogHandler := eb.Handler

	eb.Handler = func(ctx context.Context, e l.Event) {
		// Update the Browser DOM with what we've done client first
		if sourceAttr.page != nil {
			if browserTag := sourceAttr.page.GetBrowserNodeByID(e.Binding.Component.GetID()); browserTag != nil {
				browserTag.Add(l.Attrs{"disabled": ""})
			}
		}
		// Update the Page DOM
		if adder, ok := e.Binding.Component.(l.Adder); ok {
			adder.Add(l.Attrs{"disabled": ""})
		} else {
			// ???
		}
		// Call original handler
		ogHandler(ctx, e)
	}

	return l.E(eb, sourceAttr)
}

type PreemptDisableAttribute struct {
	*l.Attribute

	page *l.Page
}

func (a *PreemptDisableAttribute) Initialize(page *l.Page) {
	a.page = page
	page.Head.Add(l.T("script", l.HTML(PreemptDisableOnClickJavaScript)))
}
