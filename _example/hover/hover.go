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
	hoverState := l.Box(" ")

	hover := l.C("h2",
		l.Style{"padding": "1em", "text-align": "center", "border": "solid"},
		l.On("mouseEnter", func(ctx context.Context, e l.Event) {
			hoverState.Set("Mouse enter")
		}),
		l.On("mouseLeave", func(ctx context.Context, e l.Event) {
			hoverState.Set("Mouse leave")
		}),
		"Hover over me",
	)

	page := l.NewPage()
	page.DOM().Title().Add("Hover Example")
	page.DOM().Head().Add(
		l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

	page.DOM().Body().Add(
		l.T("header",
			l.T("h1", "Hover"),
			l.T("p", "React to hover events on the server"),
		),
		l.T("main",
			l.T("div", hover),
			l.T("hr"),
			l.T("pre", l.T("code", hoverState)),
		),
	)

	return page
}
