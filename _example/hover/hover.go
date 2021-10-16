package main

import (
	"context"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
)

func main() {
	http.Handle("/", home())

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home() *l.PageServer {
	f := func() *l.Page {
		hoverState := " "

		hover := l.C("h2",
			l.Style{"background-color": "#7fc3c3", "border-radius": "4px", "padding": "4px"},
			l.On("mouseEnter", func(ctx context.Context, e l.Event) {
				hoverState = "Mouse enter"
			}),
			l.On("mouseLeave", func(ctx context.Context, e l.Event) {
				hoverState = "Mouse leave"
			}),
			"Hover over me",
		)

		page := l.NewPage()
		page.Title.Add("Hover Example")
		page.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://classless.de/classless.css"}))

		page.Body.Add(
			l.T("h1", "Hover"),
			l.T("blockquote", "React to hover events on the server"),
			l.T("div", hover),
			l.T("hr"),
			l.T("pre", l.T("code", &hoverState)),
		)

		return page
	}

	return l.NewPageServer(f)
}
