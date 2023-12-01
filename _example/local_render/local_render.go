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
	count := l.Box(0)

	page := l.NewPage()
	page.DOM().Title().Add("Local GetNodes Example")
	page.DOM().Head().Add(l.T("link",
		l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

	page.DOM().Body().Add(
		l.T("header",
			l.T("h1", "Local Render"),
			l.T("p", "By default, the whole page if checked for difference after an event. "+
				"You can override that behaviour and chose to only render a component and it's children."),
		),
		l.T("main",
			l.T("h2", "Global Render"),
			l.T("h4", "Everything will update"),
			newCountBtn(count),
			l.Group(" The count is: ", l.T("em", count), " clicks"),
			l.T("h2", "Local Render"),
			l.T("h4", "Only the button will update"),
			newCountBtnLocal(count),
			l.Group(" The count is: ", l.T("em", count), " clicks"),
		),
	)

	return page
}

func newCountBtn(count *l.NodeBox[int]) *l.Component {
	c := l.C("button", count)

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		count.Lock(func(v int) int { return v + 1 })
	}))

	return c
}

func newCountBtnLocal(count *l.NodeBox[int]) *l.Component {
	c := l.C("button", count)

	// Don't render this component when an event binding is triggered
	c.AutoRender = false

	c.Add(l.On("click", func(ctx context.Context, e l.Event) {
		count.Lock(func(v int) int { return v + 1 })

		// Will render the passed component and it's subtree
		l.RenderComponent(ctx, c)
	}))

	return c
}
