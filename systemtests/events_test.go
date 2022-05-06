package systemtests_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

func TestEvents_Propagation(t *testing.T) {
	t.Parallel()

	var btnInner, btnOuter bool

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM.Body.Add(
			l.C("div",
				l.On("click", func(ctx context.Context, e l.Event) {
					btnOuter = true
				}),

				l.C("button", l.Attrs{"id": "btn"}, "Click Me",
					l.On("click", func(ctx context.Context, e l.Event) {
						btnInner = true
					}),
				),
			),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	if !btnInner || !btnOuter {
		t.Fail()
	}
}

func TestEvents_StopPropagation(t *testing.T) {
	t.Parallel()

	var btnInner, btnOuter bool

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM.Body.Add(
			l.C("div",
				l.On("click", func(ctx context.Context, e l.Event) {
					btnOuter = true
				}),

				l.C("button", l.Attrs{"id": "btn"}, "Click Me",
					l.StopPropagation(),
					l.On("click", func(ctx context.Context, e l.Event) {
						btnInner = true
					}),
				),
			),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	if !btnInner || btnOuter {
		t.Fail()
	}
}
