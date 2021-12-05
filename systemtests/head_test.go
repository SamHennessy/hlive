package systemtests_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

func TestHead_TitleStatic(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM.Title.Add("value 1")

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.Diff(t, "value 1", hlivetest.Title(t, h.pwpage))
}

func TestHead_TitleDynamic(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		title := "value 1"

		page := l.NewPage()

		page.DOM.Title.Add(&title)

		page.DOM.Body.Add(
			l.C("button",
				l.Attrs{"id": "btn"},
				l.On("click", func(ctx context.Context, e l.Event) {
					title = "value 2"
				}),
				"Click Me",
			),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.Diff(t, "value 1", hlivetest.Title(t, h.pwpage))

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "value 2", hlivetest.Title(t, h.pwpage))
}

func TestHead_ScriptTag(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM.Body.Add(
			l.C("button", l.Attrs{"id": "btn"}, "Click Me",
				l.On("click", func(ctx context.Context, e l.Event) {
					page.DOM.Head.Add(l.T("script", l.HTML(`document.getElementById("content").innerText = "value 2"`)))
				}),
			),
			l.T("div", l.Attrs{"id": "content"}, "value 1"),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.Diff(t, "value 1", hlivetest.TextContent(t, h.pwpage, "#content"))

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "value 2", hlivetest.TextContent(t, h.pwpage, "#content"))
}
