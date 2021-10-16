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
		page.Title.Add("Preempt Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		var count int

		btn := l.C("button",
			// Passing by reference
			&count,
		)

		btn.Add(hlivekit.PreemptDisableOn(l.On("click",
			func(_ context.Context, _ l.Event) {
				time.Sleep(2 * time.Second)
				count++
				btn.Add(l.Attrs{"disabled": nil})
			}),
		))

		page.Body.Add(
			l.T("h1", "Preempt - Client Side First Code"),
			l.T("blockquote", "Update the client side DOM before the server side."),
			l.T("p", "The handler will sleep for 2 seconds to simulate a long processing time. We will "+
				"disable the button on the client side first to prevent extra clicks."),
			"Clicks: ",
			btn,
		)

		return page
	}

	return l.NewPageServer(f)
}
