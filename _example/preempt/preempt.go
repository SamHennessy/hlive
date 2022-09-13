package main

import (
	"context"
	"log"
	"net/http"
	"time"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivekit"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve:", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("Preempt Example")
		page.DOM().Head().Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		var countWith int

		btnWith := l.C("button",
			// Passing by reference
			&countWith,
		)

		btnWith.Add(hlivekit.PreemptDisableOn(l.On("click",
			func(_ context.Context, _ l.Event) {
				time.Sleep(2 * time.Second)
				countWith++
				btnWith.Add(l.Attrs{"disabled": nil})
			}),
		))

		var countWithout int

		btnWithout := l.C("button",
			// Passing by reference
			&countWithout,
		)

		btnWithout.Add(l.On("click",
			func(_ context.Context, _ l.Event) {
				time.Sleep(2 * time.Second)
				countWithout++
			}),
		)

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "Preempt - Client Side First Code"),
				l.T("p", "Update the client side DOM before the server side."),
			),
			l.T("main",
				l.T("p", "The handler will sleep for 2 seconds to simulate a long processing time. "+
					"The first button will be disabled in the browser first to prevent extra clicks. Now click the "+
					"buttons as many times as you can to see the differance"),
				"Clicks With: ",
				btnWith,
				l.T("br"),
				"Clicks Without: ",
				btnWithout,
			),
		)

		return page
	}

	return l.NewPageServer(f)
}
