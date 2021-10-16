package systemtests_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

func TestAddRemove_AddSibling(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		page := l.NewPage()

		parent := l.T("div", l.Attrs{"id": "parent"},
			l.T("div", l.Attrs{"id": "a"}),
			l.T("div", l.Attrs{"id": "b"}),
		)

		page.Body.Add(
			l.C("button", l.Attrs{"id": "btn"}, "Click Me",
				l.On("click", func(ctx context.Context, e l.Event) {
					parent.Add(l.T("div", l.Attrs{"id": "c"}))
				}),
			),
			parent,
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.Diff(t, "b", hlivetest.GetID(t, h.pwpage, "#parent div:last-child"))

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "c", hlivetest.GetID(t, h.pwpage, "#parent div:last-child"))
}

func TestAddRemove_AddMultipleSibling(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		page := l.NewPage()

		parent := l.T("div", l.Attrs{"id": "parent"},
			l.T("div", l.Attrs{"id": "a"}),
			l.T("div", l.Attrs{"id": "b"}),
		)

		page.Body.Add(
			l.C("button", l.Attrs{"id": "btn"}, "Click Me",
				l.On("click", func(ctx context.Context, e l.Event) {
					parent.Add(
						l.T("div", l.Attrs{"id": "c"}),
						l.T("div", l.Attrs{"id": "d"}),
					)
				}),
			),
			parent,
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "c", hlivetest.GetID(t, h.pwpage, "#parent div:nth-child(3)"))
	hlivetest.Diff(t, "d", hlivetest.GetID(t, h.pwpage, "#parent div:nth-child(4)"))
}
