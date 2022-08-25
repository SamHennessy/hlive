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

// TODO: Once?
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
			l.LoggerDev.Error().Msg("PreemptDisableOn: bound Component must be an Adder")
		}

		// Call original handler
		if ogHandler != nil {
			ogHandler(ctx, e)
		}
	}

	return l.E(eb, sourceAttr)
}

type PreemptDisableAttribute struct {
	*l.Attribute

	page     *l.Page
	rendered bool
}

func (a *PreemptDisableAttribute) Initialize(page *l.Page) {
	if a.rendered {
		return
	}

	a.page = page
	page.DOM.Head.Add(l.T("script", l.HTML(PreemptDisableOnClickJavaScript)))
}

func (a *PreemptDisableAttribute) InitializeSSR(page *l.Page) {
	a.rendered = true
	a.page = page
	page.DOM.Head.Add(l.T("script", l.HTML(PreemptDisableOnClickJavaScript)))
}
