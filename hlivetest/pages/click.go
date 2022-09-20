package pages

import (
	"context"

	l "github.com/SamHennessy/hlive"
)

func Click() func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("Test Ack Click")

		count := l.Box(0)

		page.DOM().Body().Add(
			l.C("button",
				l.Attrs{"id": "btn"},
				l.On("click", func(_ context.Context, _ l.Event) {
					count.Lock(func(val int) int {
						return val + 1
					})
				}),
				"Click",
			),
			l.T("div", l.Attrs{"id": "count"}, count),
		)

		return page
	}
}
