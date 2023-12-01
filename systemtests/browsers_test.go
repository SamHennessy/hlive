package systemtests_test

import (
	"context"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

// Dealing with browser behavior

// While you may have 3 nodes in the Page dom if they result in 3 strings the browser will combine them into a
// single node. this tests we can deal with that situation.
func TestBrowser_StringConcat(t *testing.T) {
	t.Parallel()

	pageFn := func() *l.Page {
		count := l.Box(0)

		page := l.NewPage()

		page.DOM().Body().Add(
			l.T("div", l.Attrs{"id": "content"},
				"The count is ", count, ".",
			),
			l.C("button", l.Attrs{"id": "btn"}, "Click Me",
				l.On("click", func(ctx context.Context, e l.Event) {
					count.Lock(func(v int) int {
						return v + 1
					})
				})),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.Diff(t, "The count is 0.", hlivetest.TextContent(t, h.pwpage, "#content"))

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	hlivetest.Diff(t, "The count is 1.", hlivetest.TextContent(t, h.pwpage, "#content"))
}
