package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/SamHennessy/hlive/hlivekit"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func callback(container *l.Component) {
	container.Add(
		hlivekit.OnDiffApply(
			func(ctx context.Context, e l.Event) {
				container.Add(l.T("p", "Diff Applied"))
				container.RemoveEventBinding(e.Binding.ID)
			},
		),
	)
}

func home() *l.PageServer {
	f := func() *l.Page {
		container := l.C("code")

		btn := l.C("button", "Trigger Click",
			l.On("click", func(ctx context.Context, e l.Event) {
				container.Add(l.T("p", "Click"))
				callback(container)
			}),
		)

		page := l.NewPage()
		page.Title.Add("Callback Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("h1", "Callback"),
			l.T("blockquote", "Get notified when a change has been applied in the browser"),
			btn,
			l.T("h2", "Events"),
			l.T("pre", container),
		)

		return page
	}

	return l.NewPageServer(f)
}
