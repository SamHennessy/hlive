package systemtests_test

import (
	"context"
	"sync"
	"testing"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivetest"
)

func TestEvents_Propagation(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex
	var btnInner, btnOuter bool

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM().Body().Add(
			l.C("div",
				l.On("click", func(ctx context.Context, e l.Event) {
					mu.Lock()
					btnOuter = true
					mu.Unlock()
				}),

				l.C("button", l.Attrs{"id": "btn"}, "Click Me",
					l.On("click", func(ctx context.Context, e l.Event) {
						mu.Lock()
						btnInner = true
						mu.Unlock()
					}),
				),
			),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	mu.Lock()
	defer mu.Unlock()

	if !btnInner || !btnOuter {
		t.Fail()
	}
}

func TestEvents_StopPropagation(t *testing.T) {
	t.Parallel()

	var mu sync.Mutex

	var btnInner, btnOuter bool

	pageFn := func() *l.Page {
		page := l.NewPage()

		page.DOM().Body().Add(
			l.C("div",
				l.On("click", func(ctx context.Context, e l.Event) {
					mu.Lock()
					btnOuter = true
					mu.Unlock()
				}),

				l.C("button", l.Attrs{"id": "btn"}, "Click Me",
					l.StopPropagation(),
					l.On("click", func(ctx context.Context, e l.Event) {
						mu.Lock()
						btnInner = true
						mu.Unlock()
					}),
				),
			),
		)

		return page
	}

	h := setup(t, pageFn)
	defer h.teardown()

	hlivetest.ClickAndWait(t, h.pwpage, "#btn")

	mu.Lock()
	defer mu.Unlock()

	if !btnInner || btnOuter {
		t.Fail()
	}
}
