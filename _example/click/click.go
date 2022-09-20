package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", l.NewPageServer(home))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.Page {
	page := l.NewPage()
	page.DOM().Title().Add("Click Example")
	page.DOM().Head().Add(
		l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

	// A thread safe value container
	count := l.Box(0)

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
						// We need to read and write inside a single lock
						count.Lock(func(v int) int { return v + 1 })
					}),
					count,
				),
			),
		),
	)

	return page
}
