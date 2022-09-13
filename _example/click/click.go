package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", l.NewPageServer(home()))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() func() *l.Page {
	return func() *l.Page {
		page := l.NewPage()
		page.DOM().Title().Add("Click Example")
		page.DOM().Head().Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		var count int

		page.DOM().Body().Add(
			l.T("header",
				l.T("h1", "Click"),
				l.T("p", "Click the button and see the count increase"),
			),
			l.T("main",
				l.T("p",
					"Clicks: ",
					l.C("button",
						l.On("click", func(_ context.Context, _ l.Event) {
							count++
						}),
						// Passing by reference
						&count,
					),
				),
			),
		)

		return page
	}
}
