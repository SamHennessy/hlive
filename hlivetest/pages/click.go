package pages

import (
	"context"

	l "github.com/SamHennessy/hlive"
)

func Click() func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("Test Ack Click")

		var count int

		page.DOM().Body().Add(
			l.C("button",
				l.Attrs{"id": "btn"},
				l.On("click", func(_ context.Context, _ l.Event) {
					count++
				}),
				"Click",
			),
			l.T("div", l.Attrs{"id": "count"}, &count),
		)

		return page
	}
}
